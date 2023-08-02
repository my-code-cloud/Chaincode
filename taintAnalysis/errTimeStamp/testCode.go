package test

// 污点源：time()返回的结果（准确来说应该是指针参数）
/* 匹配特征：
*  1. 污点不能传播到分支跳转语句
*  2. 污点不能传播到PutState()
*/
import (
	"encoding/json"
	"fmt"	
	"time"
    "github.com/hyperledger/fabric-chaincode-go/shim"
    "github.com/hyperledger/fabric-protos-go/peer"
)

type BadChaincode struct {}

func (t *BadChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("success"))
}

func (t *BadChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	tByte, err := json.Marshal(time.Now())
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState("key", tByte)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("success"))
}

func main() {
	if err := shim.Start(new(BadChaincode)); err != nil {
		fmt.Printf("Error starting BadChaincode chaincode: %s", err)
	}
}