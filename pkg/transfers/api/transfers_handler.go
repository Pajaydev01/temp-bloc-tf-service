package api

import (
	transferModel "bloc-mfb/pkg/transfers/model"
	"bloc-mfb/pkg/transfers/usecase"
	req "bloc-mfb/utils/http"
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

	account, transaction, err := usecase.DoLocalTransfer(transfer)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}

	data := make(map[string]interface{})
	data["Account"] = account
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

	data := make(map[string]interface{})
	data["Account"] = account
	req.SendSuccessResponse(w, false, data, "Name Enquiry successful", 201)
}
