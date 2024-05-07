/****************************************************************
 * Código de smart contract correspondiente al proyecto de grado
 * Álvaro Miguel Salinas Dockar
 * Universidad Católica Boliviana "San Pablo"
 * Ingeniería Mecatrónica
 * La Paz - Bolivia, 2020
 ***************************************************************/
package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"

	peer "github.com/hyperledger/fabric-protos-go/peer"
)

// VotingUCBChaincode es la estructura que representa este smart contract
type VotingUCBChaincode struct {
}

// Voter representa la estructura de datos que tiene el votante
type Voter struct {
	ObjectType string `json:"docType"`
	Identifier string `json:"identifier"`
	HasVoted   bool   `json:"hasVoted"`
}

// Candidate es la estructura que define a la persona u objeto candidato en la elección
type Candidate struct {
	ObjectType   string `json:"docType"`
	Organization string `json:"organization"`
	VoteCounter  uint64 `json:"voteCounter"`
}

// Config almacena la configuración en la emisión de votos, siempre con la llave "config"
type Config struct {
	ElectionOpen     bool `json:"electionOpen"`
	VotersRegistered bool `json:"votersRegistered"`
}

// Count es la estructura que maneja los datos finales
type Count struct {
	Rojo     uint64 `json:"rojo"`
	Amarillo uint64 `json:"amarillo"`
	Azul     uint64 `json:"azul"`
	Blanco   uint64 `json:"blanco"`
	Nulo     uint64 `json:"nulo"`
}

// Init es la función específica para inicializar los datos en el blockchain
func (voter *VotingUCBChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {

	fmt.Println("Ingreso de datos inicializado")

	/***************************************
	*Inicializando Variables Para candidatos
	****************************************/

	// Nótese el hardcode

	var azul = Candidate{
		ObjectType:   "candidate",
		Organization: "azul",
		VoteCounter:  0,
	}

	var rojo = Candidate{
		ObjectType:   "candidate",
		Organization: "rojo",
		VoteCounter:  0,
	}

	var amarillo = Candidate{
		ObjectType:   "candidate",
		Organization: "amarillo",
		VoteCounter:  0,
	}

	var blanco = Candidate{
		ObjectType:   "candidate",
		Organization: "blanco",
		VoteCounter:  0,
	}

	var nulo = Candidate{
		ObjectType:   "candidate",
		Organization: "nulo",
		VoteCounter:  0,
	}

	/****************************************************************************
	*Datos iniciales para ccandidatos almacenados en blockchain en forma de JSON
	*****************************************************************************/

	// Datos azul

	azulJSON, err := json.Marshal(azul)
	if err != nil {
		return errorResponse("Problemas al convertir a JSON", 500)
	}
	err = stub.PutState(azul.Organization, azulJSON)
	if err != nil {
		return errorResponse("No se pudieron ingresar los datos a la base de datos", 500)
	}

	// Datos rojo

	rojoJSON, err := json.Marshal(rojo)
	if err != nil {
		return errorResponse("Problemas al convertir a JSON", 500)
	}
	err = stub.PutState(rojo.Organization, rojoJSON)
	if err != nil {
		return errorResponse("No se pudieron ingresar los datos a la base de datos", 500)
	}

	// Datos amarillo

	amarilloJSON, err := json.Marshal(amarillo)
	if err != nil {
		return errorResponse("Problemas al convertir a JSON", 500)
	}
	err = stub.PutState(amarillo.Organization, amarilloJSON)
	if err != nil {
		return errorResponse("No se pudieron ingresar los datos a la base de datos", 500)
	}

	// Datos blanco

	blancoJSON, err := json.Marshal(blanco)
	if err != nil {
		return errorResponse("Problemas al convertir a JSON", 500)
	}
	err = stub.PutState(blanco.Organization, blancoJSON)
	if err != nil {
		return errorResponse("No se pudieron ingresar los datos a la base de datos", 500)
	}

	// Datos nulo

	nuloJSON, err := json.Marshal(nulo)
	if err != nil {
		return errorResponse("Problemas al convertir a JSON", 500)
	}
	err = stub.PutState(nulo.Organization, nuloJSON)
	if err != nil {
		return errorResponse("No se pudieron ingresar los datos a la base de datos", 500)
	}

	// Se declara la mesa como cerrada

	var initConfig = Config{
		ElectionOpen:     false,
		VotersRegistered: false,
	}

	confJSON, err := json.Marshal(initConfig)
	if err != nil {
		return errorResponse("Problemas al convertir a JSON", 500)
	}
	err = stub.PutState("config", confJSON)
	if err != nil {
		return errorResponse("No se pudieron ingresar los datos a la base de datos", 500)
	}

	return shim.Success([]byte("Se registraron a todos los candidatos de la elección"))
}

// Invoke Para manejar las funciones que acepta el blockchain
func (voter *VotingUCBChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	// Get the function name and parameters
	function, args := stub.GetFunctionAndParameters()

	fmt.Println("Invoke ejecutado : ", function, ", args = ", args)

	switch {

	// Funciones Invoke
	case function == "addNewVoters":
		return addNewVoters(stub, args)
	case function == "openElection":
		return openElection(stub)
	case function == "closeElection":
		return closeElection(stub)
	case function == "voteEmition":
		return voteEmition(stub, args)
	// Funciones Query
	case function == "voterStatusInspection":
		return voterStatusInspection(stub, args)
	case function == "candidateInspection":
		return candidateInspection(stub, args)
	case function == "voteCounting":
		return voteCounting(stub)
	}

	return errorResponse("Invalid function", 1)
}

// Manejador de respuestas de error
func errorResponse(err string, code uint) peer.Response {
	codeStr := strconv.FormatUint(uint64(code), 10)
	errorString := "{\"error\":" + err + ", \"code\":" + codeStr + " \" }"
	return shim.Error(errorString)
}

// Manejador de respuestas de exito
func successResponse(dat string) peer.Response {
	success := "{\"response\": " + dat + ", \"code\": 200 }"
	return shim.Success([]byte(success))
}

// Registro del chaincode con el Runtime de Fabric
func main() {

	// Mensaje de inicialización en consola
	fmt.Println("Inicializado el smart contract para la plicación de voto")

	// Registra el chaincode en el runtime de fabric

	err := shim.Start(new(VotingUCBChaincode))

	if err != nil {
		fmt.Printf("Error al inicializar el chaincode: %s", err)
	}
}
