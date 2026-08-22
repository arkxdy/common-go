package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// Middleware is a generic JWT authentication middleware.
// It is designed to work with any net/http compatible router or framework.
type Middleware struct {
	validate       func(string) (*Claims, error)
	requiredRole   Role
	userIDKey      any
	claimsKey      any
	roleKey        any
	tokenExtractor func(*http.Request) (string, error)
	errorHandler   func(w http.ResponseWriter, r *http.Request, status int, err error)
}

// Option configures Middleware.
type Option func(*Middleware)

// WithRequiredRole enforces a specific role (e.g. RoleAdmin).
func WithRequiredRole(role Role) Option {
	return func(m *Middleware) {
		m.requiredRole = role
	}
}

// WithContextKeys sets custom context keys.
// Defaults: "auth.userID", "auth.claims", "auth.role".
func WithContextKeys(userIDKey, claimsKey, roleKey any) Option {
	return func(m *Middleware) {
		m.userIDKey = userIDKey
		m.claimsKey = claimsKey
		m.roleKey = roleKey
	}
}

// WithTokenExtractor replaces the default Bearer token extractor.
func WithTokenExtractor(extractor func(*http.Request) (string, error)) Option {
	return func(m *Middleware) {
		m.tokenExtractor = extractor
	}
}

// WithErrorHandler customizes the error response.
func WithErrorHandler(handler func(w http.ResponseWriter, r *http.Request, status int, err error)) Option {
	return func(m *Middleware) {
		m.errorHandler = handler
	}
}

// NewMiddleware creates a new Middleware with a token validation function.
func NewMiddleware(validate func(string) (*Claims, error), opts ...Option) *Middleware {
	m := &Middleware{
		validate:       validate,
		userIDKey:      "auth.userID",
		claimsKey:      "auth.claims",
		roleKey:        "auth.role",
		tokenExtractor: defaultTokenExtractor,
		errorHandler:   defaultErrorHandler,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func defaultTokenExtractor(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", errors.New("authorization header is required")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header format")
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("empty token")
	}
	return token, nil
}

func defaultErrorHandler(w http.ResponseWriter, r *http.Request, status int, err error) {
	http.Error(w, err.Error(), status)
}

// Authenticate performs token extraction, validation, and role checking.
// It returns the claims, suggested HTTP status, and error.
func (m *Middleware) Authenticate(r *http.Request) (*Claims, int, error) {
	token, err := m.tokenExtractor(r)
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}

	claims, err := m.validate(token)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, ErrExpiredToken) {
			status = http.StatusForbidden
		}
		return nil, status, err
	}

	if m.requiredRole != "" && claims.Role != m.requiredRole {
		return nil, http.StatusForbidden, errors.New("insufficient role")
	}

	return claims, http.StatusOK, nil
}

// Handler wraps an http.Handler with JWT authentication.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, status, err := m.Authenticate(r)
		if err != nil {
			m.errorHandler(w, r, status, err)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, m.userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, m.claimsKey, claims)
		ctx = context.WithValue(ctx, m.roleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HandlerFunc is a convenience wrapper for http.HandlerFunc.
func (m *Middleware) HandlerFunc(next http.HandlerFunc) http.Handler {
	return m.Handler(next)
}
