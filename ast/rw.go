package fixtures

// 污点源：Putstate()操作的对象
/* 匹配特征：
*  1. 污点不能传播到GetState()
 */

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

type b struct {
	key   string
	value int
}

type a struct {
	b
	s string
}

type BadChaincode struct{}

func (t *BadChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("success"))
}

func (t *BadChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	key := "key"
	data := "data"

	err := stub.PutState(a.b.key, []byte(data))
	if err != nil {
		return shim.Error("could not write new data")
	}
	respone, err := stub.GetState(a.b.key)
	if err != nil {
		return shim.Error("could not read data")
	}

	return shim.Success([]byte(respone))
}
