package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func MakeRequest(url string, method string, body interface{}, headers interface{}) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	//headers
	if headers != nil {
		if headerMap, ok := headers.(map[string]string); ok {
			for key, value := range headerMap {
				req.Header.Set(key, value)
			}
		}
	}
	//body
	if body != nil {
		if method == "GET" {
			// Convert the body to a query string and append it to the URL
			queryParams, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			req.URL.RawQuery = string(queryParams)
			log.Println("GET request with query params:", string(queryParams))

		} else {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			log.Println("Request body:", string(jsonBody))
			req.Body = io.NopCloser(bytes.NewBuffer(jsonBody))
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
