package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	// "strconv"
	// "time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	// "github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}

// Trade :  Define the trade structure, with 6 properties.  Structure tags are used by encoding/json library
type Trade struct {
	TradeId   string `json:"tradeId"`
	FromParty string `json:"fromParty"`
	ToParty   string `json:"toParty"`
	Amount    string `json:"amount"`
	TradeDate string `json:"tradeDate"`
	Status    string `json:"status"`
}

// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("fabtxn_cc")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "readPrivateTradebyId":
		return s.readPrivateTradebyId(APIstub, args)
	case "createPrivateTrade":
		return s.createPrivateTrade(APIstub, args)
	case "readPrivateTrade":
		return s.readPrivateTrade(APIstub, args)
	case "updatePrivateData":
		return s.updatePrivateData(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}

func (s *SmartContract) createPrivateTrade(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type TradeTransientInput struct {
		TradeId   string `json:"tradeId"`
		FromParty string `json:"fromParty"`
		ToParty   string `json:"toParty"`
		Amount    string `json:"amount"`
		TradeDate string `json:"tradeDate"`
		Status    string `json:"status"`
	}
	if len(args) != 0 {
		return shim.Error("1111111-----Incorrect number of arguments. Private trade data must be passed in transient map.")
	}

	logger.Infof("11111111111111111111111111")

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}

	tradeDataAsBytes, ok := transMap["trade"]
	if !ok {
		return shim.Error("trade must be a key in the transient map")
	}
	logger.Infof("********************8   " + string(tradeDataAsBytes))

	if len(tradeDataAsBytes) == 0 {
		return shim.Error("333333 -marble value in the transient map must be a non-empty JSON string")
	}

	logger.Infof("2222222")

	var tradeInput TradeTransientInput
	err = json.Unmarshal(tradeDataAsBytes, &tradeInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(tradeDataAsBytes) + "Error is : " + err.Error())
	}

	logger.Infof("3333")

	if len(tradeInput.TradeId) == 0 {
		return shim.Error("TradeId field must be a non-empty string")
	}
	if len(tradeInput.FromParty) == 0 {
		return shim.Error("FromParty field must be a non-empty string")
	}
	if len(tradeInput.ToParty) == 0 {
		return shim.Error("ToParty field must be a non-empty string")
	}
	if len(tradeInput.Amount) == 0 {
		return shim.Error("Amount field must be a non-empty string")
	}
	if len(tradeInput.TradeDate) == 0 {
		return shim.Error("TradeDate field must be a non-empty string")
	}
	if len(tradeInput.Status) == 0 {
		return shim.Error("Status field must be a non-empty string")
	}
	// logger.Infof(tradeInput);

	var fromP string
	fromP = tradeInput.FromParty

	var toP string
	toP = tradeInput.ToParty

	var fncName string

	logger.Infof("From party is ", fromP)
	logger.Infof("To Party is ", toP)

	logger.Infof("444444")

	// ==== Check if trade already exists ====
	if (fromP == "Org1" && toP == "Org3") || (fromP == "Org3" && toP == "Org1") {
		fncName = "collectionTx13"
	} else if (fromP == "Org1" && toP == "Org2") || (fromP == "Org2" && toP == "Org1") {
		fncName = "collectionTx12"
	} else if (fromP == "Org2" && toP == "Org3") || (fromP == "Org3" && toP == "Org2") {
		fncName = "collectionTx23"
	}
	logger.Infof(fncName)

	// ==== Check if car already exists ====
	tradeAsBytes, err := APIstub.GetPrivateData(fncName, tradeInput.TradeId)
	if err != nil {
		return shim.Error("Failed to get trade: " + err.Error())
	} else if tradeAsBytes != nil {
		fmt.Println("This trade already exists: " + tradeInput.TradeId)
		return shim.Error("This trade already exists: " + tradeInput.TradeId)
	}

	logger.Infof("55555")

	var trade = Trade{TradeId: tradeInput.TradeId, FromParty: tradeInput.FromParty, ToParty: tradeInput.ToParty, Amount: tradeInput.Amount, TradeDate: tradeInput.TradeDate, Status: tradeInput.Status}

	tradeAsBytes, err = json.Marshal(trade)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.PutPrivateData(fncName, tradeInput.TradeId, tradeAsBytes)
	if err != nil {
		logger.Infof("6666666")
		return shim.Error(err.Error())
	}

	return shim.Success(tradeAsBytes)
}

func (s *SmartContract) readPrivateTrade(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// collectionCars, collectionCarPrivateDetails, _implicit_org_Org1MSP, _implicit_org_Org2MSP
	resultsIterator, err := APIstub.GetPrivateDataByRange(args[0], "", "") //return unboundedly
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		// buffer.WriteString("{")
		// buffer.WriteString("\"")
		// buffer.WriteString(queryResponse.Key)
		// buffer.WriteString("\"")

		// buffer.WriteString(", \"Record\":")
		// // Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		// buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllPrivateTrades:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) readPrivateTradebyId(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	tradeAsBytes, err := APIstub.GetPrivateData(args[0], args[1])
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[1] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if tradeAsBytes == nil {
		jsonResp := "{\"Error\":\"Trade private details does not exist: " + args[1] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(tradeAsBytes)
}

func (s *SmartContract) updatePrivateData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	type TradeTransientInput struct {
		TradeId   string `json:"tradeId"`
		FromParty string `json:"fromParty"`
		ToParty   string `json:"toParty"`
		Amount    string `json:"amount"`
		TradeDate string `json:"tradeDate"`
		Status    string `json:"status"`
	}
	if len(args) != 0 {
		return shim.Error("1111111----Incorrect number of arguments. Private trade data must be passed in transient map.")
	}

	logger.Infof("11111111111111111111111111")

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}

	tradeDataAsBytes, ok := transMap["trade"]
	if !ok {
		return shim.Error("car must be a key in the transient map")
	}
	logger.Infof("********************8   " + string(tradeDataAsBytes))

	if len(tradeDataAsBytes) == 0 {
		return shim.Error("333333 -marble value in the transient map must be a non-empty JSON string")
	}

	logger.Infof("2222222")

	var tradeInput TradeTransientInput
	err = json.Unmarshal(tradeDataAsBytes, &tradeInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(tradeDataAsBytes) + "Error is : " + err.Error())
	}
	var fromP string
	fromP = tradeInput.FromParty

	var toP string
	toP = tradeInput.ToParty

	var fncName string

	logger.Infof("From party is ", fromP)
	logger.Infof("To Party is ", toP)

	logger.Infof("444444")

	// ==== Check if trade already exists ====
	if (fromP == "Org1" && toP == "Org3") || (fromP == "Org3" && toP == "Org1") {
		fncName = "collectionTx13"
	} else if (fromP == "Org1" && toP == "Org2") || (fromP == "Org2" && toP == "Org1") {
		fncName = "collectionTx12"
	} else if (fromP == "Org2" && toP == "Org3") || (fromP == "Org3" && toP == "Org2") {
		fncName = "collectionTx23"
	}
	logger.Infof(fncName)
	//verifying for existence
	tradeAsBytes, err := APIstub.GetPrivateData(fncName, tradeInput.TradeId)
	logger.Infof("55555")
	var trade = Trade{TradeId: tradeInput.TradeId, FromParty: tradeInput.FromParty, ToParty: tradeInput.ToParty, Amount: tradeInput.Amount, TradeDate: tradeInput.TradeDate, Status: tradeInput.Status}
	tradeAsBytes, err = json.Marshal(trade)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData(fncName, tradeInput.TradeId, tradeAsBytes)
	if err != nil {
		logger.Infof("6666666")
		return shim.Error(err.Error())
	}

	return shim.Success(tradeAsBytes)

}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {
	fmt.Printf("Inside main function")
	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
