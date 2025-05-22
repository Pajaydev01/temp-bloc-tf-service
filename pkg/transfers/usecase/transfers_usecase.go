package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bloc-transfer-service/config/database"
	txnModel "github.com/bloc-transfer-service/pkg/transactions/model"
	transactionRepo "github.com/bloc-transfer-service/pkg/transactions/repository"
	transferModel "github.com/bloc-transfer-service/pkg/transfers/model"
	"github.com/bloc-transfer-service/utils/easypay"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

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
	_, err = transactionRepo.SaveOrUpdateTransaction(&transaction)
	if err != nil {
		return txnModel.Transactions{}, err
	}
	_, err = doMainDebitAndCallNibbs(transferItem, nameEnquiryResponse, transaction)
	if err != nil {
		return txnModel.Transactions{}, err
	}
	//remove the redis cache
	_, err = redisCon.Del(context.Background(), key).Result()
	if err != nil {
		log.Println("Redis error!!!!!", err)
		return txnModel.Transactions{}, err
	}
	//call main provider to do main transfer
	return transaction, nil
}

func doMainDebitAndCallNibbs(txnItem transferModel.TransferRequest, beneficiary transferModel.NESingleResponseEasyPay, transation txnModel.Transactions) (transferModel.EASYPAYFtSingleResponse, error) {
	// proceed to debit main amount
	log.Println("====== Debiting main transfer and calling nibbs ==========", txnItem.Amount)
	result, err := easypay.EasyPayTransfer(txnItem, beneficiary, transation)
	if err != nil {
		return transferModel.EASYPAYFtSingleResponse{}, err
	}
	//err,token
	return result, nil
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

func GetInstitution() ([]transferModel.NIPInstitutions, error) {
	institutions, err := easypay.GeInstitutions()
	if err != nil {
		log.Println("Error making request to EasyPay:", err)
		return nil, err
	}
	return institutions, nil
}
