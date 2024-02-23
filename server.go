package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	contentTypeJSON            = "application/json"
	MAX_STATEMENT_TRANSCATIONS = 10
)

var ErrDebitBelowLimit = errors.New("insufficient limit for this debit")
var ErrInvalidTransaction = errors.New("invalid transaction payload")

type Server struct {
	transactionStore TransactionStore
	http.Handler
}

func NewServer(store TransactionStore) *Server {
	var server = new(Server)

	server.transactionStore = store
	server.Handler = setupRoutes(server)

	return server
}

func setupRoutes(server *Server) http.Handler {
	router := http.NewServeMux()
	router.Handle("POST /clientes/{id}/transacoes", http.HandlerFunc(server.postTransactions))
	router.Handle("GET /clientes/{id}/extrato", http.HandlerFunc(server.getStatement))

	return router
}

func (s *Server) postTransactions(w http.ResponseWriter, r *http.Request) {
	clientId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorHandler(w, "invalid client id", ErrInvalidTransaction)
	}

	transaction, err := getTransactionFromBody(r.Body)
	if err != nil {
		errorHandler(w, "getTransactionFromBody", ErrInvalidTransaction)
		return
	}
	transaction.TransactionDate = time.Now()

	clientBalance, err := s.transactionStore.AddTransactionSync(
		clientId,
		transaction,
		s.processTransaction,
	)
	if err != nil {
		errorHandler(w, "transactionStore.AddTransactionSync", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", contentTypeJSON)
	json.NewEncoder(w).Encode(&clientBalance)
}

func (s *Server) getStatement(w http.ResponseWriter, r *http.Request) {
	clientId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errorHandler(w, "invalid client id", ErrInvalidTransaction)
	}

	balance, err := s.transactionStore.GetBalance(clientId)
	if err != nil {
		errorHandler(w, "transactionStore.GetBalance", ErrClientNotFound)
		return
	}

	transactions, err := s.transactionStore.GetTransactions(clientId, MAX_STATEMENT_TRANSCATIONS)
	if err != nil {
		errorHandler(w, "transactionStore.GetTransactions", err)
		return
	}

	statement := getStatement(balance, transactions)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", contentTypeJSON)
	json.NewEncoder(w).Encode(&statement)
}

func (s *Server) processTransaction(clientBalance *ClientBalance, transaction Transaction) error {
	if !isValidTransaction(transaction) {
		return ErrInvalidTransaction
	}

	switch transaction.Type {
	case TypeCredit:
		clientBalance.Balance += transaction.Amount
	case TypeDebit:
		newBalance := clientBalance.Balance - transaction.Amount
		if newBalance < -clientBalance.AccountLimit {
			return ErrDebitBelowLimit
		}

		clientBalance.Balance = newBalance
	}

	return nil
}

func isValidTransaction(t Transaction) bool {
	if t.Amount <= 0 {
		return false
	}

	if t.Description == "" {
		return false
	}

	if len(t.Description) > 10 {
		return false
	}

	if t.Type != TypeCredit && t.Type != TypeDebit {
		return false
	}

	return true
}

func errorHandler(w http.ResponseWriter, errContext string, err error) {
	switch err {
	case ErrInvalidTransaction, ErrDebitBelowLimit:
		w.WriteHeader(http.StatusUnprocessableEntity)

	case ErrClientNotFound:
		w.WriteHeader(http.StatusNotFound)

	default:
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ERROR %s: %v\n", errContext, err)
	}
}

func getStatement(balance ClientBalance, transactions []Transaction) *ClientStatement {
	return &ClientStatement{
		Balance: ClientStatementBalance{
			Total:         balance.Balance,
			Limit:         balance.AccountLimit,
			StatementDate: time.Now(),
		},
		LatestTransactions: transactions,
	}
}

func getTransactionFromBody(body io.Reader) (Transaction, error) {
	var transaction Transaction
	err := json.NewDecoder(body).Decode(&transaction)
	if err != nil {
		return transaction, err
	}
	return transaction, nil
}
