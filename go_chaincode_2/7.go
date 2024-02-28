package main

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type Invoker struct {
	ccName string
	ccStub shim.ChaincodeStubInterface
}

func NewInvoker(ccName string, ccStub shim.ChaincodeStubInterface) *Invoker {
	return &Invoker{
		ccName: ccName,
		ccStub: ccStub,
	}
}

func (invoker *Invoker) Invoke(opName string, args ...string) ([]byte, error) {
	invokeResponse := invoker.ccStub.InvokeChaincode(
		invoker.ccName,
		getQueryArgs(opName, args...),
		invoker.ccStub.GetChannelID(),
	)

	if invokeResponse.Status != shim.OK {
		return nil, fmt.Errorf(invokeResponse.Message)
	}

	return invokeResponse.Payload, nil
}

func getQueryArgs(opName string, args ...string) [][]byte {
	queryArgs := make([][]byte, len(args)+1)

	queryArgs[0] = []byte(opName)
	for i, arg := range args {
		queryArgs[i+1] = []byte(arg)
	}

	return queryArgs
}
