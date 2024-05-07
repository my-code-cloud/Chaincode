package main

import (
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
  storeContract := new(StoreContract)
  cc, err := contractapi.NewChaincode(storeContract)

  if err != nil {
    panic(err.Error())
  }

  if err := cc.Start(); err != nil {
    panic(err.Error())
  }
}
