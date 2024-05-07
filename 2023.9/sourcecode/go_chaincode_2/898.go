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

func  TestPropertyContract(t *testing.T) {
	os.Setenv("MODE","TEST")
	
	assert := assert.New(t)
	uid := uuid.New().String()

	cc, err := contractapi.NewChaincode(new(PropertyContract))
	assert.Nil(err, "error should be nil")

	stub := shimtest.NewMockStub("TestStub", cc)
	assert.NotNil(stub, "Stub is nil, TestStub creation failed")

	// - - - test PropertyContract:Put function - - - 
	putResp := stub.MockInvoke(uid,[][]byte{
		[]byte("PropertyContract:Put"),
		[]byte("1"),
		[]byte("1"),
		[]byte("ProType"),
		[]byte("ProName"),
		[]byte("Desc"),
		[]byte("Address"),
		[]byte("Location"),
		[]byte("LocationLat"),
		[]byte("LocationLong"),
		[]byte("Views"),
		[]byte("ViewerStats"),
		[]byte("EntryDate"),
		[]byte("ExpiryDate"),
		[]byte("Status"),
	})
	assert.EqualValues(OK, putResp.GetStatus(), putResp.GetMessage())
	

	// - - - test PropertyContract:Get function - - - 
	testID := "1"
	getResp := stub.MockInvoke(uid, [][]byte{
		[]byte("PropertyContract:Get"),
		[]byte(testID),
	})
	assert.EqualValues(OK, getResp.GetStatus(), getResp.GetMessage())
	assert.NotNil(getResp.Payload, "getResp.Payload should not be nil")
	
	propertyObj := new(PropertyObj)
	err = json.Unmarshal(getResp.Payload, propertyObj)
	assert.Nil(err, "json.Unmarshal error should be nil")
	assert.NotNil(propertyObj, "propertyObj should not be nil")

	retrievedID := strconv.Itoa(propertyObj.ProID)
	assert.EqualValues(testID, retrievedID, "testID and retrievedID mismatch")
}