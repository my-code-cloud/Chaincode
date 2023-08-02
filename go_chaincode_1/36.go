package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for control the food
type SmartContract struct {
	contractapi.Contract
}

//Mask describes basic details
type Mask struct {
	Type   string  `json:"type"`
	Code   string  `json:"code"`
	Madeby string  `json:"madeby"`
	Owner  string  `json:"owner"`
	State  string  `json:"state"`
	Price  float32 `json:"price"`
}

// From Laurent question
type MaskTx struct {
	CodeId    string    `json:"codeId"`
	TxId      string    `json:"tx"`
	PrevOwner string    `json:"prevOwner"`
	NewOwner  string    `json:"newOwner"`
	Timestamp time.Time `json:"timestamp"`
}

// InitLedger adds a base set of cars to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	masks := []Mask{
		Mask{Type: "FP2", Code: "id:0:AX90", Madeby: "Spain", Owner: "Provider-A", State: "Available", Price: 1.2},
		Mask{Type: "FP1", Code: "id:0:AX91", Madeby: "Spain", Owner: "Provider-A", State: "Available", Price: 1.2},
		Mask{Type: "FP2", Code: "id:0:AX92", Madeby: "Spain", Owner: "Provider-A", State: "Available", Price: 1.2},
		Mask{Type: "FP1", Code: "id:0:AX93", Madeby: "Spain", Owner: "Provider-A", State: "Available", Price: 1.2},
		Mask{Type: "FP2", Code: "id:1:AX94", Madeby: "India", Owner: "Provider-B", State: "Available", Price: 1.1},
		Mask{Type: "FP1", Code: "id:1:AX95", Madeby: "India", Owner: "Provider-B", State: "Available", Price: 1.1},
		Mask{Type: "FP2", Code: "id:2:AX96", Madeby: "India", Owner: "Provider-B", State: "Available", Price: 1.1},
		Mask{Type: "FP1", Code: "id:2:AX97", Madeby: "India", Owner: "Provider-B", State: "Available", Price: 1.1},
		Mask{Type: "FP2", Code: "id:3:AX98", Madeby: "China", Owner: "Provider-C", State: "Available", Price: 1.05},
		Mask{Type: "FP3", Code: "id:3:AX99", Madeby: "China", Owner: "Provider-C", State: "Available", Price: 1.05},
	}

	for i, mask := range masks {
		maskAsBytes, _ := json.Marshal(mask)
		err := ctx.GetStub().PutState("id:"+strconv.Itoa(i), maskAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put mask in world state. %s", err.Error())
		}
	}

	return nil
}

// Create a new mask
func (s *SmartContract) CreateMask(ctx contractapi.TransactionContextInterface,
	maskId string,
	typeM string,
	madeBy string,
	owner string,
	code string,
	state string,
	price float32) error {

	// validate parameters if we dont want to update

	exists, err := s.MaskExists(ctx, maskId)

	if err != nil {
		return err
	}

	if !exists {
		mask := Mask{
			Type:   typeM,
			Code:   maskId + ":" + code,
			Madeby: madeBy,
			Owner:  owner,
			State:  state,
			Price:  price,
		}

		maskAsBytes, err := json.Marshal(mask)
		if err != nil {
			fmt.Printf("Marshal error: %s", err.Error())
			return err
		}
		return ctx.GetStub().PutState(maskId, maskAsBytes)
	} else {
		return fmt.Errorf("The mask %s  already exists", maskId)
	}
}

// Send a mask, change owner
func (s *SmartContract) SendMask(ctx contractapi.TransactionContextInterface,
	maskId string,
	owner string) (bool, error) {

	// validate parameters if we dont want to update

	exists, err := s.MaskExists(ctx, maskId)

	if err != nil {
		return false, err
	}

	if exists {
		maskBytes, err := ctx.GetStub().GetState(maskId)

		mask := new(Mask)
		err = json.Unmarshal(maskBytes, mask)
		if err != nil {
			return false, fmt.Errorf("Unmarshal error. %s", err.Error())
		}

		mask2 := Mask{
			Type:   mask.Type,
			Code:   mask.Code,
			Madeby: mask.Madeby,
			Owner:  owner,
			State:  "Sold",
			Price:  mask.Price,
		}

		maskAsBytes, err := json.Marshal(mask2)
		if err != nil {
			fmt.Printf("Marshal error: %s", err.Error())
			return false, err
		}
		ctx.GetStub().PutState(maskId, maskAsBytes)

		txid := ctx.GetStub().GetTxID()

		maskTx := MaskTx{
			CodeId:    mask.Code,
			TxId:      txid,
			PrevOwner: mask.Owner,
			NewOwner:  owner,
			Timestamp: time.Now(),
		}
		maskTxAsBytes, err := json.Marshal(maskTx)
		if err != nil {
			fmt.Printf("Marshal error: %s", err.Error())
			return false, err
		}
		err = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), maskTxAsBytes)
		if err != nil {
			return false, err
		}

		return true, nil
	} else {
		return false, fmt.Errorf("The mask %s  does not exist", maskId)
	}
}

// Update a mask
func (s *SmartContract) UpdateMask(ctx contractapi.TransactionContextInterface,
	maskId string,
	typeM string,
	madeBy string,
	owner string,
	code string,
	state string,
	price float32) error {

	// validate parameters if we dont want to update

	exists, err := s.MaskExists(ctx, maskId)

	if err != nil {
		return err
	}

	if exists {
		mask := Mask{
			Type:   typeM,
			Code:   maskId + ":" + code,
			Madeby: madeBy,
			Owner:  owner,
			State:  state,
			Price:  price,
		}

		maskAsBytes, err := json.Marshal(mask)
		if err != nil {
			fmt.Printf("Marshal error: %s", err.Error())
			return err
		}
		return ctx.GetStub().PutState(maskId, maskAsBytes)
	} else {
		return fmt.Errorf("The mask %s  does not exist", maskId)
	}
}

func (s *SmartContract) GetMask(ctx contractapi.TransactionContextInterface, maskId string) (*Mask, error) {

	maskAsBytes, err := ctx.GetStub().GetState(maskId)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if maskAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", maskId)
	}

	mask := new(Mask)

	err = json.Unmarshal(maskAsBytes, mask)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal error. %s", err.Error())
	}

	return mask, nil
}

func (s *SmartContract) GetMasksByState(ctx contractapi.TransactionContextInterface, state string) (int, []*Mask, error) {

	queryString := fmt.Sprintf(`{"selector":{"state":"%s"}}`, state)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return 0, nil, err
	}

	defer resultsIterator.Close()

	var maskArray []*Mask
	var counter int
	counter = 0
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return 0, nil, err
		}
		var m Mask
		err = json.Unmarshal(queryResult.Value, &m)
		if err != nil {
			return 0, nil, err
		}
		maskArray = append(maskArray, &m)
		counter = counter + 1
	}

	return counter, maskArray, nil
}

func (s *SmartContract) GetMasksTxByCode(ctx contractapi.TransactionContextInterface, code string) (int, []*Mask, error) {

	queryString := fmt.Sprintf(`{"selector":{"codeId":"%s"}}`, code)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return 0, nil, err
	}

	defer resultsIterator.Close()

	var maskArray []*Mask
	var counter int
	counter = 0
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return 0, nil, err
		}
		var m Mask
		err = json.Unmarshal(queryResult.Value, &m)
		if err != nil {
			return 0, nil, err
		}
		maskArray = append(maskArray, &m)
		counter = counter + 1
	}

	return counter, maskArray, nil
}

func (s *SmartContract) DeleteMask(ctx contractapi.TransactionContextInterface, maskId string) error {

	exists, err := s.MaskExists(ctx, maskId)

	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("The mask %s does not exist", maskId)
	}

	return ctx.GetStub().DelState(maskId)
}

// MaskExists returns true when asset with given ID exists in world state
func (s *SmartContract) MaskExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {

	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// ChangeCarOwner updates the owner field of car with given id in world state
func (s *SmartContract) ChangeMaskOwner(ctx contractapi.TransactionContextInterface, maskId string, newOwner string) (bool, error) {

	maskBytes, err := ctx.GetStub().GetState(maskId)
	if err != nil {
		return false, fmt.Errorf("Failed to read Mask Info from world state: %v", err)
	}

	if maskBytes == nil {
		return false, fmt.Errorf("The Mask object does not exist")
	}

	var mask Mask
	err = json.Unmarshal(maskBytes, &mask)
	if err != nil {
		return false, err
	}

	txid := ctx.GetStub().GetTxID()

	maskTx := MaskTx{
		CodeId:    mask.Code,
		TxId:      txid,
		PrevOwner: mask.Owner,
		NewOwner:  newOwner,
		Timestamp: time.Now(),
	}

	mask.Owner = newOwner

	maskBytes, err = json.Marshal(mask)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(maskId, maskBytes)
	if err != nil {
		return false, err
	}

	var txBytes []byte
	txBytes, err = json.Marshal(maskTx)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), txBytes)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *SmartContract) GetAllMasks(ctx contractapi.TransactionContextInterface) ([]*Mask, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all masks in the chaincode namespace.

	// Get iterator object
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var masks []*Mask
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var mask Mask
		err = json.Unmarshal(queryResponse.Value, &mask)
		if err != nil {
			return nil, err
		}
		masks = append(masks, &mask)
	}

	return masks, nil
}

func (s *SmartContract) GetTotalMasks(ctx contractapi.TransactionContextInterface) (int, error) {

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return 0, err
	}
	defer resultsIterator.Close()

	var counter int
	counter = 0
	for resultsIterator.HasNext() {
		counter = counter + 1
		_, err := resultsIterator.Next()
		if err != nil {
			return 0, err
		}
	}

	return counter, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create provider chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting provider chaincode: %s", err.Error())
	}
}
