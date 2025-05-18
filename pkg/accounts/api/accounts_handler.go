package api

import (
	"bloc-mfb/pkg/accounts/model"
	"bloc-mfb/pkg/accounts/usecase"
	req "bloc-mfb/utils/http"
	"net/http"
)

// AccountsHandler handles HTTP requests
func CreateAccount(w http.ResponseWriter, r *http.Request) {
	account := model.Accounts{}
	err := req.GetRequestBody(r, &account)
	err = model.Accounts.ValidateCreateAccount(account)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}
	//proceed to create account
	account, err = usecase.CreateAccount(account)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}
	//send notification for successful account creation
	req.SendSuccessResponse(w, true, account, "Account created successfully", 201)

}

func GetAccount(w http.ResponseWriter, r *http.Request) {
	account := model.Accounts{}
	errs := req.GetRequestBody(r, &account)
	if errs != nil {
		req.SendSuccessResponse(w, false, nil, errs.Error(), http.StatusBadRequest)
		return
	}
	account, err := usecase.GetAccountByCustomerId(account.CustomerID)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}
	req.SendSuccessResponse(w, true, account, "Account created successfully", 200)
}

func TestLock(w http.ResponseWriter, r *http.Request) {
	// account, err := usecase.GetAccountById(12)
	// acc, err := usecase.DebitOrCreditAccount(account, "debit", 1000000)
	// if err != nil {
	// 	req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
	// 	return
	// }
	// req.SendSuccessResponse(w, true, acc, "Account created successfully", 200)
}
