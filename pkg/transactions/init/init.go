package init

import (
	"bloc-mfb/pkg/transactions/api"
	"bloc-mfb/pkg/transactions/model"

	"github.com/gorilla/mux"
)

// Init function for Transaction
func InitTransaction(router *mux.Router) {
	// Init code here
	model.Init()
	api.Router(router)
}
