package main_test

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"
	"time"

	api "github.com/gustavonovaes/rinha-backend-2024-go"
	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	dbInstance, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Fail to open connection with database: %v", err)
	}

	db = dbInstance
}

func TestServerIntegration(t *testing.T) {
	defer db.Close()

	clientId := 1
	clients := map[int]api.ClientBalance{
		clientId: {AccountLimit: 5000, Balance: 0},
	}

	store := initPostgresStore(t, clients)
	server := api.NewServer(store)

	transactions := []api.Transaction{
		{1000, api.TypeCredit, "Teste", time.Now()},
		{1000, api.TypeCredit, "Teste", time.Now()},
		{1000, api.TypeCredit, "Teste", time.Now()},
		{1500, api.TypeDebit, "Teste", time.Now()},
	}

	for _, t := range transactions {
		server.ServeHTTP(
			httptest.NewRecorder(),
			newPostTransactionRequest(clientId, t),
		)

		time.Sleep(25 * time.Millisecond)
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetStatementRequest(clientId))

	var clientStatement api.ClientStatement
	json.NewDecoder(response.Body).Decode(&clientStatement)

	if clientStatement.Balance.AccountLimit != 5000 {
		t.Errorf("incorrect limit. Got %d, want %d", clientStatement.Balance.AccountLimit, 0)
	}

	if clientStatement.Balance.Total != 1500 {
		t.Errorf("incorrect balance. Got %d, want %d", clientStatement.Balance.Total, 1500)
	}

	if len(clientStatement.LatestTransactions) != 4 {
		t.Fatalf(
			"incorrect latests transactions count. Got %d, want %d",
			len(clientStatement.LatestTransactions),
			4,
		)
	}

	slices.Reverse(transactions)
	transactionsTypeOrder := []string{}
	for _, t := range transactions {
		transactionsTypeOrder = append(transactionsTypeOrder, t.Type)
	}

	for i, transactionType := range transactionsTypeOrder {
		if clientStatement.LatestTransactions[i].Type != transactionType {
			t.Errorf(
				"incorrect latests transction order. Index %d, got %s, want %s",
				i,
				clientStatement.LatestTransactions[i].Type,
				transactionType,
			)
			break
		}
	}

	response = httptest.NewRecorder()
	server.ServeHTTP(
		response,
		newPostTransactionRequest(clientId, api.Transaction{
			Amount:      clientStatement.Balance.AccountLimit + clientStatement.Balance.Total + 1,
			Description: "too much",
			Type:        api.TypeDebit,
		}),
	)

	if response.Code != http.StatusUnprocessableEntity {
		t.Errorf("should return 422 when trying to debit below limit. got %d", response.Code)
	}

	response = httptest.NewRecorder()
	server.ServeHTTP(
		response,
		newGetStatementRequest(404),
	)

	if response.Code != http.StatusNotFound {
		t.Errorf("should return 404 on inexistent client. got %d", response.Code)
	}
}

func initPostgresStore(
	t *testing.T,
	clients map[int]api.ClientBalance,
) api.TransactionStore {
	t.Helper()

	store := api.NewPostgresTransactionStore(db)

	store.Clear()
	for clientId, balance := range clients {
		err := store.AddClient(clientId, balance.Balance, balance.AccountLimit)
		if err != nil {
			t.Fatalf("fail to add client: %v", err)
		}
	}

	return store
}
