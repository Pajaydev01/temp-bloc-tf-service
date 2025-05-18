package usecase

import (
	"bloc-mfb/config/database"
	"bloc-mfb/pkg/accounts/model"
	customerUseCase "bloc-mfb/pkg/customer/usecase"
	KycUseCase "bloc-mfb/pkg/kyc/usecase"
	crypt "bloc-mfb/utils/encryption"
	"bloc-mfb/utils/exception"
	"errors"
	"log"

	"gorm.io/gorm/clause"
)

// AccountsUseCase handles business logic
func CreateAccount(account model.Accounts) (model.Accounts, error) {
	//do tracing and all later
	//check for existing account since we have not set rule for a customer to have multiple account
	log.Println("=============== creating account ===============")
	check := database.GetDB().Where(&model.Accounts{CustomerID: account.CustomerID}).First(&account)
	if check.Error == nil {

		return model.Accounts{}, errors.New("user already has an existing Nuban account")
	}
	//get the user
	customer, err := customerUseCase.GetCustomerById(account.CustomerID)
	if err != nil {

		return model.Accounts{}, err
	}
	//generate account number
	sequence := crypt.GenerateRandomDigitsWithPrefix("20", 7)
	//add nip logic to generate account here
	account.GenerateBlocAccount(sequence)

	account.Name = *customer.FullName
	account.BVN = &customer.BVN
	account.KYCTIER = customer.KYCTier
	currency := "NGN"
	account.Currency = &currency
	create := database.GetDB().Create(&account)
	//attach customer and other association
	database.GetDB().Preload(clause.Associations).First(&account)
	if create.Error != nil {

		return model.Accounts{}, exception.HandleDBError(create.Error)
	}
	return account, nil
}

func GetAccountByCustomerId(id uint) (model.Accounts, error) {
	//fmt.Println("", id)
	var account = model.Accounts{}
	result := database.GetDB().Preload(clause.Associations).Where(&model.Accounts{CustomerID: id}).First(&account)
	if result.Error != nil {
		return model.Accounts{}, exception.HandleDBError(result.Error)
	}
	return account, nil
}

func GetAccountById(id uint) (model.Accounts, error) {
	//fmt.Println("", id)
	var account = model.Accounts{}
	result := database.GetDB().Preload(clause.Associations).Where(&model.Accounts{ID: id}).First(&account)
	if result.Error != nil {
		return model.Accounts{}, exception.HandleDBError(result.Error)
	}
	return account, nil
}

func GetAccountByAccounNumber(account_no string) (model.Accounts, error) {
	var account = model.Accounts{}
	result := database.GetDB().Preload(clause.Associations).Where(&model.Accounts{AccountNumber: account_no}).First(&account)
	if result.Error != nil {
		return model.Accounts{}, exception.HandleDBError(result.Error)
	}
	return account, nil
}

func DebitOrCreditAccount(account model.Accounts, debit_credit string, amount int64) (acct model.Accounts, balance_before int64, error error) {
	log.Printf("%sing account", debit_credit)
	var old_before int64
	//lock the account first
	err := account.LockAccount()
	if err != nil {
		return model.Accounts{}, 0, err
	}
	if debit_credit == "credit" {
		account, err = KycUseCase.ValidateCreditKyc(account, amount)
		if err != nil {
			_ = account.UnlockAccount()
			return model.Accounts{}, 0, err
		}
		old, err := account.CreditAccount(amount)
		old_before = old
		if err != nil {
			_ = account.UnlockAccount()
			return model.Accounts{}, 0, err
		}

	}
	if debit_credit == "debit" {
		account, err = KycUseCase.ValidateDebitKyc(account, amount)
		if err != nil {
			_ = account.UnlockAccount()
			return model.Accounts{}, 0, err
		}
		old, err := account.DebitAccount(amount)
		old_before = old
		if err != nil {
			_ = account.UnlockAccount()
			return model.Accounts{}, 0, err
		}
	}
	err = account.UnlockAccount()
	if err != nil {
		_ = account.UnlockAccount()
		return model.Accounts{}, 0, err
	}
	return account, old_before, nil
}
