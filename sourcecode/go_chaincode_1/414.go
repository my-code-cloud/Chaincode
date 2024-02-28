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

type ExportCert struct {
	CertID string `json:"certid"`
	Time   string `json:"time"`
	Hash  string `json:"hash"`
}

type QueryResult struct {
	Key    string `json:"Key"`
	Record *ExportCert
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	certs := []ExportCert{
		//ExportCert{CertID: "imp1", ExporterName: "exp1",
	}

	for i, cert := range certs {
		certAsBytes, _ := json.Marshal(cert)
		err := ctx.GetStub().PutState("CERT"+strconv.Itoa(i), certAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

func (s *SmartContract) CreateCert(ctx contractapi.TransactionContextInterface, certNumber string, certid string, time string, hash string) error {
	cert := ExportCert{
		CertID: certid,
		Time: time,
		Hash: hash,
	}

	certAsBytes, _ := json.Marshal(cert)

	return ctx.GetStub().PutState(certNumber, certAsBytes)
}

func (s *SmartContract) QueryCert(ctx contractapi.TransactionContextInterface, certNumber string) (*ExportCert, error) {
	certAsBytes, err := ctx.GetStub().GetState(certNumber)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if certAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", certNumber)
	}

	cert := new(ExportCert)
	_ = json.Unmarshal(certAsBytes, cert)

	return cert, nil
}

func (s *SmartContract) QueryAllCerts(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := "CERT0"
	endKey := "CERT999"

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

		cert := new(ExportCert)
		_ = json.Unmarshal(queryResponse.Value, cert)

		queryResult := QueryResult{Key: queryResponse.Key, Record: cert}
		results = append(results, queryResult)
	}

	return results, nil
}


func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabsc chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabsc chaincode: %s", err.Error())
	}
}
