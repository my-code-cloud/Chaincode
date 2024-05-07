package test

// 检测go转IR后的map结构
/* 匹配特征：
*  1. 链码中不应使用map结构
*/

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

var myMap = map[int]int{
	1: 1,
	2: 5,
	3: 10,
	4: 50,
}

type BadChaincode struct {}

func (t *BadChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (t *BadChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	returnValue := 0
	for i, ii := range myMap {
		returnValue = returnValue*i - ii
	}
	return shim.Success([]byte("value: " + string(returnValue)))
}

func main() {
	if err := shim.Start(new(BadChaincode)); err != nil {
		fmt.Printf("Error starting BadChaincode chaincode: %s", err)
	}
}