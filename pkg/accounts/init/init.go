package accountInit

import (
	"github.com/bloc-transfer-service/pkg/accounts/api"
	"github.com/bloc-transfer-service/pkg/accounts/model"

	"github.com/gorilla/mux"
)

// Init function for Accounts
func InitAccount(router *mux.Router) {
	// Init code here
	model.Init()
	api.Router(router)
}
