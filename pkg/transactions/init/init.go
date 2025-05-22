package init

import (
	"github.com/bloc-transfer-service/pkg/transactions/api"
	"github.com/bloc-transfer-service/pkg/transactions/model"

	"github.com/gorilla/mux"
)

// Init function for Transaction
func InitTransaction(router *mux.Router) {
	// Init code here
	model.Init()
	api.Router(router)
}
