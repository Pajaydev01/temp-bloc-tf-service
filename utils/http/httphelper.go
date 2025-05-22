package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func MakeRequest(url string, method string, body any, headers interface{}) (*http.Response, error) {
	client := &http.Client{}
	log.Println("Making request to URL:", url)
	log.Println("Request method:", method)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	//req.Header.Set("Content-Type", "application/json")
	//headers
	if headers != nil {
		if headerMap, ok := headers.(map[string]string); ok {
			for key, value := range headerMap {
				req.Header.Set(key, value)
			}
		}
	}
	log.Println("Request headers:", req.Header)
	//body
	if body != nil {
		switch b := body.(type) {
		case string:
			req.Body = io.NopCloser(bytes.NewBufferString(b))
		case *bytes.Buffer:
			req.Body = io.NopCloser(b)
		case io.Reader:
			req.Body = io.NopCloser(b)
		// case *strings.Reader:
		// 	req.Body = io.NopCloser(b)
		default:
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			req.Body = io.NopCloser(bytes.NewBuffer(jsonBody))
		}
		log.Println("Request body:", req.Body)
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	log.Println("Response status:", resp)
	return resp, nil
}
