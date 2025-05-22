package api

import (
	"log"
	"net/http"

	"github.com/bloc-transfer-service/pkg/transactions/usecase"
	req "github.com/bloc-transfer-service/utils/http"
)

// TransactionsHandler handles HTTP requests
func GetTransactions(w http.ResponseWriter, r *http.Request) {
	//filters includes:  start_date, end_date, status, payment_type, reversal, drcr
	filter := req.GetPossibleTransactionFilters(r)
	log.Println("Transaction filters", filter)
	transactions, err := usecase.GetTransactions(filter)
	if err != nil {
		req.SendSuccessResponse(w, false, transactions, err.Error(), 412)
		return
	}
	req.SendSuccessResponseWithMetadata(w, true, transactions.Data, "Transactions retreived", 200, transactions.MetaData)
}
