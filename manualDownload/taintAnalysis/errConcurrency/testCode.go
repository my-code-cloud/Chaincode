package test

// 检测go并发的API（转IR后的结构）
/* 匹配特征：
*  1. 链码中不应使用并发操作
*/

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

type BadChaincode struct {
}

func (t *BadChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("success"))
}

func (t *BadChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	key := "key"
	data := "data"
	data2 := "data2"
	go stub.PutState(key, []byte(data))
	go stub.PutState(key, []byte(data2))
	return shim.Success([]byte("success"))
}	

func main() {
	if err := shim.Start(new(BadChaincode)); err != nil {
		fmt.Printf("Error starting BadChaincode chaincode: %s", err)
	}
}