package usecase

import (
	"bloc-mfb/config/database"
	accountModel "bloc-mfb/pkg/accounts/model"
	accountUsecase "bloc-mfb/pkg/accounts/usecase"
	"bloc-mfb/pkg/transactions/model"
	txnModel "bloc-mfb/pkg/transactions/model"
	txnUsecase "bloc-mfb/pkg/transactions/usecase"
	transferModel "bloc-mfb/pkg/transfers/model"
	"bloc-mfb/utils/easypay"
	crypt "bloc-mfb/utils/encryption"
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

func DoLocalTransfer(transferItem transferModel.TransferRequest) (acct accountModel.Accounts, txn txnModel.Transactions, err error) {
	account, err := accountUsecase.GetAccountByCustomerId(transferItem.CustomerId)
	var transaction txnModel.Transactions
	transaction.AccountID = account.ID
	transaction.AccountNumber = account.AccountNumber
	transaction.Currency = "NGN"
	transaction.AccountNumber = account.AccountNumber
	transaction.CustomerID = account.CustomerID
	if err != nil {
		return account, transaction, err
	}
	FeeAndVarResp, error := debitChargeAndFee(transferItem, account)
	//update the  account balance with the vat remaing balance, very important

	account.Balance = FeeAndVarResp.VatTransaction.CurrentAccountBalance
	account.AvailableBalance = FeeAndVarResp.VatTransaction.CurrentAccountBalance
	if error != nil {
		return account, transaction, error
	}

	account, txns, err := doMainDebitAndCallNibbs(account, transferItem, transaction, FeeAndVarResp)
	if err != nil {
		return accountModel.Accounts{}, txnModel.Transactions{}, err
	}
	//call main provider to do main transfer
	return account, txns, nil
}

func debitChargeAndFee(TransferItem transferModel.TransferRequest, account accountModel.Accounts) (tran model.FeeVatChargeResp, error error) {
	FeeAndVarResp := txnModel.FeeVatChargeResp{}
	Fee, Vat := GetTransferFeeAndVat(TransferItem)
	var transaction txnModel.Transactions
	transaction.AccountID = account.ID
	transaction.AccountNumber = account.AccountNumber
	transaction.Currency = "NGN"
	transaction.AccountNumber = account.AccountNumber
	transaction.CustomerID = account.CustomerID
	/// debit fee //
	log.Println("====== charging transfer fee ==========", Fee)
	account, old_balance, err := accountUsecase.DebitOrCreditAccount(account, "debit", Fee)
	if err != nil {
		return FeeAndVarResp, err
	}
	//create fee txn
	transaction.Amount = Fee
	transaction.Reference = crypt.GenerateTransactionRef()
	transaction.Status = string(txnModel.TRANSACTION_STATUS_SUCCESS)
	transaction.PaymentType = "charges"
	transaction.PreviousAccountBalance = old_balance
	transaction.CurrentAccountBalance = account.Balance
	transaction.DRCR = "DR"
	feeTxn, err := txnUsecase.CreateTransaction(transaction)
	if err != nil {
		//reverse the debit
		_, err = account.ReverseDebit(feeTxn)
		if err != nil {
			return FeeAndVarResp, err
		}
		return FeeAndVarResp, err
	}
	FeeAndVarResp.FeeTransactionId = feeTxn.ID
	FeeAndVarResp.FeeTransaction = feeTxn
	log.Println("====== charged transfer fee successfully ==========", Fee)

	//////////debit the charges
	log.Println("====== charging VAT fee ==========", Vat)
	acct, old_balance, err := accountUsecase.DebitOrCreditAccount(account, "debit", Vat)
	if err != nil {
		log.Println("Error creating vat debit", Vat)
		//reverse the debit and create reversal for fee
		old, er := account.ReverseDebit(FeeAndVarResp.FeeTransaction)
		if er != nil {
			log.Println("Error reversing fee debit")
			return FeeAndVarResp, er
		}
		log.Println("Reversing transfer charges from vat error block", Fee)
		FeeAndVarResp.FeeTransaction.Status = string(txnModel.TRANSACTION_STATUS_FAILED)
		_, _ = txnUsecase.UpdateTransaction(FeeAndVarResp.FeeTransaction)

		transaction.Reference = crypt.GenerateTransactionRef()
		transaction.Amount = FeeAndVarResp.FeeTransaction.Amount
		transaction.DRCR = "CR"
		transaction.PreviousAccountBalance = old
		transaction.CurrentAccountBalance = FeeAndVarResp.FeeTransaction.PreviousAccountBalance
		transaction.PaymentType = "charges"
		transaction.ReversedTransactionID = FeeAndVarResp.FeeTransactionId
		transaction.Status = string(txnModel.TRANSACTION_STATUS_SUCCESS)
		transaction.Reversal = true
		_, error := txnUsecase.CreateTransaction(transaction)
		if error != nil {
			log.Println("Error reversing fee debit")
			return FeeAndVarResp, error
		}
		return FeeAndVarResp, err
	}
	//create vat debit txn
	transaction.Amount = Vat
	transaction.Reference = crypt.GenerateTransactionRef()
	transaction.Status = string(txnModel.TRANSACTION_STATUS_SUCCESS)
	transaction.PaymentType = "VAT"
	transaction.PreviousAccountBalance = old_balance
	transaction.CurrentAccountBalance = acct.Balance
	transaction.DRCR = "DR"

	vatTxn, err := txnUsecase.CreateTransaction(transaction)
	if err != nil {
		_, err = acct.ReverseDebit(transaction)
		if err != nil {
			return FeeAndVarResp, err
		}
		//reverse the debit and create reversal for fee
		old, err := acct.ReverseDebit(FeeAndVarResp.FeeTransaction)
		if err != nil {
			return FeeAndVarResp, err
		}

		transaction.Amount = FeeAndVarResp.FeeTransaction.Amount
		transaction.DRCR = "CR"
		transaction.PreviousAccountBalance = old
		transaction.CurrentAccountBalance = FeeAndVarResp.FeeTransaction.PreviousAccountBalance
		transaction.PaymentType = "charges"
		_, err = txnUsecase.CreateTransaction(transaction)
		if err != nil {
			return FeeAndVarResp, err
		}
		return FeeAndVarResp, err
	}
	FeeAndVarResp.VatTransactionId = vatTxn.ID
	FeeAndVarResp.VatTransaction = vatTxn
	log.Println("====== charged VAT fee successfully ==========", Vat)
	return FeeAndVarResp, err
}

func ReversChargeAndFee(feeandvat txnModel.FeeVatChargeResp, account accountModel.Accounts) error {
	log.Println("====reversing transfer fee and charges======")
	fee_old_bal, error := account.ReverseDebit(feeandvat.FeeTransaction)
	if error != nil {
		log.Println("An error occured reversing transfer fee")
		return error
	}
	//update status of that transaction to failed first
	feeandvat.FeeTransaction.Status = string(txnModel.TRANSACTION_STATUS_FAILED)
	_, error = txnUsecase.UpdateTransaction(feeandvat.FeeTransaction)
	transaction := feeandvat.FeeTransaction
	transaction.ID = 0
	transaction.Amount = feeandvat.FeeTransaction.Amount
	transaction.DRCR = "CR"
	transaction.PreviousAccountBalance = feeandvat.VatTransaction.CurrentAccountBalance
	transaction.Reference = crypt.GenerateTransactionRef()
	transaction.CurrentAccountBalance = fee_old_bal + transaction.Amount
	transaction.Reversal = true
	transaction.ReversedTransactionID = feeandvat.FeeTransaction.ID
	transaction.PaymentType = "charges"
	transaction.Status = string(txnModel.TRANSACTION_STATUS_SUCCESS)
	_, error = txnUsecase.CreateTransaction(transaction)
	if error != nil {
		log.Println("Error creating charges reverse transaction")
		return error
	}

	log.Println("====reversing transfer VAT======")
	_, erro := account.ReverseDebit(feeandvat.VatTransaction)
	if erro != nil {
		return erro
	}
	feeandvat.VatTransaction.Status = string(txnModel.TRANSACTION_STATUS_FAILED)
	_, _ = txnUsecase.UpdateTransaction(feeandvat.VatTransaction)
	vat_prev_bal := transaction.CurrentAccountBalance
	vat_new_bal := vat_prev_bal + feeandvat.VatTransaction.Amount

	//update the previous and current bal before changing the transaction
	transaction = feeandvat.VatTransaction
	transaction.ID = 0
	transaction.PreviousAccountBalance = vat_prev_bal
	transaction.CurrentAccountBalance = vat_new_bal
	transaction.Amount = feeandvat.VatTransaction.Amount
	transaction.DRCR = "CR"
	transaction.Reference = crypt.GenerateTransactionRef()
	transaction.PaymentType = "VAT"
	transaction.Reversal = true
	transaction.Status = string(txnModel.TRANSACTION_STATUS_SUCCESS)
	transaction.ReversedTransactionID = feeandvat.VatTransaction.ID
	_, error = txnUsecase.CreateTransaction(transaction)
	if error != nil {
		return error
	}
	return nil
}

func doMainDebitAndCallNibbs(account accountModel.Accounts, txnItem transferModel.TransferRequest, transaction txnModel.Transactions, FeeAndVarResp txnModel.FeeVatChargeResp) (acct accountModel.Accounts, txn txnModel.Transactions, error error) {
	// proceed to debit main amount
	log.Println("====== Debiting main transfer amount ==========", txnItem.Amount)
	acct, old_balance, err := accountUsecase.DebitOrCreditAccount(account, "debit", txnItem.Amount)
	if err != nil {
		log.Println("====== error while debiting main tf amount ==========", err)
		error := ReversChargeAndFee(FeeAndVarResp, account)
		if error != nil {
			return accountModel.Accounts{}, txnModel.Transactions{}, error
		}
		return accountModel.Accounts{}, txnModel.Transactions{}, err
	}
	transaction.Amount = txnItem.Amount
	transaction.Reference = crypt.GenerateTransactionRef()
	transaction.Status = string(txnModel.TRANSACTION_STATUS_PENDING)
	transaction.PaymentType = "Transfer"
	transaction.PreviousAccountBalance = old_balance
	transaction.CurrentAccountBalance = acct.Balance
	transaction.DRCR = "DR"
	// _, err = txnUsecase.CreateTransaction(transaction)
	//here is where we call nibbs or main transfer provider

	// ////////
	if transaction.MetaData == nil {
		transaction.MetaData = &txnModel.MetaData{}
	}
	transaction.MetaData.ReceiverAccountName = txnItem.AccountNumber
	transaction.MetaData.ReceiverAccountNumber = txnItem.AccountNumber
	transaction.MetaData.ReceiverBankName = txnItem.BankCode
	transaction.MetaData.ReceiverBankCode = txnItem.BankCode
	transaction.MetaData.NipRef = crypt.GenerateRandomString(15)
	_, _ = transaction.UpdateTransactionMeta()
	log.Println("====== All debits successful, transactions submitted and waiting for tsq ==========", txnItem.Amount)
	return acct, transaction, nil
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
