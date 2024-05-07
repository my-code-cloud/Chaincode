package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

//MessagingBoardChaincode stuct
type MessagingBoardChaincode struct {
}

//Init method
func (t *MessagingBoardChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	LogMessage(INFO, "---INIT SMART INDUSTRY CONTRACT---", "")

	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *MessagingBoardChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	// Variable
	// var err error

	LogMessage(INFO, "*** MessagingBoardChaincode function invoked: ", fn)

	if fn == "postMessage" {
		return t.postMessage(stub, args) // TODO
	} else if fn == "getMessage" {
		return t.getMessage(stub, args) // TODO

	} else if fn == "getLatestVersion" {
		return t.getLatestVersionMessage(stub, args) // TODO

	} else {
		return shim.Error("Invalid invoke function name")
	}
}

//*************************************************************
// MANAGE MESSAGE
//*************************************************************

//postMessage: the input json must contain the object Message,
//The caller must be a CREATOR USER!!!

func (t *MessagingBoardChaincode) postMessage(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var err error
	var message Message = Message{}
	var messages []Message
	var byteMessage []byte
	var IDKey string
	var identityMessage IdentityMessage = IdentityMessage{}
	var descriptionError string = ""
	var found bool
	var verifyOk bool

	LogMessage(INFO, "postMessage()", time.Now().Format(time.RFC3339))

	// Control IdentityMessage into CA and check the authorization
	identityMessage, found, err = GetIdentityMessage(stub)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage GetIdentityMessage err not nil", err.Error())
		return shim.Error(err.Error())
	}
	if !found {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage GetIdentityMessage err nil but something mandatory notfound", "")
		LogMessage(ERROR, "MessagingBoardChaincode fileds mandatory are ROLE: ", identityMessage.Role)
		LogMessage(ERROR, "MessagingBoardChaincode fileds mandatory are UUID: ", identityMessage.Client)
		return shim.Error("MessagingBoardChaincode postMessage GetIdentityMessage err nil but something notfound")
	}
	if identityMessage.Role != "" && identityMessage.Role != WRITER {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage ERROR: role found but is not creator (super-user)!!!", identityMessage.Role)
		return shim.Error("MessagingBoardChaincode postMessage ERROR: role is not authorized to create a Message!!!")
	}

	// message.PkPreamble.Domain, message.PkPreamble.Environment, message.PkPreamble.Process, message.Name

	// Create Object from args parameters
	err = json.Unmarshal([]byte(args[0]), &message)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage jsonUnmarshal err not nil", err.Error())
		return shim.Error(err.Error())
	}

	LogMessage(DEBUG, "postMessage() with Domain     : ", message.PkPreamble.Domain)
	LogMessage(DEBUG, "postMessage() with Environment: ", message.PkPreamble.Environment)
	LogMessage(DEBUG, "postMessage() with Process    : ", message.PkPreamble.Process)
	LogMessage(DEBUG, "postMessage() with Name       : ", message.Name)

	if message.Name == "" {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage Name are mandatory but is NOT valorized", "")
		return shim.Error("MessagingBoardChaincode postMessage Name are mandatory but is NOT valorized")
	}

	found, descriptionError, message.PkPreamble, message.Name = CheckIdentityMessage(identityMessage, message)
	if descriptionError != "" || !found {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage CheckIdentityMessage err not nil", descriptionError)
		return shim.Error(descriptionError)
	}

	if message.PkPreamble.Domain == "" && message.PkPreamble.Environment == "" && message.PkPreamble.Process == "" && message.Name == "" {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage Fields Key of Message are mandatory but is any is valorized", "")
		return shim.Error("MessagingBoardChaincode postMessage Fields Key of Message are mandatory but is any is valorized")
	}

	// Control single field
	verifyOk, descriptionError = verifyString(message.PkPreamble.Domain)
	if !verifyOk {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage Fields Key Domain: "+descriptionError, message.PkPreamble.Domain)
		return shim.Error(descriptionError)
	}
	verifyOk, descriptionError = verifyString(message.PkPreamble.Environment)
	if !verifyOk {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage Fields Key Environment: "+descriptionError, message.PkPreamble.Environment)
		return shim.Error(descriptionError)
	}
	verifyOk, descriptionError = verifyString(message.PkPreamble.Process)
	if !verifyOk {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage Fields Key Process: "+descriptionError, message.PkPreamble.Process)
		return shim.Error(descriptionError)
	}
	verifyOk, descriptionError = verifyString(message.Name)
	if !verifyOk {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage Fields Key Name: "+descriptionError, message.Name)
		return shim.Error(descriptionError)
	}
	verifyOk, descriptionError = verifyURL(message.MessageRef)
	if !verifyOk {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage length of URL MessageRef must be from 1 to 256crt: ", message.MessageRef)
		return shim.Error(descriptionError)
	}
	verifyOk, descriptionError = verifyURL(message.Seal)
	if !verifyOk {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage length of URL Seal must be from 1 to 256crt: ", message.Seal)
		return shim.Error(descriptionError)
	}
	if message.MessageSize == 0 {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage MessageSize is 0", "")
		return shim.Error(descriptionError)
	}

	// Create the PARTIAL KEY
	IDKey, err = stub.CreateCompositeKey(MESSAGE, []string{message.PkPreamble.Domain, message.PkPreamble.Environment, message.PkPreamble.Process, message.Name})
	// IDKey, err = getKey(stub, STATE, message.Domain+message.Environment+message.Process+message.Name)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage getKey err not nil", err.Error())
		return shim.Error(err.Error())
	}

	var arrayKey []string = []string{message.PkPreamble.Domain, message.PkPreamble.Environment, message.PkPreamble.Process, message.Name}
	// GetState for partial key -> list of objectDistribution
	messages, err = GetStateByPartialCompositeKeyMessage(stub, arrayKey)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage GetStateByPartialCompositeKeyObject err not nil", err.Error())
		return shim.Error(err.Error())
	}

	if len(messages) == 0 {
		message.Version = 0
	} else {
		message.Version = messages[len(messages)-1].Version + 1
	}
	message.Created = time.Now().Format(time.RFC3339)
	if message.CreatedBy == "" {
		message.CreatedBy = identityMessage.Client
	}
	// Create the (TOTAL) KEY
	IDKey, err = stub.CreateCompositeKey(MESSAGE, []string{message.PkPreamble.Domain, message.PkPreamble.Environment, message.PkPreamble.Process, message.Name, strconv.Itoa(message.Version)})
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage getKey err not nil", err.Error())
		return shim.Error(err.Error())
	}
	// Convert Object message in []Byte
	byteMessage, err = json.Marshal(&message)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage jsonMarshal err not nil", err.Error())
		return shim.Error(err.Error())
	}

	// PUT STATE INTO THE LEDGER
	err = stub.PutState(IDKey, byteMessage)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage putState() err not nil", err.Error())
		return shim.Error(err.Error())
	}

	jsonResp = string(byteMessage)
	LogMessage(INFO, "MessagingBoardChaincode postMessage putState is OK with response:", jsonResp)
	return shim.Success([]byte(jsonResp))

}

//getMessage: the input json must contain the KEY of the object MEssage to SEARCH,
//The caller must be a ANY USER!!!

func (t *MessagingBoardChaincode) getMessage(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var message Message = Message{}
	var byteMessage []byte
	var messages []Message
	var err error
	var IDKey string
	var identityMessage IdentityMessage
	var found bool
	var descriptionError string = ""

	LogMessage(INFO, "getMessage()", time.Now().Format(time.RFC3339))

	// Control Parameter
	if len(args) != 4 && len(args) != 5 {
		LogMessage(ERROR, "MessagingBoardChaincode getMessage numbers of Args not correctly!: ", strconv.Itoa(len(args)))
		return shim.Error("MessagingBoardChaincode getMessage ERROR: numbers of Args not correctly!:\n")
	}

	message.PkPreamble.Domain = strings.TrimSpace(args[0])
	message.PkPreamble.Environment = strings.TrimSpace(args[1])
	message.PkPreamble.Process = strings.TrimSpace(args[2])
	message.Name = strings.TrimSpace(args[3])
	LogMessage(DEBUG, "getMessage() with Domain     : ", message.PkPreamble.Domain)
	LogMessage(DEBUG, "getMessage() with Environment: ", message.PkPreamble.Environment)
	LogMessage(DEBUG, "getMessage() with Process    : ", message.PkPreamble.Process)
	LogMessage(DEBUG, "getMessage() with Name       : ", message.Name)

	// Control IdentityMessage into CA and check the authorization to GetMessage
	identityMessage, found, err = GetIdentityMessage(stub)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode getMessage GetIdentityMessage err not nil", err.Error())
		return shim.Error(err.Error())
	}
	if !found {
		LogMessage(ERROR, "MessagingBoardChaincode getMessage GetIdentityMessage err nil but something mandatory notfound", "")
		LogMessage(ERROR, "MessagingBoardChaincode fileds mandatory are ROLE: ", identityMessage.Role)
		LogMessage(ERROR, "MessagingBoardChaincode fileds mandatory are UUID: ", identityMessage.Client)
		return shim.Error("MessagingBoardChaincode getMessage GetIdentityMessage err nil but something notfound")
	}
	if identityMessage.Role != "" && identityMessage.Role != WRITER && identityMessage.Role != READER {
		LogMessage(ERROR, "MessagingBoardChaincode getMessage ERROR: role found but is not creator and not user!!!", identityMessage.Role)
		return shim.Error("MessagingBoardChaincode getMessage ERROR: role is not authorized to get a Message!!!")
	}
	found, descriptionError, message.PkPreamble, message.Name = CheckIdentityMessage(identityMessage, message)
	if descriptionError != "" || !found {
		LogMessage(ERROR, "MessagingBoardChaincode getMessage CheckIdentityMessage err not nil", descriptionError)
		return shim.Error(descriptionError)
	}

	if len(args) == 5 {
		// Key complete with 5 parameters - Simple GetState
		message.Version, err = strconv.Atoi(args[4])
		if err != nil {
			LogMessage(ERROR, "MessagingBoardChaincode getMessage strconv.Atoi err not nil", err.Error())
			return shim.Error(err.Error())
		}

		// Create the KEY
		IDKey, err = stub.CreateCompositeKey(MESSAGE, []string{message.PkPreamble.Domain, message.PkPreamble.Environment, message.PkPreamble.Process, message.Name, strconv.Itoa(message.Version)})
		// IDKey, err = getKey(stub, MESSAGE, message.PkPreamble.Domain+message.PkPreamble.Environment+message.PkPreamble.Process+message.Name+versionString)
		if err != nil {
			LogMessage(ERROR, "MessagingBoardChaincode getMessage getKey err not nil", err.Error())
			return shim.Error(err.Error())
		}

		byteMessage, err = stub.GetState(IDKey)
		LogMessage(INFO, "MessagingBoardChaincode getMessage Message version found is OK", string(byteMessage))
		if err != nil {
			LogMessage(ERROR, "MessagingBoardChaincode getMessage GetState err not nil", err.Error())
			return shim.Error(err.Error())
		}

	} else {
		// Partial Key number of parameters: 4
		// GetState for partial key -> list of message

		// Create the KEY
		IDKey, err = stub.CreateCompositeKey(MESSAGE, []string{message.PkPreamble.Domain, message.PkPreamble.Environment, message.PkPreamble.Process, message.Name})
		// IDKey, err = getKey(stub, MESSAGE, message.PkPreamble.Domain+message.PkPreamble.Environment+message.PkPreamble.Process+message.Name)
		if err != nil {
			LogMessage(ERROR, "MessagingBoardChaincode getMessage getKey partial err not nil", err.Error())
			return shim.Error(err.Error())
		}
		var arrayKey = []string{message.PkPreamble.Domain, message.PkPreamble.Environment, message.PkPreamble.Process, message.Name}
		messages, err = GetStateByPartialCompositeKeyMessage(stub, arrayKey)
		if err != nil {
			LogMessage(ERROR, "MessagingBoardChaincode getMessage GetStateByPartialCompositeKeyObject err not nil", err.Error())
			return shim.Error(err.Error())
		}

		if len(messages) == 0 {
			byteMessage = nil
			LogMessage(WARNING, "MessagingBoardChaincode getMessage Message not found -return nil", IDKey)
		} else {
			byteMessage, err = json.Marshal(messages[len(messages)-1])
			LogMessage(INFO, "MessagingBoardChaincode getMessage Message latest found is OK", string(byteMessage))
			if err != nil {
				LogMessage(ERROR, "MessagingBoardChaincode getMessage Marshal err not nil", err.Error())
				return shim.Error(err.Error())
			}
		}
	}
	if byteMessage == nil {
		return shim.Success(nil)
	}
	return shim.Success(byteMessage)
}

//getLatestVersionMessage: the input json must contain the Partial KEY of the object Message: we need estract latest versiono SEARCH,
//The caller must be a ANY USER!!!
func (t *MessagingBoardChaincode) getLatestVersionMessage(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var message Message = Message{}
	var messages []Message
	var byteVersion []byte
	var version int
	var err error
	var found bool
	var descriptionError string
	var identityMessage IdentityMessage

	LogMessage(INFO, "getLatestVersionMessage()", time.Now().Format(time.RFC3339))

	// Control Parameter
	if len(args) != 4 {
		LogMessage(ERROR, "MessagingBoardChaincode getLatestVersionMessage numbers of Args not correctly!: ", strconv.Itoa(len(args)))
		return shim.Error("MessagingBoardChaincode getLatestVersionMessage ERROR: numbers of Args not correctly!:\n")
	}

	message.PkPreamble.Domain = strings.TrimSpace(args[0])
	message.PkPreamble.Environment = strings.TrimSpace(args[1])
	message.PkPreamble.Process = strings.TrimSpace(args[2])
	message.Name = strings.TrimSpace(args[3])

	// Control IdentityMessage into CA and check the authorization to GetVersion
	identityMessage, found, err = GetIdentityMessage(stub)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage GetIdentityMessage err not nil", err.Error())
		return shim.Error(err.Error())
	}
	if !found {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage GetIdentityMessage err nil but something mandatory notfound", "")
		LogMessage(ERROR, "MessagingBoardChaincode fileds mandatory are ROLE: ", identityMessage.Role)
		LogMessage(ERROR, "MessagingBoardChaincode fileds mandatory are UUID: ", identityMessage.Client)
		return shim.Error("MessagingBoardChaincode postMessage GetIdentityMessage err nil but something notfound")
	}
	if identityMessage.Role != "" && identityMessage.Role != WRITER && identityMessage.Role != READER {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage ERROR: role found but is not creator and not user!!!", identityMessage.Role)
		return shim.Error("MessagingBoardChaincode postMessage ERROR: role is not authorized to get a Message!!!")
	}

	found, descriptionError, message.PkPreamble, message.Name = CheckIdentityMessage(identityMessage, message)
	if descriptionError != "" || !found {
		LogMessage(ERROR, "MessagingBoardChaincode postMessage CheckIdentityMessage err not nil", descriptionError)
		return shim.Error(descriptionError)
	}

	var arrayKey []string = []string{message.PkPreamble.Domain, message.PkPreamble.Environment, message.PkPreamble.Process, message.Name}
	// GetState for partial key -> list of objectDistribution
	messages, err = GetStateByPartialCompositeKeyMessage(stub, arrayKey)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode getLatestVersionMessage GetStateByPartialCompositeKeyObject err not nil", err.Error())
		return shim.Error(err.Error())
	}

	if len(messages) == 0 { // NOT FOUND
		message.Version = 0
		version = 0
	} else {
		message.Version = messages[len(messages)-1].Version
		version = messages[len(messages)-1].Version
	}

	LogMessage(DEBUG, "MessagingBoardChaincode getLatestVersionMessage version response of message", strconv.Itoa(message.Version))

	byteVersion, err = json.Marshal(&version)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode getLatestVersionMessage jsonMarshal err not nil", err.Error())
		return shim.Error(err.Error())
	}

	LogMessage(INFO, "MessagingBoardChaincode getLatestVersionMessage GetState response is OK", strconv.Itoa(version))
	return shim.Success(byteVersion)
}

// **********************************************************************
// UTILITY
// **********************************************************************

// LogMessage System Manual Logger
func LogMessage(activity string, message string, parameter string) {

	fmt.Println(time.Now().Format(time.RFC3339)+" "+activity+" MESSAGE: "+message, parameter)

	/* if activity == INFO {
		fmt.Println(time.Now().Format(time.RFC3339) + " INFO MESSAGE<<<")
		fmt.Println(message, parameter)
		fmt.Println("                  ")
	}
	if activity == WARNING {
		fmt.Println(time.Now().Format(time.RFC3339) + " WARNING MESSAGE!!!")
		fmt.Println(message, parameter)
		fmt.Println("                     ")
	}
	if activity == ERROR {
		fmt.Println(time.Now().Format(time.RFC3339) + " ERROR MESSAGE!!!**")
		fmt.Println(message, parameter)
		fmt.Println("***********************")
	}
	if activity == DEBUG {
		fmt.Println(time.Now().Format(time.RFC3339) + " DEBUG MESSAGE ")
		fmt.Println(message, parameter)
	} else {
		fmt.Println(message, parameter)
	} */
}

// GetStateByPartialCompositeKeyMessage return a array of Message by key
func GetStateByPartialCompositeKeyMessage(stub shim.ChaincodeStubInterface, KEYid []string) ([]Message, error) {

	var buffer bytes.Buffer
	var err error
	var message Message
	var messages []Message

	resultsIterator, err := stub.GetStateByPartialCompositeKey(MESSAGE, KEYid)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode GetStateByPartialCompositeKeyMessage err not nil", string(err.Error()))
		return nil, err
	}
	defer resultsIterator.Close()
	// buffer is a JSON array containing QueryResults

	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	var i int

	for i = 0; resultsIterator.HasNext(); i++ {
		responseRange, err := resultsIterator.Next()
		if err != nil {
			LogMessage(ERROR, "MessagingBoardChaincode Into Loop  resultsIterator.Next() for err not nil", strconv.Itoa(i))
			return nil, err
		}

		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			LogMessage(ERROR, "MessagingBoardChaincode Into Loop  stub.SplitCompositeKey for err not nil", responseRange.Key)
			return nil, err
		}

		keyComplete, err := stub.CreateCompositeKey(MESSAGE, compositeKeyParts)

		if err != nil {
			LogMessage(ERROR, "MessagingBoardChaincode GetStateByPartialCompositeKeyMessage CreateCompositeKey err not nil", err.Error())
			return nil, err
		}

		LogMessage(DEBUG, "MessagingBoardChaincode Into Loop post stub.SplitCompositeKey", keyComplete)

		// idAsBytes, err := stub.GetState(idReturn)
		idAsBytes, err := stub.GetState(keyComplete)
		if err != nil {
			LogMessage(ERROR, "MessagingBoardChaincode GetStateByPartialCompositeKeyMessage GetState: ", err.Error())
			return nil, err
		}
		if idAsBytes == nil {
			LogMessage(ERROR, "MessagingBoardChaincode GetStateByPartialCompositeKeyMessage GetState DATA NOT FOUND for KEY: ", keyComplete)
		}

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(keyComplete)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(idAsBytes))

		err = json.Unmarshal(idAsBytes, &message)
		if err != nil {
			LogMessage(ERROR, "MessagingBoardChaincode GetStateByPartialCompositeKeyMessage Unmarshal Message", err.Error())
			return nil, err
		}
		messages = append(messages, message)

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Version < messages[j].Version
	})
	return messages, nil

}

// GetIdentityMessage return struct IdentityMessage with data of CA
func GetIdentityMessage(stub shim.ChaincodeStubInterface) (IdentityMessage, bool, error) {
	var identityMessage IdentityMessage = IdentityMessage{}
	var found bool
	var role string
	var err error
	var uuid string
	var domains string
	var environments string
	var processes string
	var names string
	var emptyList []string

	// GetAttributeValue(ROLE) - notfound is KO
	found, role, err = isInvokerOperator(stub, ROLE)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode GetIdentityMessage isInvokerOperator(role) err not nil", err.Error())
		return identityMessage, false, err
	}
	if !found {
		LogMessage(ERROR, "MessagingBoardChaincode GetIdentityMessage ERROR: ROLE is empty!!!", "")
		shim.Error("MessagingBoardChaincode GetIdentityMessage ERROR: ROLE is empty!!!")
		return identityMessage, false, nil
	}
	LogMessage(DEBUG, "MessagingBoardChaincode GetIdentityMessage isInvokerOperator by role: ", role)
	identityMessage.Role = role

	// GetAttributeValue(UUID) - notfound is KO
	found, uuid, err = isInvokerOperator(stub, UUID)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode GetIdentityMessage isInvokerOperator(uuid) err not nil", err.Error())
		return identityMessage, false, err
	}
	if !found {
		LogMessage(ERROR, "MessagingBoardChaincode GetIdentityMessage ERROR: UUID is empty!!!", "")
		shim.Error("MessagingBoardChaincode GetIdentityMessage ERROR: ROLE is empty!!!")
		return identityMessage, false, nil
	}
	LogMessage(DEBUG, "MessagingBoardChaincode GetIdentityMessage isInvokerOperator by uuid: ", uuid)
	identityMessage.Client = uuid

	// GetAttributeValue(Domain) - notfound is OK
	found, domains, err = isInvokerOperator(stub, DOMAIN)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode GetIdentityMessage isInvokerOperator(domain) err not nil", err.Error())
		return identityMessage, false, err
	}
	if !found {
		LogMessage(INFO, "MessagingBoardChaincode GetIdentityMessage: domain of CA is empty!!!", "")
		identityMessage.Domains = emptyList
	} else {
		LogMessage(DEBUG, "MessagingBoardChaincode GetIdentityMessage isInvokerOperator by domain: ", domains)
		identityMessage.Domains = strings.Split(domains, ",")
		for i := 0; i < len(identityMessage.Domains); i++ {
			identityMessage.Domains[i] = strings.TrimSpace(identityMessage.Domains[i])
		}

	}

	// GetAttributeValue(Environment) - notfound is OK
	found, environments, err = isInvokerOperator(stub, ENVIRONMENT)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode GetIdentityMessage isInvokerOperator(environment) err not nil", err.Error())
		return identityMessage, false, err
	}
	if !found {
		LogMessage(INFO, "MessagingBoardChaincode GetIdentityMessage: environment of CA is empty!!!", "")
		identityMessage.Environments = emptyList
	} else {
		LogMessage(DEBUG, "MessagingBoardChaincode GetIdentityMessage isInvokerOperator by environment: ", environments)
		identityMessage.Environments = strings.Split(environments, ",")
		for i := 0; i < len(identityMessage.Environments); i++ {
			identityMessage.Environments[i] = strings.TrimSpace(identityMessage.Environments[i])
		}
	}

	// GetAttributeValue(Process) - notfound is OK
	found, processes, err = isInvokerOperator(stub, PROCESS)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode GetIdentityMessage isInvokerOperator(process) err not nil", err.Error())
		return identityMessage, false, err
	}
	if !found {
		LogMessage(INFO, "MessagingBoardChaincode GetIdentityMessage: process of CA is empty!!!", "")
		identityMessage.Processes = emptyList
	} else {
		LogMessage(DEBUG, "MessagingBoardChaincode GetIdentityMessage isInvokerOperator by process: ", processes)
		identityMessage.Processes = strings.Split(processes, ",")
		for i := 0; i < len(identityMessage.Processes); i++ {
			identityMessage.Processes[i] = strings.TrimSpace(identityMessage.Processes[i])
		}
	}

	// GetAttributeValue(Name) - notfound is OK
	found, names, err = isInvokerOperator(stub, NAME)
	if err != nil {
		LogMessage(ERROR, "MessagingBoardChaincode GetIdentityMessage isInvokerOperator(name) err not nil", err.Error())
		return identityMessage, false, err
	}
	if !found {
		LogMessage(INFO, "MessagingBoardChaincode GetIdentityMessage: name of CA is empty!!!", "")
		identityMessage.Names = emptyList
	} else {
		LogMessage(DEBUG, "MessagingBoardChaincode GetIdentityMessage isInvokerOperator by name: ", names)
		identityMessage.Names = strings.Split(names, ",")
		for i := 0; i < len(identityMessage.Names); i++ {
			identityMessage.Names[i] = strings.TrimSpace(identityMessage.Names[i])
		}
	}

	return identityMessage, true, nil
}

// CheckIdentityValueMessageValue return a value and a descriptionError
func CheckIdentityValueMessageValue(idValues []string, msgValue string) (string, string) {
	var msgError string = ""
	var valueResponse string = ""
	// var boolResponse bool = false
	LogMessage(DEBUG, "MessagingBoardChaincode CheckIdentityValueMessageValue msgValue     : ", msgValue)
	LogMessage(DEBUG, "MessagingBoardChaincode CheckIdentityValueMessageValue idValues     : ", strings.Join(idValues, ","))
	if msgValue == "" {
		if len(idValues) == 1 {
			valueResponse = idValues[0]
		} else {
			if len(idValues) == 0 {
				LogMessage(WARNING, "MessagingBoardChaincode CheckIdentityValueMessageValue NO msgValue and NO idValues", "")
				valueResponse = ""
			} else { // len(idValues) == 0
				msgError = "Is not possible to get a unique value in identity: msgValue empty and identity with more options"
				valueResponse = ""
			}
		}
	} else { // msgValue valorized
		if len(idValues) == 0 {
			valueResponse = msgValue
		} else {
			for i := 0; i < len(idValues); i++ {
				if msgValue == idValues[i] {
					valueResponse = msgValue
					break
				}
			}
			if valueResponse == "" {
				msgError = "This identity CA don't has the authorization for this Value: " + msgValue
			}
		}
	}
	return valueResponse, msgError
}

// CheckIdentityMessage return a boolean and a description
func CheckIdentityMessage(identityMessage IdentityMessage, message Message) (bool, string, PkPreamble, string) {

	var descriptionError string = ""
	var pk PkPreamble = PkPreamble{}
	var nm string = ""

	pk.Domain, descriptionError = CheckIdentityValueMessageValue(identityMessage.Domains, message.PkPreamble.Domain)
	if descriptionError != "" {
		descriptionError = descriptionError + " -> Domain"
		LogMessage(ERROR, "MessagingBoardChaincode CheckIdentityValueMessageValue Domain error: ", descriptionError)
		LogMessage(ERROR, "Domains register for this identity: ", strings.Join(identityMessage.Domains, ","))
		LogMessage(ERROR, "Domain in INPUT parameters        : ", message.PkPreamble.Domain)
		return false, descriptionError, pk, nm
	}

	pk.Environment, descriptionError = CheckIdentityValueMessageValue(identityMessage.Environments, message.PkPreamble.Environment)
	if descriptionError != "" {
		descriptionError = descriptionError + " -> Environment"
		LogMessage(ERROR, "MessagingBoardChaincode CheckIdentityValueMessageValue Environment error: ", descriptionError)
		LogMessage(ERROR, "Environments register for this identity: ", strings.Join(identityMessage.Environments, ","))
		LogMessage(ERROR, "Environment in INPUT parameters        : ", message.PkPreamble.Environment)
		return false, descriptionError, pk, nm
	}

	pk.Process, descriptionError = CheckIdentityValueMessageValue(identityMessage.Processes, message.PkPreamble.Process)
	if descriptionError != "" {
		descriptionError = descriptionError + " -> Process"
		LogMessage(ERROR, "MessagingBoardChaincode CheckIdentityValueMessageValue Process error: ", descriptionError)
		LogMessage(ERROR, "Processes register for this identity: ", strings.Join(identityMessage.Processes, ","))
		LogMessage(ERROR, "Process in INPUT parameters         : ", message.PkPreamble.Process)
		return false, descriptionError, pk, nm
	}

	nm, descriptionError = CheckIdentityValueMessageValue(identityMessage.Names, message.Name)
	if descriptionError != "" {
		descriptionError = descriptionError + " -> Name"
		LogMessage(ERROR, "MessagingBoardChaincode CheckIdentityValueMessageValue Name error: ", descriptionError)
		LogMessage(ERROR, "Names register for this identity: ", strings.Join(identityMessage.Names, ","))
		LogMessage(ERROR, "Name in INPUT parameters        : ", message.Name)
		return false, descriptionError, pk, nm
	}
	return true, descriptionError, pk, nm
}

// verifyString: apply restriction for the key
func verifyString(value string) (bool, string) {
	var verifyOk bool = true
	var descriptionError string = ""

	for _, char := range value {
		if !strings.Contains(ALPHACHAROK, strings.ToLower(string(char))) {
			verifyOk = false
			descriptionError = "MessagingBoardChaincode postMessage Fields Key has character not allowed: " + value
		}
	}
	if len(value) > LIMITMAXSTRING {
		verifyOk = false
		descriptionError = "MessagingBoardChaincode postMessage Fields Key is too long (max 50 crt): " + value
	}

	return verifyOk, descriptionError
}

// verifyURL: apply restriction for field URL
func verifyURL(url string) (bool, string) {
	var verifyOk bool = true
	var descriptionError string = ""

	if len(url) > LIMITMAXURL {
		verifyOk = false
		descriptionError = "MessagingBoardChaincode postMessage Fields Key URL is too long (max 256 crt): " + url
	}
	if len(url) == 0 {
		verifyOk = false
		descriptionError = "MessagingBoardChaincode postMessage Fields Key URL is not valorized (min 1 crt): " + url
	}
	return verifyOk, descriptionError
}

// **********************************************************************
// END
// **********************************************************************

// main function starts up the chaincode in the container during instantiate
func main() {
	fmt.Println("---MAIN FINDUSTRY CONTRACT---")
	if err := shim.Start(new(MessagingBoardChaincode)); err != nil {

		// log.Debug("Main()")
		fmt.Printf("Error starting Contract chaincode: %s", err)
	}
}
