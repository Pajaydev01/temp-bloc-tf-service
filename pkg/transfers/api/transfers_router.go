package api

import "github.com/gorilla/mux"

func Router(router *mux.Router) {
	baseRoute := router.PathPrefix("/v1/transfer").Subrouter()

	baseRoute.HandleFunc("", Transfer).Methods("POST")
	baseRoute.HandleFunc("/name-enquiry", NameEnquiry).Methods("POST")
}
