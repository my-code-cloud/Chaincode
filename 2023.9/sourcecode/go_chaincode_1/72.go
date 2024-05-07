/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"github.com/jhl8109/blockchain-event-trace-system/used_car/chaincode"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	assetChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating used car transfer chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting used car transfer chaincode: %v", err)
	}
}
