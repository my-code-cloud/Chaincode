package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

// Device represents the structure of the IoT device information
type Device struct {
	MACID       string  `json:"macId"`
	Reputation  float64 `json:"reputation"`
	PublicKey   string  `json:"publicKey"`
	PrivateKey  string  `json:"privateKey"`
	IsLBR		bool	`json:isLBR`
	AccessLevel int     `json:accessLevel`
}

// DeviceChaincode represents the chaincode implementation
type DeviceChaincode struct {
}

// Init is called during chaincode instantiation to initialize any data
func (d *DeviceChaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke is called to perform any action on the ledger
func (d *DeviceChaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "initLedger":
		return d.initLedger(stub, args)
	case "registerDevice":
		return d.registerDevice(stub, args)
	case "retrieveDevice":
		return d.retrieveDevice(stub, args)
	case "updateDevice":
		return d.updateDevice(stub, args)
	case "deleteDevice":
		return d.deleteDevice(stub, args)
	default:
		return shim.Error("Invalid function name")
	}
}

func (s *DeviceChaincode) initLedger(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	devices := []Device{
		Device{
			MACID: "12:34:56:78:AB", 
			Reputation: 10.00,
			PublicKey: "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAExpVkyDQBa64BNFy5I5fVh43JKlzY7sCJhTnrf++7JKb6dXvTBzLgpSMZdCjRJj+5gt7CV2y1CwUlAxa2eV7uCQ==",
			PrivateKey: "MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgCIMgPCzJ7qZdSP9wQz+BfR9W4b1w3hW/wy1LBqFOz2GhRANCAATGlWTINAFrrgE0XLkjl9WHjckqXNjuwImFOet/77skpvp1e9MHMuClIxl0KNEmP7mC3sJXbLULBSUDFrZ5Xu4J",
			IsLBR: false,		
			AccessLevel: 0},
	}

	i := 0
	for i < len(devices) {
		devAsBytes, _ := json.Marshal(devices[i])
		stub.PutState("DEVICE"+strconv.Itoa(i), devAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

// registerDevice creates a new device and stores it on the ledger
func (d *DeviceChaincode) registerDevice(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3: MAC ID, Reputation, Shared Key")
	}

	macID := args[0]
	reputation := args[1]
	publicKey := args[2]
	privateKey := args[3]
	isLBR := false
	if (args[4] == "1") {
		isLBR = true
	}
	accessLevel := 0
	if (args[5] == "1") {
		accessLevel = 1
	}

	// Check if the device already exists
	deviceAsBytes, err := stub.GetState(macID)
	if err != nil {
		return shim.Error("Failed to get device: " + err.Error())
	} else if deviceAsBytes != nil {
		return shim.Error("Device already exists with MAC ID: " + macID)
	}

	// Create a new device object
	rep, err := strconv.ParseFloat(reputation, 64)
	device := Device{
		MACID:      macID,
		Reputation: rep,
		PublicKey: publicKey,
		PrivateKey: privateKey,
		IsLBR: isLBR,
		AccessLevel: accessLevel,
	}

	// Convert the device object to JSON
	deviceJSON, err := json.Marshal(device)
	if err != nil {
		return shim.Error("Failed to marshal device JSON: " + err.Error())
	}

	// Store the device on the ledger
	err = stub.PutState(macID, deviceJSON)
	if err != nil {
		return shim.Error("Failed to put device on the ledger: " + err.Error())
	}

	return shim.Success(nil)
}

// retrieveDevice retrieves a device from the ledger based on its MAC ID
func (d *DeviceChaincode) retrieveDevice(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1: MAC ID")
	}

	macID := args[0]

	// Retrieve the device from the ledger
	deviceAsBytes, err := stub.GetState(macID)
	if err != nil {
		return shim.Error("Failed to get device: " + err.Error())
	} else if deviceAsBytes == nil {
		return shim.Error("Device does not exist with MAC ID: " + macID)
	}

	return shim.Success(deviceAsBytes)
}

// updateDevice updates the reputation of a device on the ledger
func (d *DeviceChaincode) updateDevice(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2: MAC ID, Reputation")
	}

	macID := args[0]
	newReputation := args[1]

	// Retrieve the device from the ledger
	deviceAsBytes, err := stub.GetState(macID)
	if err != nil {
		return shim.Error("Failed to get device: " + err.Error())
	} else if deviceAsBytes == nil {
		return shim.Error("Device does not exist with MAC ID: " + macID)
	}

	// Update the device's reputation
	device := Device{}
	err = json.Unmarshal(deviceAsBytes, &device)
	if err != nil {
		return shim.Error("Failed to unmarshal device JSON: " + err.Error())
	}
	rep, err := strconv.ParseFloat(newReputation, 64)
	device.Reputation = rep

	// Convert the updated device object to JSON
	deviceJSON, err := json.Marshal(device)
	if err != nil {
		return shim.Error("Failed to marshal device JSON: " + err.Error())
	}

	// Update the device on the ledger
	err = stub.PutState(macID, deviceJSON)
	if err != nil {
		return shim.Error("Failed to put device on the ledger: " + err.Error())
	}

	return shim.Success(nil)
}

// deleteDevice deletes a device from the ledger based on its MAC ID
func (d *DeviceChaincode) deleteDevice(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1: MAC ID")
	}

	macID := args[0]

	// Delete the device from the ledger
	err := stub.DelState(macID)
	if err != nil {
		return shim.Error("Failed to delete device: " + err.Error())
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(DeviceChaincode))
	if err != nil {
		fmt.Printf("Error starting Device chaincode: %s", err)
	}
}
