package main

import (
	"fmt"

	factory "factory/contract"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {

	contract := new(factory.FactoryContract)
	contract.Name = "factory_contract"
	contract.Info.Version = "0.0.1"

	chaincode, err := contractapi.NewChaincode(contract)

	if err != nil {
		panic(fmt.Sprintf("Error creating chaincode. %s", err.Error()))
	}

	chaincode.Info.Title = "FactoryChaincode"
	chaincode.Info.Version = "0.0.1"

	err = chaincode.Start()

	if err != nil {
		panic(fmt.Sprintf("Error starting chaincode. %s", err.Error()))
	}
}
