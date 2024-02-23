package main

import (
	"errors"
	"time"
)

var ErrClientNotFound = errors.New("client not found")

type ClientBalance struct {
	AccountLimit int `json:"limite"`
	Balance      int `json:"saldo"`
}

type ClientStatement struct {
	Balance            ClientStatementBalance `json:"saldo"`
	LatestTransactions []Transaction          `json:"ultimas_transacoes"`
}

type ClientStatementBalance struct {
	Total         int       `json:"total"`
	Limit         int       `json:"limite"`
	StatementDate time.Time `json:"data_extrato"`
}

type TransactionStore interface {
	Clear() error
	AddClient(clientId int, balance, limit int) error
	GetBalance(clientId int) (ClientBalance, error)
	UpdateBalance(clientId int, clientBalance ClientBalance) error
	AddTransaction(clientId int, transaction Transaction) error
	AddTransactionSync(
		clientId int,
		transaction Transaction,
		processTransaction func(c *ClientBalance, t Transaction) error,
	) (ClientBalance, error)
	GetTransactions(clientId, count int) ([]Transaction, error)
}
