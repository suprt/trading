package infra

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/suprt/trading/services/user/internal/domain"
)

type JWTConfig struct {
	Secret []byte
	TTL    time.Duration
}
type JWTIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTIssuer(cfg JWTConfig) *JWTIssuer {
	return &JWTIssuer{secret: cfg.Secret, ttl: cfg.TTL}
}

func (j *JWTIssuer) IssueAccess(userID domain.UserID, roles []domain.Role) (domain.AccessToken, error) {
	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = string(role)
	}
	claims := jwt.MapClaims{
		"sub":   string(userID),
		"roles": roleStrings,
		"exp":   time.Now().Add(j.ttl).Unix(),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secret)
	if err != nil {
		return "", err
	}
	return domain.AccessToken(signed), nil
}

func (j *JWTIssuer) NewRefresh() (domain.RefreshToken, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return domain.RefreshToken(base64.URLEncoding.EncodeToString(b)), nil
}

func (j *JWTIssuer) ParseAccess(token domain.AccessToken) (domain.UserID, []domain.Role, error) {
	parsed, err := jwt.Parse(token.String(), func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil {
		return "", nil, domain.ErrInvalidToken
	}
	if !parsed.Valid {
		return "", nil, domain.ErrInvalidToken
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return "", nil, domain.ErrInvalidToken
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", nil, domain.ErrInvalidToken
	}

	var roles []domain.Role
	if raw, ok := claims["roles"].([]any); ok {
		for _, r := range raw {
			if s, ok := r.(string); ok {
				roles = append(roles, domain.Role(s))
			}
		}
	}
	return domain.UserID(sub), roles, nil
}
