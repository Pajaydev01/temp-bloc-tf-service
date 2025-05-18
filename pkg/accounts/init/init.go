package accountInit

import (
	"bloc-mfb/pkg/accounts/api"
	"bloc-mfb/pkg/accounts/model"

	"github.com/gorilla/mux"
)

// Init function for Accounts
func InitAccount(router *mux.Router) {
	// Init code here
	model.Init()
	api.Router(router)
}
