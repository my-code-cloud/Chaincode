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

type Record struct {
	Reader   string `json:"reader"`
	Url  string `json:"url"`
	DataHash string `json:"DataHash"`
	Owner  string `json:"owner"`
}

type QueryResult struct {
	Key    string `json:"Key"`
	Record *Record
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	records := []Record{
		Record{Reader: "reader1", Url: "url1", DataHash: "hash1", Owner: "owner1"},
		Record{Reader: "reader2", Url: "url2", DataHash: "hash2", Owner: "owner2"},
		Record{Reader: "reader3", Url: "url3", DataHash: "hash3", Owner: "owner3"},
	}

	for i, record := range records {
		recordAsBytes, _ := json.Marshal(record)
		err := ctx.GetStub().PutState("Record"+strconv.Itoa(i), recordAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

func (s *SmartContract) CreateRecord(ctx contractapi.TransactionContextInterface, recordNumber string,reader string, url string, dataHash string, owner string) error {
	record := Record{
		Reader:  reader,
		Url:  url,
		DataHash: dataHash,
		Owner:  owner,
	}

	recordAsBytes, _ := json.Marshal(record)

	return ctx.GetStub().PutState(recordNumber, recordAsBytes)
}

func (s *SmartContract) QueryRecord(ctx contractapi.TransactionContextInterface, recordNumber string) (*Record, error) {
	recordAsBytes, err := ctx.GetStub().GetState(recordNumber)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if recordAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", recordNumber)
	}

	record := new(Record)
	_ = json.Unmarshal(recordAsBytes, record)

	return record, nil
}

func (s *SmartContract) QueryAllRecords(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := "RECORD0"
	endKey := "RECORD99"

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

		record := new(record)
		_ = json.Unmarshal(queryResponse.Value, record)

		queryResult := QueryResult{Key: queryResponse.Key, Record: record}
		results = append(results, queryResult)
	}

	return results, nil
}



func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create mediCo chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting mediCo chaincode: %s", err.Error())
	}
}
