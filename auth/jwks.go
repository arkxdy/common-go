package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
)

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

// NewJWKS builds a JWKS from a public RSA key.
// The kid parameter identifies the key and should match the one used in token headers.
func NewJWKS(publicKey *rsa.PublicKey, kid string) JWKS {
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes())

	return JWKS{
		Keys: []JWK{
			{
				Kty: "RSA",
				Kid: kid,
				Use: "sig",
				Alg: "RS256",
				N:   n,
				E:   e,
			},
		},
	}
}

// JWKSHandler returns an http.Handler that serves the JWKS as JSON.
func JWKSHandler(publicKey *rsa.PublicKey, kid string) http.Handler {
	jwks := NewJWKS(publicKey, kid)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	})
}
