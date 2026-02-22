package main

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
)

var (
	errHatimNotFound   = errors.New("hatim not found")
	errInvalidPassword = errors.New("invalid password")
)

const (
	defaultTarget   = 50
	shareCodeLength = 8
	tokenBytes      = 32
	argonTime       = 1
	argonMemory     = 64 * 1024
	argonThreads    = 4
	argonKeyLength  = 32
)

type hatim struct {
	ID           uuid.UUID
	ShareCode    string
	Title        string
	Count        int
	Target       int
	PasswordHash *string
}

type hatimState struct {
	ShareCode        string
	Title            string
	Count            int
	Target           int
	RequiresPassword bool
}

type hatimSummary struct {
	ShareCode        string
	Title            string
	Count            int
	Target           int
	RequiresPassword bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type updateHatimInput struct {
	Title  *string
	Count  *int
	Target *int
}

func createHatim(ctx context.Context, pool *pgxpool.Pool, title string, target int, password string) (hatim, string, error) {
	if target <= 0 {
		target = defaultTarget
	}

	var passwordHash *string
	if strings.TrimSpace(password) != "" {
		hash, err := hashPassword(password)
		if err != nil {
			return hatim{}, "", err
		}
		passwordHash = &hash
	}

	var created hatim
	for attempt := 0; attempt < 5; attempt++ {
		shareCode, err := generateShareCode()
		if err != nil {
			return hatim{}, "", err
		}
		id := uuid.New()

		_, err = pool.Exec(ctx, `
			INSERT INTO hatims (id, share_code, title, target, password_hash)
			VALUES ($1, $2, $3, $4, $5)
		`, id, shareCode, title, target, passwordHash)
		if err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return hatim{}, "", err
		}

		created = hatim{
			ID:           id,
			ShareCode:    shareCode,
			Title:        title,
			Count:        0,
			Target:       target,
			PasswordHash: passwordHash,
		}
		break
	}

	if created.ShareCode == "" {
		return hatim{}, "", errors.New("unable to generate unique share code")
	}

	token, err := createToken(ctx, pool, created.ID)
	if err != nil {
		return hatim{}, "", err
	}

	return created, token, nil
}

func getHatimState(ctx context.Context, pool *pgxpool.Pool, shareCode string) (hatimState, error) {
	var state hatimState
	row := pool.QueryRow(ctx, `
		SELECT share_code, title, count, target, password_hash
		FROM hatims
		WHERE share_code = $1
	`, shareCode)

	var passwordHash *string
	if err := row.Scan(&state.ShareCode, &state.Title, &state.Count, &state.Target, &passwordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return hatimState{}, errHatimNotFound
		}
		return hatimState{}, err
	}

	state.RequiresPassword = passwordHash != nil
	return state, nil
}

func getHatimAuth(ctx context.Context, pool *pgxpool.Pool, shareCode string) (hatim, error) {
	var result hatim
	row := pool.QueryRow(ctx, `
		SELECT id, share_code, count, target, password_hash
		FROM hatims
		WHERE share_code = $1
	`, shareCode)

	if err := row.Scan(&result.ID, &result.ShareCode, &result.Count, &result.Target, &result.PasswordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return hatim{}, errHatimNotFound
		}
		return hatim{}, err
	}

	return result, nil
}

func incrementHatim(ctx context.Context, pool *pgxpool.Pool, shareCode string, amount int) (int, int, bool, error) {
	if amount <= 0 {
		amount = 1
	}

	var count int
	var target int
	row := pool.QueryRow(ctx, `
		UPDATE hatims
		SET count = LEAST(count + $2, target), updated_at = NOW()
		WHERE share_code = $1
		RETURNING count, target
	`, shareCode, amount)

	if err := row.Scan(&count, &target); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, false, errHatimNotFound
		}
		return 0, 0, false, err
	}

	return count, target, count >= target, nil
}

func listHatims(ctx context.Context, pool *pgxpool.Pool) ([]hatimSummary, error) {
	rows, err := pool.Query(ctx, `
		SELECT share_code, title, count, target, password_hash, created_at, updated_at
		FROM hatims
		ORDER BY updated_at DESC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]hatimSummary, 0)
	for rows.Next() {
		var item hatimSummary
		var passwordHash *string
		if err := rows.Scan(
			&item.ShareCode,
			&item.Title,
			&item.Count,
			&item.Target,
			&passwordHash,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.RequiresPassword = passwordHash != nil
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func updateHatim(ctx context.Context, pool *pgxpool.Pool, shareCode string, input updateHatimInput) (hatimState, error) {
	title := input.Title
	count := input.Count
	target := input.Target

	if title == nil && count == nil && target == nil {
		return getHatimState(ctx, pool, shareCode)
	}

	row := pool.QueryRow(ctx, `
		UPDATE hatims
		SET
			title = COALESCE($2, title),
			target = CASE
				WHEN $4::bool THEN GREATEST(COALESCE($3, target), 1)
				ELSE target
			END,
			count = CASE
				WHEN $5::bool AND $4::bool THEN LEAST(GREATEST(COALESCE($6, count), 0), GREATEST(COALESCE($3, target), 1))
				WHEN $5::bool THEN LEAST(GREATEST(COALESCE($6, count), 0), target)
				WHEN $4::bool THEN LEAST(count, GREATEST(COALESCE($3, target), 1))
				ELSE count
			END,
			updated_at = NOW()
		WHERE share_code = $1
		RETURNING share_code, title, count, target, password_hash
	`, shareCode, title, target, target != nil, count != nil, count)

	var state hatimState
	var passwordHash *string
	if err := row.Scan(&state.ShareCode, &state.Title, &state.Count, &state.Target, &passwordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return hatimState{}, errHatimNotFound
		}
		return hatimState{}, err
	}

	state.RequiresPassword = passwordHash != nil
	return state, nil
}

func deleteHatim(ctx context.Context, pool *pgxpool.Pool, shareCode string) error {
	cmd, err := pool.Exec(ctx, `
		DELETE FROM hatims
		WHERE share_code = $1
	`, shareCode)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errHatimNotFound
	}
	return nil
}

func createToken(ctx context.Context, pool *pgxpool.Pool, hatimID uuid.UUID) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO hatim_tokens (token, hatim_id)
		VALUES ($1, $2)
	`, token, hatimID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func validateToken(ctx context.Context, pool *pgxpool.Pool, shareCode string, token string) (bool, error) {
	if token == "" {
		return false, nil
	}

	row := pool.QueryRow(ctx, `
		SELECT 1
		FROM hatim_tokens t
		JOIN hatims h ON h.id = t.hatim_id
		WHERE t.token = $1 AND h.share_code = $2
	`, token, shareCode)

	var exists int
	if err := row.Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", argonMemory, argonTime, argonThreads, b64Salt, b64Hash)
	return encoded, nil
}

func verifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var memory uint32
	var timeCost uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	computed := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(decodedHash)))
	return subtleCompare(decodedHash, computed), nil
}

func subtleCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

func generateShareCode() (string, error) {
	buf := make([]byte, 5)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
	encoded = strings.ToLower(encoded)
	if len(encoded) < shareCodeLength {
		return "", errors.New("short share code")
	}
	return encoded[:shareCodeLength], nil
}

func generateToken() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func joinHatim(ctx context.Context, pool *pgxpool.Pool, shareCode string, password string) (string, error) {
	h, err := getHatimAuth(ctx, pool, shareCode)
	if err != nil {
		return "", err
	}

	if h.PasswordHash != nil {
		ok, err := verifyPassword(password, *h.PasswordHash)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", errInvalidPassword
		}
	}

	return createToken(ctx, pool, h.ID)
}

func getHatimForWS(ctx context.Context, pool *pgxpool.Pool, shareCode string) (hatim, error) {
	return getHatimAuth(ctx, pool, shareCode)
}
