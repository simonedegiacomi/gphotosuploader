package auth

import "net/http"

type Credentials interface {
	GetClient() *http.Client
	GetAPIToken() string
}