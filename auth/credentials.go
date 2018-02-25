package auth

// Struct that holds the persistent parameters relative to an user
type PersistentParameters struct {
	UserId string `json:"userId"`
}

// Struct that contains all the parameters that changes
type RuntimeParameters struct {
	AtToken string
}

// type Credentials interface {
// 	// Returns an authenticated client
// 	GetClient() *http.Client

// 	// Returns the user parameters that do not change
// 	GetPersistentParameters() (*PersistentParameters, error)
// 	SetPersistentParameters(*PersistentParameters)

// 	// Returns authentication tokens and ids scraped at runtime
// 	GetRuntimeParameters() *RuntimeParameters
// }
