package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for control the food
type SmartContract struct {
	contractapi.Contract
}

// Information about the productor 
type Productor struct {
	// introducir id del productor
	Name string `json:"name"`  // Nombre productor <= outchain DB 
	Location string `json:"location"` // Ubicacion productor <= outchain DB
}

// Informacion del producto
type Product struct {
	Name string `json:"name"` // Nombre del producto
	Kind string `json:"variety"`
	Quantity int `json:"quantity"` // Cantidad del producto en gramos
}

//Data es la estructura a guardar en la base de datos
type Data struct {
	Owner  Productor `json:"productor"` // contiene informacion del prductor
	Product Product `json:"product"`	// Contiene informacion del producto
}

func (s *SmartContract) Set(ctx contractapi.TransactionContextInterface, DataId string, ProductorName string, Location string, ProductName string, ProductKind string , Quantity int) error {

	// Validar si la data ya esta en la red.
	dataAsBytes, err := ctx.GetStub().GetState(DataId)
	if err != nil {
		return fmt.Errorf("Failed to read from world state. %s", err.Error())
	}
	if dataAsBytes == nil {
		
		//validaciones de negocio

		productor := Productor {
			Name: ProductorName,
			Location: Location,
		}
		
		product := Product {
			Name: ProductName,
			Kind: ProductKind,
			Quantity: Quantity,
		}
		
		data := Data {
			Owner: productor,
			Product: product,
		}
		
		dataAsBytes, err := json.Marshal(data)
		
		if err != nil {
			fmt.Printf("Marshal error: %s", err.Error())
			return err
		}
		
		return ctx.GetStub().PutState(DataId, dataAsBytes)
	} else {
		fmt.Printf("Data id exists, to modify state please use method Update: %s", err.Error())
		return err
	}
}

func (s *SmartContract) Edit(ctx contractapi.TransactionContextInterface, DataId string, ProductorName string, Location string, ProductName string, ProductKind string , Quantity int) error {

	//Validaciones de sintaxis

	productor := Productor {
		Name: ProductorName,
		Location: Location,
	}

	product := Product {
		Name: ProductName,
		Kind: ProductKind,
		Quantity: Quantity,
	}

	data := Data {
		Owner: productor,
		Product: product,
	}

	dataAsBytes, err := json.Marshal(data)

	if err != nil {
		fmt.Printf("Marshal error: %s", err.Error())
		return err
	}

	return ctx.GetStub().PutState(DataId, dataAsBytes)
}


func (s *SmartContract) Query(ctx contractapi.TransactionContextInterface, dataId string) (*Data, error) {

	dataAsBytes, err := ctx.GetStub().GetState(dataId)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if dataAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", dataId)
	}

	data := new(Data)

	err = json.Unmarshal(dataAsBytes, data)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal error. %s", err.Error())
	}

	return data, nil
}



func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create foodcontrol chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting foodcontrol chaincode: %s", err.Error())
	}
}
