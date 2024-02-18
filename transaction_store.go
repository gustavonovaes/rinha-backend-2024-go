package rinhabackend202401

import "time"

type TransactionStatus struct {
	Limit, Balance int
}

type TransactionsBalance struct {
	Total              int
	Limit              int
	BalanceDate        time.Time
	LatestTransactions []Transaction
}

type TransactionStore interface {
	CreateTransaction(clientId string, transaction Transaction) (TransactionStatus, error)
	GetBalance(clientId string) (TransactionsBalance, error)
}
