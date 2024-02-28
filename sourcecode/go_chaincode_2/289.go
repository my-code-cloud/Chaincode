package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// RequestDOChaincode represents the chaincode implementation
type RequestDOChaincode struct {
}

// RequestDetails represents the structure for storing DO request details
type RequestDetails struct {
	Requestor Requestor `json:"requestor"`
}

// Requestor represents the structure for storing requestor details
type Requestor struct {
	RequestorType   string `json:"requestorType"`
	URLFile         string `json:"urlFile"`
	NPWP            string `json:"npwp"`
	NIB             string `json:"nib"`
	RequestorName   string `json:"requestorName"`
	RequestorAddr   string `json:"requestorAddress"`
}

// Init initializes the chaincode
func (cc *RequestDOChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke is the entry point for chaincode invocations
func (cc *RequestDOChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "requestDO" {
		return cc.requestDO(stub, args)
	} else if function == "queryDO" {
		return cc.queryDO(stub, args)
	}

	return shim.Error("Invalid function name.")
}

// RequestDO allows a requestor to submit a DO request
func (cc *RequestDOChaincode) requestDO(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}

	var requestDetails RequestDetails
	err := json.Unmarshal([]byte(args[0]), &requestDetails)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to unmarshal request details: %s", err.Error()))
	}

	// Save the request details on the ledger
	err = stub.PutState("DORequest", []byte(args[0]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to save DO request: %s", err.Error()))
	}

	return shim.Success(nil)
}

// QueryDO retrieves the data of a specific DO request
func (cc *RequestDOChaincode) queryDO(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Expecting 0.")
	}

	// Retrieve the DO request data from the ledger
	doRequestBytes, err := stub.GetState("DORequest")
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get DO request: %s", err.Error()))
	}
	if doRequestBytes == nil {
		return shim.Error("DO request does not exist.")
	}

	return shim.Success(doRequestBytes)
}

func main() {
	err := shim.Start(new(RequestDOChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
