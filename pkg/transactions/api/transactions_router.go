package api

import "github.com/gorilla/mux"

func Router(router *mux.Router) {
	baseRoute := router.PathPrefix("/v1/transactions").Subrouter()
	baseRoute.HandleFunc("/{customerID}", GetTransactions).Methods("GET")
}
