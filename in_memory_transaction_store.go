package rinhabackend202401

type InMemoryTractionStore struct {
	transactions []Transaction
}

func (i *InMemoryTractionStore) CreateTransaction(
	id string,
	transaction Transaction,
) (TransactionStatus, error) {
	return TransactionStatus{}, nil
}

func (i *InMemoryTractionStore) GetBalance(id string) (TransactionsBalance, error) {
	return TransactionsBalance{}, nil
}
