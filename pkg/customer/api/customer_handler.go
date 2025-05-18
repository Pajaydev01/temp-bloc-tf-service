package api

import (
	"bloc-mfb/pkg/customer/model"
	"bloc-mfb/pkg/customer/usecase"
	req "bloc-mfb/utils/http"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// CustomerHandler handles HTTP requests
func CreateCustomer(w http.ResponseWriter, r *http.Request) {
	customer := model.Customer{}
	err := req.GetRequestBody(r, &customer)
	err = model.Customer.ValidateCreateCustomer(customer)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}
	fullName := customer.FirstName + " " + customer.LastName
	customer.FullName = &fullName
	customer, err = usecase.CreateCustomer(customer)
	if err != nil {
		fmt.Println("error", err)
		req.SendSuccessResponse(w, false, nil, "Unable to create user, one or more record exists", 412)
		return
	}

	//send notification, webhook and all
	req.SendSuccessResponse(w, true, customer, "User created", 201)
}

func UpdateCustomerToT1(w http.ResponseWriter, r *http.Request) {
	vars, err := req.GetRequestParams(r, "customerID")
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(vars, 10, 32)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, "Invalid customer ID", http.StatusBadRequest)
		return
	}
	customer, error := usecase.GetCustomerById(uint(id))
	if error != nil {
		req.SendSuccessResponse(w, false, nil, error.Error(), http.StatusBadRequest)
		return
	}
	request := model.TierOneUpgradeRequest{}
	err = req.GetRequestBody(r, &request)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}

	err = model.TierOneUpgradeRequest.ValidateCreateCustomerUpgradeTierOne(request)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}
	customer.AddressDetails = request.AddressDetails
	customer.Gender = request.Gender
	customer.PlaceOfBirth = request.PlaceOfBirth
	customer.Country = request.Country
	dateOfBirth, err := time.Parse("2006-01-02", request.DOB)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, "Invalid date format", http.StatusBadRequest)
		return
	}
	customer.DateOfBirth = dateOfBirth

	customer, err = usecase.UpdateCustomerToT1(customer)
	if err != nil {
		fmt.Println("error", err)
		req.SendSuccessResponse(w, false, nil, "Unable to create user, one or more record exists", 412)
		return
	}

	//send notification, webhook and all
	req.SendSuccessResponse(w, true, customer, "User account tier upgraded to tier one", 200)
}

func GetAllCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := usecase.GetAllCustomers()
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), 412)
	}
	req.SendSuccessResponse(w, true, customers, "Users retrieved", 200)
}

func GetCustomerById(w http.ResponseWriter, r *http.Request) {
	vars, err := req.GetRequestParams(r, "customerID")
	if err != nil {
		req.SendSuccessResponse(w, false, nil, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(vars, 10, 32)
	if err != nil {
		req.SendSuccessResponse(w, false, nil, "Invalid customer ID", http.StatusBadRequest)
		return
	}
	customer, error := usecase.GetCustomerById(uint(id))
	if error != nil {
		req.SendSuccessResponse(w, false, nil, error.Error(), http.StatusBadRequest)
		return
	}
	req.SendSuccessResponse(w, true, customer, "User retrieved", 200)
}
