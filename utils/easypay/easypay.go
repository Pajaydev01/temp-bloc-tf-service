package easypay

import (
	"bloc-mfb/config/database"
	transferModel "bloc-mfb/pkg/transfers/model"
	helper "bloc-mfb/utils/http"
	"bloc-mfb/utils/misc"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
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
		redisCLient := database.GetRedisClient()
		cachedToken, err := redisCLient.Get(context.Background(), "easypay_token").Result()
		if err == nil {
			return cachedToken, nil
		}
		log.Println("Token not found in cache, or cache expired, making call for another", err)
	}

	// If not cached or expired, make a new request to get the token
	url := fmt.Sprintf("%sreset", os.Getenv("EASYPAY_BASE_URL"))

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	// payload := map[string]string{
	// 	"client_id":     os.Getenv("EASY_PAY_CLIENT_ID"),
	// 	"scope":         fmt.Sprintf("%s/.default", os.Getenv("EASY_PAY_CLIENT_ID")),
	// 	"client_secret": os.Getenv("EASY_PAY_CLIENT_SECRET"),
	// 	"grant_type":    "client_credentials",
	// }
	payload := fmt.Sprintf("client_id=%s&scope=%s/.default&client_secret=%s&grant_type=client_credentials", os.Getenv("EASY_PAY_CLIENT_ID"), os.Getenv("EASY_PAY_CLIENT_ID"), os.Getenv("EASY_PAY_CLIENT_SECRET"))

	response, err := helper.MakeRequest(url+"?"+payload, "GET", payload, headers)
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
	token := result["token"].(string)
	if useCache {
		redisCLient := database.GetRedisClient()
		redisCLient.Set(context.Background(), "easypay_token", token, 0)
	}

	return token, nil
}

func EasyPayNameEnquiry(accountNumber, bankCode string) (transferModel.NESingleResponseEasyPay, error) {
	token, err := EasypayAuth(true)
	if err != nil {
		log.Println("Error getting EasyPay token:", err)
		return transferModel.NESingleResponseEasyPay{}, err
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
