package domain

type Role string

const (
	RoleUser   Role = "USER"
	RoleTrader Role = "TRADER"
	RoleAdmin  Role = "ADMIN"
)
