package init

import (
	"bloc-mfb/pkg/transfers/api"

	"github.com/gorilla/mux"
)

// Init function for Transfers
func InitTransfer(router *mux.Router) {
	api.Router(router)
}
