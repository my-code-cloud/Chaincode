/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

//--- arnaud.bart@sita.aero ----

package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type SimpleAsset struct {
}

func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "set" {
		result, err = set(stub, args)
	} else if fn == "setFPL" {
		result, err = setFPL(stub, args)
	} else if fn == "getFPL" {
		result, err = getFPL(stub, args)
	} else if fn == "getFPLasBA" {
		result, err = getFPLasBA(stub, args)
	} else if fn == "getFPLasEZ" {
		result, err = getFPLasEZ(stub, args)
	} else { // assume 'get' even if fn is nil
		result, err = get(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

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

// SetFPL stores the values of the specified asset key in State | PDC | Implicit NATS
func setFPL(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 5 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}
	// Store Public info in State database so available to all members of cc
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	// Store Private info in Implicit nats database for any airline so available only to nats
	implicit := "restricited nats: "+ args[2]
	err1 := stub.PutPrivateData("_implicit_org_natsMSP", args[0], []byte(implicit))
	if err1 != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	// Store Private info in PDC between airline and nats so available to the airline and nats
	pdc := "nats" + args[4]
	err2 := stub.PutPrivateData(string(pdc), args[0], []byte(args[2]))
	if err2 != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}

	return "The asset is " + args[0] + " where public info is " + args[1] + " and private info is " + args[2] + "visible only from nats & " + args[4] + " and restricted info is "+ args[3] + " visible only from nats.", nil
}

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

// GetFPL returns the value of the specified asset key stored in State | PDC | Implicit when NATS
func getFPL(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetPrivateData("_implicit_org_natsMSP", args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	pdc := "nats" + args[1]
	value1, err1 := stub.GetPrivateData(string(pdc), args[0])
	if err1 != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value1 == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	// return string(value), nil
	value2, err2 := stub.GetState(args[0])
	if err2 != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err2)
	}
	if value2 == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return "Public info in State DB is " + string(value2) + " // Private info in PDC is " + string(value1) + " // Private info in NATS DB only is " + string(value), nil
}

func getFPLasBA(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}
	value, err := stub.GetPrivateData("natsba", args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	value1, err1 := stub.GetState(args[0])
	if err1 != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err1)
	}
	if value1 == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return "Public info is " + string(value1) + " | Private info only visible by BA & NATS is " + string(value), nil
}

func getFPLasEZ(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}
	value, err := stub.GetPrivateData("natsez", args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	// return string(value), nil
	value1, err1 := stub.GetState(args[0])
	if err1 != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err1)
	}
	if value1 == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return "Public info is " + string(value1) + " | Private info only visible by EZ & NATS is " + string(value), nil
}

func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}