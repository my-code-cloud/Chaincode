package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}

// IC :  Define the ic structure, with 4 properties.  Structure tags are used by encoding/json library
type IC struct {
	Identifier   string `json:"identifier"`
	Type  string `json:"type"`
	CRP string `json:"crp"`
	Owner  string `json:"owner"`
}

type icPrivateDetails struct {
	Owner string `json:"owner"`
	Price string `json:"price"`
}

// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("traceic_cc")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "queryIC":
		return s.queryIC(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	case "createIC":
		return s.createIC(APIstub, args)
	case "queryAllICs":
		return s.queryAllICs(APIstub)
	case "changeICOwner":
		return s.changeICOwner(APIstub, args)
	case "getHistoryForAsset":
		return s.getHistoryForAsset(APIstub, args)
	case "queryICsByOwner":
		return s.queryICsByOwner(APIstub, args)
	case "restictedMethod":
		return s.restictedMethod(APIstub, args)
	case "test":
		return s.test(APIstub, args)
	case "createPrivateIC":
		return s.createPrivateIC(APIstub, args)
	case "readPrivateIC":
		return s.readPrivateIC(APIstub, args)
	case "updatePrivateData":
		return s.updatePrivateData(APIstub, args)
	case "readICPrivateDetails":
		return s.readICPrivateDetails(APIstub, args)
	case "createPrivateICImplicitForOrg1":
		return s.createPrivateICImplicitForOrg1(APIstub, args)
	case "createPrivateICImplicitForOrg2":
		return s.createPrivateICImplicitForOrg2(APIstub, args)
	case "queryPrivateDataHash":
		return s.queryPrivateDataHash(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}

	// return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryIC(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	icAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(icAsBytes)
}

func (s *SmartContract) readPrivateIC(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	// collectionICs, collectionICPrivateDetails, _implicit_org_Org1MSP, _implicit_org_Org2MSP
	icAsBytes, err := APIstub.GetPrivateData(args[0], args[1])
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[1] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if icAsBytes == nil {
		jsonResp := "{\"Error\":\"IC private details does not exist: " + args[1] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(icAsBytes)
}

func (s *SmartContract) readPrivateICIMpleciteForOrg1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	icAsBytes, _ := APIstub.GetPrivateData("_implicit_org_Org1MSP", args[0])
	return shim.Success(icAsBytes)
}

func (s *SmartContract) readICPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	icAsBytes, err := APIstub.GetPrivateData("collectionICPrivateDetails", args[0])

	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[0] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if icAsBytes == nil {
		jsonResp := "{\"Error\":\"Marble private details does not exist: " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(icAsBytes)
}

func (s *SmartContract) test(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	icAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(icAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	ics := []IC{
		IC{Identifier: "Aa", Type: "Laptop", CRP: "1234", Owner: "Org1"},
		IC{Identifier: "Bb", Type: "Desktop", CRP: "2345", Owner: "Org1"},
		IC{Identifier: "Cc", Type: "Gaming", CRP: "3456", Owner: "Org1"},
		IC{Identifier: "Dd", Type: "Server", CRP: "4567", Owner: "Org1"},
		IC{Identifier: "Ee", Type: "Mobile", CRP: "5678", Owner: "Org2"},
		IC{Identifier: "Ff", Type: "Iot", CRP: "6789", Owner: "Org2"},
		IC{Identifier: "Gg", Type: "Network", CRP: "7890", Owner: "Org2"},
		IC{Identifier: "Hh", Type: "Neuro", CRP: "8901", Owner: "Org3"},
		IC{Identifier: "Ii", Type: "Auto", CRP: "9012", Owner: "Org3"},
		IC{Identifier: "Jj", Type: "Tablet", CRP: "0123", Owner: "Org3"},
	}

	i := 0
	for i < len(ics) {
		icAsBytes, _ := json.Marshal(ics[i])
		APIstub.PutState("IC"+strconv.Itoa(i), icAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createPrivateIC(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type icTransientInput struct {
		Identifier  string `json:"identifier"` //the fieldtags are needed to keep case from bouncing around
		Type string `json:"type"`
		CRP string `json:"crp"`
		Owner string `json:"owner"`
		Price string `json:"price"`
		Key   string `json:"key"`
	}
	if len(args) != 0 {
		return shim.Error("1111111----Incorrect number of arguments. Private marble data must be passed in transient map.")
	}

	logger.Infof("11111111111111111111111111")

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}

	icDataAsBytes, ok := transMap["ic"]
	if !ok {
		return shim.Error("ic must be a key in the transient map")
	}
	logger.Infof("********************8   " + string(icDataAsBytes))

	if len(icDataAsBytes) == 0 {
		return shim.Error("333333 -marble value in the transient map must be a non-empty JSON string")
	}

	logger.Infof("2222222")

	var icInput icTransientInput
	err = json.Unmarshal(icDataAsBytes, &icInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(icDataAsBytes) + "Error is : " + err.Error())
	}

	logger.Infof("3333")

	if len(icInput.Key) == 0 {
		return shim.Error("name field must be a non-empty string")
	}
	if len(icInput.Identifier) == 0 {
		return shim.Error("crp field must be a non-empty string")
	}
	if len(icInput.Type) == 0 {
		return shim.Error("type field must be a non-empty string")
	}
	if len(icInput.CRP) == 0 {
		return shim.Error("crp field must be a non-empty string")
	}
	if len(icInput.Owner) == 0 {
		return shim.Error("owner field must be a non-empty string")
	}
	if len(icInput.Price) == 0 {
		return shim.Error("price field must be a non-empty string")
	}

	logger.Infof("444444")

	// ==== Check if ic already exists ====
	icAsBytes, err := APIstub.GetPrivateData("collectionICs", icInput.Key)
	if err != nil {
		return shim.Error("Failed to get marble: " + err.Error())
	} else if icAsBytes != nil {
		fmt.Println("This ic already exists: " + icInput.Key)
		return shim.Error("This ic already exists: " + icInput.Key)
	}

	logger.Infof("55555")

	var ic = IC{Identifier: icInput.Identifier, Type: icInput.Type, CRP: icInput.CRP, Owner: icInput.Owner}

	icAsBytes, err = json.Marshal(ic)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.PutPrivateData("collectionICs", icInput.Key, icAsBytes)
	if err != nil {
		logger.Infof("6666666")
		return shim.Error(err.Error())
	}

	icPrivateDetails := &icPrivateDetails{Owner: icInput.Owner, Price: icInput.Price}

	icPrivateDetailsAsBytes, err := json.Marshal(icPrivateDetails)
	if err != nil {
		logger.Infof("77777")
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("collectionICPrivateDetails", icInput.Key, icPrivateDetailsAsBytes)
	if err != nil {
		logger.Infof("888888")
		return shim.Error(err.Error())
	}

	return shim.Success(icAsBytes)
}

func (s *SmartContract) updatePrivateData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	type icTransientInput struct {
		Owner string `json:"owner"`
		Price string `json:"price"`
		Key   string `json:"key"`
	}
	if len(args) != 0 {
		return shim.Error("1111111----Incorrect number of arguments. Private marble data must be passed in transient map.")
	}

	logger.Infof("11111111111111111111111111")

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}

	icDataAsBytes, ok := transMap["ic"]
	if !ok {
		return shim.Error("ic must be a key in the transient map")
	}
	logger.Infof("********************8   " + string(icDataAsBytes))

	if len(icDataAsBytes) == 0 {
		return shim.Error("333333 -marble value in the transient map must be a non-empty JSON string")
	}

	logger.Infof("2222222")

	var icInput icTransientInput
	err = json.Unmarshal(icDataAsBytes, &icInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(icDataAsBytes) + "Error is : " + err.Error())
	}

	icPrivateDetails := &icPrivateDetails{Owner: icInput.Owner, Price: icInput.Price}

	icPrivateDetailsAsBytes, err := json.Marshal(icPrivateDetails)
	if err != nil {
		logger.Infof("77777")
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("collectionICPrivateDetails", icInput.Key, icPrivateDetailsAsBytes)
	if err != nil {
		logger.Infof("888888")
		return shim.Error(err.Error())
	}

	return shim.Success(icPrivateDetailsAsBytes)

}

func (s *SmartContract) createIC(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var ic = IC{Identifier: args[1], Type: args[2], CRP: args[3], Owner: args[4]}

	icAsBytes, _ := json.Marshal(ic)
	APIstub.PutState(args[0], icAsBytes)

	indexName := "owner~key"
	crpNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{ic.Owner, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(crpNameIndexKey, value)

	return shim.Success(icAsBytes)
}

func (S *SmartContract) queryICsByOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	owner := args[0]

	ownerAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("owner~key", []string{owner})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIdResultIterator.Close()

	var i int
	var id string

	var ics []byte
	bArrayMemberAlreadyWritten := false

	ics = append([]byte("["))

	for i = 0; ownerAndIdResultIterator.HasNext(); i++ {
		responseRange, err := ownerAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositeKeyParts[1]
		assetAsBytes, err := APIstub.GetState(id)

		if bArrayMemberAlreadyWritten == true {
			newBytes := append([]byte(","), assetAsBytes...)
			ics = append(ics, newBytes...)

		} else {
			// newBytes := append([]byte(","), icsAsBytes...)
			ics = append(ics, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	ics = append(ics, []byte("]")...)

	return shim.Success(ics)
}

func (s *SmartContract) queryAllICs(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "IC0"
	endKey := "IC999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
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
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllICs:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) restictedMethod(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// get an ID for the client which is guaranteed to be unique within the MSP
	//id, err := cid.GetID(APIstub) -

	// get the MSP ID of the client's identity
	//mspid, err := cid.GetMSPID(APIstub) -

	// get the value of the attribute
	//val, ok, err := cid.GetAttributeValue(APIstub, "attr1") -

	// get the X509 certificate of the client, or nil if the client's identity was not based on an X509 certificate
	//cert, err := cid.GetX509Certificate(APIstub) -

	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "approver" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as APPROVER have access this method!")
	} else {
		if len(args) != 1 {
			return shim.Error("Incorrect number of arguments. Expecting 1")
		}

		icAsBytes, _ := APIstub.GetState(args[0])
		return shim.Success(icAsBytes)
	}

}

func (s *SmartContract) changeICOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	icAsBytes, _ := APIstub.GetState(args[0])
	ic := IC{}

	json.Unmarshal(icAsBytes, &ic)
	ic.Owner = args[1]

	icAsBytes, _ = json.Marshal(ic)
	APIstub.PutState(args[0], icAsBytes)

	return shim.Success(icAsBytes)
}

func (t *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	icName := args[0]

	resultsIterator, err := stub.GetHistoryForKey(icName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForAsset returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) createPrivateICImplicitForOrg1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect arguments. Expecting 5 arguments")
	}

	var ic = IC{Identifier: args[1], Type: args[2], CRP: args[3], Owner: args[4]}

	icAsBytes, _ := json.Marshal(ic)
	// APIstub.PutState(args[0], icAsBytes)

	err := APIstub.PutPrivateData("_implicit_org_Org1MSP", args[0], icAsBytes)
	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	return shim.Success(icAsBytes)
}

func (s *SmartContract) createPrivateICImplicitForOrg2(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect arguments. Expecting 5 arguments")
	}

	var ic = IC{Identifier: args[1], Type: args[2], CRP: args[3], Owner: args[4]}

	icAsBytes, _ := json.Marshal(ic)
	APIstub.PutState(args[0], icAsBytes)

	err := APIstub.PutPrivateData("_implicit_org_Org2MSP", args[0], icAsBytes)
	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	return shim.Success(icAsBytes)
}

func (s *SmartContract) queryPrivateDataHash(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	icAsBytes, _ := APIstub.GetPrivateDataHash(args[0], args[1])
	return shim.Success(icAsBytes)
}

// func (s *SmartContract) CreateICAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
// 	if len(args) != 1 {
// 		return shim.Error("Incorrect number of arguments. Expecting 1")
// 	}

// 	var ic IC
// 	err := json.Unmarshal([]byte(args[0]), &ic)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	icAsBytes, err := json.Marshal(ic)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	err = APIstub.PutState(ic.ID, icAsBytes)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	return shim.Success(nil)
// }

// func (s *SmartContract) addBulkAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
// 	logger.Infof("Function addBulkAsset called and length of arguments is:  %d", len(args))
// 	if len(args) >= 500 {
// 		logger.Errorf("Incorrect number of arguments in function CreateAsset, expecting less than 500, but got: %b", len(args))
// 		return shim.Error("Incorrect number of arguments, expecting 2")
// 	}

// 	var eventKeyValue []string

// 	for i, s := range args {

// 		key :=s[0];
// 		var ic = IC{Identifier: s[1], Type: s[2], CRP: s[3], Owner: s[4]}

// 		eventKeyValue = strings.SplitN(s, "#", 3)
// 		if len(eventKeyValue) != 3 {
// 			logger.Errorf("Error occured, Please identifier sure that you have provided the array of strings and each string should be  in \"EventType#Key#Value\" format")
// 			return shim.Error("Error occured, Please identifier sure that you have provided the array of strings and each string should be  in \"EventType#Key#Value\" format")
// 		}

// 		assetAsBytes := []byte(eventKeyValue[2])
// 		err := APIstub.PutState(eventKeyValue[1], assetAsBytes)
// 		if err != nil {
// 			logger.Errorf("Error coocured while putting state for asset %s in APIStub, error: %s", eventKeyValue[1], err.Error())
// 			return shim.Error(err.Error())
// 		}
// 		// logger.infof("Adding value for ")
// 		fmt.Println(i, s)

// 		indexName := "Event~Id"
// 		eventAndIDIndexKey, err2 := APIstub.CreateCompositeKey(indexName, []string{eventKeyValue[0], eventKeyValue[1]})

// 		if err2 != nil {
// 			logger.Errorf("Error coocured while putting state in APIStub, error: %s", err.Error())
// 			return shim.Error(err2.Error())
// 		}

// 		value := []byte{0x00}
// 		err = APIstub.PutState(eventAndIDIndexKey, value)
// 		if err != nil {
// 			logger.Errorf("Error coocured while putting state in APIStub, error: %s", err.Error())
// 			return shim.Error(err.Error())
// 		}
// 		// logger.Infof("Created Composite key : %s", eventAndIDIndexKey)

// 	}

// 	return shim.Success(nil)
// }

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}