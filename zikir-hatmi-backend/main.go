package main

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
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
	clients   map[*websocket.Conn]*clientInfo
	shareCode string
	db        *pgxpool.Pool
	count     int
	target    int
}

type clientInfo struct {
	id   string
	name string
}

type presenceUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func newHub(db *pgxpool.Pool, shareCode string, count int, target int) *hub {
	return &hub{
		clients:   make(map[*websocket.Conn]*clientInfo),
		shareCode: shareCode,
		db:        db,
		count:     count,
		target:    target,
	}
}

var fallbackCounter uint64

func randomID() string {
	const alphabet = "23456789abcdefghjkmnpqrstuvwxyz"
	b := make([]byte, 8)
	randomBytes := make([]byte, 8)
	if _, err := crand.Read(randomBytes); err != nil {
		// Fallback to time-based unique ID if crypto/rand fails
		log.Printf("crypto/rand error: %v", err)
		fallbackCounter++
		ts := time.Now().UnixNano()
		fallbackID := fmt.Sprintf("%x%x", ts, fallbackCounter)
		if len(fallbackID) > 8 {
			return fallbackID[:8]
		}
		return fallbackID
	}
	for i := range b {
		b[i] = alphabet[int(randomBytes[i])%len(alphabet)]
	}
	return string(b)
}

func normalizeName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if len([]rune(name)) > 24 {
		r := []rune(name)
		name = string(r[:24])
	}
	return name
}

func (h *hub) add(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = &clientInfo{id: randomID(), name: ""}
	h.mu.Unlock()
}

func (h *hub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	_ = conn.Close()
}

func (h *hub) setClientName(conn *websocket.Conn, name string) {
	name = normalizeName(name)
	h.mu.Lock()
	if info, ok := h.clients[conn]; ok {
		info.name = name
	}
	h.mu.Unlock()
}

func (h *hub) presenceSnapshot() []presenceUser {
	h.mu.RLock()
	users := make([]presenceUser, 0, len(h.clients))
	for _, info := range h.clients {
		name := normalizeName(info.name)
		if name == "" {
			name = "Misafir"
		}
		users = append(users, presenceUser{ID: info.id, Name: name})
	}
	h.mu.RUnlock()

	sort.Slice(users, func(i, j int) bool {
		if users[i].Name == users[j].Name {
			return users[i].ID < users[j].ID
		}
		return users[i].Name < users[j].Name
	})
	return users
}

func (h *hub) broadcastPresence() {
	users := h.presenceSnapshot()
	payload, err := json.Marshal(Message{Type: "presence", Users: users})
	if err != nil {
		log.Printf("presence marshal error: %v", err)
		return
	}
	h.broadcast(websocket.TextMessage, payload)
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

func (h *hub) increment(ctx context.Context, amount int) (int, int, bool, error) {
	if amount <= 0 {
		amount = 1
	}
	if amount > 1000 {
		amount = 1000
	}

	count, target, completed, err := incrementHatim(ctx, h.db, h.shareCode, amount)
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

// rateLimiter provides simple rate limiting for authentication attempts
type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int           // max attempts
	window   time.Duration // time window
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove old attempts
	filtered := make([]time.Time, 0)
	for _, t := range rl.attempts[key] {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= rl.limit {
		rl.attempts[key] = filtered
		return false
	}

	rl.attempts[key] = append(filtered, now)
	return true
}

// trustProxy controls whether to trust X-Forwarded-For headers.
// Set TRUST_PROXY=true when running behind a trusted reverse proxy.
var trustProxy = os.Getenv("TRUST_PROXY") == "true"

func getClientIP(r *http.Request) string {
	// Only trust proxy headers if explicitly configured
	if trustProxy {
		// Check X-Forwarded-For header (for proxied requests)
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// Take the first IP in the list (client IP as added by the first proxy)
			if idx := strings.Index(xff, ","); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}
		// Check X-Real-IP header
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return xri
		}
	}
	// Fall back to RemoteAddr (direct connection IP)
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

var allowedOrigins = parseAllowedOrigins()

func parseAllowedOrigins() map[string]bool {
	origins := make(map[string]bool)
	// Default allowed origins for development
	origins["localhost"] = true
	origins["127.0.0.1"] = true

	// Parse ALLOWED_ORIGINS environment variable (comma-separated)
	if envOrigins := os.Getenv("ALLOWED_ORIGINS"); envOrigins != "" {
		for _, origin := range strings.Split(envOrigins, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				origins[origin] = true
			}
		}
	}
	return origins
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Allow requests without Origin header (same-origin requests)
		return true
	}

	// Parse the origin URL to extract the host
	// Origin format: scheme://host[:port]
	origin = strings.TrimPrefix(origin, "http://")
	origin = strings.TrimPrefix(origin, "https://")
	// Remove port if present
	if colonIdx := strings.Index(origin, ":"); colonIdx != -1 {
		origin = origin[:colonIdx]
	}

	return allowedOrigins[origin]
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
	EnableCompression: true,
}

type Message struct {
	Type   string         `json:"type"`
	Count  int            `json:"count,omitempty"`
	Target int            `json:"target,omitempty"`
	Amount int            `json:"amount,omitempty"`
	Name   string         `json:"name,omitempty"`
	Users  []presenceUser `json:"users,omitempty"`
}

type createHatimRequest struct {
	Title    string `json:"title"`
	Target   int    `json:"target"`
	Password string `json:"password"`
}

type updateHatimRequest struct {
	Title  *string `json:"title"`
	Count  *int    `json:"count"`
	Target *int    `json:"target"`
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

type hatimSummaryResponse struct {
	ShareCode        string    `json:"shareCode"`
	Title            string    `json:"title"`
	Count            int       `json:"count"`
	Target           int       `json:"target"`
	RequiresPassword bool      `json:"requiresPassword"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func (s *hubStore) get(shareCode string) *hub {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hubs[shareCode]
}

func (s *hubStore) remove(shareCode string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.hubs, shareCode)
}

func registerRoutes(hubs *hubStore, db *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	// Rate limiter: 10 join attempts per minute per IP
	joinLimiter := newRateLimiter(10, time.Minute)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/hatims", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
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
		case http.MethodGet:
			// Require admin key for listing all hatims
			adminKey := os.Getenv("ADMIN_KEY")
			if adminKey == "" {
				writeError(w, http.StatusForbidden, "admin listing disabled")
				return
			}
			providedKey := extractBearerToken(r)
			if providedKey != adminKey {
				writeError(w, http.StatusUnauthorized, "invalid admin key")
				return
			}

			items, err := listHatims(r.Context(), db)
			if err != nil {
				log.Printf("list hatims error: %v", err)
				writeError(w, http.StatusInternalServerError, "unable to list hatims")
				return
			}

			result := make([]hatimSummaryResponse, 0, len(items))
			for _, item := range items {
				result = append(result, hatimSummaryResponse{
					ShareCode:        item.ShareCode,
					Title:            item.Title,
					Count:            item.Count,
					Target:           item.Target,
					RequiresPassword: item.RequiresPassword,
					CreatedAt:        item.CreatedAt,
					UpdatedAt:        item.UpdatedAt,
				})
			}
			writeJSON(w, http.StatusOK, result)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
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
			switch r.Method {
			case http.MethodGet:
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
			case http.MethodPatch:
				// Require valid token for updates
				token := extractBearerToken(r)
				ok, err := validateToken(r.Context(), db, shareCode, token)
				if err != nil {
					log.Printf("token validation error: %v", err)
					writeError(w, http.StatusInternalServerError, "unable to validate token")
					return
				}
				if !ok {
					writeError(w, http.StatusUnauthorized, "invalid or missing token")
					return
				}

				var payload updateHatimRequest
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
					writeError(w, http.StatusBadRequest, "invalid body")
					return
				}

				state, err := updateHatim(r.Context(), db, shareCode, updateHatimInput{
					Title:  payload.Title,
					Count:  payload.Count,
					Target: payload.Target,
				})
				if err != nil {
					if errors.Is(err, errHatimNotFound) {
						writeError(w, http.StatusNotFound, "hatim not found")
						return
					}
					log.Printf("update hatim error: %v", err)
					writeError(w, http.StatusInternalServerError, "unable to update hatim")
					return
				}

				if hb := hubs.get(shareCode); hb != nil {
					hb.setState(state.Count, state.Target)
					statePayload, err := json.Marshal(Message{Type: "state", Count: state.Count, Target: state.Target})
					if err == nil {
						hb.broadcast(websocket.TextMessage, statePayload)
					}
				}

				writeJSON(w, http.StatusOK, hatimResponse{
					ShareCode:        state.ShareCode,
					Title:            state.Title,
					Count:            state.Count,
					Target:           state.Target,
					RequiresPassword: state.RequiresPassword,
				})
			case http.MethodDelete:
				// Require valid token for deletion
				token := extractBearerToken(r)
				ok, err := validateToken(r.Context(), db, shareCode, token)
				if err != nil {
					log.Printf("token validation error: %v", err)
					writeError(w, http.StatusInternalServerError, "unable to validate token")
					return
				}
				if !ok {
					writeError(w, http.StatusUnauthorized, "invalid or missing token")
					return
				}

				err = deleteHatim(r.Context(), db, shareCode)
				if err != nil {
					if errors.Is(err, errHatimNotFound) {
						writeError(w, http.StatusNotFound, "hatim not found")
						return
					}
					log.Printf("delete hatim error: %v", err)
					writeError(w, http.StatusInternalServerError, "unable to delete hatim")
					return
				}

				hubs.remove(shareCode)
				writeJSON(w, http.StatusNoContent, nil)
			default:
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}

		if len(parts) == 2 && parts[1] == "join" {
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}

			// Rate limit join attempts to prevent brute force attacks
			// Using "/" as delimiter since shareCode is base32 encoded (no "/" chars)
			clientIP := getClientIP(r)
			if !joinLimiter.allow(clientIP + "/" + shareCode) {
				writeError(w, http.StatusTooManyRequests, "too many attempts, please try again later")
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
		hb.broadcastPresence()

		defer func() {
			hb.remove(conn)
			hb.broadcastPresence()
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

			switch incoming.Type {
			case "increment":
				// handled below
			case "hello", "set_name":
				hb.setClientName(conn, incoming.Name)
				hb.broadcastPresence()
				continue
			default:
				continue
			}

			count, target, completed, err := hb.increment(r.Context(), incoming.Amount)
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

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(auth) < len(prefix) || auth[:len(prefix)] != prefix {
		return ""
	}
	return auth[len(prefix):]
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
