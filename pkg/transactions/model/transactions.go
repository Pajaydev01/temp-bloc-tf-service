package model

import (
	"bloc-mfb/config/database"
	"bloc-mfb/utils/exception"
	"time"
)

type transaction_status string

const (
	TRANSACTION_STATUS_PENDING transaction_status = "pending"
	TRANSACTION_STATUS_SUCCESS transaction_status = "successful"
	TRANSACTION_STATUS_FAILED  transaction_status = "failed"
)

type Transactions struct {
	ID                    uint      `bson:"_id,omitempty" json:"id,omitempty"`
	CreatedAt             time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt             time.Time `bson:"updated_at" json:"updated_at"`
	Amount                int64     `gorm:"not null" bson:"amount,omitempty" json:"amount,omitempty"`
	AccountNumber         string    `json:"account_number," bson:"account_number,omitempty"`
	Reference             string    `gorm:"not null;uniqueIndex" bson:"reference" json:"reference"`
	Status                string    `gorm:"type:enum('pending', 'failed','successful');default:'successful'" bson:"status,omitempty" json:"status,omitempty"`
	Shared                bool      `json:"-" bson:"shared"`
	Currency              string    `gorm:"not null" bson:"currency,omitempty" json:"currency,omitempty"`
	Environment           string    `gorm:"type:enum('production', 'sandbox');default:'sandbox'" json:"environment" bson:"environment"` //test or live
	PaymentMethod         string    `json:"payment_method,omitempty" bson:"payment_method,omitempty"`
	ProviderID            uint      `bson:"provider,omitempty" json:"-"`
	Group                 string    `bson:"group,omitempty" json:"group,omitempty"`
	ProviderName          string    `bson:"provider_name,omitempty" json:"provider_name,omitempty"`
	ProviderRef           string    `bson:"provider_ref,omitempty" json:"provider_ref,omitempty"`
	PaymentType           string    `json:"payment_type,omitempty" json:"payment_type,omitempty"`
	Source                string    `json:"source"`
	Fee                   int64     `bson:"fee"`
	Reversal              bool      `bson:"reversal" json:"reversal"`
	ReversedTransactionID uint      `bson:"reversed_transaction_id,omitempty" json:"reversed_transaction_id,omitempty"`
	Narration             string    `json:"narration,omitempty"`
	DRCR                  string    `gorm:"type:enum('dr', 'cr'); not null" json:"drcr,omitempty"`
	Migrated              bool      `json:"-" bson:"migrated,omitempty"`
	ExchangeRate          float64   `json:"exchange_rate,omitempty" bson:"exchange_rate,omitempty"`
	Category              string    `json:"category,omitempty" bson:"-"`
	Name                  string    `json:"name,omitempty" bson:"-"`
	MetaData              *MetaData `gorm:"foreignKey:MetaDataID" json:"meta_data,omitempty"`
	MetaDataID            *uint     `json:"-"`
}

type MetaData struct {
	ID            uint      `json:"id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     time.Time `json:"deleted_at"`
	TransactionId uint      `json:"transaction_id,omitempty"`
	//nip transactions
	NipRef string `json:"nip_ref,omitempty"`
	//transfers
	ReceiverAccountName   string `json:"receiver_account_name,omitempty"`
	ReceiverAccountNumber string `json:"receiver_account_number,omitempty"`
	ReceiverBankCode      string `json:"receiver_bank_code,omitempty"`
	ReceiverBankName      string `json:"receiver_bank_name,omitempty"`

	//inflow
	SenderAccountName   string `json:"sender_account_name,omitempty"`
	SenderAccountNumber string `json:"sender_account_number,omitempty"`
	SenderBankName      string `json:"sender_bank_name,omitempty"`
	//fee and others
	FeeTransactionId uint `json:"fee_transaction_id,omitempty"`
	VatTransactionId uint `json:"vat_transaction_id,omitempty"`
}

type FeeVatChargeResp struct {
	VatTransactionId uint
	VatTransaction   Transactions
	FeeTransactionId uint
	FeeTransaction   Transactions
}

type PaginatedResponse struct {
	MetaData PaginatedMetaData `json:"metadata"` // Pagination metadata
	Data     []Transactions    `json:"data"`     // Paginated data
}

type PaginatedMetaData struct {
	TotalCount int64 `json:"total_count"` // Total number of records
	Page       int   `json:"page"`        // Current page number
	PageSize   int   `json:"page_size"`   // Number of records per page
	HasNext    bool  `json:"has_next"`    // True if there's another page
}

func Init() {
	database.GetDB().AutoMigrate(&Transactions{})

	// Ensure the unique index is created for the Reference field
	//database.GetDB().Migrator().CreateIndex(&Transactions{}, "Reference")

	database.GetDB().Migrator().AlterColumn(&Transactions{}, "Reference")
}

func (t *Transactions) UpdateTransactionMeta() (*Transactions, error) {
	//save the transaction meta and update the id
	t.MetaDataID = &t.MetaData.ID
	save := database.GetDB().Save(&t)
	if save.Error != nil {
		return nil, exception.HandleDBError(save.Error)
	}
	return t, nil
}
