package main

import (
	"log"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func main() {
	if err := shim.Start(&exampleCc{}); err != nil {
		log.Fatal(err)
	}
}

type exampleCc struct {
}

func (*exampleCc) Init(stub shim.ChaincodeStubInterface) peer.Response {
	t, err := stub.GetTransient()
	if err != nil {
		return shim.Error(err.Error())
	}

	log.Println(t)

	if err = stub.PutState(`key`, t[`key`]); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (*exampleCc) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	t, err := stub.GetTransient()
	if err != nil {
		return shim.Error(err.Error())
	}

	log.Println(t)

	if err = stub.PutState(`key`, t[`key`]); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
