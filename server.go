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
	contentTypeJSON                    = "application/json"
	MAX_STATEMENT_TRANSCATIONS         = 10
	MAX_TRANSACTION_DESCRIPTION_LENGTH = 10
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

	clientBalance, err := s.addTransaction(clientId, transaction)
	if err != nil {
		errorHandler(w, "addTransaction", err)
		return
	}

	writeResponse(w, http.StatusOK, &clientBalance)
}

func (s Server) addTransaction(clientId int, transaction Transaction) (ClientBalance, error) {
	return s.transactionStore.AddTransactionSync(
		clientId,
		transaction,
		processTransaction,
	)
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

	statement := buildStatement(balance, transactions)

	writeResponse(w, http.StatusOK, &statement)
}

func processTransaction(
	clientBalance ClientBalance,
	transaction Transaction,
) (ClientBalance, error) {
	if !isValidTransaction(transaction) {
		return clientBalance, ErrInvalidTransaction
	}

	switch transaction.Type {
	case TypeCredit:
		clientBalance.Balance += transaction.Amount
	case TypeDebit:
		newBalance := clientBalance.Balance - transaction.Amount
		if newBalance < -clientBalance.AccountLimit {
			return clientBalance, ErrDebitBelowLimit
		}

		clientBalance.Balance = newBalance
	}

	return clientBalance, nil
}

func isValidTransaction(t Transaction) bool {
	if t.Amount <= 0 {
		return false
	}

	if t.Description == "" {
		return false
	}

	if len(t.Description) > MAX_TRANSACTION_DESCRIPTION_LENGTH {
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

func buildStatement(balance ClientBalance, transactions []Transaction) ClientStatement {
	return ClientStatement{
		Balance: ClientStatementBalance{
			Total:         balance.Balance,
			AccountLimit:  balance.AccountLimit,
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

func writeResponse[T any](w http.ResponseWriter, statusCode int, data *T) {
	w.WriteHeader(statusCode)
	w.Header().Set("content-type", contentTypeJSON)

	json.NewEncoder(w).Encode(&data)
}
