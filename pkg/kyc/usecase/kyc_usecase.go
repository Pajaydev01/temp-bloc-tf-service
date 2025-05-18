package usecase

import (
	accountModel "bloc-mfb/pkg/accounts/model"
	customerModel "bloc-mfb/pkg/customer/model"
	kycModel "bloc-mfb/pkg/kyc/model"
	transactionUsecase "bloc-mfb/pkg/transactions/usecase"
	"errors"
	"log"
)

// KycUseCase handles business logic
func ValidateCreditKyc(account accountModel.Accounts, amount int64) (accountModel.Accounts, error) {
	log.Println("------validating precredit kyc----------")
	newbal := account.AvailableBalance + amount
	switch account.Customer.KYCTier {
	case "1":
		if amount > kycModel.T1_MAX_CREDIT_AMT || newbal > kycModel.T1_MAX_ACCT_BAL {
			var note accountModel.Note
			note.Action = "Frozen KYC Tier 1 Account"
			note.Reason = "Performed a transaction that failed to comply with our KYC Level 1 requirements"
			err := account.FreezeAccount(note)
			if err != nil {
				return accountModel.Accounts{}, err
			}
			return account, nil
		}

	case "2":
		if amount > kycModel.T2_MAX_CREDIT_AMT || newbal > kycModel.T2_MAX_ACCT_BAL {
			var note accountModel.Note
			note.Action = "Frozen KYC Tier 2 Account"
			note.Reason = "Performed a transaction that failed to comply with our KYC Level 2 requirements"
			err := account.FreezeAccount(note)
			if err != nil {
				return accountModel.Accounts{}, err
			}
			return account, nil
		}
	}
	return account, nil
}

// KycUseCase handles business logic
func ValidateDebitKyc(account accountModel.Accounts, amount int64) (accountModel.Accounts, error) {
	log.Println("------validating predebit kyc----------")
	newbal := account.AvailableBalance
	//get the account daily transactions before doing the check
	DailyTrans, err := transactionUsecase.GetAccountDayTransactions(account)
	log.Println("Total trans for the day:", len(DailyTrans))
	if err != nil {
		return accountModel.Accounts{}, err
	}
	//calculate total day
	totalDailyTrans := transactionUsecase.GetTotalAmount(DailyTrans) + amount
	switch account.Customer.KYCTier {
	//kyc zero should not do transfers yet
	case "0":
		return accountModel.Accounts{}, errors.New("please upgrade your account to Tier 1 and try again")

	case "1":
		if amount > kycModel.T1_MAX_DEBIT_AMT || newbal > kycModel.T1_MAX_ACCT_BAL || totalDailyTrans > kycModel.T1_MAX_DAILY_CUM_DR_AMT {
			return accountModel.Accounts{}, errors.New("please upgrade your account to Tier 2 and try again or check the transaction amount and your limits")
		}

	case "2":
		if amount > kycModel.T2_MAX_DEBIT_AMT || newbal > kycModel.T2_MAX_ACCT_BAL || totalDailyTrans > kycModel.T2_MAX_DAILY_CUM_DR_AMT {
			return accountModel.Accounts{}, errors.New("please upgrade your account to Tier 3 and try again or check the transaction amount and your limits")
		}

	case "3":
		if amount > kycModel.T3_MAX_DEBIT_AMT || totalDailyTrans > kycModel.T3_MAX_DAILY_CUM_DR_AMT {
			return accountModel.Accounts{}, errors.New("please upgrade your account try again or check the transaction amount and your limits")
		}
	}
	return account, nil
}

func UpgradeCustomerToTierOne(customer customerModel.Customer) (customerModel.Customer, error) {
	//call provider and do all bvn check which is the basics for tier one
	if customer.KYCTier == string(customerModel.KYCTierLevel1) {
		return customer, nil
	}
	customer.KYCTier = string(customerModel.KYCTierLevel1)
	//update the customer
	err := customer.UpdateCustomer()
	address, err := customer.GetCustomerAddress()
	if err != nil {
		return customerModel.Customer{}, err
	}
	customer.AddressDetailsId = address.ID
	//delete the address details before updating
	_ = customer.UpdateCustomer()
	if err != nil {
		return customerModel.Customer{}, err
	}
	return customer, nil
}

func UpgradeCustomerToTierTwo(customer customerModel.Customer) (customerModel.Customer, error) {
	//call provider and do all checks for tier zero checks then update the customer
	if customer.KYCTier == string(customerModel.KYCTierLevel2) {
		return customer, nil
	}
	customer.KYCTier = string(customerModel.KYCTierLevel2)
	//update the customer
	err := customer.UpdateCustomer()
	if err != nil {
		return customerModel.Customer{}, err
	}
	return customer, nil
}

func UpgradeCustomerToTierThree(customer customerModel.Customer) (customerModel.Customer, error) {
	//call provider and do all checks for tier zero checks then update the customer
	if customer.KYCTier == string(customerModel.KYCTierLevel3) {
		return customer, nil
	}
	customer.KYCTier = string(customerModel.KYCTierLevel3)
	//update the customer
	err := customer.UpdateCustomer()
	if err != nil {
		return customerModel.Customer{}, err
	}
	return customer, nil
}
