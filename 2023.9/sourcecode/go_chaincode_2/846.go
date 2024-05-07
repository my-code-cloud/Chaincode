// SPDX-License-Identifier: Apache-2.0

/*
  Sample Chaincode based on Demonstrated Scenario

 This code is based on code written by the Hyperledger Fabric community.
  Original code can be found here: https://github.com/hyperledger/fabric-samples/blob/release/chaincode/fabcar/fabcar.go
*/

package main

/* Imports
* 5 utility libraries for handling bytes, reading and writing JSON,
formatting, and string manipulation
* 2 specific Hyperledger Fabric specific libraries for Smart Contracts
*/
import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"strconv"
)

// Define the Smart Contract structure
type SmartContract struct {
}

/* Define Telephonenumber structure, with 3 properties.
Structure tags are used by encoding/json library
*/
type Telephonenumber struct {
	Phonenumber string `json:"tn"`
	UseIntent string `json:"use_intent"`
	AssignedTo string `json:"assignedTo"`
	Owner string `json:"owner"`
}

/*
 * The queryItem queryItem method *
Used to view the records of one particular item
It takes one argument -- the key for the item in question
*/
func (s *SmartContract) queryItem(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	itemAsBytes, _ := APIstub.GetState(args[0])
	if itemAsBytes == nil {
		return shim.Error("Could not locate item")
	}
	return shim.Success(itemAsBytes)
}


/*
 * The recordTelephonenumber method *
Can be used to add new Telephonenumbers to the DLT.s
*/
func (s *SmartContract) recordTelephonenumber(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	var newTelephonenumber = Telephonenumber{ Phonenumber: args[1], UseIntent: args[2], AssignedTo: args[3], Owner: args[4]}

	newTelephonenumberAsBytes, _ := json.Marshal(newTelephonenumber)
	err := APIstub.PutState(args[0], newTelephonenumberAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record new Telephonenumber: %s", args[0]))
	}

	return shim.Success(nil)
}

/*
 * The Init method *
 called when the Smart Contract is instantiated by the network
 * Best practice is to have any Ledger initialization in separate function
 -- see initLedger()
*/
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method *
 called when an application requests to run the Smart Contract
 The app also specifies the specific smart contract function to call with args
*/
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger
	if function == "queryItem" {
		return s.queryItem(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "recordTelephonenumber" {
		return s.recordTelephonenumber(APIstub, args)
	} else if function == "queryAllItems" {
		return s.queryAllItems(APIstub)
	}

	return shim.Error("Invalid Smart Contract function name.")
}


/*
 * The initLedger method *
Will add test data (5 Telephonenumber)to our network
*/
func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	telephonenumber := []Telephonenumber{
		Telephonenumber{Phonenumber: "+573202366543", UseIntent: "good", AssignedTo: "1", Owner: "1"},
		Telephonenumber{Phonenumber: "+4419428683", UseIntent: "good", AssignedTo: "2", Owner: "2"},
		Telephonenumber{Phonenumber: "+4934128756", UseIntent: "good", AssignedTo: "3", Owner: "3"},
		Telephonenumber{Phonenumber: "+491512899154", UseIntent: "good", AssignedTo: "4", Owner: "4"},
		Telephonenumber{Phonenumber: "+34563412421", UseIntent: "bad", AssignedTo: "5", Owner: "5"},
	}

	i := 0
	for i < len(telephonenumber) {
		fmt.Println("i is ", i)
		telephonenumberAsBytes, _ := json.Marshal(telephonenumber[i])
		APIstub.PutState(strconv.Itoa(i+1), telephonenumberAsBytes)
		fmt.Println("Added Telephonenumber", telephonenumber[i])
		i = i + 1
	}

	return shim.Success(nil)
}

/*
 * The queryAllItems method *
allows for assessing all the records added to the ledger(all Telephonenumber entries)
This method does not take any arguments. Returns JSON string containing results.
*/
func (s *SmartContract) queryAllItems(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "0"
	endKey := "999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add comma before array members,suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
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

	fmt.Printf("- queryAllItems:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

/*
 * main function *
calls the Start function
The main function starts the chaincode in the container during instantiation.
*/
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
