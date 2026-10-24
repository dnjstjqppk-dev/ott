package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken = errors.New("invalid or expired access token")
)

type AccessClaims struct {
	UserID      string `json:"uid"`
	ProfileID   string `json:"pid"`
	DeviceID    string `json:"did"`
	Role        string `json:"rol"`
	MaxMaturity string `json:"mat"`
	CountryCode string `json:"cnt"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	signingKey []byte
	accessTTL  time.Duration
}

func NewTokenManager(secret string, accessTTL time.Duration) *TokenManager {
	return &TokenManager{
		signingKey: []byte(secret),
		accessTTL:  accessTTL,
	}
}

func (m *TokenManager) IssueAccessToken(uid, pid, did, role, maturity, country string) (string, int64, error) {
	now := time.Now().UTC()
	exp := now.Add(m.accessTTL)
	claims := AccessClaims{
		UserID:      uid,
		ProfileID:   pid,
		DeviceID:    did,
		Role:        role,
		MaxMaturity: maturity,
		CountryCode: country,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "netflix-iam-svc",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.signingKey)
	return signed, int64(m.accessTTL.Seconds()), err
}

func (m *TokenManager) VerifyAccessToken(rawToken string) (*AccessClaims, error) {
	parsed, err := jwt.ParseWithClaims(rawToken, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.signingKey, nil
	})
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := parsed.Claims.(*AccessClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func GenerateRefreshTokenPair() (rawToken string, sha256Hex string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	rawToken = hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(rawToken))
	return rawToken, hex.EncodeToString(sum[:]), nil
}

func HashSHA256(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}

func HashSecretBcrypt(raw string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), 12)
	return string(hashed), err
}

func CompareBcrypt(hashed, raw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(raw)) == nil
}

