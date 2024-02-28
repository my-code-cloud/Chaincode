package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// / SmartContract provides functions for managing food supply chain
type SmartContract struct {
	contractapi.Contract
}

type User struct {
	Name     string  `json:"Name"`
	UserID   string  `json:"UserID"`
	Email    string  `json:"Email"`
	UserType string  `json:"UserType"`
	Balance  float64 `json:"Balance"`
}

// FoodProduct describes basic details of food supply chain management
type Product struct {
	ID            string  `json:"Id"`
	Name          string  `json:"Name"`
	Owner         string  `json:"Owner"`
	ProductType   string  `json:"ProductType"`
	DNA           string  `json:"DNA"`
	Status        string  `json:"Status"`
	Weight        float64 `"json:Weight"`
	Price         float64 `"json:Price"`
	ShippingPrice float64 `"json:ShippingPrice"`
}

func main() {
	Chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Print("Error creating chaincode")
	}
	if err := Chaincode.Start(); err != nil {
		fmt.Print("Error starting chaincode")
	}
}

func (s *SmartContract) CreatePackage(ctx contractapi.TransactionContextInterface, id string, name string, owner string, productType string, dna string, status string, weight float64, price float64, shippingPrice float64) error {
	exists, err := s.IDExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the product %s already exists", id)
	}
	product := Product{
		ID:            id,
		Name:          name,
		Owner:         owner,
		ProductType:   productType,
		DNA:           dna,
		Status:        status,
		Weight:        weight,
		Price:         price,
		ShippingPrice: shippingPrice,
	}
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, productJSON)
}

func (s *SmartContract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]*User, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all users in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var users []*User
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var user User
		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (s *SmartContract) GetAllProducts(ctx contractapi.TransactionContextInterface) ([]*Product, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all products in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var products []*Product
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var product Product
		err = json.Unmarshal(queryResponse.Value, &product)
		if err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, nil
}

func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, name string, userID string, email string, userType string, balance float64) error {
	exists, err := s.IDExists(ctx, userID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the product %s already exists", userID)
	}
	product := User{
		Name:     name,
		UserID:   userID,
		Email:    email,
		UserType: userType,
		Balance:  balance,
	}
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(userID, productJSON)
}

func (s *SmartContract) IDExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return productJSON != nil, nil
}

func (s *SmartContract) ReadUser(ctx contractapi.TransactionContextInterface, id string) (*User, error) {
	userJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		return nil, fmt.Errorf("the user %s does not exist", id)
	}

	var user User
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *SmartContract) ReadProduct(ctx contractapi.TransactionContextInterface, id string) (*Product, error) {
	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if productJSON == nil {
		return nil, fmt.Errorf("the product %s does not exist", id)
	}

	var product Product
	err = json.Unmarshal(productJSON, &product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *SmartContract) DepositMoneyForSeed(ctx contractapi.TransactionContextInterface, userID string, value float64) error {
	user, err := s.ReadUser(ctx, userID)
	if err != nil {
		return err
	}
	user.Balance -= value
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(userID, userJSON)
}

func (s *SmartContract) GenomicMarkerDataSequenceOriginal(ctx contractapi.TransactionContextInterface, productID string, dna string) error {
	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}
	product.DNA = dna
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(productID, productJSON)

}

func (s *SmartContract) StartSeedShipment(ctx contractapi.TransactionContextInterface, productID string, shipowner string) error {
	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}
	product.Status = "Shipping"
	product.Owner = shipowner
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(productID, productJSON)
}

func (s *SmartContract) UnlockSeedShipment(ctx contractapi.TransactionContextInterface, productID string, newowner string) error {
	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}
	product.Status = "Arriverd"
	s.GetSeedShipmentMoney(ctx, product.Owner, product.ShippingPrice)
	product.Owner = newowner
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(productID, productJSON)
}

func (s *SmartContract) GenomicMarkerDataSequenceReceived(ctx contractapi.TransactionContextInterface, productID string, dna string) error {
	product, err := s.ReadProduct(ctx, productID)
	if err != nil {
		return err
	}
	product.DNA = dna
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(productID, productJSON)
}

func (s *SmartContract) GetSeedShipmentMoney(ctx contractapi.TransactionContextInterface, userID string, value float64) error {
	user, err := s.ReadUser(ctx, userID)
	if err != nil {
		return err
	}
	user.Balance += value
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(userID, userJSON)
}
func (s *SmartContract) Refund(ctx contractapi.TransactionContextInterface, userID string, value float64) error {
	user, err := s.ReadUser(ctx, userID)
	if err != nil {
		return err
	}
	user.Balance += value
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(userID, userJSON)
}
