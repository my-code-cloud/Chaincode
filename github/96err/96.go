package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"time"

	"github.com/hyperledger/fabric/common/flogging"

	// "github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	// "github.com/hyperledger/fabric-protos-go/common"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

// Chaincode Define the Smart Contract structure
type Chaincode struct {
}

// Loan define the loan structure
type Loan struct {
	Account string `json:"Account"`
	Amount  string `json:"Amount"`
	Name    string `json:"Name"`
	Mobile  string `json:"Mobile"`
}

type privateloan struct {
	Name   string `json:"Name"`
	Amount string `json:"Amount"`
}

var logg = flogging.MustGetLogger("loan")

// Init ;  Method for initializing smart contract
func (ch *Chaincode) Init(APIstub shim.ChaincodeStubInterface) peer.Response {
	// fmt.Printf("Inside Init method")
	return shim.Success(nil)
}

// Invoke ;  Method for initializing smart contract
func (ch *Chaincode) Invoke(APIstub shim.ChaincodeStubInterface) peer.Response {

	fmt.Printf("Inside Invoke method")

	// /* Print the transaction ID */
	// logg.Infof("Transaction id: ", APIstub.GetTxID())

	// /* Print the Channel ID */
	// logg.Infof("Channel ID : ", APIstub.GetChannelID())

	/* Print the Timestamp */
	// timestamp, _ := APIstub.GetTxTimestamp()
	// timestring := time.Unix(timestamp.GetSeconds(), 0)
	// logg.Infof("Timestamp: ", timestring)

	// creator, _ := APIstub.GetCreator()
	// logg.Infof("Creator: ", string(creator))

	/* Fetch the sign proposal */
	// Get the SignedProposal
	// SignedProposal has 2 parts
	// 1. ProposalBytes
	// 2. Signature
	// signproposal, _ := APIstub.GetSignedProposal()
	// data := signproposal.GetProposalBytes()
	// proposal := &peer.Proposal{}
	// proto.Unmarshal(data, proposal)

	// Proposal has 2 parts
	// 1. Header
	// 2. Payload - the structure for this depends on the type in the ChannelHeader
	// header := &common.Header{}
	// proto.Unmarshal(proposal.GetHeader(), header)

	// Header has 2 parts
	// 1. ChannelHeader
	// 2. SignatureHeader
	// channelheader := &common.ChannelHeader{}
	// proto.Unmarshal(header.GetChannelHeader(), channelheader)

	// logg.Infof("channelHeader.GetType() => ", common.HeaderType(channelheader.GetType()))
	// logg.Infof("channelHeader.GetChannelId() => ", channelheader.GetChannelId())

	function, args := APIstub.GetFunctionAndParameters()

	logg.Infof("Funtion Name : %d", function)
	logg.Infof("Length of Argument : %d", len(args))

	switch function {
	case "createloan":
		return ch.createloan(APIstub, args)
	case "queryloan":
		return ch.queryloan(APIstub, args)
	case "initLedger":
		return ch.initLedger(APIstub)
	case "queryallloan":
		return ch.queryallloan(APIstub)
	case "querlyloanbyName":
		return ch.querlyloanbyName(APIstub, args)
	case "changeName":
		return ch.changeName(APIstub, args)
	case "deleteloan":
		return ch.deleteloan(APIstub, args)
	case "gethistorydata":
		return ch.gethistorydata(APIstub, args)
	case "createprivateloan":
		return ch.createprivateloan(APIstub, args)
	case "readPrivateloan":
		return ch.readPrivateloan(APIstub, args)
	case "readPrivateloanamount":
		return ch.readPrivateloanamount(APIstub, args)
	case "createloanImplicitHDFC":
		return ch.createloanImplicitHDFC(APIstub, args)
	case "createloanImplicitICICI":
		return ch.createloanImplicitICICI(APIstub, args)
	case "readloanImplicitHDFC":
		return ch.readloanImplicitHDFC(APIstub, args)
	case "readloanImplicitICICI":
		return ch.readloanImplicitICICI(APIstub, args)
	case "queryPrivateDataHash":
		return ch.queryPrivateDataHash(APIstub, args)
	default:
		return shim.Error("Invalid chaincode name")
	}
}

func (ch *Chaincode) createprivateloan(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	logg.Infof("Inside createprivateloan method")

	type transientloan struct {
		Key     string `json:"key"`
		Account string `json:"Account"`
		Amount  string `json:"Amount"`
		Name    string `json:"Name"`
		Mobile  string `json:"Mobile"`
	}

	if len(args) != 0 {
		return shim.Error("Private data must be passed in transient map")
	}

	transmap, err := APIstub.GetTransient()

	if err != nil {
		return shim.Error("Error getting transient map: " + err.Error())
	}

	loandetail, err1 := transmap["loan"]

	if !err1 {
		return shim.Error("loan must be a key in the transient map")
	}

	if len(loandetail) == 0 {
		return shim.Error("333333 -loan value in the transient map must be a non-empty JSON string")
	}

	var loaninput transientloan
	err2 := json.Unmarshal(loandetail, &loaninput)

	if err2 != nil {
		return shim.Error("Failed to decode JSON of: " + string(loandetail) + "Error is : " + err2.Error())
	}

	//Check the private data is already exist or not
	loanprivatedata, err3 := APIstub.GetPrivateData("privateloan", loaninput.Key)

	if err3 != nil {
		return shim.Error("Failed to get loan: " + err3.Error())
	} else if loanprivatedata != nil {
		fmt.Println("This loan already exists: " + loaninput.Key)
		return shim.Error("This loan already exists: " + loaninput.Key)
	}

	var loan = Loan{Account: loaninput.Account, Amount: loaninput.Amount, Name: loaninput.Name, Mobile: loaninput.Mobile}

	privateloandt, err4 := json.Marshal(loan)

	if err4 != nil {
		return shim.Error(err.Error())
	}

	err5 := APIstub.PutPrivateData("privateloan", loaninput.Key, privateloandt)

	if err5 != nil {
		return shim.Error(err.Error())
	}

	privateloan := &privateloan{Name: loaninput.Name, Amount: loaninput.Amount}

	privateloandetails, err6 := json.Marshal(privateloan)

	if err6 != nil {
		return shim.Error(err.Error())
	}

	err7 := APIstub.PutPrivateData("privateloanamount", loaninput.Key, privateloandetails)

	if err7 != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(privateloandt)

}

func (ch *Chaincode) readPrivateloan(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	loandetails, err := APIstub.GetPrivateData(args[0], args[1])

	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[1] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if loandetails == nil {
		jsonResp := "{\"Error\":\"Loan private details does not exist: " + args[1] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(loandetails)
}

func (ch *Chaincode) readPrivateloanamount(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	loanamtdetails, err := APIstub.GetPrivateData("privateloanamount", args[1])

	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[1] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if loanamtdetails == nil {
		jsonResp := "{\"Error\":\"Loan private details does not exist: " + args[1] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(loanamtdetails)
}
func (ch *Chaincode) createloan(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	logg.Infof("Inside createloan method and the argument is : %d", args[0])

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments")
	}
	var loan = Loan{Account: args[1], Amount: args[2], Name: args[3], Mobile: args[4]}

	loandetails, _ := json.Marshal(loan)
	APIstub.PutState(args[0], loandetails)

	/* Create Composity key */

	indexname := "loan~name"
	loanindex, err := APIstub.CreateCompositeKey(indexname, []string{loan.Name, args[0]})

	if err != nil {
		return shim.Error(err.Error())
	}

	value := []byte{0x00}
	APIstub.PutState(loanindex, value)

	return shim.Success(loandetails)
}

func (ch *Chaincode) createloanImplicitHDFC(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	logg.Infof("Inside createloanImplicitHDFC method and the argument is : %d", args[0])

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments")
	}
	var loan = Loan{Account: args[1], Amount: args[2], Name: args[3], Mobile: args[4]}

	loandetails, _ := json.Marshal(loan)

	err := APIstub.PutPrivateData("_implicit_org_HDFCMSP", args[0], loandetails)

	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	return shim.Success(loandetails)
}

func (ch *Chaincode) readloanImplicitHDFC(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	logg.Infof("Inside readloanImplicitHDFC method and the argument is : %d", args[0])

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	loandetails, _ := APIstub.GetPrivateData("_implicit_org_HDFCMSP", args[0])
	return shim.Success(loandetails)
}

func (ch *Chaincode) createloanImplicitICICI(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	logg.Infof("Inside createloanImplicitICICI method and the argument is : %d", args[0])

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments")
	}
	var loan = Loan{Account: args[1], Amount: args[2], Name: args[3], Mobile: args[4]}

	loandetails, _ := json.Marshal(loan)

	err := APIstub.PutPrivateData("_implicit_org_ICICIMSP", args[0], loandetails)

	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	return shim.Success(loandetails)
}

func (ch *Chaincode) readloanImplicitICICI(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	logg.Infof("Inside readloanImplicitICICI method and the argument is : %d", args[0])

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	loandetails, _ := APIstub.GetPrivateData("_implicit_org_ICICIMSP", args[0])
	return shim.Success(loandetails)
}

func (ch *Chaincode) queryloan(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	logg.Infof("Inside queryloan method and argument is: %d", args[0])
	loandetails, _ := APIstub.GetState(args[0])

	return shim.Success(loandetails)
}

func (ch *Chaincode) querlyloanbyName(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("One argument is expected")
	}

	name := args[0]
	logg.Infof("Name is : ", name)

	indexresult, err := APIstub.GetStateByPartialCompositeKey("loan~name", []string{name})

	if err != nil {
		return shim.Error(err.Error())
	}

	defer indexresult.Close()

	var i int
	var id string
	var loan []byte
	flag := false

	loan = append([]byte("["))

	for i = 0; indexresult.HasNext(); i++ {
		response, err := indexresult.Next()

		if err != nil {
			return shim.Error(err.Error())
		}

		object, compositekeyparts, err := APIstub.SplitCompositeKey(response.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositekeyparts[1]
		logg.Infof("ID is : ", name)
		value, err := APIstub.GetState(id)

		if flag == true {
			newvalue := append([]byte(","), value...)
			loan = append(loan, newvalue...)
		} else {
			loan = append(loan, value...)
		}

		fmt.Printf("Found a asset for index : %s loan id : ", object, compositekeyparts[0], compositekeyparts[1])
		flag = true

	}
	loan = append(loan, []byte("]")...)
	return shim.Success(loan)
}

func (ch *Chaincode) queryallloan(APIstub shim.ChaincodeStubInterface) peer.Response {

	startKey := "LOAN0"
	endKey := "LOAN99"

	var store bytes.Buffer
	flag := false

	iterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer iterator.Close()

	store.WriteString("[")

	for iterator.HasNext() {
		loandetails, err := iterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if flag == true {
			store.WriteString(",")
		}
		store.WriteString("{\"Key\":")
		store.WriteString("\"")
		store.WriteString(loandetails.Key)
		store.WriteString("\"")

		store.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		store.WriteString(string(loandetails.Value))
		store.WriteString("}")
		flag = true
	}

	return shim.Success(store.Bytes())
}

func (ch *Chaincode) initLedger(APIstub shim.ChaincodeStubInterface) peer.Response {
	loan := []Loan{
		Loan{Account: "6210", Amount: "200000", Name: "Vikas", Mobile: "9932809261"},
		Loan{Account: "8210", Amount: "500000", Name: "Sumit", Mobile: "8961383344"},
		Loan{Account: "9753", Amount: "500000", Name: "Tuhin", Mobile: "8765876567"},
		Loan{Account: "6784", Amount: "400000", Name: "Risabh", Mobile: "9674888324"},
		Loan{Account: "5210", Amount: "660000", Name: "Manish", Mobile: "9051864567"},
		Loan{Account: "8211", Amount: "100000", Name: "Arijit", Mobile: "8961376890"},
		Loan{Account: "9911", Amount: "500000", Name: "Arijit", Mobile: "8961376890"},
	}

	i := 0
	for i < len(loan) {
		loandetails, _ := json.Marshal(loan[i])
		APIstub.PutState("LOAN"+strconv.Itoa(i), loandetails)
		i = i + 1
	}
	logg.Infof("Inside Init ledger")
	return shim.Success(nil)
}

func (ch *Chaincode) changeName(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments")
	}
	logg.Infof("Inside changeName method and argument is: ", args[1])

	loandetails, _ := APIstub.GetState(args[0])
	loan := Loan{}
	json.Unmarshal(loandetails, &loan)
	loan.Name = args[1]
	newloan, _ := json.Marshal(loan)
	APIstub.PutState(args[0], newloan)

	// Emit event
	eventpayload := "{ \"value\": " + loan.Name + " }"
	APIstub.SetEvent("Name changed", []byte(eventpayload))

	return shim.Success([]byte(eventpayload))
}

func (ch *Chaincode) deleteloan(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	logg.Infof("Inside deleteloan method and argument is: ", args[0])

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}

	loandata, _ := APIstub.GetState(args[0])
	if loandata == nil {
		return shim.Error("No data present")
	}

	err := APIstub.DelState(args[0])
	if err != nil {
		fmt.Println("Delete Failed!!! ", err.Error())
		return shim.Error(("Delete Failed!! " + err.Error() + "!!!"))
	}

	return shim.Success([]byte("Data is deleted"))
}

func (ch *Chaincode) gethistorydata(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	logg.Infof("Inside gethistorydata method and argument is: ", args[0])

	iterator, _ := APIstub.GetHistoryForKey(args[0])
	defer iterator.Close()

	var data bytes.Buffer
	flag := false

	data.WriteString("[")

	for iterator.HasNext() {
		loandata, err := iterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if flag == true {
			data.WriteString(",")
		}
		data.WriteString("{\"Transaction ID\":")
		data.WriteString("\"")
		data.WriteString(loandata.TxId)
		data.WriteString("\"")

		data.WriteString(", \"Value\":")
		data.WriteString(string(loandata.Value))

		data.WriteString(", \"Timestamp\":")
		data.WriteString("\"")
		data.WriteString(time.Unix(loandata.Timestamp.Seconds, int64(loandata.Timestamp.Nanos)).String())
		data.WriteString("\"")

		data.WriteString(", \"IsDelete\":")
		data.WriteString("\"")
		data.WriteString(strconv.FormatBool(loandata.IsDelete))
		data.WriteString("\"")
		data.WriteString("}")

		flag = true
	}

	data.WriteString("]")
	logg.Infof("History data for loan: ", args[0], "Data: ", data.String())
	// fmt.Printf("- getHistoryForloan returning:\n%s\n", data.String())

	return shim.Success(data.Bytes())
}

func (ch *Chaincode) queryPrivateDataHash(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	loandetails, _ := APIstub.GetPrivateDataHash(args[0], args[1])
	return shim.Success(loandetails)
}

func main() {
	err := shim.Start(new(Chaincode))

	if err != nil {
		fmt.Printf("Error invoking Chaincode : %s", err)
	}

}