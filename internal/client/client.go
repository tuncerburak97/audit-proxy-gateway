package client

import (
	"audit-proxy-gateway/internal/config"
	"net/http"
	"time"
)

var (
	client *http.Client
)

func InitClient() {
	cfg := config.GetConfig()
	client = &http.Client{
		Timeout: time.Duration(cfg.Application.Http.Timeout) * time.Second,
	}
}

func GetClient() *http.Client {
	return client
}
