package test

import (
	"fmt"
    "github.com/hyperledger/fabric-chaincode-go/shim"
    "github.com/hyperledger/fabric-protos-go/peer"
)

type BadChaincode struct {}

func (t *BadChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("success"))
}

func (t *BadChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()
	result, err := stub.GetPrivateData(fn, string(args[0]))

	if err != nil {
		return shim.Error(err.Error())
	}
	// Return the result as success payload
	output := []byte(result)
	return shim.Success(output)
}

func main() {
	if err := shim.Start(new(BadChaincode)); err != nil {
		fmt.Printf("Error starting BadChaincode chaincode: %s", err)
	}
}