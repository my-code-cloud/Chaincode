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
)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}

// Car :  Define the car structure.  Structure tags are used by encoding/json library
type Land struct {
	Address       string `json:"address"`
	Location      string `json:"location"`
	Type          string `json:"type"`
	Area          string `json:"area"`
	City          string `json:"city"`
	Country       string `json:"country"`
	OwnerCnic     string `json:"ownerCnic"`
	PropertyImage string `json:"propertyImage"`
	Status        string `json:"status"`
}

// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("fabcar_cc")

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	lands := []Land{
		Land{Address: "R-1028 sector 15/A", Location: "Buffer zone", Type: "House", Area: "120 yards", City: "Karachi", Country: "Pakistan", OwnerCnic: "42101-2696589-3", PropertyImage: "https://carhistorypictures.s3-ap-southeast-1.amazonaws.com/2018_Toyota_Corolla.jpg", Status: "created"},
		Land{Address: "R-1027 sector 15/A", Location: "Buffer zone", Type: "House", Area: "120 yards", City: "Karachi", Country: "Pakistan", OwnerCnic: "42101-2696589-3", PropertyImage: "https://carhistorypictures.s3-ap-southeast-1.amazonaws.com/2018_Toyota_Corolla.jpg", Status: "created"},
	}

	i := 0
	for i < len(lands) {
		landAsBytes, _ := json.Marshal(lands[i])
		APIstub.PutState(lands[i].Address, landAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "queryLand":
		return s.queryLand(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	case "createLand":
		return s.createLand(APIstub, args)
	case "changeLandOwner":
		return s.changeLandOwner(APIstub, args)
	case "getHistoryForAsset":
		return s.getHistoryForAsset(APIstub, args)
	case "queryLandsByOwner":
		return s.queryLandsByOwner(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}

func (s *SmartContract) queryLand(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	landAsBytes, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to read from world state")
	}

	if landAsBytes == nil {
		return shim.Error("does not exist")
	}
	return shim.Success(landAsBytes)
}

func (s *SmartContract) createLand(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}
	var land = Land{Address: args[0], Location: args[1], Type: args[2], Area: args[3], City: args[4], Country: args[5], OwnerCnic: args[6], PropertyImage: args[7], Status: "created"}

	landAsBytes, _ := json.Marshal(land)
	APIstub.PutState(args[0], landAsBytes)
	APIstub.SetEvent("CreateLand",landAsBytes)
	indexName := "ownerCnic~address"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{land.OwnerCnic, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)

	return shim.Success(landAsBytes)
}

func (S *SmartContract) queryLandsByOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	owner := args[0]

	ownerAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("ownerCnic~address", []string{owner})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIdResultIterator.Close()

	var i int
	var id string

	var lands []byte
	bArrayMemberAlreadyWritten := false

	lands = append([]byte("["))

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
			lands = append(lands, newBytes...)

		} else {
			lands = append(lands, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	lands = append(lands, []byte("]")...)

	return shim.Success(lands)
}

func (t *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	landAddress := args[0]

	resultsIterator, err := stub.GetHistoryForKey(landAddress)
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

func (s *SmartContract) changeLandOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	indexName := "ownerCnic~address"
	landAsBytes, _ := APIstub.GetState(args[0])
	land := Land{}

	json.Unmarshal(landAsBytes, &land)
	// delete old key
	ownerLandidIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{land.OwnerCnic, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.DelState(ownerLandidIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// add new record
	land.OwnerCnic = args[1]
	land.Status = "Transfered"

	landAsBytes, _ = json.Marshal(land)
	APIstub.PutState(args[0], landAsBytes)
	APIstub.SetEvent("OwnerChanged",landAsBytes) 
	// add new key
	newOwnerLandidIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{args[1], args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(newOwnerLandidIndexKey, value)

	return shim.Success(landAsBytes)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {
	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
