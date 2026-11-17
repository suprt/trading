package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/suprt/trading/user-service/internal/core/domain"
	er "github.com/suprt/trading/user-service/internal/core/errors"
)

// 32 байта = 256 бит энтропии.
//
//	После RawURLEncoding → 43 символа base64.
//	Если меняешь — обнови (buf.validate.field).string.len в user.proto.
const refreshEntropyBytes = 32

type Config struct {
	Secret []byte
	TTL    time.Duration
}
type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTIssuer(cfg Config) *Issuer {
	return &Issuer{secret: cfg.Secret, ttl: cfg.TTL}
}

func (j *Issuer) IssueAccess(userID domain.UserID, roles []domain.Role, rolesVersion int64) (domain.AccessToken, error) {
	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = string(role)
	}
	claims := jwt.MapClaims{
		"sub":   string(userID),
		"roles": roleStrings,
		"rv":    rolesVersion,
		"exp":   time.Now().Add(j.ttl).Unix(),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secret)
	if err != nil {
		return "", err
	}
	return domain.AccessToken(signed), nil
}

func (j *Issuer) NewRefresh() (domain.RefreshToken, error) {
	b := make([]byte, refreshEntropyBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return domain.RefreshToken(base64.RawURLEncoding.EncodeToString(b)), nil
}

func (j *Issuer) ParseAccess(token domain.AccessToken) (domain.UserID, []domain.Role, int64, error) {
	parsed, err := jwt.Parse(token.String(), func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	}, jwt.WithJSONNumber())
	if err != nil {
		return "", nil, 0, er.ErrInvalidToken
	}
	if !parsed.Valid {
		return "", nil, 0, er.ErrInvalidToken
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return "", nil, 0, er.ErrInvalidToken
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", nil, 0, er.ErrInvalidToken
	}

	var rv int64
	if num, ok := claims["rv"].(json.Number); ok {
		if v, err := num.Int64(); err == nil {
			rv = v
		}
	}

	var roles []domain.Role
	if raw, ok := claims["roles"].([]any); ok {
		for _, r := range raw {
			if s, ok := r.(string); ok {
				roles = append(roles, domain.Role(s))
			}
		}
	}

	return domain.UserID(sub), roles, rv, nil
}
