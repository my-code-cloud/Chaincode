

package main

import (
	"fmt"
	"encoding/json"
	"bytes"
	"strconv"
//	"time"

//	"github.com/sparrc/go-ping"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"

        /*** SSH ***/
	"crypto/sha256"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"strings"
	"encoding/hex"
)


type NodesContract struct {
}


// Define Status codes for the response
const (
	OK    = 200
	ERROR = 500
)

//key=port, tipo regola = iptables , oggetto=nodo su cui fai le regole, value=opzionale 
type Node struct {
	Name		string	`json:"name"` //nome della regola, univoca altrimenti aggiorno il valore anziché crearlo
	Timestamp	string 	`json:"timestamp"`
	ip 		string 	`json:"event"`
}


// Init is called when the smart contract is instantiated
func (s *NodesContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	fmt.Println("Chaincode istanziato")
	return shim.Success(nil)
}



func (s *NodesContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "sshandaddnode" {			
		return s.sshAndAddNode(APIstub, args)
	} else if function == "getnodes" {
		return s.getNodes(APIstub, args)
	} else if function == "initledger" {
		return s.initLedger(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name on Nodes smart contract.")
}





func (s *NodesContract) initLedger (APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 0{ 
		return shim.Error("Incorrect number of arguments. Expecting 0")
	}

	var nodes [6]Node;	
	for i := 0; i< len(nodes); i++ {

		nodes[i].Name="E"+strconv.Itoa(i)
		fmt.Println("Nodes:",nodes[i])
	}

	i := 0
	for i < len(nodes) {
	//	fmt.Println("i is ", i)
		nodeAsBytes, _ := json.Marshal(nodes[i])
		APIstub.PutState(nodes[i].Name, nodeAsBytes)
		fmt.Println("Added", nodes[i])
		i = i + 1
	}
	return shim.Success(nil)
}




/****************** SSH ************************/

func VerifyAndAdd(cmd string, host string, sshConfig *ssh.ClientConfig) ([]byte, error) {

	conn, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		fmt.Println("Failed to dial: "+ err.Error())
		return nil, err
	}

	sess, err := conn.NewSession()
	if err != nil {
		fmt.Println("Failed to create session: "+ err.Error())
		return nil, err
	}
	defer sess.Close()

	var b bytes.Buffer
	sess.Stdout = &b
	if err := sess.Run(cmd); err != nil {
		fmt.Println("Failed to run: "+ err.Error())
		return nil, err
	}
	array_listed := strings.Split(b.String(), "\n")
	array_listed = array_listed[:len(array_listed)-1]

	listed_string := strings.Join(array_listed[:]," ")
	h := sha256.New()
	h.Write([]byte(listed_string))
	return h.Sum(nil), nil
}


func Find(slice []string, val string) (int, bool) {
    for i, item := range slice {
        if item == val {
            return i, true
        }
    }
    return -1, false
}


func GetLegitHashList() []string {
	return []string{"62c59fb8b70c2b3643d8a37974d5504a8b60e40429eb58f6c7d382b5e590f552", 
			"0dda1445fe66ba542b0d0e7ea701f6f1e41db1db9cfad2ee3f8cfe26a3b003fb", 
			"d00d2ec7bf2398427bc1aa28e8b9e92b559ea1bb35521213363a772acda5beb9",
                        "6f91bdfa397fa9438c04f039f3e10dfe2efa111e4e06b922f7a91ff2de6ff666",
                        "171777dd870a9bd5a0b9d0785dde946e58ab6fe7da5ce58fd2cb2b9eee43b84f"}
}

/****************** FINE SSH ************************/





func (s *NodesContract) sshAndAddNode (APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3{ 
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	name := args[0]
	timestamp := args[1]
	ip := args[2] //"192.168.1.107:6022"

	//##### ssh to node #####
	log.Info("test")
	sshConfig := &ssh.ClientConfig {
		User: "ubuntu", //pi
		Auth: []ssh.AuthMethod{
			ssh.Password("ssh_password"), //raspberry
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	//			 data 2021-03-22                 salto le prime 3    stampo alcune info
	var cmd string = "ls -al --time-style=long-iso /usr/bin/ | awk 'NR>=4' |  awk '{print $1, $5, $6, $7, $8;}'"

	hash, err := VerifyAndAdd(cmd, ip, sshConfig)
	if (err != nil){
		return shim.Error( fmt.Sprintf("ERROR: %s", err.Error()) )
	}
	//fmt.Printf(" %x\n", hash)

	str_hash := hex.EncodeToString(hash)
	_, found := Find(GetLegitHashList(),str_hash)
	if !found {
		return shim.Error( fmt.Sprintf("ERROR: hash is not correct. Ip host: %s; hash: %s", ip, str_hash) )
	}

	//##### add node #####
	getState, err := APIstub.GetState(name)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error from getState into sshAndAddNode: %s", err.Error()))
	}
	if bytes.Equal(getState,[]byte("")) {//then create new node
		node := Node{name, timestamp, ip}
		nodeAsBytes, marshalErr := json.Marshal(node)
		if marshalErr != nil {
			return shim.Error(fmt.Sprintf("Could not marshal new %s node: %s", name, marshalErr.Error()))
		}
		putErr := APIstub.PutState(name, nodeAsBytes)
		if putErr != nil {
			return shim.Error(fmt.Sprintf("Could not put new %s node in the ledger: %s", name, putErr.Error()))
		}

		//emit add node event
		eventPayload := "Node "+name+" with ip "+ip+" is added"
		payloadAsBytes := []byte(eventPayload)
		eventErr := APIstub.SetEvent("sshAndAddNodeEvent",payloadAsBytes)
		if (eventErr != nil) {
		  return shim.Error(fmt.Sprintf("Failed to emit sshAndAddNode event"))
		}

		fmt.Println("Added new node: ", node)
		return shim.Success([]byte(fmt.Sprintf("Successfully added %s node",  name )))
	}

	return shim.Error("Error in sshAndAddNode.")

}






func (s *NodesContract) getNodes(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Check we have a valid number of args
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments, expecting 0")
	}


	resultsIterator, err := APIstub.GetStateByRange("","")
	if err != nil {
		fmt.Println("Errore getStateByRange")
		return shim.Error(fmt.Sprintf("Errore getStateByRange -> %s",err.Error()))
	}
	defer resultsIterator.Close()

	var nodes []Node //[]byte //il risultato di tutte le macchine
	var node Node //byte //variabile machine temporanea per poi assegnarla all'array con append
 
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		json.Unmarshal(queryResponse.Value, &node)
		nodes = append(nodes, node)
	//	fmt.Println("Added into machines array:", machine.Name)
	}

	if len(nodes) == 0 {
		return shim.Error("Errore, nodes array is empty.")
	} else {
		fmt.Println("Len Nodes array: ",len(nodes))
	}

	nodesAsBytes, _ := json.Marshal(nodes)
	return shim.Success(nodesAsBytes)
}













// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(NodesContract))
	if err != nil {
		fmt.Printf("Error creating new Nodes Smart Contract: %s", err)
	}
}



