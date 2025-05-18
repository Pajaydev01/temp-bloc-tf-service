package http

import (
	"encoding/json"
	"net/http"
)

// GeneralResponse returns a standard response format
type generalResponse struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	Message  string      `json:"message,omitempty"`
	Token    string      `json:"token,omitempty"`
	Error    interface{} `json:"error,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
}

func SendSuccessResponse(w http.ResponseWriter, success bool, data interface{}, message string, status int) {
	responseSuccess := generalResponse{
		Success: success,
		Data:    data,
		Message: message,
	}
	// Convert to JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // Set status first
	json.NewEncoder(w).Encode(responseSuccess)
}

func SendSuccessResponseWithMetadata(w http.ResponseWriter, success bool, data interface{}, message string, status int, metadata interface{}) {
	responseSuccess := generalResponse{
		Success:  success,
		Data:     data,
		Message:  message,
		Metadata: metadata,
	}
	// Convert to JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // Set status first
	json.NewEncoder(w).Encode(responseSuccess)
}

func SendErrorResponse(w http.ResponseWriter, error string, status int) {
	responseError := generalResponse{
		Success: false,
		Message: error,
	}
	// Convert to JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status) // Set status first
	json.NewEncoder(w).Encode(responseError)
}
