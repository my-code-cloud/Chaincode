/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"tallysolutions.com/TallyScoreProject/chaincode/tallyscore/chaincode"
)

func main() {
	assetChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating tallyscore chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting tallyscore chaincode: %v", err)
	}
}