package main

import (
	"sort"
	"sync"
)

type InMemoryTractionStore struct {
	mu             sync.Mutex
	transactions   map[int][]Transaction
	clientBalances map[int]ClientBalance
}

func (i *InMemoryTractionStore) Clear() error {
	clear(i.transactions)
	clear(i.clientBalances)
	return nil
}

func (i *InMemoryTractionStore) AddClient(clientId int, balance, limit int) error {
	i.clientBalances[clientId] = ClientBalance{limit, balance}
	return nil
}

func (i *InMemoryTractionStore) GetBalance(clientId int) (ClientBalance, error) {
	clientBalance, ok := i.clientBalances[clientId]
	if !ok {
		return clientBalance, ErrClientNotFound
	}

	return clientBalance, nil
}

func (i *InMemoryTractionStore) UpdateBalance(clientId int, clientBalance ClientBalance) error {
	i.clientBalances[clientId] = clientBalance
	return nil
}

func (i *InMemoryTractionStore) AddTransaction(
	clientId int,
	transaction Transaction,
) error {
	i.transactions[clientId] = append(i.transactions[clientId], transaction)
	return nil
}

func (i *InMemoryTractionStore) GetTransactions(clientId, count int) ([]Transaction, error) {
	sort.Slice(i.transactions[clientId], func(a, b int) bool {
		return i.transactions[clientId][a].TransactionDate.Compare(
			i.transactions[clientId][b].TransactionDate,
		) == 1
	})

	if len(i.transactions[clientId]) > count {
		return i.transactions[clientId][:count], nil
	}

	return i.transactions[clientId], nil
}

func (i *InMemoryTractionStore) AddTransactionSync(
	clientId int,
	transaction Transaction,
	processTransaction func(c *ClientBalance, t Transaction) error,
) (ClientBalance, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	clientBalance, err := i.GetBalance(clientId)
	if err != nil {
		return clientBalance, err
	}

	err = processTransaction(&clientBalance, transaction)
	if err != nil {
		return clientBalance, err
	}

	err = i.UpdateBalance(clientId, clientBalance)
	if err != nil {
		return clientBalance, err
	}

	return clientBalance, nil
}

func NewInMemoryTractionStore(clientBalances map[int]ClientBalance) *InMemoryTractionStore {
	return &InMemoryTractionStore{
		transactions:   map[int][]Transaction{},
		clientBalances: clientBalances,
	}
}
