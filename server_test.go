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
		server, _ := newServer(1, api.ClientBalance{1000, 0})
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
				1,
				`{"valor": 1.2, "tipo": "c", "descricao": "teste"}`,
				http.StatusUnprocessableEntity,
			},
			{
				"invalid type",
				1,
				`{"valor": 1, "tipo": "x", "descricao": "teste"}`,
				http.StatusUnprocessableEntity,
			},
			{
				"empty description",
				1,
				`{"valor": 1, "tipo": "c", "descricao": ""}`,
				http.StatusUnprocessableEntity,
			},
			{
				"big description",
				1,
				`{"valor": 1, "tipo": "c", "descricao": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
				http.StatusUnprocessableEntity,
			},
			{
				"negative value",
				1,
				`{"valor": -10, "tipo": "c", "descricao": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
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
		server, response := newServer(1, api.ClientBalance{1000, 0})

		server.ServeHTTP(
			response,
			newGetStatementRequest(1),
		)

		assertStatusCode(t, response.Code, http.StatusOK)
		assertClientStatement(t, response.Body, api.ClientStatement{
			Balance: api.ClientStatementBalance{
				Total: 0,
				Limit: 1000,
			},
		})
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

func newGetStatementRequest(id int) *http.Request {
	request, _ := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/clientes/%d/extrato", id),
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

	got.Balance.StatementDate = want.Balance.StatementDate // fixme

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
