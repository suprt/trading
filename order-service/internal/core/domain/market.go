package domain

import "time"

type Market struct {
	ID            MarketID
	Name          string
	RequiredRoles []Role
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

func (m *Market) AvailableFor(userRoles []Role) bool {

	if len(m.RequiredRoles) == 0 {
		return true
	}
	for _, required := range m.RequiredRoles {
		for _, userRole := range userRoles {
			if RoleIncludes(userRole, required) {
				return true
			}
		}
	}
	return false
}
