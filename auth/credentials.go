package auth

import "net/http"

type PersistentParameters struct {
	UserId       string `json:"userId"`
}

type RuntimeParameters struct {
	AtToken string
}

type Credentials interface {
	// Returns an authenticated client
	GetClient() *http.Client

	// Returns the user parameters that do not change
	GetPersistentParameters() *PersistentParameters

	// Returns authentication tokens and ids scraped at runtime
	GetRuntimeParameters() *RuntimeParameters
}