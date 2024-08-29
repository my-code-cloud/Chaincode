/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	// Set up any variables or assets here by calling stub.PutState()

	// We store the key and the value on the ledger
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()
	log.Printf("fn=%v, args=%v", fn, args)

	var result string
	var err error
	if fn == "set" {
		err = stub.PutState(args[0], []byte(args[1]))
	} else if fn == "delete" {
		err = stub.DelState(args[0])
	} else if fn == "get" { // assume 'get' even if fn is nil
		result, err = get(stub, args)
	} else if fn == "state" {
		err = state(stub)
	} else if fn == "history" {
		result, err = history(stub, args)
	} else if fn == "getSecured" {
		result, err = GetSecured(stub, args)
	} else if fn == "setSecured" {
		err = SetSecured(stub, args)
		if err != nil {
			log.Printf("SetSecured finished successuflly")
		}
	} else {
		err = getArgs(stub)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

//======================================================================================================================================

type SecuredValue struct {
	OwnerId string `json:"OwnerId"`
	Value   string `json:"Value"`
}

func SaveSecuredValue(stub shim.ChaincodeStubInterface, mspId, key string, securedValue SecuredValue) error {
	resJson, err := json.Marshal(securedValue)
	if err != nil {
		return err
	}

	err = stub.PutPrivateData(mspId, key, resJson)
	return err
}

func SetSecured(stub shim.ChaincodeStubInterface, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("incorrect arguments. Expecting a key and a value")
	}

	// Get MSPID
	mspId, err := cid.GetMSPID(stub)
	if err != nil {
		return err
	}

	// Get client id
	ownerId, err := cid.GetID(stub)
	if err != nil {
		return err
	}

	key := args[0]
	value := args[1]

	// Check for existing stored value
	valueJson, err := stub.GetPrivateData(mspId, key)
	if err != nil {
		return err
	}

	// If there exists value
	if valueJson != nil {
		var securedValue SecuredValue
		err = json.Unmarshal(valueJson, &securedValue)
		if err != nil {
			return err
		}

		// Check its owner
		if securedValue.OwnerId != ownerId {
			return fmt.Errorf("there is already another owner of the key")
		}

		// Update value
		securedValue.Value = value

		return SaveSecuredValue(stub, mspId, key, securedValue)
	}

	// Create new value
	securedValue := SecuredValue{
		ownerId,
		value,
	}

	return SaveSecuredValue(stub, mspId, key, securedValue)
}

func GetSecured(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("incorrect arguments. Expecting one argument, the key")
	}

	key := args[0]

	// Get MSPID
	mspId, err := cid.GetMSPID(stub)
	if err != nil {
		return "", err
	}

	// Get stored value
	valueJson, err := stub.GetPrivateData(mspId, key)
	if err != nil {
		return "", err
	}

	// There can be no value with key
	if valueJson == nil {
		return "", fmt.Errorf("state doesn't contain key: %s", key)
	}

	var securedValue SecuredValue
	err = json.Unmarshal(valueJson, &securedValue)
	if err != nil {
		return "", err
	}

	return securedValue.Value, nil
}

//======================================================================================================================================

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return args[1], nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return string(value), nil
}

func getArgs(stub shim.ChaincodeStubInterface) error {
	args := stub.GetArgs()
	fmt.Println(args)

	argsString := stub.GetStringArgs()
	fmt.Println(argsString)

	fn, args2 := stub.GetFunctionAndParameters()
	fmt.Println(fmt.Sprintf("GetFunctionAndParameters(): fn=[%v], args[%v]", fn, args2))

	argsBytes, err := stub.GetArgsSlice()
	if err != nil {
		return err
	}
	fmt.Println(argsBytes)

	return nil
}

type HistoryItem struct {
	TxId      string
	Timestamp int64
	Value     string
}

func history(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("expected slice len = 1, got = [%v]", len(args))
	}

	i, err := stub.GetHistoryForKey(args[0])
	if err != nil {
		return "", err
	}
	var result []HistoryItem
	for i.HasNext() {
		obj, err := i.Next()
		if err != nil {
			return "", err
		}
		result = append(result, HistoryItem{
			TxId:      obj.TxId,
			Timestamp: obj.Timestamp.Seconds,
			Value:     string(obj.Value),
		})
	}

	buf, err := json.Marshal(&result)
	if err != nil {
		return "", err
	}
	fmt.Println(string(buf))

	return string(buf), nil
}

func state(stub shim.ChaincodeStubInterface) error {
	txID := stub.GetTxID()
	fmt.Println(txID)

	return nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
