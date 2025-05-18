package initCustomer

import (
	"bloc-mfb/pkg/customer/model"

	"bloc-mfb/pkg/customer/api"

	"github.com/gorilla/mux"
)

// Init function for Customer
// Init function for Accounts
func InitCustomer(router *mux.Router) {
	// Init code here
	model.Init()
	api.Router(router)
}
