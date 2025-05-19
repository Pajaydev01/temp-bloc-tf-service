package model

import validation "github.com/go-ozzo/ozzo-validation"

type TransferRequest struct {
	Amount          int64  `json:"amount,omitempty"`
	Currency        string `json:"currency"`
	BankCode        string `json:"bank_code,omitempty"`
	Source          string `json:"-"`
	Method          string `json:"-"`
	AccountNumber   string `json:"account_number,omitempty"`
	SenderName      string `json:"sender_name" bson:"sender_name"`
	AccountCurrency string `json:"-"`
	BeneficiaryID   uint   `json:"beneficiary_id"`
	Fees            int64  `json:"organisation_transfer_fees,omitempty"`
	Reference       string `json:"reference" bson:"reference"`
	Environment     string `json:"environment,omitempty"`
	Narration       string `json:"narration,omitempty"`
	InternalStatus  string `json:"-" bson:"internal_status"`
	SessionID       string `json:"-" bson:"session_id"`
	PurposeCode     string `json:"purpose_code,omitempty"`
	Type            string `json:"-"`
	ApprovalID      uint   `json:"-" bson:"approval_id"`
	RefId           string `json:"ref_id,omitempty"`
	//CustomerId      uint
}

type NameEnquiryRequest struct {
	AccountNumber string `json:"account_number,omitempty"`
	BankCode      string `json:"bank_code,omitempty"`
}

type NESingleRequestEasyPay struct {
	AccountNumber              string `json:"accountNumber"`
	ChannelCode                string `json:"channelCode"`
	DestinationInstitutionCode string `json:"destinationInstitutionCode"`
	TransactionId              string `json:"transactionId"`
}

type NESingleResponseEasyPay struct {
	ResponseCode               string `json:"responseCode"`
	SessionID                  string `json:"sessionID"`
	TransactionId              string `json:"transactionId"`
	ChannelCode                int    `json:"channelCode"`
	DestinationInstitutionCode string `json:"destinationInstitutionCode"`
	AccountNumber              string `json:"accountNumber"`
	AccountName                string `json:"accountName"`
	BankVerificationNumber     string `json:"bankVerificationNumber"`
	KycLevel                   int    `json:"kycLevel"`
	Message                    string `json:"message"`
}

func (a NameEnquiryRequest) ValidateNameEnquiry() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.AccountNumber, validation.Required),
		validation.Field(&a.BankCode, validation.Required),
	)
}

func (a TransferRequest) ValidateTransfer() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Amount, validation.Required),
		validation.Field(&a.RefId, validation.Required),
		validation.Field(&a.Reference, validation.Required),
		validation.Field(&a.SenderName, validation.Required),
	)
}
