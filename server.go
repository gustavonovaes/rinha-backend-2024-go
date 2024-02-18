package rinhabackend202401

import "net/http"

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
	router.HandleFunc("POST /clientes/:id/transacoes", server.postTransactions)
	router.HandleFunc("GET /clientes/:id/extrato", server.postTransactions)

	return router
}

func (s *Server) postTransactions(response http.ResponseWriter, request *http.Request) {

}

func (s *Server) getBalance(response http.ResponseWriter, request *http.Request) {

}
