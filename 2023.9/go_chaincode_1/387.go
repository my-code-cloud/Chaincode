package main

import (
	"github.com/hyperledger/fabric-samples/chaincode/thinh-chaincode/dummy"
	"github.com/hyperledger/fabric-samples/chaincode/thinh-chaincode/product"
	"github.com/hyperledger/fabric-samples/chaincode/thinh-chaincode/productBatch"
	"github.com/hyperledger/fabric-samples/chaincode/thinh-chaincode/utils"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	productContract := new(product.ProductContract)
	productContract.Name = "hamburg.example.com.ProductContract"
	productContract.TransactionContextHandler = new(utils.RoleBasedTransactionContext)
	productContract.BeforeTransaction = utils.GetClientRoleAndRefuseDistributor

	productBatchContract := new(productBatch.ProductBatchContract)
	productBatchContract.Name = "hamburg.example.com.ProductBatchContract"
	productBatchContract.TransactionContextHandler = new(utils.RoleBasedTransactionContext)
	productBatchContract.BeforeTransaction = utils.GetClientRole

	dummyContract := new(dummy.DummyContract)
	dummyContract.Name = "hamburg.example.com.DummyContract"

	cc, err := contractapi.NewChaincode(productContract, productBatchContract, dummyContract)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
