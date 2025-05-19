package usecase

import (
	"bloc-mfb/config/database"
	txnModel "bloc-mfb/pkg/transactions/model"
	transactionRepo "bloc-mfb/pkg/transactions/repository"
	transferModel "bloc-mfb/pkg/transfers/model"
	"bloc-mfb/utils/easypay"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

// TransfersUseCase handles business logic
func GetTransferFeeAndVat(transferItem transferModel.TransferRequest) (Fee int64, Vat int64) {
	// cant write the logic now
	return 5000, 100
}

func DoLocalTransfer(transferItem transferModel.TransferRequest) (txn txnModel.Transactions, err error) {
	//pull the details from the ref id in the redis cache
	redisCon := database.GetRedisClient()
	key := transferItem.RefId
	val, err := redisCon.Get(context.Background(), key).Result()
	if err != nil {
		log.Println("Redis error!!!!! or no name enquiry for this particular transfer", err)
		return txnModel.Transactions{}, fmt.Errorf("invalid ref_id or expired ref_id")
	}
	log.Println("Redis data for name enquiry", val)
	var nameEnquiryResponse transferModel.NESingleResponseEasyPay
	err = json.Unmarshal([]byte(val), &nameEnquiryResponse)
	if err != nil {
		log.Println("Error unmarshalling redis data", err)
		return txnModel.Transactions{}, err
	}

	var transaction txnModel.Transactions
	transaction.AccountNumber = nameEnquiryResponse.AccountNumber
	transaction.Currency = "NGN"
	transaction.Amount = transferItem.Amount
	transaction.Category = transferItem.Type
	transaction.Narration = transferItem.Narration
	transaction.Status = string(txnModel.TRANSACTION_STATUS_PENDING)
	transaction.Reference = transferItem.Reference
	transaction.DRCR = "DR"
	//transaction.Environment = os.Getenv("ENVIRONMENT")
	//save transaction
	_, err = transactionRepo.SaveTransaction(&transaction)
	if err != nil {
		return txnModel.Transactions{}, err
	}
	txns, err := doMainDebitAndCallNibbs(transferItem, transaction)
	if err != nil {
		return txnModel.Transactions{}, err
	}
	//call main provider to do main transfer
	return txns, nil
}

func doMainDebitAndCallNibbs(txnItem transferModel.TransferRequest, transaction txnModel.Transactions) (txn txnModel.Transactions, error error) {
	// proceed to debit main amount
	log.Println("====== Debiting main transfer and calling nibbs ==========", txnItem.Amount)
	//err,token
	return transaction, nil
}

func DoNameEnquiry(nameEnquiryItem transferModel.NameEnquiryRequest) (bson.M, error) {

	//test
	redisCon := database.GetRedisClient()
	key := fmt.Sprintf("sess_%s", primitive.NewObjectID().Hex())
	if os.Getenv("ENVIRONMENT") == "production" {
		response, err := easypay.EasyPayNameEnquiry(nameEnquiryItem.AccountNumber, nameEnquiryItem.BankCode)
		if err != nil {
			log.Println("Error making request to EasyPay:", err)
			return bson.M{}, err
		}
		if response.ResponseCode != "00" {
			log.Println("Error making request to EasyPay:", response.Message)
			return bson.M{}, fmt.Errorf("failed to authenticate: %s", response.Message)
		}
		log.Println("Response from EasyPay:", response)
		//generate a key and cache the response to use for 10 minutes to the transfer to avoid re-making name enquiry for transfer
		jsn, err := json.Marshal(response)
		_, err = redisCon.Set(context.Background(), key, jsn, 10*time.Minute).Result()
		if err != nil {
			log.Println("Redis error!!!!!", err)
			return bson.M{}, err
		}
		return bson.M{"account_number": response.AccountNumber, "account_name": response.AccountName, "ref_id": key}, nil
	}

	item := transferModel.NESingleResponseEasyPay{
		ResponseCode:               "00",
		TransactionId:              key,
		ChannelCode:                2,
		DestinationInstitutionCode: nameEnquiryItem.BankCode,
		AccountNumber:              nameEnquiryItem.AccountNumber,
		AccountName:                "test user",
		BankVerificationNumber:     "210982389230",
		KycLevel:                   1,
		Message:                    "Success",
	}
	jsn, err := json.Marshal(item)
	_, err = redisCon.Set(context.Background(), key, jsn, 10*time.Minute).Result()
	if err != nil {
		log.Println("Redis error!!!!!", err)
		return bson.M{}, err

	}
	return bson.M{"account_number": "123678947", "account_name": "test user", "ref_id": key}, nil
}
