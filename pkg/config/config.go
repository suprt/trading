//минимальный конфиг, потом расширю

package config

import "time"

type GRPC struct {
	Addr string
}

type JWT struct {
	Secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type Config struct {
	GRPC GRPC
	JWT  JWT
}

func Default() Config {
	return Config{
		GRPC: GRPC{Addr: ":50051"},
		JWT: JWT{
			Secret:     []byte("dev-secret"),
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 7 * 24 * time.Hour,
		},
	}
}
