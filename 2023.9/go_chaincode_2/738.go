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

// Car :  Define the car structure, with 4 properties.  Structure tags are used by encoding/json library
type Token struct {
	Type  string `json:"type`
	Name  string `json:"name"`
	Rate string `json:"rate"`
	Owner  string `json:"owner"`
	Amount string `json:"amount"`
}

type tokenPrivateDetails struct {
	Owner string `json:"owner"`
	Value string `json:"value"`
}

// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("fabcar_cc")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "queryToken":
		return s.queryToken(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	case "createToken":
		return s.createToken(APIstub, args)
	case "queryAllTokens":
		return s.queryAllTokens(APIstub)
	case "changeTokenOwner":
		return s.changeTokenOwner(APIstub, args)
	case "getHistoryForAsset":
		return s.getHistoryForAsset(APIstub, args)
	case "queryTokensByOwner":
		return s.queryTokensByOwner(APIstub, args)
	case "restictedMethod":
		return s.restictedMethod(APIstub, args)
	case "test":
		return s.test(APIstub, args)
	case "createPrivateToken":
		return s.createPrivateToken(APIstub, args)
	case "readPrivateToken":
		return s.readPrivateToken(APIstub, args)
	case "updatePrivateData":
		return s.updatePrivateData(APIstub, args)
	case "readTokenPrivateDetails":
		return s.readTokenPrivateDetails(APIstub, args)
	case "createPrivateTokenImplicitForOrg1":
		return s.createPrivateTokenImplicitForOrg1(APIstub, args)
	case "createPrivateTokenImplicitForOrg2":
		return s.createPrivateTokenImplicitForOrg2(APIstub, args)
	case "queryPrivateDataHash":
		return s.queryPrivateDataHash(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}

	// return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryToken(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	tokenAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(tokenAsBytes)
}

func (s *SmartContract) readPrivateToken(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	// collectionTokens, collectionTokenPrivateDetails, _implicit_org_Org1MSP, _implicit_org_Org2MSP
	tokenAsBytes, err := APIstub.GetPrivateData(args[0], args[1])
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[1] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if tokenAsBytes == nil {
		jsonResp := "{\"Error\":\"token private details does not exist: " + args[1] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(tokenAsBytes)
}

func (s *SmartContract) readPrivateTokenIMpleciteForOrg1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	tokenAsBytes, _ := APIstub.GetPrivateData("_implicit_org_Org1MSP", args[0])
	return shim.Success(tokenAsBytes)
}

func (s *SmartContract) readTokenPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	tokenAsBytes, err := APIstub.GetPrivateData("collectionTokenPrivateDetails", args[0])

	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[0] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if tokenAsBytes == nil {
		jsonResp := "{\"Error\":\"Marble private details does not exist: " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(tokenAsBytes)
}

func (s *SmartContract) test(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	tokenAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(tokenAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	tokens := []Token{
		Token{Type: "chickenPoints", Name: "Sysco-Chicken-points", Rate: "10 points = 1 sysco token", Owner: "Tomoko" ,Amount:"100"},
		Token{Type: "applePoints", Name: "Sysco-Apple-points", Rate: "20 points = 1 sysco token", Owner: "Brad",Amount:"100"},
		Token{Type: "pizzaPoints", Name: "PizzaHut-Pizza-points", Rate: "5 points = 1 sysco token", Owner: "Jin Soo",Amount:"100"},
		Token{Type: "mangoPoints", Name: "mangoSupplier-mango-points", Rate: "40 points = 1 sysco token", Owner: "Max",Amount:"100"},
		Token{Type: "coffeePoints", Name: "Starbucks-Coffee-points", Rate: "50 points = 1 sysco token", Owner: "Adriana",Amount:"100"},
		Token{Type: "teaPoints", Name: "Sysco-tea-points", Rate: "40 points = 1 sysco token", Owner: "Michel",Amount:"100"},
		Token{Type: "utensilPoints", Name: "Sysco-Utensil-points", Rate: "7 points = 1 sysco token", Owner: "Aarav",Amount:"100"},
		Token{Type: "bananaPoints", Name: "bananaFarm-banana-points", Rate: "60 points = 1 sysco token", Owner: "Pari",Amount:"100"},
	
	}

	i := 0
	for i < len(tokens) {
		tokenAsBytes, _ := json.Marshal(tokens[i])
		APIstub.PutState("TOKEN"+strconv.Itoa(i), tokenAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createPrivateToken(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type tokenTransientInput struct {
		Type  string `json:"type"` //the fieldtags are needed to keep case from bouncing around
		Name string `json:"name"`
		Rate string `json:"rate"`
		Owner string `json:"owner"`
		Value string `json:"value"`
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

	tokenDataAsBytes, ok := transMap["token"]
	if !ok {
		return shim.Error("token must be a key in the transient map")
	}
	logger.Infof("********************8   " + string(tokenDataAsBytes))

	if len(tokenDataAsBytes) == 0 {
		return shim.Error("333333 -marble value in the transient map must be a non-empty JSON string")
	}

	logger.Infof("2222222")

	var tokenInput tokenTransientInput
	err = json.Unmarshal(tokenDataAsBytes, &tokenInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(tokenDataAsBytes) + "Error is : " + err.Error())
	}

	logger.Infof("3333")

	if len(tokenInput.Key) == 0 {
		return shim.Error("key field must be a non-empty string")
	}
	if len(tokenInput.Name) == 0 {
		return shim.Error("token name field must be a non-empty string")
	}
	if len(tokenInput.Type) == 0 {
		return shim.Error("type field must be a non-empty string")
	}
	if len(tokenInput.Rate) == 0 {
		return shim.Error("rate field must be a non-empty string")
	}
	if len(tokenInput.Owner) == 0 {
		return shim.Error("owner field must be a non-empty string")
	}
	if len(tokenInput.Value) == 0 {
		return shim.Error("value field must be a non-empty string")
	}

	logger.Infof("444444")

	// ==== Check if token already exists ====
	tokenAsBytes, err := APIstub.GetPrivateData("collectionTokens", tokenInput.Key)
	if err != nil {
		return shim.Error("Failed to get marble: " + err.Error())
	} else if tokenAsBytes != nil {
		fmt.Println("This token already exists: " + tokenInput.Key)
		return shim.Error("This token already exists: " + tokenInput.Key)
	}

	logger.Infof("55555")

	var token= Token{Name: tokenInput.Name, Type: tokenInput.Type, Rate: tokenInput.Rate, Owner: tokenInput.Owner}

	tokenAsBytes, err = json.Marshal(token)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.PutPrivateData("collectionTokens", tokenInput.Key, tokenAsBytes)
	if err != nil {
		logger.Infof("6666666")
		return shim.Error(err.Error())
	}
	tokenPrivateDetails := &tokenPrivateDetails{Owner: tokenInput.Owner, Value: tokenInput.Value}

	
	tokenPrivateDetailsAsBytes, err := json.Marshal(tokenPrivateDetails)
	if err != nil {
		logger.Infof("77777")
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("collectionTokenPrivateDetails", tokenInput.Key, tokenPrivateDetailsAsBytes)
	if err != nil {
		logger.Infof("888888")
		return shim.Error(err.Error())
	}

	return shim.Success(tokenAsBytes)
}

func (s *SmartContract) updatePrivateData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	type tokenTransientInput struct {
		Owner string `json:"owner"`
		Value string `json:"value`
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

	tokenDataAsBytes, ok := transMap["token"]
	if !ok {
		return shim.Error("token must be a key in the transient map")
	}
	logger.Infof("********************8   " + string(tokenDataAsBytes))

	if len(tokenDataAsBytes) == 0 {
		return shim.Error("333333 -marble value in the transient map must be a non-empty JSON string")
	}

	logger.Infof("2222222")

	var tokenInput tokenTransientInput
	err = json.Unmarshal(tokenDataAsBytes, &tokenInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(tokenDataAsBytes) + "Error is : " + err.Error())
	}

	tokenPrivateDetails := &tokenPrivateDetails{Owner: tokenInput.Owner, Value: tokenInput.Value}

	tokenPrivateDetailsAsBytes, err := json.Marshal(tokenPrivateDetails)
	if err != nil {
		logger.Infof("77777")
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("collectionTokenPrivateDetails", tokenInput.Key, tokenPrivateDetailsAsBytes)
	if err != nil {
		logger.Infof("888888")
		return shim.Error(err.Error())
	}

	return shim.Success(tokenPrivateDetailsAsBytes)

}

func (s *SmartContract) createToken(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6")
	}

	var token = Token{Name: args[1], Type: args[2], Rate: args[3], Owner: args[4], Amount: args[5]}

	tokenAsBytes, _ := json.Marshal(token)
	APIstub.PutState(args[0], tokenAsBytes)

	indexName := "owner~key"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{token.Owner, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)

	return shim.Success(tokenAsBytes)
}

func (S *SmartContract) queryTokensByOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

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

	var tokens []byte
	bArrayMemberAlreadyWritten := false

	tokens = append([]byte("["))

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
			tokens = append(tokens, newBytes...)

		} else {
			// newBytes := append([]byte(","), carsAsBytes...)
			tokens = append(tokens, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	tokens = append(tokens, []byte("]")...)

	return shim.Success(tokens)
}

func (s *SmartContract) queryAllTokens(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "TOKEN0"
	endKey := "TOKEN999"

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

	fmt.Printf("- queryAllTokens:\n%s\n", buffer.String())

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

		tokenAsBytes, _ := APIstub.GetState(args[0])
		return shim.Success(tokenAsBytes)
	}

}

func (s *SmartContract) changeTokenOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	tokenAsBytes, _ := APIstub.GetState(args[0])
	token := Token{}

	json.Unmarshal(tokenAsBytes, &token)
	token.Owner = args[1]

	tokenAsBytes, _ = json.Marshal(token)
	APIstub.PutState(args[0], tokenAsBytes)

	return shim.Success(tokenAsBytes)
}

func (t *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	tokenName := args[0]

	resultsIterator, err := stub.GetHistoryForKey(tokenName)
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

func (s *SmartContract) createPrivateTokenImplicitForOrg1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect arguments. Expecting 5 arguments")
	}

	var token = Token{Name: args[1], Type: args[2], Rate: args[3], Owner: args[4]}

	tokenAsBytes, _ := json.Marshal(token)
	// APIstub.PutState(args[0], carAsBytes)

	err := APIstub.PutPrivateData("_implicit_org_Org1MSP", args[0], tokenAsBytes)
	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	return shim.Success(tokenAsBytes)
}

func (s *SmartContract) createPrivateTokenImplicitForOrg2(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect arguments. Expecting 5 arguments")
	}

	var token = Token{Name: args[1], Type: args[2], Rate: args[3], Owner: args[4]}

	tokenAsBytes, _ := json.Marshal(token)
	APIstub.PutState(args[0], tokenAsBytes)

	err := APIstub.PutPrivateData("_implicit_org_Org2MSP", args[0], tokenAsBytes)
	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	return shim.Success(tokenAsBytes)
}

func (s *SmartContract) queryPrivateDataHash(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	tokenAsBytes, _ := APIstub.GetPrivateDataHash(args[0], args[1])
	return shim.Success(tokenAsBytes)
}

// func (s *SmartContract) CreateCarAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
// 	if len(args) != 1 {
// 		return shim.Error("Incorrect number of arguments. Expecting 1")
// 	}

// 	var car Car
// 	err := json.Unmarshal([]byte(args[0]), &car)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	carAsBytes, err := json.Marshal(car)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	err = APIstub.PutState(car.ID, carAsBytes)
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
// 		var car = Car{Make: s[1], Model: s[2], Colour: s[3], Owner: s[4]}

// 		eventKeyValue = strings.SplitN(s, "#", 3)
// 		if len(eventKeyValue) != 3 {
// 			logger.Errorf("Error occured, Please make sure that you have provided the array of strings and each string should be  in \"EventType#Key#Value\" format")
// 			return shim.Error("Error occured, Please make sure that you have provided the array of strings and each string should be  in \"EventType#Key#Value\" format")
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
