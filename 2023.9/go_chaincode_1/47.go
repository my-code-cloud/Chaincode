package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type CampaignSmartContract struct {
	contractapi.Contract
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&CampaignSmartContract{})
	if err != nil {
		log.Panicf("Error creating chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}