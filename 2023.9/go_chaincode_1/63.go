/*
SPDX-License-Identifier: Apache-2.0
*/
package main

import (
	"encoding/json"
	"fmt"
	
	"bytes"
	"math/big"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	//"github.com/hyperledger/fabric-chaincode-go/shim/ext/cid"
	//"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	
)
type SmartContract struct {
	contractapi.Contract
}
/*-------- record ---------*/
type MedicalRecord struct {
	Name string				`json:"name"`
	In_time string			`json:"in_time"`
	Out_time string			`json:"out_time"`
	Fee string				`json:"fee"`
	Stage       string      `json:"status"`
	CreatedBy   string      `json:"createdBy"`
	MaxStage    string      `json:"maxStage"`
}

/*----------new query---------------*/
type QueryResult struct {
	Key    string `json:"Key"`
	Record *MedicalRecord
}

/*--------------init---------------*/
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	 
	 
	
	
	
	var newrecord = MedicalRecord{Name: "jack", In_time: "2000.01.01",Out_time:"2000.02.02",Fee:"500", Stage: "0", CreatedBy: "admin", MaxStage: "3"}
	recordAsBytes, _ := json.Marshal(newrecord)
	err := ctx.GetStub().PutState("1", recordAsBytes)
	
	
	
	
	if err != nil {
		return fmt.Errorf("Failed to put to world state. %s", err.Error())
	}
	startingId := new(big.Int).SetInt64(2)
	startingIdAsByte := startingId.Bytes()
	err = ctx.GetStub().PutState("MaxRecordId", startingIdAsByte) //set starting id as 2
	if err != nil {
		return fmt.Errorf("Failed to put to world state. %s", err.Error())
	}
	return nil
}
/*--------------create-----------------*/
func (s *SmartContract) CreateRecord(ctx contractapi.TransactionContextInterface, name string, in_time string, out_time string, fee string,maxStage string) error {
	
	
	

	
	var APIstub = ctx.GetStub()
	var cid = ctx.GetClientIdentity()
	canCreate, found, err := cid.GetAttributeValue("canCreate")
	if err != nil {
		return fmt.Errorf("Error when getting user rights")
	}
	if !found {
		return fmt.Errorf("User does not have right to perform this action.")
	}
	if canCreate != "true" {
		return fmt.Errorf("User does not have right to perform this action.")
	}

	username, found, _ := cid.GetAttributeValue("username")
	var newrecord = MedicalRecord{Name: name, In_time: in_time, Out_time: out_time, Fee: fee, Stage: "0", CreatedBy: username, MaxStage: maxStage}

	recordIdAsBytes, _ := APIstub.GetState("MaxRecordId")
	recordId := new(big.Int).SetBytes(recordIdAsBytes)
	recordAsString := recordId.String()

	recordAsBytes, _ := json.Marshal(newrecord)

	APIstub.PutState(recordAsString, recordAsBytes)

	increment := new(big.Int).SetInt64(1)
	newRecordId := new(big.Int).Add(recordId, increment)
	APIstub.PutState("MaxRecordId", newRecordId.Bytes())

	return nil

}
/*-----------------query--------------------------*/
func (s *SmartContract) QueryRecord(ctx contractapi.TransactionContextInterface, recordId string) (*MedicalRecord, error) {

	recordAsBytes, err := ctx.GetStub().GetState(recordId)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if recordAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", recordId)
	}

	mrecord := new(MedicalRecord)
	_ = json.Unmarshal(recordAsBytes, mrecord)

	return mrecord, nil
}


func (s *SmartContract) QueryAllRecords(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("0", "999999999")
	if err != nil {
		return nil, err
	}

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		medicalrecord := new(MedicalRecord)
		_ = json.Unmarshal(queryResponse.Value, medicalrecord)

		
		queryResult := QueryResult{Key: queryResponse.Key, Record: medicalrecord}
		results = append(results, queryResult)
	}
	// Task 3
	return results, nil
}

/*--------delete-------------*/
func (s *SmartContract) DeleteARecords(ctx contractapi.TransactionContextInterface, recordId string) error{

	return ctx.GetStub().DelState(recordId)
}

func buildJSON(resultsIterator shim.StateQueryIteratorInterface, buffer bytes.Buffer) bytes.Buffer {
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, _ := resultsIterator.Next()
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"ProductId\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}

	buffer.WriteString("]")
	return buffer
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %s", err.Error())
	}
}
