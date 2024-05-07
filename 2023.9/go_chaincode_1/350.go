package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type CanonicalTransactionContract struct {
	contractapi.Contract
}

type TransactionInfo struct {
	ChaincodeName   string   `json:"chaincodeName"`
	TransactionName string   `json:"transactionName"`
	Args            []string `json:"args"`
}

type Batch struct {
	Timestamp     int64             `json:"timestamp"`
	Transactions  []TransactionInfo `json:"transactions"`
	PrevStateHash []byte            `json:"prevStateHash"`
	NewStateHash  []byte            `json:"newStateHash"`
}

type BatchReadyForCommit struct {
	Timestamp     int64             `json:"timestamp"`
	Transactions  []TransactionInfo `json:"transactions"`
	PrevStateHash []byte            `json:"prevStateHash"`
	NewStateHash  []byte            `json:"newStateHash"`
	PrevBatchHash []byte            `json:"prevBatchHash"`
}

const latestBatchHashKey = "latestBatchHash"

func (ctc *CanonicalTransactionContract) CommitBatch(ctx contractapi.TransactionContextInterface, payload string) error {
	var batch Batch
	if err := json.Unmarshal([]byte(payload), &batch); err != nil {
		return fmt.Errorf("failed to unmarshal batch payload. Error: %s", err.Error())
	}

	var batchReadyForCommit BatchReadyForCommit;

	latestBatchHash, err := ctx.GetStub().GetState(latestBatchHashKey)
	if err != nil {
		return fmt.Errorf("failed to check if latestBatchKey already exists in world state. Error: %s", err.Error())
	}

	//TODO: check if prevStateHash matches newStateHash of latest batch
	batchReadyForCommit = BatchReadyForCommit {
		Timestamp:     batch.Timestamp,
		Transactions:  batch.Transactions,
		PrevStateHash: batch.PrevStateHash,
		NewStateHash:  batch.NewStateHash,
		PrevBatchHash: latestBatchHash,
	}

	batchJSON, err := json.Marshal(batchReadyForCommit)
	if err != nil {
		return fmt.Errorf("failed to marshal batch struct. Error: %s", err.Error())
	}

	batchHash := crypto.Keccak256(batchJSON)
	err = ctx.GetStub().PutState(hex.EncodeToString(batchHash), batchJSON)
	if err != nil {
		return fmt.Errorf("failed to add batch into world state. Error: %s", err.Error())
	}

	err = ctx.GetStub().PutState(latestBatchHashKey, batchHash)
	if err != nil {
		return fmt.Errorf("failed to add latestBatchHash into world state. Error: %s", err.Error())
	}
	return nil;
}


func (ctc *CanonicalTransactionContract) GetBatchHistory(ctx contractapi.TransactionContextInterface) (string, error) {
	var history []string

	//TODO: handle errors

	latestBatchHash, err := ctx.GetStub().GetState(latestBatchHashKey)
	if err != nil {
		return "", fmt.Errorf("failed to check if latestBatchKey already exists in world state. Error: %s", err.Error())
	}

	if latestBatchHash == nil {
		return "", fmt.Errorf("latest batch hash does not exist in world state. Are you sure you added a batch?")
	}

	rawBatch, err := ctx.GetStub().GetState(hex.EncodeToString(latestBatchHash))
	if err != nil {
		return "", fmt.Errorf("failed to check if latest batch already exists in world state. Error: %s", err.Error())
	}

	for rawBatch != nil {
		var batch BatchReadyForCommit
		json.Unmarshal(rawBatch, &batch)
		history = append(history, string(rawBatch))

		rawBatch, _ = ctx.GetStub().GetState(hex.EncodeToString(batch.PrevBatchHash))
	}

	return strings.Join(history, ";"), nil
}

func (ctc *CanonicalTransactionContract) DeleteBatchChain (ctx contractapi.TransactionContextInterface) {
	ctx.GetStub().DelState(latestBatchHashKey);
}


func main() {
	chaincode, err := contractapi.NewChaincode(new(CanonicalTransactionContract))

	if err != nil {
		fmt.Printf("Error create statecontract chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting statecontract chaincode: %s", err.Error())
	}
}
