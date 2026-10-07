package domain

import "time"

type RefreshSession struct {
	Token     RefreshToken
	UserID    UserID
	ExpiresAt time.Time
}

func (s *RefreshSession) Expired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}
