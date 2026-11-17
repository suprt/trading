package domain

import (
	"time"
)

type OrderID string
type UserID string
type MarketID string

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type OrderType string

const (
	OrderTypeLimit  OrderType = "LIMIT"
	OrderTypeMarket OrderType = "MARKET"
)

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "CREATED"
	OrderStatusFilled    OrderStatus = "FILLED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusRejected  OrderStatus = "REJECTED"
)

type Order struct {
	ID        OrderID
	UserID    UserID
	MarketID  MarketID
	Side      OrderSide
	Type      OrderType
	Price     int64
	Quantity  int64
	Status    OrderStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
