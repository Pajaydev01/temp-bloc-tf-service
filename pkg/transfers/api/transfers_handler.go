package api

import (
	transferModel "bloc-mfb/pkg/transfers/model"
	"bloc-mfb/pkg/transfers/usecase"
	req "bloc-mfb/utils/http"
	"encoding/json"
	"net/http"
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
	req.SendSuccessResponse(w, false, data, "Transfer successful", 201)
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

	req.SendSuccessResponse(w, false, data, "Name Enquiry successful", 201)
}
