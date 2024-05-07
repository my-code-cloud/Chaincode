package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ABstore Chaincode implementation
type ABstore struct {
	contractapi.Contract
}

// Init init smart contract
func (t *ABstore) Init(ctx contractapi.TransactionContextInterface, a string, aVal int, b string, bVal int) error {
	fmt.Println("ABstore Init")
	var err error
	// Initialize the chaincode
	fmt.Printf("Aval = %d, Bval = %d\n", aVal, bVal)
	// Write the state to the ledger
	err = ctx.GetStub().PutState(a, []byte(strconv.Itoa(aVal)))
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(b, []byte(strconv.Itoa(bVal)))
	if err != nil {
		return err
	}

	return nil
}

// Invoke transaction makes payment of X units from A to B
func (t *ABstore) Invoke(ctx contractapi.TransactionContextInterface, a, b string, x int) error {
	var (
		err  error
		aVal int
		bVal int
	)

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	aValbytes, err := ctx.GetStub().GetState(a)
	if err != nil {
		return fmt.Errorf("Failed to get state")
	}
	if aValbytes == nil {
		return fmt.Errorf("Entity not found")
	}
	aVal, _ = strconv.Atoi(string(aValbytes))

	bValbytes, err := ctx.GetStub().GetState(b)
	if err != nil {
		return fmt.Errorf("Failed to get state")
	}
	if bValbytes == nil {
		return fmt.Errorf("Entity not found")
	}
	bVal, _ = strconv.Atoi(string(bValbytes))

	// Perform the execution
	aVal = aVal - x
	bVal = bVal + x
	fmt.Printf("Aval = %d, Bval = %d\n", aVal, bVal)

	// Write the state back to the ledger
	err = ctx.GetStub().PutState(a, []byte(strconv.Itoa(aVal)))
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(b, []byte(strconv.Itoa(bVal)))
	if err != nil {
		return err
	}

	return nil
}

// Delete  an entity from state
func (t *ABstore) Delete(ctx contractapi.TransactionContextInterface, a string) error {
	// Delete the key from the state in ledger
	err := ctx.GetStub().DelState(a)
	if err != nil {
		return fmt.Errorf("Failed to delete state")
	}

	return nil
}

// Query callback representing the query of a chaincode
func (t *ABstore) Query(ctx contractapi.TransactionContextInterface, a string) (string, error) {
	var err error
	// Get the state from the ledger
	aValBytes, err := ctx.GetStub().GetState(a)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + a + "\"}"
		return "", errors.New(jsonResp)
	}

	if aValBytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + a + "\"}"
		return "", errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + a + "\",\"Amount\":\"" + string(aValBytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return string(aValBytes), nil
}

func main() {
	cc, err := contractapi.NewChaincode(new(ABstore))
	if err != nil {
		panic(err.Error())
	}
	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting ABstore chaincode: %s", err)
	}
}
