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

func  TestLoanDocContract(t *testing.T) {
	os.Setenv("MODE","TEST")
	
	assert := assert.New(t)
	uid := uuid.New().String()

	cc, err := contractapi.NewChaincode(new(LoanDocContract))
	assert.Nil(err, "error should be nil")

	stub := shimtest.NewMockStub("TestStub", cc)
	assert.NotNil(stub, "Stub is nil, TestStub creation failed")

	// - - - test LoanDocContract:Put function - - - 
	putResp := stub.MockInvoke(uid,[][]byte{
		[]byte("LoanDocContract:Put"),
		[]byte("1"),
		[]byte("1"),
		[]byte("docname"),
		[]byte("docdesc"),
		[]byte("doclink"),
	})
	assert.EqualValues(OK, putResp.GetStatus(), putResp.GetMessage())
	

	// - - - test LoanDocContract:Get function - - - 
	testID := "1"
	getResp := stub.MockInvoke(uid, [][]byte{
		[]byte("LoanDocContract:Get"),
		[]byte(testID),
	})
	assert.EqualValues(OK, getResp.GetStatus(), getResp.GetMessage())
	assert.NotNil(getResp.Payload, "getResp.Payload should not be nil")
	
	loanDocObj := new(LoanDocObj)
	err = json.Unmarshal(getResp.Payload, loanDocObj)
	assert.Nil(err, "json.Unmarshal error should be nil")
	assert.NotNil(loanDocObj, "loanDocObj should not be nil")

	retrievedID := strconv.Itoa(loanDocObj.DocID)
	assert.EqualValues(testID, retrievedID, "testID and retrievedID mismatch")
}