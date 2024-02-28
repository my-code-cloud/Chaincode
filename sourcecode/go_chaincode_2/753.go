package main

import (
	//"bytes"
	"encoding/json"
	"fmt"

	//"github.com/hyperledger/fabric/core/chaincode/shim"
	"log"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

func main() {
	err := shim.Start(new(Chaincode))
	if err != nil {
		fmt.Printf("Error on start: %s", err)
	}
}

// Chaincode is the definition of the chaincode structure.
type Chaincode struct {
}

func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()
	fmt.Println("Init()", fcn, params)
	return shim.Success(nil)
}

func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()
	fmt.Println("Invoke()", fcn, params)

	if fcn == "set" {
		return cc.SaveLedger(stub, params)
	} else if fcn == "get" {
		return cc.GetLedger(stub, params)
	} else if fcn == "history" {
		return cc.GetHistory(stub, params)
	} else {
		return shim.Error("INVALID FUNCTION NAME")
	}
	return shim.Success(nil)
}

func (cc *Chaincode) SaveLedger(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("ARGUMENTS NOT MATCHING")
	}

	log.Println("Inside createIngestable, args received", args)

	// Get Epoch Time
	now := time.Now()
	secs := now.Unix()

	strJson := args[0]

	//strJson := `{"Title":"titanic","city":"hyd"}`

	fmt.Println("strJson", strJson)

	var recObj interface{}
	errRead := json.Unmarshal([]byte(strJson), &recObj)

	if errRead != nil {
		log.Println("Error in read")
		//return shim.Error("Error in Putstate")
	}

	itemid := recObj.(map[string]interface{})

	fmt.Println("\nITEMID: ", itemid)
	fmt.Printf("\n%+v\n ", recObj)

	recordationeBytes, err2 := json.Marshal(recObj)

	if err2 != nil {
		fmt.Println("error in compliance creation", err2)
	} else {
		fmt.Println(string(recordationeBytes))
	}

	epid := strconv.FormatInt(secs, 10)
	err1 := stub.PutState("DLT"+epid, recordationeBytes)

	/*strid := itemid["userId"].(string)
	fmt.Println("strid == >", strid)
	err1 := stub.PutState(strid, recordationeBytes)
	*/
	if err1 != nil {
		log.Println("Error in putstate")
		return shim.Error("Error in Putstate")
	} else {
		err4 := stub.SetEvent("lineage", recordationeBytes)
		if err4 != nil {
			return shim.Error(fmt.Sprintf("Faild to emit event!"))
		} else {
			fmt.Println("\n lineage event fired.. !")
		}
	}

	fmt.Println("Putstate completed!")
	printmsg := "DLT" + epid + " item created!"
	return shim.Success([]byte(printmsg))
}

func (cc *Chaincode) GetLedger(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Insufficient arguments")
	}

	var userId, jsonResp string
	var err error

	userId = args[0]
	ComplainceDetails, err := stub.GetState(userId)

	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + userId + "\"}"
		return shim.Error(jsonResp)
	} else if ComplainceDetails == nil {
		jsonResp = "{\"Error\":\" does not exist: " + userId + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(ComplainceDetails)
}

func (cc *Chaincode) GetHistory(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("incorrect number of arguments")
	}

	repid := args[0]

	resultIterator, err := stub.GetHistoryForKey(repid)
	var jsonResp string
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + repid + "\"}"
		return shim.Error(jsonResp)
	} else if resultIterator == nil {
		jsonResp = "{\"Error\":\" does not exist: " + repid + "\"}"
		return shim.Error(jsonResp)
	}

	defer resultIterator.Close()

	fmt.Println("\nResult Iterator: ", resultIterator)
	var recObjs interface{}
	recArray := []interface{}{}
	n := 0

	for resultIterator.HasNext() {
		var recObj interface{}
		resObj, err2 := resultIterator.Next()
		if err2 != nil {
			fmt.Println(err2)
		}
		fmt.Println("resObj:==>  ", resObj)
		strJson := resObj.Value
		//strTrx := resObj.TxId
		errRead := json.Unmarshal([]byte(strJson), &recObj)
		if errRead != nil {
			log.Println("Error in read")
		}
		fmt.Println("recObj:==>  ", recObj)
		recObjs = recObj.(map[string]interface{})
		fmt.Println("\n counter:==>  ", n)
		fmt.Println("\n recObjs:==>  ", recObjs)
		//recArray[0] = recObj.(map[string]interface{})
		recArray = append(recArray, recObjs)
		n = n + 1
	}

	resultBytes, err := json.Marshal(recArray)

	if err != nil {
		fmt.Println("Error in marshalling")
		return shim.Error("Error in marshalling")
	}
	fmt.Println("ResultBytes:==>  ", string(resultBytes))
	return shim.Success(resultBytes)
}
