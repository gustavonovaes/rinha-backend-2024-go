package rinhabackend202401_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	rinhabackend202401 "github.com/gustavonovaes/rinha-backend-202401"
)

func TestPOSTTransaction(t *testing.T) {
	server := NewServer()

	t.Run("creates transaction successfully ", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := newPostTransactionRequest("1", rinhabackend202401.Transaction{
			Value:       1000,
			Type:        "c",
			Description: "",
		})
	})
}

func TestGETBalance(t *testing.T) {

}

func newPostTransactionRequest(id string, transaction Transaction) *http.Request {
	body := bufio.NewWriter()
	json.NewEncoder(body).Encode(transaction)
	request, _ := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/clientes/%s/transacoes"),
		transaction,
	)
	return request
}
