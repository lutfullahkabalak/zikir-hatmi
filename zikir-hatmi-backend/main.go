package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

// hub coordinates websocket clients and broadcasts messages for a hatim.
type hub struct {
	mu        sync.RWMutex
	clients   map[*websocket.Conn]struct{}
	shareCode string
	db        *pgxpool.Pool
	count     int
	target    int
}

func newHub(db *pgxpool.Pool, shareCode string, count int, target int) *hub {
	return &hub{
		clients:   make(map[*websocket.Conn]struct{}),
		shareCode: shareCode,
		db:        db,
		count:     count,
		target:    target,
	}
}

func (h *hub) add(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

func (h *hub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	if _, ok := h.clients[conn]; ok {
		delete(h.clients, conn)
	}
	h.mu.Unlock()
	_ = conn.Close()
}

func (h *hub) broadcast(msgType int, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		client.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := client.WriteMessage(msgType, payload); err != nil {
			log.Printf("broadcast error: %v", err)
		}
	}
}

func (h *hub) getState() (int, int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count, h.target
}

func (h *hub) setState(count int, target int) {
	h.mu.Lock()
	h.count = count
	h.target = target
	h.mu.Unlock()
}

func (h *hub) increment(ctx context.Context) (int, int, bool, error) {
	count, target, completed, err := incrementHatim(ctx, h.db, h.shareCode)
	if err != nil {
		return 0, 0, false, err
	}

	h.mu.Lock()
	h.count = count
	h.target = target
	h.mu.Unlock()
	return count, target, completed, nil
}

type hubStore struct {
	mu   sync.Mutex
	hubs map[string]*hub
	db   *pgxpool.Pool
}

func newHubStore(db *pgxpool.Pool) *hubStore {
	return &hubStore{hubs: make(map[string]*hub), db: db}
}

func (s *hubStore) getOrCreate(ctx context.Context, shareCode string) (*hub, error) {
	s.mu.Lock()
	if existing, ok := s.hubs[shareCode]; ok {
		s.mu.Unlock()
		return existing, nil
	}
	s.mu.Unlock()

	state, err := getHatimState(ctx, s.db, shareCode)
	if err != nil {
		return nil, err
	}

	created := newHub(s.db, shareCode, state.Count, state.Target)

	s.mu.Lock()
	if existing, ok := s.hubs[shareCode]; ok {
		s.mu.Unlock()
		return existing, nil
	}
	s.hubs[shareCode] = created
	s.mu.Unlock()
	return created, nil
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Simplify local development; tighten for production.
	},
	EnableCompression: true,
}

type Message struct {
	Type   string `json:"type"`
	Count  int    `json:"count,omitempty"`
	Target int    `json:"target,omitempty"`
}

type createHatimRequest struct {
	Title    string `json:"title"`
	Target   int    `json:"target"`
	Password string `json:"password"`
}

type joinHatimRequest struct {
	Password string `json:"password"`
}

type hatimResponse struct {
	ShareCode        string `json:"shareCode"`
	Title            string `json:"title"`
	Count            int    `json:"count"`
	Target           int    `json:"target"`
	RequiresPassword bool   `json:"requiresPassword"`
	Token            string `json:"token,omitempty"`
}

func registerRoutes(hubs *hubStore, db *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/hatims", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var payload createHatimRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "invalid body")
			return
		}

		created, token, err := createHatim(r.Context(), db, payload.Title, payload.Target, payload.Password)
		if err != nil {
			log.Printf("create hatim error: %v", err)
			writeError(w, http.StatusInternalServerError, "unable to create hatim")
			return
		}

		writeJSON(w, http.StatusCreated, hatimResponse{
			ShareCode:        created.ShareCode,
			Title:            created.Title,
			Count:            created.Count,
			Target:           created.Target,
			RequiresPassword: created.PasswordHash != nil,
			Token:            token,
		})
	})

	mux.HandleFunc("/hatims/", func(w http.ResponseWriter, r *http.Request) {
		trimmed := strings.TrimPrefix(r.URL.Path, "/hatims/")
		if trimmed == "" {
			writeError(w, http.StatusNotFound, "not found")
			return
		}

		parts := strings.Split(trimmed, "/")
		shareCode := parts[0]
		if shareCode == "" {
			writeError(w, http.StatusNotFound, "not found")
			return
		}

		if len(parts) == 1 {
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}

			state, err := getHatimState(r.Context(), db, shareCode)
			if err != nil {
				if errors.Is(err, errHatimNotFound) {
					writeError(w, http.StatusNotFound, "hatim not found")
					return
				}
				log.Printf("get hatim error: %v", err)
				writeError(w, http.StatusInternalServerError, "unable to fetch hatim")
				return
			}

			writeJSON(w, http.StatusOK, hatimResponse{
				ShareCode:        state.ShareCode,
				Title:            state.Title,
				Count:            state.Count,
				Target:           state.Target,
				RequiresPassword: state.RequiresPassword,
			})
			return
		}

		if len(parts) == 2 && parts[1] == "join" {
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}

			var payload joinHatimRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
				writeError(w, http.StatusBadRequest, "invalid body")
				return
			}

			token, err := joinHatim(r.Context(), db, shareCode, payload.Password)
			if err != nil {
				if errors.Is(err, errHatimNotFound) {
					writeError(w, http.StatusNotFound, "hatim not found")
					return
				}
				if errors.Is(err, errInvalidPassword) {
					writeError(w, http.StatusUnauthorized, "invalid password")
					return
				}
				log.Printf("join hatim error: %v", err)
				writeError(w, http.StatusInternalServerError, "unable to join hatim")
				return
			}

			writeJSON(w, http.StatusOK, map[string]string{"token": token})
			return
		}

		writeError(w, http.StatusNotFound, "not found")
	})

	mux.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		shareCode := strings.TrimPrefix(r.URL.Path, "/ws/")
		if shareCode == "" {
			writeError(w, http.StatusNotFound, "not found")
			return
		}

		token := r.URL.Query().Get("token")
		ok, err := validateToken(r.Context(), db, shareCode, token)
		if err != nil {
			log.Printf("token validation error: %v", err)
			writeError(w, http.StatusInternalServerError, "unable to validate token")
			return
		}
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		hb, err := hubs.getOrCreate(r.Context(), shareCode)
		if err != nil {
			if errors.Is(err, errHatimNotFound) {
				writeError(w, http.StatusNotFound, "hatim not found")
				return
			}
			log.Printf("hub creation error: %v", err)
			writeError(w, http.StatusInternalServerError, "unable to load hatim")
			return
		}

		state, err := getHatimState(r.Context(), db, shareCode)
		if err == nil {
			hb.setState(state.Count, state.Target)
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("upgrade error: %v", err)
			return
		}

		hb.add(conn)
		log.Printf("client connected: %s", conn.RemoteAddr())

		defer func() {
			hb.remove(conn)
			log.Printf("client disconnected: %s", conn.RemoteAddr())
		}()

		count, target := hb.getState()
		statePayload, err := json.Marshal(Message{Type: "state", Count: count, Target: target})
		if err != nil {
			log.Printf("state marshal error: %v", err)
			return
		}
		if err := conn.WriteMessage(websocket.TextMessage, statePayload); err != nil {
			log.Printf("state write error: %v", err)
			return
		}

		conn.SetReadLimit(512)
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		for {
			msgType, payload, err := conn.ReadMessage()
			if err != nil {
				log.Printf("read error: %v", err)
				return
			}

			if msgType != websocket.TextMessage {
				continue
			}

			var incoming Message
			if err := json.Unmarshal(payload, &incoming); err != nil {
				log.Printf("invalid message: %v", err)
				continue
			}

			if incoming.Type != "increment" {
				continue
			}

			count, target, completed, err := hb.increment(r.Context())
			if err != nil {
				log.Printf("increment error: %v", err)
				continue
			}
			statePayload, err := json.Marshal(Message{Type: "state", Count: count, Target: target})
			if err != nil {
				log.Printf("state marshal error: %v", err)
				continue
			}
			hb.broadcast(websocket.TextMessage, statePayload)

			if completed {
				completedPayload, err := json.Marshal(Message{Type: "completed"})
				if err != nil {
					log.Printf("completed marshal error: %v", err)
					continue
				}
				hb.broadcast(websocket.TextMessage, completedPayload)
			}
		}
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("response encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := initDB(ctx)
	if err != nil {
		log.Fatalf("database init error: %v", err)
	}
	defer db.Close()

	hubs := newHubStore(db)

	srv := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           registerRoutes(hubs, db),
	}

	go func() {
		log.Printf("server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	hubs.mu.Lock()
	for _, hb := range hubs.hubs {
		hb.mu.Lock()
		for conn := range hb.clients {
			conn.Close()
		}
		hb.mu.Unlock()
	}
	hubs.mu.Unlock()

	log.Println("server stopped cleanly")
}
