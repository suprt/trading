package domain

type Role string

const (
	RoleUser   Role = "USER"
	RoleTrader Role = "TRADER"
	RoleAdmin  Role = "ADMIN"
)

func RoleIncludes(r Role, other Role) bool {
	if r == other {
		return true
	}
	switch r {
	case RoleAdmin:
		return other == RoleTrader || other == RoleUser
	case RoleTrader:
		return other == RoleUser
	}
	return false
}
