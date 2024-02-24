package main

type TransactionStore interface {
	Clear() error
	AddClient(clientId int, balance, limit int) error
	GetBalance(clientId int) (ClientBalance, error)
	UpdateBalance(clientId int, clientBalance ClientBalance) error
	AddTransaction(clientId int, transaction Transaction) error
	AddTransactionSync(
		clientId int,
		transaction Transaction,
		processTransaction func(c ClientBalance, t Transaction) (ClientBalance, error),
	) (ClientBalance, error)
	GetTransactions(clientId, count int) ([]Transaction, error)
}
