package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	mspprotos "github.com/hyperledger/fabric-protos-go/msp"
)

func getTransientFundArrayFromQuery(
	ctx contractapi.TransactionContextInterface,
	queryString string) ([]TransientFund, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	transientFundArr := []TransientFund{}
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		jsonByteObj := queryResponse.Value
		transientFund := TransientFund{}
		json.Unmarshal(jsonByteObj, &transientFund)
		transientFundArr = append(transientFundArr, transientFund)
	}
	return transientFundArr, nil
}

func resetAllTransientFund(
	ctx contractapi.TransactionContextInterface) error {

	queryString := fmt.Sprintf(
		`{"selector":{"docType":"%s"}}`,
		transientFundObjectType)
	transientFundArr, err := getTransientFundArrayFromQuery(ctx, queryString)
	if err != nil {
		return err
	}
	for _, transientFund := range transientFundArr {
		err = ctx.GetStub().DelState(transientFund.RefID)
		if err != nil {
			return err
		}
	}
	return nil
}

func crossChannelQuery(
	ctx contractapi.TransactionContextInterface,
	queryArgs [][]byte,
	targetChannel string,
	targetChaincode string) ([]byte, error) {

	response := ctx.GetStub().InvokeChaincode(targetChaincode, queryArgs, targetChannel)
	if response.Status != 200 {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
		return nil, fmt.Errorf(errStr)
	}
	responseAsBytes := response.Payload

	return responseAsBytes, nil
}

func checkArgArrayLength(
	args []string,
	expectedArgLength int) error {

	argArrayLength := len(args)
	if argArrayLength != expectedArgLength {
		errMsg := fmt.Sprintf(
			"Incorrect number of arguments: Received %d, expecting %d",
			argArrayLength,
			expectedArgLength)
		return errors.New(errMsg)
	}
	return nil
}

func getSigner(
	ctx contractapi.TransactionContextInterface) (string, error) {

	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return "", err
	}
	id := &mspprotos.SerializedIdentity{}
	err = proto.Unmarshal(creator, id)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(id.GetIdBytes())
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	// mspID := id.GetMspid() // if you need the mspID
	signer := cert.Subject.CommonName
	return signer, nil
}

func verifyIdentity(
	ctx contractapi.TransactionContextInterface,
	identity string) error {

	creatorString, err := getSigner(ctx)
	if err != nil {
		return err
	}
	if creatorString != identity {
		errMsg := fmt.Sprintf(
			"Error: Identity of creator (%s) does not match %s",
			creatorString,
			identity)
		return errors.New(errMsg)
	}
	return nil
}

func getTxTimeStampAsTime(
	ctx contractapi.TransactionContextInterface) (time.Time, error) {

	timestampTime := time.Time{}
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return timestampTime, err
	}
	timestampTime, err = ptypes.Timestamp(timestamp)
	if err != nil {
		return timestampTime, err
	}

	return timestampTime, nil
}
