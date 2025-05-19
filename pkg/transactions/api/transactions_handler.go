package api

// TransactionsHandler handles HTTP requests
// func GetTransactions(w http.ResponseWriter, r *http.Request) {
// 	vars, err := req.GetRequestParams(r, "customerID")
// 	if err != nil {
// 		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	id, err := strconv.ParseUint(vars, 10, 32)
// 	if err != nil {
// 		req.SendSuccessResponse(w, false, nil, "Invalid customer ID", http.StatusBadRequest)
// 		return
// 	}
// 	//filters includes:  start_date, end_date, status, payment_type, reversal, drcr
// 	filter := req.GetPossibleTransactionFilters(r)
// 	log.Println("Transaction filters", filter)
// 	transactions, err := usecase.GetCustomerTransactionsById(uint(id), filter)
// 	if err != nil {
// 		req.SendSuccessResponse(w, false, transactions, err.Error(), 412)
// 		return
// 	}
// 	req.SendSuccessResponseWithMetadata(w, true, transactions.Data, "Transactions retreived", 200, transactions.MetaData)
// }
