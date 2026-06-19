package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Errors returned by token operations.
var (
	ErrInvalidToken = errors.New("auth: invalid token")
	ErrExpiredToken = errors.New("auth: token expired")
)

// Claims are the custom JWT claims embedded in access tokens.
type Claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

// TokenManager issues and verifies HS256 JWTs.
type TokenManager struct {
	secret    []byte
	accessTTL time.Duration
}

// NewTokenManager constructs a TokenManager. The secret must be non-empty.
func NewTokenManager(secret string, accessTTL time.Duration) (*TokenManager, error) {
	if secret == "" {
		return nil, errors.New("auth: jwt secret must be set")
	}
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	return &TokenManager{secret: []byte(secret), accessTTL: accessTTL}, nil
}

// Issue creates a signed access token for the given user id.
func (tm *TokenManager) Issue(userID string) (token string, expiresAt time.Time, err error) {
	now := time.Now()
	expiresAt = now.Add(tm.accessTTL)
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "ventiqra",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(tm.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

// Parse validates the token signature and expiry, returning its claims.
func (tm *TokenManager) Parse(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method %v", ErrInvalidToken, t.Header["alg"])
		}
		return tm.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	if claims.UserID == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
