package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}

// TokenConfig holds everything needed to generate and validate tokens.
type TokenConfig struct {
	PrivateKey    *rsa.PrivateKey
	PublicKey     *rsa.PublicKey
	Issuer        string
	Audience      jwt.ClaimStrings
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	KeyID         string
}

// Validator validates JWTs using a static public key and expected claims.
type Validator struct {
	PublicKey *rsa.PublicKey
	Issuer    string
	Audience  jwt.ClaimStrings
}

// NewValidator creates a Validator.
func NewValidator(publicKey *rsa.PublicKey, issuer string, audience []string) *Validator {
	return &Validator{
		PublicKey: publicKey,
		Issuer:    issuer,
		Audience:  audience,
	}
}

// ValidateToken parses and verifies a JWT string.
func (v *Validator) ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		// Ensure signing method is RSA
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		return v.PublicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Verify issuer
	if claims.Issuer != v.Issuer {
		return nil, errors.New("invalid token issuer")
	}

	// Verify audience manually (does not rely on jwt.VerifyAudience)
	if len(v.Audience) > 0 {
		validAudience := false
		for _, aud := range claims.Audience {
			for _, expected := range v.Audience {
				if aud == expected {
					validAudience = true
					break
				}
			}
			if validAudience {
				break
			}
		}
		if !validAudience {
			return nil, errors.New("invalid token audience")
		}
	}

	return claims, nil
}

// TokenGenerator creates signed JWTs.
type TokenGenerator struct {
	cfg TokenConfig
}

// NewTokenGenerator returns a TokenGenerator using the given config.
func NewTokenGenerator(cfg TokenConfig) *TokenGenerator {
	return &TokenGenerator{cfg: cfg}
}

// GenerateTokens creates an access/refresh pair for a regular user.
func (g *TokenGenerator) GenerateTokens(userID string) (access, refresh, accessJTI, refreshJTI string, err error) {
	return g.generatePair(userID, RoleUser)
}

// GenerateAdminTokens creates an access/refresh pair for an admin.
func (g *TokenGenerator) GenerateAdminTokens(adminID string) (access, refresh, accessJTI, refreshJTI string, err error) {
	return g.generatePair(adminID, RoleAdmin)
}

func (g *TokenGenerator) generatePair(userID string, role Role) (string, string, string, string, error) {
	accessJTI := uuid.New().String()
	refreshJTI := uuid.New().String()

	access, err := g.generateToken(userID, accessJTI, g.cfg.AccessExpiry, role)
	if err != nil {
		return "", "", "", "", err
	}
	refresh, err := g.generateToken(userID, refreshJTI, g.cfg.RefreshExpiry, role)
	if err != nil {
		return "", "", "", "", err
	}
	return access, refresh, accessJTI, refreshJTI, nil
}

func (g *TokenGenerator) generateToken(userID, jti string, expiration time.Duration, role Role) (string, error) {
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    g.cfg.Issuer,
			Audience:  g.cfg.Audience,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = g.cfg.KeyID
	return token.SignedString(g.cfg.PrivateKey)
}

// LoadKeys reads RSA private and public keys from PEM files.
func LoadKeys(privateKeyPath, publicKeyPath string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	priv, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load private key: %w", err)
	}
	pub, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load public key: %w", err)
	}
	return priv, pub, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	return key, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return key, nil
}

// DefaultTokenConfigFromEnv builds a TokenConfig using environment variables.
func DefaultTokenConfigFromEnv() TokenConfig {
	accessExpiry := 15 * time.Minute
	if v := os.Getenv("JWT_ACCESS_EXPIRY"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			accessExpiry = d
		}
	}
	refreshExpiry := 24 * time.Hour
	if v := os.Getenv("JWT_REFRESH_EXPIRY"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			refreshExpiry = d
		}
	}
	return TokenConfig{
		Issuer:        os.Getenv("AUTH_SERVICE_URL"),
		Audience:      jwt.ClaimStrings{os.Getenv("WEB_URL"), os.Getenv("MOBILE_URL")},
		AccessExpiry:  accessExpiry,
		RefreshExpiry: refreshExpiry,
		KeyID:         "key1",
	}
}
