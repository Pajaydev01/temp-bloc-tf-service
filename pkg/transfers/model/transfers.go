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

type EASYPAYFtSingleResponse struct {
	ResponseCode                      string `json:"responseCode"`
	SessionID                         string `json:"sessionID"`
	TransactionId                     string `json:"transactionId"`
	ChannelCode                       string `json:"channelCode"`
	NameEnquiryRef                    string `json:"nameEnquiryRef"`
	DestinationInstitutionCode        string `json:"destinationInstitutionCode"`
	BeneficiaryAccountName            string `json:"beneficiaryAccountName"`
	BeneficiaryAccountNumber          string `json:"beneficiaryAccountNumber"`
	BeneficiaryKYCLevel               string `json:"beneficiaryKYCLevel"`
	BeneficiaryBankVerificationNumber string `json:"beneficiaryBankVerificationNumber"`
	OriginatorAccountName             string `json:"originatorAccountName"`
	OriginatorAccountNumber           string `json:"originatorAccountNumber"`
	OriginatorBankVerificationNumber  string `json:"originatorBankVerificationNumber"`
	OriginatorKYCLevel                string `json:"originatorKYCLevel"`
	TransactionLocation               string `json:"transactionLocation"`
	Narration                         string `json:"narration"`
	PaymentReference                  string `json:"paymentReference"`
	Amount                            string `json:"amount"`
}

type TSQuerySingleResponse struct {
	SourceInstitutionCode string `json:"sourceInstitutionCode"`
	ChannelCode           string `json:"channelCode"`
	SessionID             string `json:"sessionID"`
	ResponseCode          string `json:"responseCode"`
	TransactionId         string `json:"transactionId"`
}

type FTSingleCreditOutward struct {
	Amount                            string `json:"amount"`
	BeneficiaryAccountName            string `json:"beneficiaryAccountName"`
	BeneficiaryAccountNumber          string `json:"beneficiaryAccountNumber"`
	BeneficiaryBankVerificationNumber string `json:"beneficiaryBankVerificationNumber"`
	BeneficiaryKYCLevel               string `json:"beneficiaryKYCLevel"`
	ChannelCode                       string `json:"channelCode"`
	OriginatorAccountName             string `json:"originatorAccountName"`
	OriginatorAccountNumber           string `json:"originatorAccountNumber"`
	OriginatorKYCLevel                string `json:"originatorKYCLevel"`
	MandateReferenceNumber            string `json:"mandateReferenceNumber"`
	PaymentReference                  string `json:"paymentReference"`
	TransactionLocation               string `json:"transactionLocation"`
	OriginatorNarration               string `json:"originatorNarration"`
	BeneficiaryNarration              string `json:"beneficiaryNarration"`
	BillerId                          string `json:"billerId"`
	DestinationInstitutionCode        string `json:"destinationInstitutionCode"`
	SourceInstitutionCode             string `json:"sourceInstitutionCode"`
	TransactionId                     string `json:"transactionId"`
	OriginatorBankVerificationNumber  string `json:"originatorBankVerificationNumber"`
	NameEnquiryRef                    string `json:"nameEnquiryRef"`
	InitiatorAccountName              string `json:"InitiatorAccountName"`
	InitiatorAccountNumber            string `json:"InitiatorAccountNumber"`
}

type NIPInstitutions struct {
	InstitutionCode string `json:"institutionCode"`
	InstitutionName string `json:"institutionName"`
	Category        int    `json:"category"`
	CategoryCode    string `json:"categoryCode"`
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
