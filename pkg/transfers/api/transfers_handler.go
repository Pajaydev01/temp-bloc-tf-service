package api

import (
	"encoding/json"
	"net/http"

	transferModel "github.com/bloc-transfer-service/pkg/transfers/model"
	"github.com/bloc-transfer-service/pkg/transfers/usecase"
	req "github.com/bloc-transfer-service/utils/http"
)

// TransfersHandler handles HTTP requests
func Transfer(w http.ResponseWriter, r *http.Request) {
	transfer := transferModel.TransferRequest{}
	err := req.GetRequestBody(r, &transfer)
	err = transferModel.TransferRequest.ValidateTransfer(transfer)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}

	transaction, err := usecase.DoLocalTransfer(transfer)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}

	data := make(map[string]interface{})
	data["Transaction"] = transaction
	req.SendSuccessResponse(w, true, data, "Transfer initiated", 201)
}

func NameEnquiry(w http.ResponseWriter, r *http.Request) {
	nameEnquiry := transferModel.NameEnquiryRequest{}
	err := req.GetRequestBody(r, &nameEnquiry)
	err = transferModel.NameEnquiryRequest.ValidateNameEnquiry(nameEnquiry)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}

	account, err := usecase.DoNameEnquiry(nameEnquiry)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}

	var data map[string]interface{}
	accountBytes, _ := json.Marshal(account)
	json.Unmarshal(accountBytes, &data)

	req.SendSuccessResponse(w, true, data, "Name Enquiry successful", 201)
}

func GetInstitutions(w http.ResponseWriter, r *http.Request) {
	institutions, err := usecase.GetInstitution()
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}
	req.SendSuccessResponse(w, true, institutions, "Institutions fetched successfully", 200)
}
