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
	AccountLimit  int       `json:"limite"`
	StatementDate time.Time `json:"data_extrato"`
}
