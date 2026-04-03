package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tf-agent/tf-agent/internal/db"
)

type userContextKey struct{}

// generateToken returns a tfa-prefixed cryptographically random token.
// Format: tfa-<40 hex chars>  (e.g. tfa-3a7f2b...)
func generateToken() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "tfa-" + hex.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hex of a raw token.
// Tokens are long random strings so SHA-256 is sufficient without bcrypt.
func HashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// authMiddleware validates the Bearer token and injects the user into context.
func authMiddleware(store db.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		user, err := store.GetUserByTokenHash(r.Context(), HashToken(token))
		if err != nil || user == nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// adminMiddleware requires role=admin.
func adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := userFromContext(r.Context())
		if user == nil || user.Role != "admin" {
			writeError(w, http.StatusForbidden, "admin required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func userFromContext(ctx context.Context) *db.User {
	u, _ := ctx.Value(userContextKey{}).(*db.User)
	return u
}

func bearerToken(r *http.Request) string {
	// Check Authorization header first.
	h := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(h, "Bearer "); ok {
		return strings.TrimSpace(after)
	}
	// Fall back to ?token= query param (needed for EventSource which can't set headers).
	if t := r.URL.Query().Get("token"); t != "" {
		return t
	}
	return ""
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
