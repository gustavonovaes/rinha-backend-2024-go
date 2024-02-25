package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	api "github.com/gustavonovaes/rinha-backend-2024-go"
)

func TestPOSTTransaction(t *testing.T) {
	t.Run("returns 200 on credit", func(t *testing.T) {
		clientId := 1
		server, response := newServer(clientId, api.ClientBalance{1000, 0})
		server.ServeHTTP(response, newPostTransactionRequest(1, api.Transaction{
			Amount:      42,
			Type:        api.TypeCredit,
			Description: "Credit",
		}))

		assertStatusCode(t, response.Code, http.StatusOK)
		assertClientBalance(t, response.Body, api.ClientBalance{
			AccountLimit: 1000,
			Balance:      42,
		})
	})

	t.Run("returns 200 on debit", func(t *testing.T) {
		clientId := 1
		server, response := newServer(clientId, api.ClientBalance{1000, 0})
		server.ServeHTTP(response, newPostTransactionRequest(clientId, api.Transaction{
			Amount:      42,
			Type:        api.TypeDebit,
			Description: "Debit",
		}))

		assertStatusCode(t, response.Code, http.StatusOK)
		assertClientBalance(t, response.Body, api.ClientBalance{
			AccountLimit: 1000,
			Balance:      -42,
		})
	})

	t.Run("validation cases", func(t *testing.T) {
		clientId := 1
		server, _ := newServer(clientId, api.ClientBalance{1000, 0})
		cases := []struct {
			CaseName       string
			ClientId       int
			Body           string
			ExpectedStatus int
		}{
			{
				"invalid json",
				1,
				"{invalid, 1::2 }",
				http.StatusUnprocessableEntity,
			},
			{
				"amount float value",
				clientId,
				`{"valor": 1.2, "tipo": "c", "descricao": "teste"}`,
				http.StatusUnprocessableEntity,
			},
			{
				"invalid type",
				clientId,
				`{"valor": 1, "tipo": "x", "descricao": "teste"}`,
				http.StatusUnprocessableEntity,
			},
			{
				"empty description",
				clientId,
				`{"valor": 1, "tipo": "c", "descricao": ""}`,
				http.StatusUnprocessableEntity,
			},
			{
				"big description",
				clientId,
				`{"valor": 1, "tipo": "c", "descricao": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
				http.StatusUnprocessableEntity,
			},
			{
				"negative value",
				clientId,
				`{"valor": -10, "tipo": "c", "descricao": "negative"}`,
				http.StatusUnprocessableEntity,
			},
			{
				"inexistent client",
				404,
				`{"valor": 1, "tipo": "c", "descricao": "teste"}`,
				http.StatusNotFound,
			},
		}

		for _, c := range cases {
			t.Run(c.CaseName, func(t *testing.T) {
				response := httptest.NewRecorder()
				server.ServeHTTP(
					response,
					newPostTransactionRequestWithBody(c.ClientId, c.Body),
				)
				assertStatusCode(t, response.Code, c.ExpectedStatus)
			})
		}
	})
}

func TestGETStatement(t *testing.T) {
	t.Run("returns 404 when the client does not exist", func(t *testing.T) {
		server, response := newServer(1, api.ClientBalance{1000, 0})

		server.ServeHTTP(
			response,
			newGetStatementRequest(404),
		)

		assertStatusCode(t, response.Code, http.StatusNotFound)
	})

	t.Run("returns 200", func(t *testing.T) {
		clientId := 1
		server, response := newServer(clientId, api.ClientBalance{1000, 0})

		server.ServeHTTP(
			response,
			newGetStatementRequest(clientId),
		)

		assertStatusCode(t, response.Code, http.StatusOK)
		assertClientStatement(t, response.Body, api.ClientStatement{
			Balance: api.ClientStatementBalance{
				Total:        0,
				AccountLimit: 1000,
			},
			LatestTransactions: nil,
		})
	})

	t.Run("returns all transaction with the latest first", func(t *testing.T) {
		clientId := 1
		server, response := newServer(clientId, api.ClientBalance{1000, 0})

		transactions := []api.Transaction{
			{42, api.TypeCredit, "Credit", time.Now()},
			{42, api.TypeDebit, "Debit", time.Now()},
			{42, api.TypeDebit, "Debit", time.Now()},
		}
		for _, t := range transactions {
			server.ServeHTTP(httptest.NewRecorder(), newPostTransactionRequest(clientId, t))
		}

		server.ServeHTTP(
			response,
			newGetStatementRequest(clientId),
		)

		assertStatusCode(t, response.Code, http.StatusOK)
		assertClientStatement(t, response.Body, api.ClientStatement{
			Balance: api.ClientStatementBalance{
				Total:        -42,
				AccountLimit: 1000,
			},
			LatestTransactions: []api.Transaction{
				{42, api.TypeDebit, "Debit", time.Now()},
				{42, api.TypeDebit, "Debit", time.Now()},
				{42, api.TypeCredit, "Credit", time.Now()},
			},
		})
	})

	t.Run("returns only max transactions", func(t *testing.T) {
		clientId := 1
		server, response := newServer(clientId, api.ClientBalance{1000, 0})

		var total int

		for i := range 100 {
			amount := 42 * i
			total += amount
			server.ServeHTTP(
				httptest.NewRecorder(),
				newPostTransactionRequest(clientId, api.Transaction{
					amount, api.TypeCredit, "Credit", time.Now(),
				}),
			)
		}

		server.ServeHTTP(
			response,
			newGetStatementRequest(clientId),
		)

		assertStatusCode(t, response.Code, http.StatusOK)

		statement := getClientStatementFromResponse(response.Body)

		got := len(statement.LatestTransactions)

		if statement.Balance.Total != total {
			t.Errorf(
				"incorrect balance total: got: %d, want: %d",
				statement.Balance.Total,
				total,
			)
		}

		if got != api.MAX_STATEMENT_TRANSCATIONS {
			t.Errorf(
				"incorrect transactions count: got: %d, want: %d",
				got,
				api.MAX_STATEMENT_TRANSCATIONS,
			)
		}
	})
}

func newServer(
	clientId int,
	client api.ClientBalance,
) (*api.Server, *httptest.ResponseRecorder) {
	server := api.NewServer(api.NewInMemoryTractionStore(
		map[int]api.ClientBalance{
			clientId: client,
		},
	))

	return server, httptest.NewRecorder()
}

func newPostTransactionRequest(
	clientId int,
	transaction api.Transaction,
) *http.Request {
	body := &bytes.Buffer{}
	json.NewEncoder(body).Encode(transaction)

	request, _ := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/clientes/%d/transacoes", clientId),
		body,
	)
	return request
}

func newPostTransactionRequestWithBody(id int, body string) *http.Request {
	request, _ := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/clientes/%d/transacoes", id),
		bytes.NewBufferString(body),
	)
	return request
}

func newGetStatementRequest(clientId int) *http.Request {
	request, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/clientes/%d/extrato", clientId),
		nil,
	)
	return request
}

func assertStatusCode(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func assertClientBalance(t *testing.T, body io.Reader, want api.ClientBalance) {
	t.Helper()

	got := getClientBalanceFromResponse(body)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func assertClientStatement(t *testing.T, body io.Reader, want api.ClientStatement) {
	t.Helper()

	got := getClientStatementFromResponse(body)

	// fixme mocking date
	got.Balance.StatementDate = want.Balance.StatementDate
	for i, _ := range got.LatestTransactions {
		got.LatestTransactions[i].TransactionDate = want.LatestTransactions[i].TransactionDate
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func getClientBalanceFromResponse(body io.Reader) (clientBalance api.ClientBalance) {
	json.NewDecoder(body).Decode(&clientBalance)
	return
}

func getClientStatementFromResponse(body io.Reader) (clientStatement api.ClientStatement) {
	json.NewDecoder(body).Decode(&clientStatement)
	return
}
