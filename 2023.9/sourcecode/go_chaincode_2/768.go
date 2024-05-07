package main

import (
	"os"
	"testing"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func  TestTransactionContract(t *testing.T) {
	os.Setenv("MODE","TEST")
	
	assert := assert.New(t)
	uid := uuid.New().String()

	cc, err := contractapi.NewChaincode(new(TransactionContract))
	assert.Nil(err, "error should be nil")

	stub := shimtest.NewMockStub("TestStub", cc)
	assert.NotNil(stub, "Stub is nil, TestStub creation failed")

	// - - - test TransactionContract:Put function - - - 
	putResp := stub.MockInvoke(uid,[][]byte{
		[]byte("TransactionContract:Put"),
		[]byte("1"),
		[]byte("TxnDate"),
		[]byte("2"),
		[]byte("3"),
		[]byte("1.1"),
		[]byte("2.2"),
		[]byte("3.3"),
		[]byte("4.4"),
		[]byte("DueDate"),
		[]byte("Bank"),
		[]byte("LoanStatus"),
	})
	assert.EqualValues(OK, putResp.GetStatus(), putResp.GetMessage())
	

	// - - - test TransactionContract:Get function - - - 
	testID := "1"
	getResp := stub.MockInvoke(uid, [][]byte{
		[]byte("TransactionContract:Get"),
		[]byte(testID),
	})
	assert.EqualValues(OK, getResp.GetStatus(), getResp.GetMessage())
	assert.NotNil(getResp.Payload, "getResp.Payload should not be nil")
	
	transactionObj := new(TransactionObj)
	err = json.Unmarshal(getResp.Payload, transactionObj)
	assert.Nil(err, "json.Unmarshal error should be nil")
	assert.NotNil(transactionObj, "transactionObj should not be nil")

	retrievedID := strconv.Itoa(transactionObj.TxnID)
	assert.EqualValues(testID, retrievedID, "testID and retrievedID mismatch")
}