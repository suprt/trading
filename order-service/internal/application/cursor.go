package application

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/suprt/trading/order-service/internal/core/domain"
)

func encode(createdAt time.Time, id domain.OrderID) string {
	raw := createdAt.UTC().Format(time.RFC3339Nano) + "|" + string(id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decode(token string) (time.Time, domain.OrderID, error) {
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid token: %w", err)
	}
	parts := strings.SplitN(string(data), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("invalid token format")
	}
	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid created_at: %w", err)
	}
	return createdAt, domain.OrderID(parts[1]), nil
}
