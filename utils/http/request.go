package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func GetRequestBody(r *http.Request, item interface{}) error {
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		return err
	}
	return nil
}

func GetRequestParams(r *http.Request, param string) (string, error) {
	vars := mux.Vars(r) // Get path variables
	item := vars[param]
	if item == "" {
		return "", fmt.Errorf(fmt.Sprintf("%s not found", param))
	}
	return item, nil
}

func GetRequestQueryItem(r *http.Request, item string) string {
	query := r.URL.Query()
	res := query.Get(item)
	return res
}

func GetPossibleTransactionFilters(r *http.Request) map[string]string {
	query := r.URL.Query()
	filters := map[string]string{
		"start_date":   query.Get("start_date"),
		"end_date":     query.Get("end_date"),
		"status":       query.Get("status"),
		"payment_type": query.Get("payment_type"),
		"reversal":     query.Get("reversal"),
		"drcr":         query.Get("drcr"),
		"page":         query.Get("page"),
	}

	for key, value := range filters {
		if value == "" {
			delete(filters, key)
		}
	}
	return filters
}

func GetRequestHeaderItem(r *http.Request, item string) string {
	headers := r.Header
	res := headers.Get(item)
	return res
}
