package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

//简单asset
type SimpleAsset struct{
	
}

type PrivateData struct {
	Key string `json:"key"`
	Value string `json:"value"`
}

func(t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response{
	//Get the args from the transaction proposal
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments.Expecting a key and a value")
	}

	//store key and value on the ledger
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

func(t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	//Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()
	var result string
	var err error
	if fn == "set" {
		result, err = set(stub, args)
	}else if fn == "publicAsset" {
		result, err = publicAsset(stub)
	}else if fn == "privateAsset" {
		result, err = privateAsset(stub)
	}else if fn == "readPublicAsset" {
		result, err = readPublicAsset(stub, args)
	}else if fn == "readPrivateAsset" {
		result, err = readPrivateAsset(stub, args)
	}else {
		result, err = get(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(result))
}


func set(stub shim.ChaincodeStubInterface, args []string) (string, error){
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return args[1], nil
}

func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrecct arguments. Expecting a key")
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

func publicAsset(stub shim.ChaincodeStubInterface)(string, error) {
	transMap, err := stub.GetTransient()
	if err != nil {
			return "", fmt.Errorf("Error getting transient: " + err.Error())
		}
	DataJson, ok := transMap["Data"]
	if !ok {
		return "", fmt.Errorf("Expecting a key.")
	}

	var DataInput PrivateData
	err = json.Unmarshal(DataJson, &DataInput)
	if err != nil {
		return "", fmt.Errorf("Failed to unmarshal JSON: %s", err)
	}
	err = stub.PutPrivateData("publicAssets", DataInput.Key, []byte(DataInput.Value))
	if err != nil {
		return "", fmt.Errorf("Failed to put public Asset: %s with %v: %v", err, DataInput.Key, DataInput.Value)
	}
	return string(DataInput.Value), nil
}

func privateAsset(stub shim.ChaincodeStubInterface) (string, error){
	transMap, err := stub.GetTransient()
	if err != nil {
			return "", fmt.Errorf("Error getting transient: " + err.Error())
		}
	DataJson, ok := transMap["Data"]
	if !ok {
		return "", fmt.Errorf("Expecting a key.")
	}

	var DataInput PrivateData
	err = json.Unmarshal(DataJson, &DataInput)
	if err != nil {
		return "", fmt.Errorf("Failed to unmarshal JSON: %s", err)
	}
	err = stub.PutPrivateData("privateAssets", DataInput.Key, []byte(DataInput.Value))
	if err != nil {
		return "", fmt.Errorf("Failed to put private Asset: %s", err)
	}
	return string(DataInput.Value), nil
}

func readPublicAsset(stub shim.ChaincodeStubInterface, args []string)(string, error){
	if len(args) != 1 {
			return "", fmt.Errorf("Incorrect arguments. Expecting a key.")
	}
	value,	err := stub.GetPrivateData("publicAssets", args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get Asset: %s with error %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return string(value), nil
}

func readPrivateAsset(stub shim.ChaincodeStubInterface, args[]string) (string, error){
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key.")
	}
	value, err := stub.GetPrivateData("privateAssets", args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get Asset: %s with error %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return string(value), nil
}
func main(){
	err := shim.Start(new(SimpleAsset))
	if err != nil {
		fmt.Printf("Error start simple chaincode: %s", err)
	}
}