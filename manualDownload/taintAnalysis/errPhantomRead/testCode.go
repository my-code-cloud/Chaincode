package test

// 污点源：范围查询函数的返回结果
// 如GetQueryResult()，GetHistoryForKey()和GetPrivateDataQueryResult()
/* 匹配特征：
*  1. 污点不能传播到PutState()操作
*/

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

type BadChaincode struct {}

func (t *BadChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("success"))
}

func (t *BadChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	iterator, _ := stub.GetHistoryForKey("key")
	data, _ := iterator.Next()

	err := stub.PutState("key", data.Value)
	if err != nil {
		return shim.Error("could not write new data")
	}

	return shim.Success([]byte("stored"))
}

func main() {
	if err := shim.Start(new(BadChaincode)); err != nil {
		fmt.Printf("Error starting BadChaincode chaincode: %s", err)
	}
}