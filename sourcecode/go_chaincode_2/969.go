package main

import (
	"fmt"
	"encoding/json"
	"bytes"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)


type RulesContract struct {
}


// Define Status codes for the response
const (
	OK    = 200
	ERROR = 500
	//le macchine hanno il nome: M1000000 in poi
)

//key=port, tipo regola = iptables , oggetto=nodo su cui fai le regole, value=opzionale 
type Rule struct {
	Name		string	`json:"name"` //nome della regola, univoca altrimenti aggiorno il valore anziché crearlo
	Timestamp	string 	`json:"timestamp"`
	Event 		string 	`json:"event"`
	Value 		string 	`json:"value"`
}


// Init is called when the smart contract is instantiated
func (s *RulesContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	fmt.Println("Chaincode istanziato")
	return shim.Success(nil)
}


func (s *RulesContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "addrule" {			//add regola del sistemista
		return s.addRule(APIstub, args)
	} else if function == "getrule" {
		return s.getRule(APIstub, args)
	} else if function == "getruledelta" {
		return s.getRuleDelta(APIstub, args)
	} else if function == "prunesafe" {		// eliminare delta
		return s.pruneSafe(APIstub, args)
	}
	/*else if function == "gettopology" { 		//controller poiché si trova in ogni canale e ha il ledger di tutti
		return s.getTopology(APIstub, args)
	} else if function == "addevent" { 		//evento anomalo, si comporta di conseguenza e salva un log
		return s.addEvent(APIstub, args)
	}  else if function == "delete" {		//da vedere se può servire
		return s.delete(APIstub, args)
	} else if function == "getlastevent" {		//monitoring
		return s.getLastEvent(APIstub, args)
	}*/

	return shim.Error("Invalid Smart Contract function name on Rules smart contract.")
}



func (s *RulesContract) addRule (APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 4{ //rule_name, event, value and timestamp
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	name := args[0]
	timestamp := args[1]
	event := args[2]
	value  := args[3]


	getState, err := APIstub.GetState(name)

	if err != nil {
		return shim.Error(fmt.Sprintf("Error from getState into addRule: %s", err.Error()))
	}

	if bytes.Equal(getState,[]byte("")) {//then create new rule
		rule := Rule{name, timestamp, event, value}
		ruleAsBytes, marshalErr := json.Marshal(rule)
		if marshalErr != nil {
			return shim.Error(fmt.Sprintf("Could not marshal new %s rule: %s", name, marshalErr.Error()))
		}
		putErr := APIstub.PutState(name, ruleAsBytes)
		if putErr != nil {
			return shim.Error(fmt.Sprintf("Could not put new %s rule in the ledger: %s", name, putErr.Error()))
		}
		fmt.Println("Added new rule: ", rule)
		return shim.Success([]byte(fmt.Sprintf("Successfully added %s rule",  name )))
	} else { //updating rule
		// Retrieve info needed for the update procedure
		txid := APIstub.GetTxID()
		compositeIndexName := "name~timestamp~event~value~txID"

		// Create the composite key that will allow us to query for all deltas on a particular variable
		compositeKey, compositeErr := APIstub.CreateCompositeKey(compositeIndexName, []string{name,timestamp,event,value,txid})
		if compositeErr != nil {
			return shim.Error(fmt.Sprintf("Could not create a composite key for %s rule: %s", name, compositeErr.Error()))
		}		
		// Save the composite key index
		compositePutErr := APIstub.PutState(compositeKey, []byte{0x00})
		if compositePutErr != nil {
			return shim.Error(fmt.Sprintf("Could not put operation for %s in the ledger: %s", name, compositePutErr.Error()))
		}
		fmt.Println("Updated new rule: ", name)
		return shim.Success([]byte(fmt.Sprintf("Successfully updated %s",  name )))
	}
}






func (s *RulesContract) getRule(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Check we have a valid number of args
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1")
	}

	name := args[0]
//	rule := Rule{}

	//get rule to update
	getStateAsBytes, getErr := APIstub.GetState(name)
	if getErr != nil {
		return shim.Error(fmt.Sprintf("Could not get rule for %s: %s", name, getErr.Error()))
	}

	if bytes.Equal(getStateAsBytes, []byte("")){
		return shim.Error(fmt.Sprintf("Rule %s doesn't found", name))
	}
	fmt.Println("GetState of get function: "+string(getStateAsBytes[:]))
	return shim.Success(getStateAsBytes)
}






func (s *RulesContract) getRuleDelta(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Check we have a valid number of args
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1")
	}

	name := args[0]
	rule := Rule{}

	//get rule to update
	getStateAsBytes, getErr := APIstub.GetState(name)
	if getErr != nil {
		return shim.Error(fmt.Sprintf("Could not get rule for %s: %s", name, getErr.Error()))
	}

	if bytes.Equal(getStateAsBytes, []byte("")){
		return shim.Error(fmt.Sprintf("Rule %s doesn't found", name))
	}
	fmt.Println("GetState of get function: "+string(getStateAsBytes[:]))

	if err := json.Unmarshal(getStateAsBytes, &rule); err != nil{
		return shim.Error(err.Error())
	}

	// Get all deltas for the variable
	deltaResultsIterator, deltaErr := APIstub.GetStateByPartialCompositeKey("name~timestamp~event~value~txID", []string{name})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", name, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	// Check the variable existed
	if !deltaResultsIterator.HasNext() {
		return shim.Error(fmt.Sprintf("No Machine by the name %s exists", name))
	}

	// Iterate through result set and compute final value

	for deltaResultsIterator.HasNext() {
		// Get the next row
		responseRange, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error(nextErr.Error())
		}

		// Split the composite key into its component parts
		_, keyParts, splitKeyErr := APIstub.SplitCompositeKey(responseRange.Key)
		if splitKeyErr != nil {
			return shim.Error(splitKeyErr.Error())
		}

		fmt.Println("KeyParts:", keyParts)

		// Retrieve the delta state of rule
		rule.Timestamp =  keyParts[1]
		rule.Event = keyParts[2]
		rule.Value = keyParts[3]

	}
	
	ruleAsByte, marshallErr := json.Marshal(rule)
	if marshallErr != nil {
		return shim.Error(fmt.Sprintf("Marshall error into getRuleDelta: %s",marshallErr.Error()))
	}
	return shim.Success(ruleAsByte)

}






func (s *RulesContract) pruneSafe(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Verify there are a correct number of arguments
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, expecting 1 (the name of the variable to prune)")
	}

	// Get the var name
	name := args[0]

	// Get the var's value and process it
	getResp := s.getRuleDelta(APIstub, args)
	if getResp.Status == ERROR {
		return shim.Error(fmt.Sprintf("Could not retrieve the value of %s before pruning, pruning aborted: %s", name, getResp.Message))
	}
	
	rule := Rule{}
	unmarshalErr := json.Unmarshal(getResp.Payload, &rule)
	if unmarshalErr != nil {
		return shim.Error(fmt.Sprintf("Could not unmarshalErr the payload of %s, pruning aborted: %s", name, unmarshalErr.Error()))
	}
	fmt.Println("Rule after deltas: ",rule)

	// Store the var's value temporarily
	backupPutErr := APIstub.PutState(fmt.Sprintf("%s_PRUNE_BACKUP", name), getResp.Payload)
	if backupPutErr != nil {
		return shim.Error(fmt.Sprintf("Could not backup the value of %s before pruning, pruning aborted: %s", name, backupPutErr.Error()))
	}

	// Get all deltas for the variable
	deltaResultsIterator, deltaErr := APIstub.GetStateByPartialCompositeKey("name~timestamp~event~value~txID", []string{name})
	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve value for %s: %s", name, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	
	// Delete each row
	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {
		responseRange, nextErr := deltaResultsIterator.Next()

		if nextErr != nil {
			return shim.Error(fmt.Sprintf("Could not retrieve next row for pruning: %s", nextErr.Error()))
		}

		deltaRowDelErr := APIstub.DelState(responseRange.Key)
		if deltaRowDelErr != nil {
			return shim.Error(fmt.Sprintf("Could not delete delta row: %s", deltaRowDelErr.Error()))
		}
	}

	// Insert new row for the final value
	updateResp := s.addRule(APIstub, []string{name, rule.Event, rule.Value, rule.Timestamp})
	if updateResp.Status == ERROR {
		return shim.Error(fmt.Sprintf("Could not insert the final state of the rule after pruning, variable backup is stored in %s_PRUNE_BACKUP: %s", name, updateResp.Message))
	}

	// Delete the backup value
	delErr := APIstub.DelState(fmt.Sprintf("%s_PRUNE_BACKUP", name))
	if delErr != nil {
		return shim.Error(fmt.Sprintf("Could not delete backup value %s_PRUNE_BACKUP, this does not affect the ledger but should be removed manually", name))
	}

	return shim.Success([]byte(fmt.Sprintf("Successfully pruned rule %s, final state is %s with TS %s, %d rows pruned", name, rule.Value, rule.Timestamp, i)))
}










// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(RulesContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}

