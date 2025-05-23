package easypay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bloc-transfer-service/config/database"
	txnModel "github.com/bloc-transfer-service/pkg/transactions/model"
	txnRepo "github.com/bloc-transfer-service/pkg/transactions/repository"
	transferModel "github.com/bloc-transfer-service/pkg/transfers/model"
	asyncQueue "github.com/bloc-transfer-service/utils/AsyncQueue"
	helper "github.com/bloc-transfer-service/utils/http"
	"github.com/bloc-transfer-service/utils/misc"
	"github.com/hibiken/asynq"
)

func getDataFormat() string {
	currentTime := time.Now()
	timeFormat := strings.Split(currentTime.Format("15:04:05"), ":")
	date := strings.Split(currentTime.Format("06-01-02"), "-")
	year := date[0]
	month := date[1]
	day := date[2]
	hour := timeFormat[0]
	minute := timeFormat[1]
	seconds := timeFormat[2]
	return fmt.Sprintf("%v%v%v%v%v%v", year, month, day, hour, minute, seconds)
}

func generateSessionId(bankCode string) string {
	date := getDataFormat()
	random := misc.GenerateRandomDigits(12)
	sessionId := fmt.Sprintf("%v%v%v", bankCode, date, random)
	return sessionId
}

func EasypayAuth(useCache bool) (string, error) {
	// Check if the token is already cached and not expired
	if useCache {
		log.Println("Checking for cached EasyPay token")
		redisCLient := database.GetRedisClient()
		cachedToken, err := redisCLient.Get(context.Background(), "easypay_token").Result()
		if err == nil {
			return cachedToken, nil
		}
		log.Println("Token not found in cache, or cache expired, making call for another", err)
	}

	if os.Getenv("ENVIRONMENT") == "dev" {
		log.Println("Simulating EasyPay token generation in dev environment")
		token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
		redisCLient := database.GetRedisClient()
		redisCLient.Set(context.Background(), "easypay_token", token, 55*time.Minute)
		return token, nil
	}

	// If not cached or expired, make a new request to get the token
	url := fmt.Sprintf("%sreset", os.Getenv("EASYPAY_BASE_URL"))

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	// payload := map[string]string{
	// 	"client_id":     os.Getenv("EASY_PAY_CLIENT_ID"),
	// 	"scope":         fmt.Sprintf("%s/.default", os.Getenv("EASY_PAY_CLIENT_ID")),
	// 	"client_secret": os.Getenv("EASY_PAY_CLIENT_SECRET"),
	// 	"grant_type":    "client_credentials",
	// }
	payload := strings.NewReader(fmt.Sprintf("client_id=%s&scope=%s/.default&client_secret=%s&grant_type=client_credentials", os.Getenv("EASY_PAY_CLIENT_ID"), os.Getenv("EASY_PAY_CLIENT_ID"), os.Getenv("EASY_PAY_CLIENT_SECRET")))

	response, err := helper.MakeRequest(url, "POST", payload, headers)
	if err != nil {
		log.Println("Error making generating token:", err)
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", fmt.Errorf("failed to authenticate: %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}
	log.Println("Response from EasyPay:", string(body))
	token := result["access_token"].(string)
	if useCache {
		redisCLient := database.GetRedisClient()
		redisCLient.Set(context.Background(), "easypay_token", token, 55*time.Minute)
	}

	return token, nil
}

func EasyPayNameEnquiry(accountNumber, bankCode string) (transferModel.NESingleResponseEasyPay, error) {
	log.Println("============ EasyPay name enquiry request initiated here ==========================")
	token, err := EasypayAuth(true)
	if err != nil {
		log.Println("Error getting EasyPay token:", err)
		return transferModel.NESingleResponseEasyPay{}, fmt.Errorf("unable to make name enquiry at this time, please try again later")
	}
	url := fmt.Sprintf("%snipservice/v1/nip/nameenquiry", os.Getenv("EASYPAY_BASE_URL"))
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
	payload := map[string]string{
		"accountNumber":              accountNumber,
		"destinationInstitutionCode": bankCode,
		"channelCode":                "1",
		"transactionId":              generateSessionId(os.Getenv("EASY_PAY_BANK_CODE")),
	}

	//handle test cases here
	if os.Getenv("ENVIRONMENT") == "dev" {
		//simulate success
		result := transferModel.NESingleResponseEasyPay{
			ResponseCode:               "00",
			TransactionId:              payload["transactionId"],
			ChannelCode:                1,
			DestinationInstitutionCode: payload["destinationInstitutionCode"],
			AccountNumber:              payload["accountNumber"],
			AccountName:                "test user",
			BankVerificationNumber:     "210982389230",
			KycLevel:                   1,
			Message:                    "Success",
		}
		return result, nil
	}
	response, err := helper.MakeRequest(url, "POST", payload, headers)
	if err != nil {
		log.Println("Error making request to EasyPay:", err)
		return transferModel.NESingleResponseEasyPay{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return transferModel.NESingleResponseEasyPay{}, fmt.Errorf("failed to authenticate: %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return transferModel.NESingleResponseEasyPay{}, err
	}

	var result transferModel.NESingleResponseEasyPay
	err = json.Unmarshal(body, &result)
	if err != nil {
		return transferModel.NESingleResponseEasyPay{}, err
	}
	log.Println("Response from EasyPay:", string(body))

	return result, nil
}

func EasyPayTransfer(transferItem transferModel.TransferRequest, beneficiary transferModel.NESingleResponseEasyPay, transaction txnModel.Transactions) (transferModel.EASYPAYFtSingleResponse, error) {
	log.Println("============ EasyPay transfer request initiated here ==========================")
	token, err := EasypayAuth(true)
	if err != nil {
		log.Println("Error getting EasyPay token:", err)
		return transferModel.EASYPAYFtSingleResponse{}, err
	}
	log.Println("EasyPay token:", token)

	url := fmt.Sprintf("%snipservice/v1/nip/fundtransfer", os.Getenv("EASYPAY_BASE_URL"))
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
	transferItem.SessionID = generateSessionId(os.Getenv("EASY_PAY_BANK_CODE"))

	payload := map[string]interface{}{
		"amount":                            convertAmountTSF(transferItem.Amount),
		"beneficiaryAccountName":            beneficiary.AccountName,
		"beneficiaryAccountNumber":          beneficiary.AccountNumber,
		"beneficiaryBankVerificationNumber": beneficiary.BankVerificationNumber,
		"beneficiaryKYCLevel":               strconv.Itoa(beneficiary.KycLevel),
		"channelCode":                       1,
		"originatorAccountName":             fmt.Sprintf("Bloc-%s", transferItem.SenderName),
		"originatorAccountNumber":           os.Getenv("EASY_PAY_ORIGINATOR_ACCOUNT"),
		"originatorKYCLevel":                "3",
		"paymentReference":                  generateSessionId(os.Getenv("EASY_PAY_BANK_CODE")),
		"transactionLocation":               "1.39716,3.07117",
		"originatorNarration":               transferItem.Narration,
		"beneficiaryNarration":              transferItem.Narration,
		"destinationInstitutionCode":        transferItem.BankCode,
		"transactionId":                     generateSessionId(os.Getenv("EASY_PAY_BANK_CODE")),
		"originatorBankVerificationNumber":  "33333333333",
		"nameEnquiryRef":                    beneficiary.SessionID,
		"sessionID":                         transferItem.SessionID,
	}
	log.Println("Payload for EasyPay transfer:", payload)
	//handle test cases here
	if os.Getenv("ENVIRONMENT") == "dev" {
		//simulate success
		result := transferModel.EASYPAYFtSingleResponse{
			ResponseCode:                      "00",
			SessionID:                         transferItem.SessionID,
			TransactionId:                     generateSessionId(os.Getenv("EASY_PAY_BANK_CODE")),
			ChannelCode:                       "1",
			NameEnquiryRef:                    payload["nameEnquiryRef"].(string),
			DestinationInstitutionCode:        payload["destinationInstitutionCode"].(string),
			BeneficiaryAccountName:            payload["beneficiaryAccountName"].(string),
			BeneficiaryAccountNumber:          payload["beneficiaryAccountNumber"].(string),
			BeneficiaryKYCLevel:               payload["beneficiaryKYCLevel"].(string),
			BeneficiaryBankVerificationNumber: payload["beneficiaryBankVerificationNumber"].(string),
			OriginatorAccountName:             payload["originatorAccountName"].(string),
			OriginatorAccountNumber:           payload["originatorAccountNumber"].(string),
			OriginatorBankVerificationNumber:  payload["originatorBankVerificationNumber"].(string),
			OriginatorKYCLevel:                payload["originatorKYCLevel"].(string),
			TransactionLocation:               payload["transactionLocation"].(string),
			Narration:                         payload["originatorNarration"].(string),
			PaymentReference:                  payload["paymentReference"].(string),
			Amount:                            payload["amount"].(string),
		}
		//successful, add to tsq quue
		err := addTransferToQueue(transferItem, beneficiary, transaction, "0", 20*time.Second, transferItem.SessionID)
		if err != nil {
			log.Println("Error adding transfer to queue:", err)
			return transferModel.EASYPAYFtSingleResponse{}, err
		}
		return result, nil
		//submit  to queue runner to do tsq
	}
	response, err := helper.MakeRequest(url, "POST", payload, headers)
	if err != nil {
		log.Println("Error making transfer request to EasyPay:", err)
		return transferModel.EASYPAYFtSingleResponse{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return transferModel.EASYPAYFtSingleResponse{}, err
	}

	var result transferModel.EASYPAYFtSingleResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return transferModel.EASYPAYFtSingleResponse{}, err
	}
	log.Println("Response from EasyPay transfer:", string(body))
	//submit to queue runner to do tsq
	err = addTransferToQueue(transferItem, beneficiary, transaction, "0", 20*time.Second, transferItem.SessionID)
	if err != nil {
		log.Println("Error adding transfer to queue:", err)
		return transferModel.EASYPAYFtSingleResponse{}, err
	}
	return result, nil
}

func addTransferToQueue(transferItem transferModel.TransferRequest, beneficiary transferModel.NESingleResponseEasyPay, transaction txnModel.Transactions, attempt string, duration time.Duration, sessionId string) error {
	log.Println("Adding transfer to queue for processing:", transferItem.SessionID)
	// Create the payload
	item := map[string]interface{}{
		"transferItem": transferItem,
		"beneficiary":  beneficiary,
		"transaction":  transaction,
		"sessionId":    sessionId,
		"attempt":      attempt,
	}

	payload := asyncQueue.TaskPayload{
		Name:  "processTransfer",
		Items: item,
	}
	// Enqueue the task
	err := asyncQueue.EnqueueTask(payload, duration)
	if err != nil {
		log.Println("Error enqueuing task:", err)
		return err
	}
	return nil
}

func convertAmountTSF(amount int64) string {
	f := math.Ceil(float64(amount*100)) / 100

	s := fmt.Sprintf("%.2f", f/100)
	return s
}

func HandleTsqQueue(ctx context.Context, task *asynq.Task) error {
	log.Println("================= Queue processor now Processing task:============ ")
	// Enqueue the task
	// EnqueueTask(payload, 0)
	transferItem, beneficiary, transaction, attempt, sessionId, err := getQueuedData(task)
	log.Println("Session id for transfer:", sessionId)
	if err != nil {
		log.Println("Error getting queued data:", err)
	}
	intervalStr := os.Getenv("EASYPAY_TSQ_INTERVAL_TIME_MINUTES")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Println("Error converting EASYPAY_TSQ_INTERVAL_TIME_MINUTES to integer:", err)
		return err
	}
	//do tsq
	result, err := DoTsqEasyPayTransfer(sessionId)
	if err != nil {
		log.Println("Error making request to EasyPay:, retry in 10 minutes", err)
		// Enqueue the task again with a delay
		addTransferToQueue(transferItem, beneficiary, transaction, "0", time.Duration(interval)*time.Minute, sessionId)
		return nil
	}
	if result.ResponseCode != "00" {
		log.Println("Error making request to EasyPay:, retry in 10 minutes, response code: ", result.ResponseCode)
		if attempt == "0" {
			log.Println("Attempt 2 for this transaction in 10 minute", result.ResponseCode)
			addTransferToQueue(transferItem, beneficiary, transaction, "1", time.Duration(interval)*time.Minute, sessionId)
		} else if attempt == "1" {
			log.Println("Attempt 3 for this transaction in 10 minute", result.ResponseCode)
			addTransferToQueue(transferItem, beneficiary, transaction, "2", time.Duration(interval)*time.Minute, sessionId)
		} else {
			//there are certain response codes that should not be failed, just leave and submit the job to run the next day and fail it

			log.Println("All attempts exhausted for this transaction, fail the transaction", transferItem.SessionID)
			//generate metadata  and save to transaction table
			transaction.Status = string(txnModel.TRANSACTION_STATUS_FAILED)
			var metadata txnModel.MetaData
			metadata.NipRef = result.TransactionId
			metadata.TransactionId = transaction.ID
			metadata.ReceiverAccountName = beneficiary.AccountName
			metadata.ReceiverAccountNumber = beneficiary.AccountNumber
			metadata.ReceiverBankCode = beneficiary.DestinationInstitutionCode
			metadata.ReceiverBankName = beneficiary.DestinationInstitutionCode
			metadata.SenderAccountName = transferItem.SenderName
			metadata.SenderAccountNumber = transferItem.AccountNumber
			metadata.SenderBankName = transferItem.BankCode
			metadata.NibbsResponsecode = result.ResponseCode
			metadata.SessionId = sessionId
			transaction.MetaData = &metadata
			txnRepo.SaveOrUpdateTransaction(&transaction)
			//call the webhooks and notification
		}
	} else {
		log.Println("Transaction successful, update the transaction table", transferItem.SessionID)
		transaction.Status = string(txnModel.TRANSACTION_STATUS_SUCCESS)
		var metadata txnModel.MetaData
		metadata.NipRef = result.TransactionId
		metadata.TransactionId = transaction.ID
		metadata.ReceiverAccountName = beneficiary.AccountName
		metadata.ReceiverAccountNumber = beneficiary.AccountNumber
		metadata.ReceiverBankCode = beneficiary.DestinationInstitutionCode
		metadata.ReceiverBankName = beneficiary.DestinationInstitutionCode
		metadata.SenderAccountName = transferItem.SenderName
		metadata.SenderAccountNumber = transferItem.AccountNumber
		metadata.SenderBankName = transferItem.BankCode
		metadata.NibbsResponsecode = result.ResponseCode
		metadata.SessionId = sessionId
		transaction.MetaData = &metadata
		txnRepo.SaveOrUpdateTransaction(&transaction)
	}
	return nil
}

func DoTsqEasyPayTransfer(sessionId string) (transferModel.TSQuerySingleResponse, error) {
	url := fmt.Sprintf("%snipservice/v1/nip/tsq", os.Getenv("EASYPAY_BASE_URL"))
	token, err := EasypayAuth(true)
	if err != nil {
		log.Println("Error getting EasyPay token:", err)
		return transferModel.TSQuerySingleResponse{}, err
	}
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
	payload := map[string]string{
		"transactionId": sessionId,
	}
	if os.Getenv("ENVIRONMENT") == "dev" {
		//simulate success
		responseCode := getResponseTestCode()
		result := transferModel.TSQuerySingleResponse{
			ResponseCode:          responseCode,
			TransactionId:         sessionId,
			ChannelCode:           "1",
			SourceInstitutionCode: "000000",
		}
		return result, nil
	}
	response, err := helper.MakeRequest(url, "POST", payload, headers)
	if err != nil {
		log.Println("Error making request to EasyPay:", err)
		return transferModel.TSQuerySingleResponse{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return transferModel.TSQuerySingleResponse{}, err
	}

	var result transferModel.TSQuerySingleResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return transferModel.TSQuerySingleResponse{}, err
	}
	return result, nil
}

func getResponseTestCode() string {
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(6) + 1 // Random number from 1 to 6
	if num%2 == 1 {
		return "00"
	}
	return "25"
}

func getQueuedData(task *asynq.Task) (req transferModel.TransferRequest, benef transferModel.NESingleResponseEasyPay, txn txnModel.Transactions, attempt string, sessionId string, err error) {
	var transferItem transferModel.TransferRequest
	var beneficiary transferModel.NESingleResponseEasyPay
	var transaction txnModel.Transactions

	var savedItems asyncQueue.TaskPayload
	if err := json.Unmarshal(task.Payload(), &savedItems); err != nil {
		return transferItem, beneficiary, transaction, attempt, sessionId, fmt.Errorf("failed to deserialize payload: %v", err)
	}
	// Deserialize the items
	log.Println("Items in the queue:", savedItems.Items["transferItem"])
	transferItemData, err := json.Marshal(savedItems.Items["transferItem"])
	if err != nil {
		return transferItem, beneficiary, transaction, attempt, sessionId, fmt.Errorf("failed to serialize transferItem: %v", err)
	}
	err = json.Unmarshal(transferItemData, &transferItem)
	if err != nil {
		return transferItem, beneficiary, transaction, attempt, sessionId, fmt.Errorf("failed to deserialize transferItem: %v", err)
	}
	beneficiaryData, err := json.Marshal(savedItems.Items["beneficiary"])
	if err != nil {
		return transferItem, beneficiary, transaction, attempt, sessionId, fmt.Errorf("failed to serialize beneficiary: %v", err)
	}
	err = json.Unmarshal(beneficiaryData, &beneficiary)
	if err != nil {
		return transferItem, beneficiary, transaction, attempt, sessionId, fmt.Errorf("failed to deserialize beneficiary: %v", err)
	}

	transactionData, err := json.Marshal(savedItems.Items["transaction"])
	if err != nil {
		return transferItem, beneficiary, transaction, attempt, sessionId, fmt.Errorf("failed to serialize transaction: %v", err)
	}
	err = json.Unmarshal(transactionData, &transaction)
	if err != nil {
		return transferItem, beneficiary, transaction, attempt, sessionId, fmt.Errorf("failed to deserialize transaction: %v", err)
	}

	// Deserialize the attempt
	attemptData, ok := savedItems.Items["attempt"].(string)
	if !ok {
		return transferItem, beneficiary, transaction, attempt, sessionId, fmt.Errorf("failed to serialize attempt: %v", err)
	}

	session_id, ok := savedItems.Items["sessionId"].(string)
	if !ok {
		return transferItem, beneficiary, transaction, attempt, sessionId, fmt.Errorf("failed to serialize attempt: %v", err)
	}

	return transferItem, beneficiary, transaction, attemptData, session_id, nil
}

func GeInstitutions() ([]transferModel.NIPInstitutions, error) {
	log.Println("============ EasyPay institutions request initiated here ==========================")
	redisClient := database.GetRedisClient()
	cachekey := "easypay_institutions"
	// Check if the token is already cached and not expired
	cachedInstitutions, err := redisClient.Get(context.Background(), cachekey).Result()
	log.Println("Cached institutions:", cachedInstitutions)
	log.Println("Cached institutions error:", err)
	if err == nil {
		var institutions []transferModel.NIPInstitutions
		err = json.Unmarshal([]byte(cachedInstitutions), &institutions)
		if err == nil {
			log.Println("Institutions from cache:", institutions)
			return institutions, nil
		}
		log.Println("Error unmarshalling cached institutions:", err)
	}
	url := fmt.Sprintf("%snipservice/v1/nip/institutions", os.Getenv("EASYPAY_BASE_URL"))
	token, err := EasypayAuth(true)
	if err != nil {
		log.Println("Error getting EasyPay token:", err)
		return []transferModel.NIPInstitutions{}, err
	}
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
	if os.Getenv("ENVIRONMENT") == "dev" {
		//simulate success
		result := []transferModel.NIPInstitutions{
			{
				InstitutionCode: "000000",
				InstitutionName: "Bloc Bank",
				Category:        1,
				CategoryCode:    "C",
			},
			{
				InstitutionCode: "000001",
				InstitutionName: "Bloc Bank 2",
				Category:        1,
				CategoryCode:    "C",
			},
			{
				InstitutionCode: "000012",
				InstitutionName: "Bloc Bank 3",
				Category:        1,
				CategoryCode:    "C",
			},
			{
				InstitutionCode: "000023",
				InstitutionName: "Bloc Bank 2",
				Category:        1,
				CategoryCode:    "C",
			},
		}
		//save to redis
		jsn, err := json.Marshal(result)
		_, err = redisClient.Set(context.Background(), cachekey, jsn, 30*time.Minute).Result()
		if err != nil {
			log.Println("Redis error!!!!!", err)
			return []transferModel.NIPInstitutions{}, err
		}
		log.Println("Institutions from EasyPay:", result)
		return result, nil
	}
	response, err := helper.MakeRequest(url, "GET", nil, headers)
	if err != nil {
		log.Println("Error making request to EasyPay:", err)
		return []transferModel.NIPInstitutions{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return []transferModel.NIPInstitutions{}, err
	}

	var result []transferModel.NIPInstitutions
	err = json.Unmarshal(body, &result)
	if err != nil {
		return []transferModel.NIPInstitutions{}, err
	}
	//save to redis
	jsn, err := json.Marshal(result)
	_, err = redisClient.Set(context.Background(), cachekey, jsn, 30*time.Minute).Result()
	if err != nil {
		log.Println("Redis error!!!!!", err)
		return []transferModel.NIPInstitutions{}, err
	}

	log.Println("Response from EasyPay:", string(body))
	return result, nil
}
