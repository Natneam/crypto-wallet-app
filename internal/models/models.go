package models

type TransactionRequest struct {
	FromAddress string `json:"fromAddress"`
	ToAddress   string `json:"toAddress"`
	Value       string `json:"value"`
}

type TransactionResult struct {
	TransactionHash string `json:"transactionHash"`
	From            string `json:"from"`
	To              string `json:"to"`
	GasPrice        string `json:"gasPrice"`
	Value           string `json:"value"`
	GasUsed         uint64 `json:"gasUsed"`
	BlockNumber     uint64 `json:"blockNumber"`
	ID              string `json:"id" bson:"_id,omitempty"`
	UserID          string `json:"user_id"`
}

// Wallet represents a user wallet.
type Wallet struct {
	ID        string `json:"id,omitempty" bson:"_id,omitempty"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
	Balance   string `json:"balance"`
	KMSKeyID  string `json:"kms_key_id"`
	UserID    string `json:"user_id"`
}

// User represents a user account.
type User struct {
	ID           string `json:"id,omitempty" bson:"_id,omitempty"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
