package models

type TransactionRequest struct {
	FromAddress string `json:"fromAddress"`
	ToAddress   string `json:"toAddress"`
	PrivateKey  string `json:"privateKey"`
}

type TransactionResult struct {
	TransactionHash string `json:"transactionHash"`
	From            string `json:"from"`
	To              string `json:"to"`
	GasPrice        string `json:"gasPrice"`
	Value           string `json:"value"`
}
