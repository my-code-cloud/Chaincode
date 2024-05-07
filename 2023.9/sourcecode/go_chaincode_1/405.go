/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a car
type SmartContract struct {
	contractapi.Contract
}

// Car describes basic details of what makes up a car
type Car struct {
	Name                string `json:"name"`
	Birthday            string `json:"birthday"`
	Vaccine_name        string `json:"vaccine_name"`
	Vaccine_batchNumber string `json:"vaccine_batchNumber"`
	Vaccination_date    string `json:"vaccination_date"`
	Vaccination_org     string `json:"vaccination_org"`
}

type PubKey struct {
	Publickey string `json:"publickey"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Car
}

// InitLedger adds a base set of cars to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	cars := []Car{
		Car{Name: "YU MIN CHIANG", Birthday: "MALE", Vaccine_name: "Moderna COVID-19 Vaccine", Vaccine_batchNumber: "a123456", Vaccination_date: "2021-01-01", Vaccination_org: "gov_taiwan"},
		Car{Name: "XIO HUA WANG", Birthday: "FEMALE", Vaccine_name: "Pfizer-BioNTech COVID-19 Vaccine", Vaccine_batchNumber: "b123456", Vaccination_date: "2021-08-01", Vaccination_org: "gov_taiwan"},
	}

	for i, car := range cars {
		carAsBytes, _ := json.Marshal(car)
		err := ctx.GetStub().PutState("88812345"+strconv.Itoa(i), carAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

// CreateCar adds a new car to the world state with given details
func (s *SmartContract) CreateCar(ctx contractapi.TransactionContextInterface, carNumber string, name string, birthday string, vaccine_name string, vaccine_batchNumber string, vaccination_date string, vaccination_org string) error {
	car := Car{
		Name:                name,
		Birthday:            birthday,
		Vaccine_name:        vaccine_name,
		Vaccine_batchNumber: vaccine_batchNumber,
		Vaccination_date:    vaccination_date,
		Vaccination_org:     vaccination_org,
	}

	carAsBytes, _ := json.Marshal(car)

	return ctx.GetStub().PutState(carNumber, carAsBytes)
}

// CreatePubkey adds a new pubkey to the world state with given details
func (s *SmartContract) CreatePubkey(ctx contractapi.TransactionContextInterface, owner string, publickey string) error {
	pubkey := PubKey{
		Publickey: publickey,
	}

	pubkeyAsBytes, _ := json.Marshal(pubkey)

	return ctx.GetStub().PutState(owner, pubkeyAsBytes)
}

// QueryPubkey returns the pubkey stored in the world state with given id
func (s *SmartContract) QueryPubkey(ctx contractapi.TransactionContextInterface, owner string) (*PubKey, error) {
	pubkeyAsBytes, err := ctx.GetStub().GetState(owner)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if pubkeyAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", owner)
	}

	pubkey := new(PubKey)
	_ = json.Unmarshal(pubkeyAsBytes, pubkey)

	return pubkey, nil
}

// QueryCar returns the car stored in the world state with given id
func (s *SmartContract) QueryCar(ctx contractapi.TransactionContextInterface, carNumber string) (*Car, error) {
	carAsBytes, err := ctx.GetStub().GetState(carNumber)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if carAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", carNumber)
	}

	car := new(Car)
	_ = json.Unmarshal(carAsBytes, car)

	return car, nil
}

// QueryAllCars returns all cars found in world state
func (s *SmartContract) QueryAllCars(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
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

		car := new(Car)
		_ = json.Unmarshal(queryResponse.Value, car)

		queryResult := QueryResult{Key: queryResponse.Key, Record: car}
		results = append(results, queryResult)
	}

	return results, nil
}

// ChangeCarOwner updates the owner field of car with given id in world state
// func (s *SmartContract) ChangeCarOwner(ctx contractapi.TransactionContextInterface, carNumber string, newOwner string) error {
// 	car, err := s.QueryCar(ctx, carNumber)

// 	if err != nil {
// 		return err
// 	}

// 	car.Owner = newOwner

// 	carAsBytes, _ := json.Marshal(car)

// 	return ctx.GetStub().PutState(carNumber, carAsBytes)
// }

// GetHistoryForCar gets the historical data of a car from the ledger
func (s *SmartContract) GetHistoryForCar(ctx contractapi.TransactionContextInterface, carNumber string) (string, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(carNumber)
	if err != nil {
		return "", err
	}
	defer resultsIterator.Close()
	// buffer is a JSON array containing historic values for the car
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}
		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return buffer.String(), nil
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
