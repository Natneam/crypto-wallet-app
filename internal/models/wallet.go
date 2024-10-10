package models

// Wallet represents a user wallet.
type Wallet struct {
	ID         string `json:"id,omitempty" bson:"_id,omitempty"`
	Name       string `json:"name"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"-"`
	Balance    string `json:"balance"`
}
