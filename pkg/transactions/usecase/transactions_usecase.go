package usecase

import (
	txnModel "github.com/bloc-transfer-service/pkg/transactions/model"
	txnRepo "github.com/bloc-transfer-service/pkg/transactions/repository"
)

// // TransactionsUseCase handles business logic
func GetTransactions(filters map[string]string) (txnModel.PaginatedResponse, error) {
	transactions, error := txnRepo.GetTransactions(filters)
	if error != nil {
		return txnModel.PaginatedResponse{}, error
	}
	return transactions, nil
}

// // get the day transaction for an accoubt excluding fee and vat
// func GetAccountDayTransactions(account accountModel.Accounts) ([]txnModel.Transactions, error) {
// 	startOfDay := time.Now().Truncate(24 * time.Hour)                // Midnight today
// 	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond) // 23:59:59 today

// 	var transactions []txnModel.Transactions
// 	result := database.GetDB().
// 		Where("payment_type <> ? AND payment_type <> ? AND account_id = ? AND created_at BETWEEN ? AND ?", "VAT", "charges", account.ID, startOfDay, endOfDay).
// 		Find(&transactions)
// 	if result.Error != nil {
// 		return []txnModel.Transactions{}, exception.HandleDBError(result.Error)
// 	}
// 	return transactions, nil
// }

// func GetTotalAmount(transactions []txnModel.Transactions) int64 {
// 	var totalAmount int64 // Change type if Amount is int

// 	for _, txn := range transactions {
// 		totalAmount += txn.Amount
// 	}
// 	return totalAmount
// }

// func CreateTransaction(transaction txnModel.Transactions) (txnModel.Transactions, error) {
// 	//we assume the transaction is properly filled
// 	create := database.GetDB().Create(&transaction)
// 	if create.Error != nil {
// 		return txnModel.Transactions{}, exception.HandleDBError(create.Error)
// 	}
// 	return transaction, nil
// }

// func UpdateTransaction(transaction txnModel.Transactions) (txnModel.Transactions, error) {
// 	//update
// 	update := database.GetDB().Save(&transaction)
// 	if update.Error != nil {
// 		return txnModel.Transactions{}, exception.HandleDBError(update.Error)
// 	}
// 	return transaction, nil
// }

// func GetTransactionById(id uint) (txnModel.Transactions, error) {
// 	var tranctions txnModel.Transactions
// 	result := database.GetDB().Where(&txnModel.Transactions{ID: id}).First(&tranctions)
// 	if result.Error != nil {
// 		return txnModel.Transactions{}, exception.HandleDBError(result.Error)
// 	}
// 	return tranctions, nil
// }

// func GetTransactionByReference(ref string) (txnModel.Transactions, error) {
// 	var tranctions txnModel.Transactions
// 	result := database.GetDB().Where(&txnModel.Transactions{Reference: ref}).First(&tranctions)
// 	if result.Error != nil {
// 		log.Println("Error getting transaction by reference")
// 		return txnModel.Transactions{}, exception.HandleDBError(result.Error)
// 	}
// 	return tranctions, nil
// }

// func GetAccountTransactionsById(account_id uint) ([]txnModel.Transactions, error) {
// 	var tranctions []txnModel.Transactions
// 	result := database.GetDB().Where(&txnModel.Transactions{AccountID: account_id}).First(&tranctions)
// 	if result.Error != nil {
// 		return nil, exception.HandleDBError(result.Error)
// 	}
// 	return tranctions, nil
// }

// func UpdateTransactionStatus(transaction txnModel.Transactions, status string) (txnModel.Transactions, error) {
// 	transaction.Status = status
// 	update := database.GetDB().Save(&transaction)
// 	if update.Error != nil {
// 		return txnModel.Transactions{}, exception.HandleDBError(update.Error)
// 	}
// 	return transaction, nil
// }

// func ReverseTransaction(transaction txnModel.Transactions, VatAndFee txnModel.FeeVatChargeResp) {

// }
