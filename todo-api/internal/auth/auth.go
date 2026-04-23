package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

// userIDKey is the context key under which the authenticated user id is stored.
const userIDKey ctxKey = "userID"

// TokenTTL is the lifetime of issued JWTs.
const TokenTTL = 24 * time.Hour

// Secret returns the signing secret, falling back to "devsecret" if unset.
func Secret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "devsecret"
	}
	return []byte(s)
}

// IssueToken returns a signed JWT for the given user id with a 24h expiry.
func IssueToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(TokenTTL)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(Secret())
}

// ParseToken validates a signed JWT and returns the subject (userID).
func ParseToken(raw string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return Secret(), nil
	})
	if err != nil {
		return "", err
	}
	if claims.Subject == "" {
		return "", errors.New("missing subject")
	}
	return claims.Subject, nil
}

// UserIDFromContext returns the user id injected by the auth middleware.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok && v != ""
}

// WithUserID returns a new context carrying the given user id.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// Middleware wraps the supplied handler, requiring a valid bearer token.
// On failure it writes a 401 response and does not invoke next.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			writeUnauthorized(w)
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			writeUnauthorized(w)
			return
		}
		userID, err := ParseToken(parts[1])
		if err != nil {
			writeUnauthorized(w)
			return
		}
		ctx := WithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}
