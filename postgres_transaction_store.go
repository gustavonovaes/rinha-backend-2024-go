package rinhabackend202401

import "database/sql"

type PostgresTractionStore struct {
	db sql.DB
}

func (p *PostgresTractionStore) CreateTransaction(
	id string,
	transaction Transaction,
) (TransactionStatus, error) {
	return TransactionStatus{}, nil
}

func (p *PostgresTractionStore) GetBalance(id string) (TransactionsBalance, error) {
	return TransactionsBalance{}, nil
}
