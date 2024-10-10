package main

// Wallet represents a cryptocurrency wallet
type Wallet struct {
	ID         string `json:"id,omitempty" bson:"_id,omitempty"`
	Name       string `json:"name" bson:"name"`
	PrivateKey string `json:"privateKey" bson:"privateKey"`
	PublicKey  string `json:"publicKey" bson:"publicKey"`
	Balance    string `json:"balance" bson:"balance"`
}

type SignTransactionRequest struct {
	FromAddress string `json:"fromAddress"`
	ToAddress   string `json:"toAddress"`
	PrivateKey  string `json:"privateKey"`
}
