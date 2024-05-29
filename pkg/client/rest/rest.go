package rest

import (
	"github.com/go-resty/resty/v2"
)

// Client .
type Client struct {
	client *resty.Client
}

// New .
func New() *Client {
	return &Client{}
}
