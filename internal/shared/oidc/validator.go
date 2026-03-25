package oidc

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kleff/platform/internal/shared/middleware"
)

// Validator validates JWTs using OIDC JWKS public keys.
// It caches keys and refreshes them on cache miss.
type Validator struct {
	jwksURI string
	client  *http.Client

	mu   sync.RWMutex
	keys map[string]*rsa.PublicKey
}

// NewValidator creates a Validator that fetches keys from the given JWKS URI.
func NewValidator(jwksURI string) *Validator {
	return &Validator{
		jwksURI: jwksURI,
		client:  &http.Client{Timeout: 10 * time.Second},
		keys:    make(map[string]*rsa.PublicKey),
	}
}

// Verify validates a JWT and returns its claims. Implements middleware.TokenVerifier.
func (v *Validator) Verify(ctx context.Context, token string) (*middleware.VerifyResult, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed JWT")
	}

	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := decodeSegment(parts[0], &header); err != nil {
		return nil, fmt.Errorf("decode JWT header: %w", err)
	}
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported JWT algorithm %q", header.Alg)
	}

	key, err := v.getKey(ctx, header.Kid)
	if err != nil {
		return nil, err
	}

	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode JWT signature: %w", err)
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], sigBytes); err != nil {
		return nil, fmt.Errorf("invalid JWT signature: %w", err)
	}

	var claims struct {
		Sub         string   `json:"sub"`
		Exp         int64    `json:"exp"`
		Roles       []string `json:"roles"`
		RealmAccess struct {
			Roles []string `json:"roles"`
		} `json:"realm_access"`
	}
	if err := decodeSegment(parts[1], &claims); err != nil {
		return nil, fmt.Errorf("decode JWT claims: %w", err)
	}
	if claims.Sub == "" {
		return nil, fmt.Errorf("JWT missing sub claim")
	}
	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("JWT expired")
	}

	roles := append(claims.Roles, claims.RealmAccess.Roles...)
	if roles == nil {
		roles = []string{}
	}

	return &middleware.VerifyResult{Subject: claims.Sub, Roles: roles}, nil
}

func (v *Validator) getKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	key, ok := v.keys[kid]
	v.mu.RUnlock()
	if ok {
		return key, nil
	}
	if err := v.fetchKeys(ctx); err != nil {
		return nil, fmt.Errorf("fetch JWKS: %w", err)
	}
	v.mu.RLock()
	key, ok = v.keys[kid]
	v.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown JWKS key ID %q", kid)
	}
	return key, nil
}

type jwkSet struct {
	Keys []struct {
		Kid string `json:"kid"`
		Kty string `json:"kty"`
		Use string `json:"use"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

func (v *Validator) fetchKeys(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURI, nil)
	if err != nil {
		return err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned %d", resp.StatusCode)
	}

	var set jwkSet
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return err
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	for _, k := range set.Keys {
		if k.Kty != "RSA" || k.Use != "sig" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		pub := &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: int(new(big.Int).SetBytes(eBytes).Int64()),
		}
		v.keys[k.Kid] = pub
	}
	return nil
}

func decodeSegment(seg string, v any) error {
	b, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
