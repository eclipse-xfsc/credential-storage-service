package model

type TransactionModel struct {
	Nonce  string `json:"nonce"`
	Expire int64  `json:"expire"`
}
