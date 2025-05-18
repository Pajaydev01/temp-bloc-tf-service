package model

import (
	"bloc-mfb/config/database"
	customerModel "bloc-mfb/pkg/customer/model"
	txnModel "bloc-mfb/pkg/transactions/model"
	"bloc-mfb/utils/exception"
	global "bloc-mfb/utils/state"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gorm.io/gorm"
)

type account_status string
type debit_credit string

const (
	ACCOUNT_STATUS_OPEN   account_status = "Open"   // Default
	ACCOUNT_STATUS_FROZEN account_status = "Frozen" //frozen
	ACCOUNT_STATUS_CLOSED account_status = "Closed" //closed
	ACCOUNT_CREDIT        debit_credit   = "credit"
	ACCOUNT_DEBIT         debit_credit   = "debit"
)

type Accounts struct {
	ID                uint                   `gorm:"primaryKey" json:"id" bson:"id"`
	Name              string                 `json:"name" bson:"name"`
	BVN               *string                `json:"bvn" bson:"bvn"`
	KYCTIER           string                 `gorm:"type:ENUM('0', '1', '2', '3');default:0;not null" json:"kyc_tier"`
	CreatedAt         time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" bson:"updated_at"`
	Status            string                 `gorm:"type:ENUM('Frozen', 'Open','Closed');default:'Open'"  json:"status"`
	Environment       string                 `gorm:"type:enum('production', 'sandbox');default:'sandbox'" json:"environment" bson:"environment"`
	Balance           int64                  `json:"balance" bson:"balance"`
	AvailableBalance  int64                  `json:"available_balance" bson:"available_balance"`
	Currency          *string                `json:"currency" bson:"currency"`
	Frequency         int64                  `json:"frequency" bson:"frequency"`
	Hold              *int64                 `json:"hold,omitempty" bson:"hold"`
	Locked            sql.NullBool           `json:"-" bson:"locked"`
	MetaDataId        uint                   `json:"-"`
	CustomerID        uint                   `gorm:"unique" json:"customer_id" bson:"customer_id"`
	AccountNumber     string                 `json:"account_number" bson:"account_number"`
	BankName          string                 `json:"bank_name" bson:"bank_name"`
	Type              string                 `json:"type,omitempty" bson:"type"`
	Alias             *string                `json:"alias,omitempty" bson:"alias"`
	InstantSettlement sql.NullBool           `json:"-" bson:"instant_settlement"`
	AutoSettle        sql.NullBool           `json:"-" bson:"auto_settle"`
	Customer          customerModel.Customer `gorm:"foreign-key:CustomerID"`
}

type Note struct {
	ID     uint   `gorm:"primaryKey" json:"id" bson:"id"`
	Action string `json:"action" bson:"action"`

	Reason    string    `json:"reason" bson:"reason"`
	AccountID uint      `json:"account_id"`
	Account   Accounts  `gorm:"foreignKey:AccountID"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

func Init() {
	database.GetDB().AutoMigrate(&Accounts{})
	database.GetDB().AutoMigrate(&Note{})

	//changes to db object can come here, please if any column is altered, please remember to update the model
	database.GetDB().Migrator().AlterColumn(&Accounts{}, "KYCTIER")
	database.GetDB().Migrator().AlterColumn(&Accounts{}, "Environment")
	database.GetDB().Migrator().CreateIndex(&Accounts{}, "CustomerID")
	database.GetDB().Migrator().CreateIndex(&Accounts{}, "Status")
}

// this is a default hooks called before creating anything into the account, it ensures debit and credit is coming from a single source
func (a *Accounts) BeforeSave(tx *gorm.DB) (err error) {
	if a.Locked.Bool || a.Locked.Valid {
		if global.GetMasterToUpdate() == "master" {
			return nil
		}
		return errors.New("account locked")
	}
	return nil
}

func (a Accounts) ValidateCreateAccount() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.CustomerID, validation.Required),
	)
}

func intToSlice(n int64, sequence []int64) []int64 {
	if n != 0 {
		i := n % 10
		sequence = append([]int64{i}, sequence...)
		return intToSlice(n/10, sequence)
	}
	return sequence
}

func (a *Accounts) GenerateBlocAccount(sequence int64) {
	log.Println("=============== creating sample nuban account number ===============")
	accountSeries := intToSlice(sequence, []int64{})
	ofiCode := []int64{9, 5, 0, 1, 1, 7}

	ofiCode = append(ofiCode, accountSeries...)

	checkSumNumbers := []int64{3, 7, 3, 3, 7, 3, 3, 7, 3, 3, 7, 3, 3, 7, 3}
	var sum int64
	var checkSum int64
	for index, code := range ofiCode {
		sum += code * checkSumNumbers[index]
	}
	delta := sum % 10

	if delta == 0 {
		checkSum = 0
	} else {
		checkSum = 10 - delta
	}

	accountNumber := fmt.Sprintf("%v%v", sequence, checkSum)

	// if os.Getenv("ENVIRONMENT") == "staging" && a.Environment == "live" {
	// 	a.AccountNumber = "2000000000"
	// 	a.BankName = "Bloc MFB"
	// }
	//before calling nibbs, check if the account created already exist
	check := database.GetDB().Where(&Accounts{AccountNumber: accountNumber}).First(&a)
	if check.Error == nil {
		//generate again till something different
		a.GenerateBlocAccount(sequence)
		return
	}
	a.AccountNumber = accountNumber
	a.BankName = "Bloc MFB"
	log.Println("=============== Bloc account generated ===============")
}

func (a *Accounts) LockAccount() error {
	a.Locked = sql.NullBool{Bool: true, Valid: true}
	global.SetMasterToUpdate("master")
	update := database.GetDB().Save(&a)
	if update.Error != nil {
		global.SetMasterToUpdate("")
		return exception.HandleDBError(update.Error)
	}
	global.SetMasterToUpdate("")
	return nil
}

func (a *Accounts) UnlockAccount() error {
	a.Locked = sql.NullBool{Bool: false, Valid: false}
	update := database.GetDB().Save(&a)
	global.SetMasterToUpdate("")
	if update.Error != nil {
		return exception.HandleDBError(update.Error)
	}
	return nil
}

func (a *Accounts) FreezeAccount(note Note) error {
	//update the account status and the note
	note.AccountID = a.ID
	create := database.GetDB().Create(&note)
	if create.Error != nil {
		return exception.HandleDBError(create.Error)
	}
	a.Status = string(ACCOUNT_STATUS_FROZEN)
	update := database.GetDB().Save(&a)
	if update.Error != nil {
		return exception.HandleDBError(update.Error)
	}
	//send notifications and webhook
	return nil
}

func (a *Accounts) UnfreezeAccount(note Note) error {
	//update the account status and the note
	note.AccountID = a.ID
	create := database.GetDB().Create(&note)
	if create.Error != nil {
		return exception.HandleDBError(create.Error)
	}
	a.Status = string(ACCOUNT_STATUS_OPEN)
	update := database.GetDB().Save(&a)
	if update.Error != nil {
		return exception.HandleDBError(update.Error)
	}
	//send notifications and webhook
	return nil
}

func (a *Accounts) CloseAccount(note Note) error {
	//update the account status and the note
	note.AccountID = a.ID
	create := database.GetDB().Create(&note)
	if create.Error != nil {
		return exception.HandleDBError(create.Error)
	}
	a.Status = string(ACCOUNT_STATUS_CLOSED)
	update := database.GetDB().Save(&a)
	if update.Error != nil {
		return exception.HandleDBError(update.Error)
	}
	//send notifications and
	return nil
}

func (a *Accounts) AccountChecks(amount int64) error {
	//first check for sufficient balance
	if amount > a.AvailableBalance {
		return errors.New("insufficient balance")
	}
	if a.AvailableBalance < 0 {
		return errors.New("account can not perform debit at this time")
	}
	if amount < 0 {
		return errors.New("invalid amount")
	}
	if a.Status != string(ACCOUNT_STATUS_OPEN) {
		return errors.New(string(a.Status))
	}
	return nil
	//return ValidateDebitKyc(a, amount)
	//there will be a fraud checker algorithm after this, subject the account to it
}

// this is the only funtion that should debit an account, if the db is called directly, it won't be honored
func (a *Accounts) DebitAccount(amount int64) (old_balanace int64, err error) {
	if err := a.AccountChecks(amount); err != nil {
		return 0, err
	}
	//account is locked while debiting, this is to ensure no other transaction is happening at the same time, and only allows this
	global.SetMasterToUpdate("master")
	newBal := a.Balance - amount
	newAvailableBal := a.AvailableBalance - amount
	log.Println("new balance", newBal)
	balance_before := a.Balance
	a.Balance = newBal
	a.AvailableBalance = newAvailableBal

	update := database.GetDB().Save(&a)
	if update.Error != nil {
		return balance_before, exception.HandleDBError(update.Error)
	}
	return balance_before, nil
}

func (a *Accounts) CreditAccount(amount int64) (old_balance int64, error error) {
	global.SetMasterToUpdate("master")
	newBal := a.Balance + amount
	newAvailableBal := a.AvailableBalance + amount
	frequency := a.Frequency + 1
	balance_before := a.Balance
	a.Balance = newBal
	a.AvailableBalance = newAvailableBal
	a.Frequency = frequency

	update := database.GetDB().Save(&a)
	if update.Error != nil {
		return balance_before, exception.HandleDBError(update.Error)
	}
	return balance_before, nil
}

func (a *Accounts) ChargeFeeAndVat(fee int64, vat int64) (balance_before int64, error error) {
	amount := fee + vat
	balance_before, err := a.DebitAccount(amount)
	if err != nil {
		return balance_before, err
	}
	return balance_before, nil
}

func (a *Accounts) ReverseDebit(transaction txnModel.Transactions) (balance_before int64, error error) {
	//credit account back with the transaction amount
	balance_before, err := a.CreditAccount(transaction.Amount)
	if err != nil {
		return balance_before, err
	}
	return balance_before, nil
}
