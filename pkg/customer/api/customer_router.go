package api

import (
	//"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	//"github.com/urfave/negroni"
	"github.com/gorilla/mux"
	//"github.com/newrelic/go-agent/v3/newrelic"
)

// route customer api here
func Router(router *mux.Router) {
	baseRoute := router.PathPrefix("/v1/customers").Subrouter()
	//baseRoute.Use(nrgorilla.Middleware(relic))

	//test
	baseRoute.HandleFunc("/create", CreateCustomer).Methods("POST")
	baseRoute.HandleFunc("/upgrade/t1/{customerID}", UpdateCustomerToT1).Methods("PUT")
	baseRoute.HandleFunc("/{customerID}", GetCustomerById).Methods("GET")
	//
	baseRoute.HandleFunc("", GetAllCustomers).Methods("GET")
}
