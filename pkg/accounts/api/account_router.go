package api

import "github.com/gorilla/mux"

func Router(router *mux.Router) {
	baseRoute := router.PathPrefix("/v1/accounts").Subrouter()

	baseRoute.HandleFunc("", CreateAccount).Methods("POST")
	baseRoute.HandleFunc("/get", GetAccount).Methods("POST")
	baseRoute.HandleFunc("", TestLock).Methods("GET")
}
