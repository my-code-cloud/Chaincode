package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

type serverConfig struct {
	CCID    string
	Address string
}

type Scanfile struct {
	Txid         string `json:"txid"`
	Hash         string `json:"hash"`
	UserId       string `json:"userid"`
	FileName     string `json:"filename"`
	Last6Txid    string `json:"last6txid"`
}

const (
	MESSAGE_0   string = "success"                                              // 應用: 成功
	MESSAGE_101 string = "Incorrect number of parameters"                       // 應用: 參數數量有誤
	MESSAGE_102 string = "An exception occurred while accessing data"           // 系統: 存取資料時發生異常(使用fabric shim sdk時)
	MESSAGE_103 string = "This key value already exists"                        // 應用: uniqueKey重複，表示檔案已經儲存過
	MESSAGE_104 string = "This key value has not been stored in the blockchain" // 應用: 無此uniqueKey，表示尚未將此檔案儲存到區塊鏈中
	MESSAGE_105 string = "An exception occurred while deleting data"            // 系統: 刪除資料時發生異常(使用fabric shim sdk時)
	// MESSAGE_109 string = "The Last6Txid field is different"                     // 應用: Last6Txid跟原本記錄的不同
)

const (
	RTNC_0   int64 = 0
	RTNC_101 int64 = 101
	RTNC_102 int64 = 102
	RTNC_103 int64 = 103
	RTNC_104 int64 = 104
	RTNC_105 int64 = 105
)

type SmartContract struct{}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()

	if function == "set" {
		return s.Set(APIstub, args)
	} else if function == "get" {
		return s.Get(APIstub, args)
	} else if function == "gethist" {
		return s.getHistory(APIstub, args)
	} else if function == "del" {
		return s.delete(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name: " + function)
}

func (s *SmartContract) Set(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// 回應的json結構
	var setResp = make(map[string]interface{})

	if len(args) != 3 {
		setResp["rtnc"] = RTNC_101
		setResp["message"] = MESSAGE_101
		mapResp, _ := json.Marshal(setResp)
		return shim.Success(mapResp)
	}

	//輸入參數
	filename := args[0]
	hash := args[1]
	userid := args[2]

	txid := APIstub.GetTxID()
	last6 := string(txid[len(txid)-6:])
	

	setResp["filename"] = filename
	setResp["hash"] = hash
	setResp["userid"] = userid

	//檢查是否有這個hash值
	hashKeyAsBytes, err := APIstub.GetState(hash)
	if err != nil {
		setResp["rtnc"] = RTNC_102
		setResp["message"] = MESSAGE_102
		mapResp, _ := json.Marshal(setResp)
		return shim.Success(mapResp)
	} else if hashKeyAsBytes != nil {
		// key值重複，表示此key值已存在
		setResp["rtnc"] = RTNC_103
		setResp["message"] = MESSAGE_103
		mapResp, _ := json.Marshal(setResp)
		return shim.Success(mapResp)
	}

	// 將檔案資訊全部儲存在storeOrderInfo結構中，轉成json字串，最後將hash當成key值儲存在區塊鏈中
	var storeInfo = Scanfile{Txid: txid, Hash: hash, UserId:userid, FileName:filename, Last6Txid:last6}
	storeInfoAsBytes, _ := json.Marshal(storeInfo)
	APIstub.PutState(hash, storeInfoAsBytes)

	// 回應的json結構
	setResp["rtnc"] = RTNC_0
	setResp["message"] = MESSAGE_0
	setResp["txid"] = txid
	mapResp, _ := json.Marshal(setResp)
	return shim.Success(mapResp)

}

func (s *SmartContract) Get(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	var vResp = make(map[string]interface{})

	if len(args) != 1 {
		vResp["rtnc"] = RTNC_101
		vResp["message"] = MESSAGE_101
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}

	//輸入參數
	hash := args[0]

	vResp["hash"] = hash

	//檢核key值是否存在
	keyAsBytes, err := APIstub.GetState(hash)
	var scanfile Scanfile
	json.Unmarshal(keyAsBytes, &scanfile) //反序列化

	if err != nil {
		vResp["rtnc"] = RTNC_102
		vResp["message"] = MESSAGE_102
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	} else if keyAsBytes == nil {
		// 查無此key，表示尚未將此檔案儲存到區塊鏈中
		vResp["rtnc"] = RTNC_104
		vResp["message"] = MESSAGE_104
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}

	// 表示查詢到檔案，做json格式回應
	vResp["rtnc"] = RTNC_0
	vResp["message"] = MESSAGE_0
	vResp["txid"] = scanfile.Txid
	vResp["hash"] = scanfile.Hash
	vResp["filename"] = scanfile.FileName
	vResp["userid"] = scanfile.UserId
	mapResp, _ := json.Marshal(vResp)
	return shim.Success(mapResp)

}

func (s *SmartContract) getHistory(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// getHistory中，錯誤用map回，正確用buffer組成

	// 回應的json結構
	var vResp = make(map[string]interface{})

	if len(args) != 1 {
		vResp["rtnc"] = RTNC_101
		vResp["message"] = MESSAGE_101
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}

	//輸入參數
	hash := args[0]

	vResp["hash"] = hash

	resultsIterator, err := APIstub.GetHistoryForKey(hash)

	if err != nil {
		vResp["rtnc"] = RTNC_102
		vResp["message"] = MESSAGE_102
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}
	defer resultsIterator.Close()

	var jsonArr []interface{}
	for resultsIterator.HasNext() {
		// response包含所有資訊，response.Value為儲存的數值
		response, err := resultsIterator.Next()
		if err != nil {
			vResp["rtnc"] = RTNC_102
			vResp["message"] = MESSAGE_102
			mapResp, _ := json.Marshal(vResp)
			return shim.Success(mapResp)
		}

		valueStr := []byte(response.Value)
		valueStruct := Scanfile{}
		json.Unmarshal(valueStr, &valueStruct)
		jsonArr = append(jsonArr, valueStruct)

	}

	if jsonArr == nil {
		vResp["rtnc"] = RTNC_104
		vResp["message"] = MESSAGE_104
		mapResp, _ := json.Marshal(vResp)
		return shim.Success(mapResp)
	}

	vResp["rtnc"] = RTNC_0
	vResp["message"] = MESSAGE_0
	vResp["info"] = jsonArr
	mapResp, _ := json.Marshal(vResp)
	return shim.Success(mapResp)

}

func (s *SmartContract) delete(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// 回應的json結構
	var setResp = make(map[string]interface{})

	if len(args) != 1 {
		setResp["rtnc"] = RTNC_101
		setResp["message"] = MESSAGE_101
		mapResp, _ := json.Marshal(setResp)
		return shim.Success(mapResp)
	}

	//輸入參數
	hash := args[0]	

	//檢查是否有這個hash值
	hashKeyAsBytes, err := APIstub.GetState(hash)
	if err != nil {
		setResp["rtnc"] = RTNC_102
		setResp["message"] = MESSAGE_102
		mapResp, _ := json.Marshal(setResp)
		return shim.Success(mapResp)
	} else if hashKeyAsBytes == nil {
		// 無此key值 故無法刪除資料
		setResp["rtnc"] = RTNC_104
		setResp["message"] = MESSAGE_104
		mapResp, _ := json.Marshal(setResp)
		return shim.Success(mapResp)
	}

	err = APIstub.DelState(hash)
	if err != nil {
		setResp["rtnc"] = RTNC_105
		setResp["message"] = MESSAGE_105
		mapResp, _ := json.Marshal(setResp)
		return shim.Success(mapResp)
	}
	
	// 回應的json結構
	setResp["rtnc"] = RTNC_0
	setResp["message"] = MESSAGE_0
	mapResp, _ := json.Marshal(setResp)
	return shim.Success(mapResp)

}

// func main() {
// 	err := shim.Start(new(SmartContract))
// 	if err != nil {
// 		fmt.Printf("Error creating new Smart Contract: %s", err)
// 	}
// }

func main() {
	// See chaincode.env.example
	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      new(SmartContract),
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	if err := server.Start(); err != nil {
		fmt.Printf("error starting scanfile chaincode: %s", err)
	}
}
