package test

// 污点源：所有全局变量和chaincode对象中的字段
/* 匹配特征：
*  1. 污点不能传播到分支跳转语句
*  2. 污点不能传播到PutState()操作
 */

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

var globalValue = ""

type BadChaincode struct{}

func (t *BadChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("success"))
}

func (t BadChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()
	if fn == "setValue" {
		globalValue = args[0]
		stub.PutState("key", []byte(globalValue))
		return shim.Success([]byte("success"))
	} else if fn == "getValue" {
		value, _ := stub.GetState("key")
		return shim.Success(value)
	}
	return shim.Error("not a valid function")
}
