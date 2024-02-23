package main

import "time"

const (
	TypeCredit = "c"
	TypeDebit  = "d"
)

type Transaction struct {
	Amount          int       `json:"valor"`
	Type            string    `json:"tipo"`
	Description     string    `json:"descricao"`
	TransactionDate time.Time `json:"realizada_em"`
}
