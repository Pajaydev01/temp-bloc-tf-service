package model

import (
	"bloc-mfb/config/database"
	"bloc-mfb/utils/exception"
	"time"

	"github.com/go-ozzo/ozzo-validation/is"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type KYCTier string // Define Enum as an integer type

const (
	KYCTierUnverified KYCTier = "0" // Default
	KYCTierLevel1     KYCTier = "1"
	KYCTierLevel2     KYCTier = "2"
	KYCTierLevel3     KYCTier = "3"
)

type Customer struct {
	ID                       uint           `gorm:"primaryKey" json:"id" bson:"id"`
	FullName                 *string        `json:"full_name" bson:"full_name"`
	PhoneNumber              *string        `gorm:"unique" json:"phone_number" bson:"phone_number"`
	Environment              string         `gorm:"type:enum('production', 'sandbox');default:'sandbox'" json:"environment" bson:"environment"`
	Email                    string         `gorm:"unique" json:"email" bson:"email"`
	Country                  string         `json:"country,omitempty"`
	Group                    string         `json:"group,omitempty"`
	Status                   string         `json:"status,omitempty"`
	CreatedAt                time.Time      `json:"created_at" bson:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at" bson:"updated_at"`
	IsDeleted                bool           `json:"is_deleted" bson:"is_deleted"`
	FirstName                string         `json:"first_name" bson:"first_name"`
	LastName                 string         `json:"last_name" bson:"last_name"`
	KYCTier                  string         `gorm:"type:ENUM('0', '1', '2', '3');default:0;not null" json:"kyc_tier"`
	KYCStatus                string         `json:"kyc_status,omitempty" bson:"kyc_status"`
	KYCVerificationSessionID string         `json:"kyc_verification_session_id,,omitempty" bson:"kyc_verification_session_id"`
	BVN                      string         `gorm:"unique" json:"-"`
	NIN                      string         `json:"nin,omitempty" bson:"nin"`
	PlaceOfBirth             string         `json:"place_of_birth,omitempty" bson:"place_of_birth"`
	Gender                   string         `json:"gender,omitempty" bson:"gender"`
	IdentityUrl              string         `json:"identity_url,omitempty" bson:"identity_url"`
	ImageUrl                 string         `json:"image_url,omitempty" bson:"image_url"`
	MeansOfIdentity          string         `json:"means_of_identity,omitempty" bson:"means_of_identity"`
	DateOfBirth              time.Time      `json:"date_of_birth,omitempty" time_format:"2006-01-02" `
	MeansOfIdentityUrl       string         `json:"means_of_identity_url,omitempty" bson:"means_of_identity_url"`
	Archived                 bool           `json:"-" bson:"archived"`
	CustomerType             string         `json:"customer_type,omitempty" bson:"customer_type"`
	Source                   string         `json:"source,omitempty"`
	Migrated                 bool           `json:"migrated,omitempty"`
	AddressDetails           AddressDetails `gorm:"foreign-key:AddressDetailsId" json:"address_details,omitempty"`
	KycDocumentUrl           string         `json:"-" bson:"kyc_document_url"`
	WhatsAppNumber           string         `json:"whatsapp_number,omitempty" bson:"whatsapp_number"`
	AddressDetailsId         uint           `gorm:"unique" json:"-"`
}

type AddressDetails struct {
	ID         uint   `gorm:"primaryKey" json:"-" bson:"id"`
	Street     string `json:"street,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	Country    string `json:"country,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	CustomerId uint   `gorm:"unique" json:"-"`
}

type TierOneUpgradeRequest struct {
	Gender         string         `json:"gender"`
	AddressDetails AddressDetails ` json:"address_details,omitempty"`
	PlaceOfBirth   string         `json:"place_of_birth"`
	Country        string         `json:"country" bson:"country"`
	DOB            string         `json:"dob"`
}

func Init() {
	//migrate all models
	database.GetDB().AutoMigrate(&Customer{})
	database.GetDB().AutoMigrate(&AddressDetails{})

	//pass migrations
	database.GetDB().Migrator().CreateIndex(&Customer{}, "Email")
	database.GetDB().Migrator().CreateIndex(&Customer{}, "PhoneNumber")
	database.GetDB().Migrator().AlterColumn(&Customer{}, "KYCTIER")
	database.GetDB().Migrator().AlterColumn(&Customer{}, "BVN")
}

func (c Customer) UpdateCustomer() error {
	save := database.GetDB().Save(&c)
	if save.Error != nil {
		return exception.HandleDBError(save.Error)
	}
	return nil
}

func (c Customer) GetCustomerAddress() (AddressDetails, error) {
	save := database.GetDB().Preload("AddressDetails").Where(&Customer{ID: c.ID}).First(&c)
	if save.Error != nil {
		return AddressDetails{}, exception.HandleDBError(save.Error)
	}
	return c.AddressDetails, nil
}

func (c Customer) ValidateCreateCustomer() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.FirstName, validation.Required),
		validation.Field(&c.LastName, validation.Required),
		validation.Field(&c.Email, validation.Required),
		validation.Field(&c.PhoneNumber, validation.Required, is.Digit),
		validation.Field(&c.BVN, validation.Required),
		validation.Field(&c.Environment, validation.When(c.Environment != "", validation.In("production", "sandbox").Error("Environment must be one of either sandbox or production"))),
	)
}

func (c TierOneUpgradeRequest) ValidateCreateCustomerUpgradeTierOne() error {
	return validation.ValidateStruct(
		&c,
		validation.Field(&c.Gender, validation.Required),
		validation.Field(&c.AddressDetails, validation.Required),
		validation.Field(&c.PlaceOfBirth, validation.Required),
		validation.Field(&c.Country, validation.Required),
		validation.Field(&c.DOB, validation.Required, validation.Date("2006-01-02")),
	)
}
