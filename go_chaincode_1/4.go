package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Product struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       float32 `json:"price"`
	Image       string  `json:"image"`
	Stock       int     `json:"stock"`
	Owner       string  `json:"owner"`
	BatchNumber int     `json:"batchnumber"`
	Qrcode      string  `json:"Qrcode"`
	Trace       string  `json:"trace"`
}
type QueryResult struct {
	Key    string `json:"key"`
	Record *Product
}

type Transaction struct {
	CreatedAt time.Time `json:"created_at"`
	From string `json:"from"`
	To string `json:"to"`
	Product string `json:"product"`
	Stock int `json:"stock"`
	Payment float64 `json:"payment"`
}

type QueryTransactionResult struct {
	Key string 'json:"key"'
	Record *Transaction
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	Products := []Product{
		Product{Name: "Coconuts oil", Category: "Prius", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 401, Qrcode: "None", Owner: "Tomoko"},
		Product{Name: "Unga", Category: "Mustang", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 402, Qrcode: "None", Owner: "Brad"},
		Product{Name: "Omo", Category: "Tucson", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 403, Qrcode: "None", Owner: "Jin Soo"},
		Product{Name: "Harpic", Category: "Passat", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 404, Qrcode: "None", Owner: "Max"},
		Product{Name: "Yoghurt", Category: "S", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 405, Qrcode: "None", Owner: "Adriana"},
		Product{Name: "Milk", Category: "205", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 406, Qrcode: "None", Owner: "Michel"},
		Product{Name: "Kiwi", Category: "S22L", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 407, Qrcode: "None", Owner: "Aarav"},
	}

	for i, product := range Products {
		productAsByte, _ := json.Marshal(product)
		err := ctx.GetStub().PutState("Product"+strconv.Itoa(i), productAsByte)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}
	return nil
}

func (s *SmartContract) CreateProduct(ctx contractapi.TransactionContextInterface, productNumber string, name string, category string, price float32, image string, stock int32, trace string, batchnumber int32, qrcode string, owner string) error {
	product := Product{
		Name:        name,
		Category:    category,
		Price:       price,
		Image:       image,
		Stock:       stock,
		Owner:       owner,
		BatchNumber: batchnumber,
		Qrcode:      qrcode,
		Trace:       trace,
	}

	productAsByte, _ := json.Marshal(product)
	return ctx.GetStub().PutState(productNumber, productAsByte)

}

func (s *SmartContract) QueryProduct(ctx contractapi.TransactionContextInterface, productNumber string) (*product, error) {
	productAsBytes, err := ctx.GetStub().GetState(productNumber)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if productAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", productNumber)
	}

	product := new(product)
	_ = json.Unmarshal(productAsBytes, product)

	return product, nil
}

func (s *SmartContract) QueryAllProducts(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := "PRODUCT0"
	endKey := "PRODUCT99"

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		product := new(Product)
		_ = json.Unmarshal(queryResponse.Value, product)

		queryResult := QueryResult{Key: queryResponse.Key, Record: product}
		results = append(results, queryResult)
	}

	return results, nil
}

func (s *SmartContract) ChangeProductOwner(ctx contractapi.TransactionContextInterface, productNumber string, newOwner string) error {
	product, err := s.QueryProduct(ctx, productNumber)

	if err != nil {
		return err
	}

	product.Owner = newOwner

	productAsBytes, _ := json.Marshal(product)

	return ctx.GetStub().PutState(productNumber, productAsBytes)
}
func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create pickngo chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting pickngo chaincode: %s", err.Error())
	}
}
