package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/suprt/trading/order-service/internal/core/domain"
	er "github.com/suprt/trading/order-service/internal/core/errors"
)

type Parser struct {
	secret []byte
}

func NewParser(secret []byte) *Parser {
	return &Parser{secret: secret}
}

func (j *Parser) ParseAccess(token domain.AccessToken) (domain.UserID, error) {
	parsed, err := jwt.Parse(token.String(), func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil {
		return "", er.ErrInvalidToken
	}
	if !parsed.Valid {
		return "", er.ErrInvalidToken
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return "", er.ErrInvalidToken
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", er.ErrInvalidToken
	}

	return domain.UserID(sub), nil
}
