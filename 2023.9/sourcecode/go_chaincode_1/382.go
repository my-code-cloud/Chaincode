/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

// SmartContract provides functions for managing a airline part
type SmartContract struct {
	contractapi.Contract
}

// Car describes basic details of what makes up a airline part
type Asset struct {
	ProductID   int `json:"productID"`
	Quantity  int `json:"quantity"`
	Owner  string `json:"owner"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Asset
}

// InitLedger adds a base set of airline parts to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		Asset{ProductID: 0, Quantity: 10, Owner: "United Airlines"},
		Asset{ProductID: 1, Quantity: 5, Owner: "Delta"},
		Asset{ProductID: 2, Quantity: 3, Owner: "Spirit"},
		Asset{ProductID: 3, Quantity: 7, Owner: "Frontier"},
		Asset{ProductID: 4, Quantity: 9, Owner: "Alaska Airlines"},
		Asset{ProductID: 5, Quantity: 6, Owner: "Southwest Airlines"},
		Asset{ProductID: 6, Quantity: 12, Owner: "JetBlue"},
		Asset{ProductID: 7, Quantity: 3, Owner: "Hawaiian Airlines"},
		Asset{ProductID: 8, Quantity: 3, Owner: "Allegiant Air"},
		Asset{ProductID: 9, Quantity: 7, Owner: "Boeing"},
	}

	for i, parts := range assets {
		partsAsBytes, _ := json.Marshal(parts)
		err := ctx.GetStub().PutState("PART"+strconv.Itoa(i), partsAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

// CreateAirlinePart adds a new airline part to the world state with given details
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, airlinePartNumber string, productID int, quantity int, owner string) (string, error) {
	asset := Asset{
		ProductID:   productID,
		Quantity: quantity,
		Owner:  owner,
	}

	airlinePartAsBytes, err := json.Marshal(asset)

	if err != nil {
    	return "", err
	}

	message := "{\"message\": \"addProduct\", \"productID\": " + strconv.Itoa(asset.ProductID) + ", \"quantity\": " + strconv.Itoa(asset.Quantity) + "}\n"

	return message, ctx.GetStub().PutState(airlinePartNumber, airlinePartAsBytes)
}

// QueryAirlinePart returns the airline part stored in the world state with given id
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, airlinePartNumber string) (*Asset, error) {
	airlinePartAsBytes, err := ctx.GetStub().GetState(airlinePartNumber)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if airlinePartAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", airlinePartNumber)
	}

	airlinePart := new(Asset)
	_ = json.Unmarshal(airlinePartAsBytes, airlinePart)

	return airlinePart, nil
}

func (s *SmartContract) QueryAllParts(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""

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

		asset := new(Asset)
		_ = json.Unmarshal(queryResponse.Value, asset)

		queryResult := QueryResult{Key: queryResponse.Key, Record: asset}
		results = append(results, queryResult)
	}

	return results, nil
}

// ChangeAirlinePartOwner updates the owner field of airline part with given id in world state
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, airlinePartNumber string, newOwner string) (string, error) {
	asset, err := s.ReadAsset(ctx, airlinePartNumber)

	if err != nil {
		return "", err
	}

	asset.Owner = newOwner

	airlinePartAsBytes, _ := json.Marshal(asset)

	message := "{\"message\": \"sellProduct\", \"productID\": " + strconv.Itoa(asset.ProductID) + ", \"quantity\": " + strconv.Itoa(asset.Quantity) + ", \"buyer\": " + asset.Owner + "}\n"

	return message, ctx.GetStub().PutState(airlinePartNumber, airlinePartAsBytes)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}

