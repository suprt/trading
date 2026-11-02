package domain

import "slices"

import "time"

type MarketID string

type Market struct {
	ID            MarketID
	Name          string
	RequiredRoles []Role
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

func (m *Market) IsActive() bool {
	return m.Enabled && m.DeletedAt == nil
}

func (m *Market) AvailableFor(userRoles []Role) bool {
	if !m.IsActive() {
		return false
	}

	if len(m.RequiredRoles) == 0 {
		return true
	}
	for _, required := range m.RequiredRoles {
		if slices.Contains(userRoles, required) {
			return true
		}
	}
	return false
}
