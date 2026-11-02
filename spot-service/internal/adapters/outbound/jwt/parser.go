package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
)

type Parser struct {
	secret []byte
}

func NewParser(secret []byte) *Parser {
	return &Parser{secret: secret}
}

func (j *Parser) ParseAccess(token domain.AccessToken) (domain.UserID, []domain.Role, int64, error) {
	parsed, err := jwt.Parse(token.String(), func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})
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
	if v, ok := claims["rv"].(int64); ok {
		rv = v
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
