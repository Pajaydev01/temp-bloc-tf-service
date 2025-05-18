package usecase

import (
	"bloc-mfb/config/database"
	"bloc-mfb/pkg/customer/model"
	kycUsecase "bloc-mfb/pkg/kyc/usecase"
	"bloc-mfb/utils/exception"

	"gorm.io/gorm/clause"
)

// CustomerUseCase handles business logic
func CreateCustomer(customer model.Customer) (model.Customer, error) {
	//create a customer here and send back the customer data
	create := database.GetDB().Create(&customer)
	if create.Error != nil {
		return model.Customer{}, create.Error
	}
	return customer, nil
}

func GetCustomerById(id uint) (model.Customer, error) {
	var customer = model.Customer{ID: id}
	result := database.GetDB().Preload(clause.Associations).First(&customer)
	if result.Error != nil {
		return model.Customer{}, exception.HandleDBError(result.Error)
	}
	return customer, nil
}

func GetAllCustomers() ([]model.Customer, error) {
	var customers []model.Customer
	get := database.GetDB().Preload(clause.Associations).Order("ID desc").Find(&customers)
	if get.Error != nil {
		return []model.Customer{}, exception.HandleDBError(get.Error)
	}
	return customers, nil
}

func UpdateCustomerToT1(customer model.Customer) (model.Customer, error) {
	//this should be the first update, just call to update to kyc tier
	customer, err := kycUsecase.UpgradeCustomerToTierOne(customer)
	if err != nil {
		return model.Customer{}, err
	}
	return customer, nil
}
