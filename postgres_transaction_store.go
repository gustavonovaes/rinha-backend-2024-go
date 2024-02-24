package main

import (
	"database/sql"
)

type PostgresTransactionStore struct {
	db *sql.DB
}

func (s *PostgresTransactionStore) Clear() error {
	query := `
		DELETE FROM transactions;
		DELETE FROM clients;
	`
	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresTransactionStore) AddClient(clientId int, balance, limit int) error {
	query := `
		insert into clients
			(id, balance, credit_limit)
		values 
			($1, $2, $3)
	`
	_, err := s.db.Exec(query, clientId, balance, limit)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresTransactionStore) GetBalance(clientId int) (ClientBalance, error) {
	query := `
		select 
			balance,
			credit_limit
		from clients
		where id = $1
	`

	clientBalance := ClientBalance{}
	err := s.db.QueryRow(query, clientId).Scan(&clientBalance.Balance, &clientBalance.AccountLimit)
	if err != nil {
		return clientBalance, err
	}

	return clientBalance, nil
}

func (s *PostgresTransactionStore) UpdateBalance(
	clientId int,
	clientBalance ClientBalance,
) error {
	query := `
		update clients 
		set balance = $2
		where id = $1
	`
	_, err := s.db.Exec(query, clientId, clientBalance.Balance)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresTransactionStore) AddTransaction(clientId int, transaction Transaction) error {
	query := `
		insert into transactions
			(client_id, amount, transaction_type, description, created_at)
		values 
			($1, $2, $3, $4, $5)
	`
	_, err := s.db.Exec(
		query,
		clientId,
		transaction.Amount,
		transaction.Type,
		transaction.Description,
		transaction.TransactionDate,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresTransactionStore) AddTransactionSync(
	clientId int,
	transaction Transaction,
	processTransaction func(clientBalance ClientBalance, transaction Transaction) (ClientBalance, error),
) (ClientBalance, error) {
	var query string

	tx, err := s.db.Begin()
	if err != nil {
		return ClientBalance{}, err
	}
	defer tx.Rollback()

	query = `
		select 
			balance,
			credit_limit
		from clients
		where id = $1
		for update 
		limit 1
	`

	clientBalance := ClientBalance{}
	err = tx.QueryRow(query, clientId).Scan(&clientBalance.Balance, &clientBalance.AccountLimit)
	if err != nil {
		tx.Rollback()
		return clientBalance, err
	}

	clientBalanceUpdated, err := processTransaction(clientBalance, transaction)
	if err != nil {
		tx.Rollback()
		return clientBalance, err
	}

	query = `
		insert into transactions
			(client_id, amount, transaction_type, description, created_at)
		values 
			($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(
		query,
		clientId,
		transaction.Amount,
		transaction.Type,
		transaction.Description,
		transaction.TransactionDate,
	)
	if err != nil {
		tx.Rollback()
		return clientBalanceUpdated, err
	}

	query = `
		update clients 
		set balance = $2
		where id = $1
	`
	_, err = tx.Exec(query, clientId, clientBalanceUpdated.Balance)
	if err != nil {
		tx.Rollback()
		return clientBalanceUpdated, err
	}

	err = tx.Commit()
	if err != nil {
		return clientBalanceUpdated, err
	}

	return clientBalanceUpdated, nil
}

func (s *PostgresTransactionStore) GetTransactions(clientId int, count int) ([]Transaction, error) {
	query := `
		select amount, description, transaction_type, created_at
		from transactions
		where client_id = $1
		order by created_at desc
		limit $2
	`

	rows, err := s.db.Query(query, clientId, count)
	if err != nil {
		return nil, err
	}

	transactions := []Transaction{}
	for rows.Next() {
		transaction := Transaction{}
		rows.Scan(
			&transaction.Amount,
			&transaction.Description,
			&transaction.Type,
			&transaction.TransactionDate,
		)
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func NewPostgresTransactionStore(db *sql.DB) *PostgresTransactionStore {
	return &PostgresTransactionStore{
		db,
	}
}
