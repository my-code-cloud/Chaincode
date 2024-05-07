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

// User Basic data  Define the Datavalut structure, with 4 properties.  Structure tags are used by encoding/json library
type Datavalut struct {
	Area  string `json:"area"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Owner string `json:"owner"`
}

type DatavalutPrivateDetails struct {
	Owner string `json:"owner"`
	Phone string `json:"price"`
}

// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("datavalut_cc")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "queryDatavalut":
		return s.queryDatavalut(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	case "createDatavalut":
		return s.createDatavalut(APIstub, args)
	case "queryAllDatavaluts":
		return s.queryAllDatavaluts(APIstub)
	case "changeDatavalutOwner":
		return s.changeDatavalutOwner(APIstub, args)
	case "getHistoryForAsset":
		return s.getHistoryForAsset(APIstub, args)
	case "queryDatavalutByOwner":
		return s.queryDatavalutsByOwner(APIstub, args)
	case "restictedMethod":
		return s.restictedMethod(APIstub, args)
	case "test":
		return s.test(APIstub, args)
	case "createPrivateDatavalut":
		return s.createPrivateDatavalut(APIstub, args)
	case "readPrivateDatavalut":
		return s.readPrivateDatavalut(APIstub, args)
	case "updatePrivateDatavalut":
		return s.updatePrivateDatavalut(APIstub, args)
	case "readDatavalutPrivateDetails":
		return s.readDatavalutPrivateDetails(APIstub, args)
	case "createPrivateDatavalutImplicitForOrg1":
		return s.createPrivateDatavalutImplicitForOrg1(APIstub, args)
	case "createPrivateDatavalutImplicitForOrg2":
		return s.createPrivateDatavalutImplicitForOrg2(APIstub, args)
	case "queryPrivateDataHash":
		return s.queryPrivateDataHash(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}

	// return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryDatavalut(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	DatavalutAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(DatavalutAsBytes)
}

//test method for upgrade chaincode

func (s *SmartContract) test(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	DatavalutAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(DatavalutAsBytes)
}

//end of test method

func (s *SmartContract) readPrivateDatavalut(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	// collectionDatavaluts, collectionDatavalutPrivateDetails, _implicit_org_Org1MSP, _implicit_org_Org2MSP
	DatavalutAsBytes, err := APIstub.GetPrivateData(args[0], args[1])
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[1] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if DatavalutAsBytes == nil {
		jsonResp := "{\"Error\":\"Datavalut private details does not exist: " + args[1] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(DatavalutAsBytes)
}

func (s *SmartContract) readPrivateDatavalutIMpleciteForOrg1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	DatavalutAsBytes, _ := APIstub.GetPrivateData("_implicit_org_Org1MSP", args[0])
	return shim.Success(DatavalutAsBytes)
}

func (s *SmartContract) readDatavalutPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	DatavalutAsBytes, err := APIstub.GetPrivateData("collectionDatavalutPrivateDetails", args[0])

	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[0] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if DatavalutAsBytes == nil {
		jsonResp := "{\"Error\":\"Marble private details does not exist: " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(DatavalutAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	Datavaluts := []Datavalut{
		Datavalut{Area: "SE0", Email: "se0@gmail.com", Phone: "+91 9988998899", Owner: "Geo"},
		Datavalut{Area: "SE1", Email: "se1@gmail.com", Phone: "+91 9922992299", Owner: "Manuel"},
		Datavalut{Area: "SE2", Email: "se2@gmail.com", Phone: "+91 9933993399", Owner: "Faizal"},
	}

	i := 0
	for i < len(Datavaluts) {
		DatavalutAsBytes, _ := json.Marshal(Datavaluts[i])
		APIstub.PutState("Datavalut"+strconv.Itoa(i), DatavalutAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createPrivateDatavalut(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type DatavalutTransientInput struct {
		Area  string `json:"area"` //the fieldtags are needed to keep case from bouncing around
		Email string `json:"email"`
		Phone string `json:"phone"`
		Owner string `json:"owner"`
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

	DatavalutDataAsBytes, ok := transMap["Datavalut"]
	if !ok {
		return shim.Error("Datavalut must be a key in the transient map")
	}
	logger.Infof("********************8   " + string(DatavalutDataAsBytes))

	if len(DatavalutDataAsBytes) == 0 {
		return shim.Error("333333 -marble value in the transient map must be a non-empty JSON string")
	}

	logger.Infof("2222222")

	var DatavalutInput DatavalutTransientInput
	err = json.Unmarshal(DatavalutDataAsBytes, &DatavalutInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(DatavalutDataAsBytes) + "Error is : " + err.Error())
	}

	logger.Infof("3333")

	if len(DatavalutInput.Key) == 0 {
		return shim.Error("name field must be a non-empty string")
	}
	if len(DatavalutInput.Area) == 0 {
		return shim.Error("Phone field must be a non-empty string")
	}
	if len(DatavalutInput.Email) == 0 {
		return shim.Error("Email field must be a non-empty string")
	}
	if len(DatavalutInput.Phone) == 0 {
		return shim.Error("Phone field must be a non-empty string")
	}
	if len(DatavalutInput.Owner) == 0 {
		return shim.Error("owner field must be a non-empty string")
	}

	logger.Infof("444444")

	// ==== Check if Datavalut already exists ====
	DatavalutAsBytes, err := APIstub.GetPrivateData("collectionDatavaluts", DatavalutInput.Key)
	if err != nil {
		return shim.Error("Failed to get marble: " + err.Error())
	} else if DatavalutAsBytes != nil {
		fmt.Println("This Datavalut already exists: " + DatavalutInput.Key)
		return shim.Error("This Datavalut already exists: " + DatavalutInput.Key)
	}

	logger.Infof("55555")

	var Datavalut = Datavalut{Area: DatavalutInput.Area, Email: DatavalutInput.Email, Phone: DatavalutInput.Phone, Owner: DatavalutInput.Owner}

	DatavalutAsBytes, err = json.Marshal(Datavalut)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.PutPrivateData("collectionDatavaluts", DatavalutInput.Key, DatavalutAsBytes)
	if err != nil {
		logger.Infof("6666666")
		return shim.Error(err.Error())
	}

	DatavalutPrivateDetails := &DatavalutPrivateDetails{Owner: DatavalutInput.Owner}

	DatavalutPrivateDetailsAsBytes, err := json.Marshal(DatavalutPrivateDetails)
	if err != nil {
		logger.Infof("77777")
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("collectionDatavalutPrivateDetails", DatavalutInput.Key, DatavalutPrivateDetailsAsBytes)
	if err != nil {
		logger.Infof("888888")
		return shim.Error(err.Error())
	}

	return shim.Success(DatavalutAsBytes)
}

func (s *SmartContract) updatePrivateDatavalut(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	type DatavalutTransientInput struct {
		Owner string `json:"owner"`
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

	DatavalutDataAsBytes, ok := transMap["Datavalut"]
	if !ok {
		return shim.Error("Datavalut must be a key in the transient map")
	}
	logger.Infof("********************8   " + string(DatavalutDataAsBytes))

	if len(DatavalutDataAsBytes) == 0 {
		return shim.Error("333333 -marble value in the transient map must be a non-empty JSON string")
	}

	logger.Infof("2222222")

	var DatavalutInput DatavalutTransientInput
	err = json.Unmarshal(DatavalutDataAsBytes, &DatavalutInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(DatavalutDataAsBytes) + "Error is : " + err.Error())
	}

	DatavalutPrivateDetails := &DatavalutPrivateDetails{Owner: DatavalutInput.Owner}

	DatavalutPrivateDetailsAsBytes, err := json.Marshal(DatavalutPrivateDetails)
	if err != nil {
		logger.Infof("77777")
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("collectionDatavalutPrivateDetails", DatavalutInput.Key, DatavalutPrivateDetailsAsBytes)
	if err != nil {
		logger.Infof("888888")
		return shim.Error(err.Error())
	}

	return shim.Success(DatavalutPrivateDetailsAsBytes)

}

func (s *SmartContract) createDatavalut(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var Datavalut = Datavalut{Area: args[1], Email: args[2], Phone: args[3], Owner: args[4]}

	DatavalutAsBytes, _ := json.Marshal(Datavalut)
	APIstub.PutState(args[0], DatavalutAsBytes)

	indexName := "owner~key"
	PhoneNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{Datavalut.Owner, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(PhoneNameIndexKey, value)

	return shim.Success(DatavalutAsBytes)
}

func (S *SmartContract) queryDatavalutsByOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

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

	var Datavaluts []byte
	bArrayMemberAlreadyWritten := false

	Datavaluts = append([]byte("["))

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
			Datavaluts = append(Datavaluts, newBytes...)

		} else {
			// newBytes := append([]byte(","), DatavalutsAsBytes...)
			Datavaluts = append(Datavaluts, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	Datavaluts = append(Datavaluts, []byte("]")...)

	return shim.Success(Datavaluts)
}

func (s *SmartContract) queryAllDatavaluts(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "Datavalut0"
	endKey := "Datavalut999"

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

	fmt.Printf("- queryAllDatavaluts:\n%s\n", buffer.String())

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

		DatavalutAsBytes, _ := APIstub.GetState(args[0])
		return shim.Success(DatavalutAsBytes)
	}

}

func (s *SmartContract) changeDatavalutOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	DatavalutAsBytes, _ := APIstub.GetState(args[0])
	Datavalut := Datavalut{}

	json.Unmarshal(DatavalutAsBytes, &Datavalut)
	Datavalut.Owner = args[1]

	DatavalutAsBytes, _ = json.Marshal(Datavalut)
	APIstub.PutState(args[0], DatavalutAsBytes)

	return shim.Success(DatavalutAsBytes)
}

func (t *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	DatavalutName := args[0]

	resultsIterator, err := stub.GetHistoryForKey(DatavalutName)
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

func (s *SmartContract) createPrivateDatavalutImplicitForOrg1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect arguments. Expecting 5 arguments")
	}

	var Datavalut = Datavalut{Area: args[1], Email: args[2], Phone: args[3], Owner: args[4]}

	DatavalutAsBytes, _ := json.Marshal(Datavalut)
	// APIstub.PutState(args[0], DatavalutAsBytes)

	err := APIstub.PutPrivateData("_implicit_org_Org1MSP", args[0], DatavalutAsBytes)
	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	return shim.Success(DatavalutAsBytes)
}

func (s *SmartContract) createPrivateDatavalutImplicitForOrg2(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect arguments. Expecting 5 arguments")
	}

	var Datavalut = Datavalut{Area: args[1], Email: args[2], Phone: args[3], Owner: args[4]}

	DatavalutAsBytes, _ := json.Marshal(Datavalut)
	APIstub.PutState(args[0], DatavalutAsBytes)

	err := APIstub.PutPrivateData("_implicit_org_Org2MSP", args[0], DatavalutAsBytes)
	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	return shim.Success(DatavalutAsBytes)
}

func (s *SmartContract) queryPrivateDataHash(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	DatavalutAsBytes, _ := APIstub.GetPrivateDataHash(args[0], args[1])
	return shim.Success(DatavalutAsBytes)
}

// func (s *SmartContract) CreateDatavalutAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
// 	if len(args) != 1 {
// 		return shim.Error("Incorrect number of arguments. Expecting 1")
// 	}

// 	var Datavalut Datavalut
// 	err := json.Unmarshal([]byte(args[0]), &Datavalut)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	DatavalutAsBytes, err := json.Marshal(Datavalut)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	err = APIstub.PutState(Datavalut.ID, DatavalutAsBytes)
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
// 		var Datavalut = Datavalut{Area: s[1], Email: s[2], Phone: s[3], Owner: s[4]}

// 		eventKeyValue = strings.SplitN(s, "#", 3)
// 		if len(eventKeyValue) != 3 {
// 			logger.Errorf("Error occured, Please Area sure that you have provided the array of strings and each string should be  in \"EventType#Key#Value\" format")
// 			return shim.Error("Error occured, Please Area sure that you have provided the array of strings and each string should be  in \"EventType#Key#Value\" format")
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
