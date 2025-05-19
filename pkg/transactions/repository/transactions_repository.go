package repository

import (
	"bloc-mfb/config/database"
	txnModel "bloc-mfb/pkg/transactions/model"
	"bloc-mfb/utils/exception"
)

// TransactionsRepository handles database operations
// func GetCustomerTransactionsById(customer_id uint, filters map[string]string) (txnModel.PaginatedResponse, error) {
// 	var tranctions []txnModel.Transactions
// 	if customer_id == 0 {
// 		return txnModel.PaginatedResponse{}, fmt.Errorf("customer_id is required")
// 	}

// 	page := 1
// 	if filters["page"] != "" {
// 		page, _ = strconv.Atoi(filters["page"])
// 		delete(filters, "page")
// 	}
// 	pageSize := 10
// 	offset := (page - 1) * pageSize
// 	//check if filter has start_date and end_date
// 	if filters["start_date"] != "" {
// 		startDate, err := time.Parse("2006-01-02", filters["start_date"])
// 		if err == nil {
// 			startDate = startDate.Truncate(24 * time.Hour) // Set start time to 12 midnight
// 		}
// 		if err != nil {
// 			return txnModel.PaginatedResponse{}, fmt.Errorf("Invalid start date")
// 		}
// 		var endDate time.Time
// 		if filters["end_date"] != "" {
// 			endDate, err = time.Parse("2006-01-02", filters["end_date"])
// 			if err == nil {
// 				endDate = endDate.Add(24*time.Hour - time.Nanosecond) // Set end time to 23:59:59
// 			}
// 		} else {
// 			endDate = time.Now()
// 		}
// 		if err != nil {
// 			return txnModel.PaginatedResponse{}, fmt.Errorf("Invalid end date")
// 		}
// 		tranctions = []txnModel.Transactions{}
// 		//delete the start and end date in the filter
// 		delete(filters, "start_date")
// 		delete(filters, "end_date")

// 		var totalRecords int64
// 		database.GetDB().Model(&txnModel.Transactions{}).Where(&txnModel.Transactions{CustomerID: customer_id}).Where(filters).Count(&totalRecords)

// 		//run query
// 		result := database.GetDB().
// 			Preload(clause.Associations).
// 			Where(&txnModel.Transactions{CustomerID: customer_id}).
// 			Where("created_at BETWEEN ? AND ?", startDate, endDate).
// 			Where(filters).Order("id desc").
// 			Limit(pageSize).
// 			Offset(offset).
// 			Find(&tranctions)

// 		totalPages := int(totalRecords) / pageSize
// 		if int(totalRecords)%pageSize != 0 {
// 			totalPages++
// 		}

// 		hasNextPage := (offset + len(tranctions)) < int(totalRecords)

// 		if result.Error != nil {
// 			return txnModel.PaginatedResponse{}, exception.HandleDBError(result.Error)
// 		}
// 		return txnModel.PaginatedResponse{
// 			MetaData: txnModel.PaginatedMetaData{
// 				TotalCount: totalRecords,
// 				Page:       page,
// 				PageSize:   pageSize,
// 				HasNext:    hasNextPage,
// 			},
// 			Data: tranctions,
// 		}, nil
// 	}
// 	if filters["start_date"] == "" && filters["end_date"] != "" {
// 		return txnModel.PaginatedResponse{}, fmt.Errorf("start_date is required if end_date is provided")
// 	}

// 	var totalRecords int64
// 	database.GetDB().Model(&txnModel.Transactions{}).Where(&txnModel.Transactions{CustomerID: customer_id}).Where(filters).Count(&totalRecords)

// 	result := database.GetDB().
// 		Preload(clause.Associations).
// 		Where(&txnModel.Transactions{CustomerID: customer_id}).
// 		Where(filters).
// 		Order("id desc").
// 		Limit(pageSize).
// 		Offset(offset).
// 		Find(&tranctions)

// 	totalPages := int(totalRecords) / pageSize
// 	if int(totalRecords)%pageSize != 0 {
// 		totalPages++
// 	}

// 	hasNextPage := (offset + len(tranctions)) < int(totalRecords)
// 	if result.Error != nil {
// 		return txnModel.PaginatedResponse{}, exception.HandleDBError(result.Error)
// 	}

// 	return txnModel.PaginatedResponse{
// 		MetaData: txnModel.PaginatedMetaData{
// 			TotalCount: totalRecords,
// 			Page:       page,
// 			PageSize:   pageSize,
// 			HasNext:    hasNextPage,
// 		},
// 		Data: tranctions,
// 	}, nil
// }

func SaveTransaction(transaction *txnModel.Transactions) (*txnModel.Transactions, error) {
	//save the transaction
	save := database.GetDB().Save(&transaction)
	if save.Error != nil {
		return nil, exception.HandleDBError(save.Error)
	}
	return transaction, nil
}
