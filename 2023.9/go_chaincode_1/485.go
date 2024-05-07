package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	"strconv"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
  	"github.com/streadway/amqp"
)

type serverConfig struct {
	CCID    string
	Address string
}

// SmartContract provides functions for managing an asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Asset struct {
	ProductID   int `json:"productID"`
	Quantity  int `json:"quantity"`
	Owner  string `json:"owner"`
}

// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResult struct {
	Record    *Asset    `json:"record"`
	TxId     string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Asset
}

// InitLedger adds a base set of cars to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ProductID: 0, Quantity: 10, Owner: "United Airlines"},
		{ProductID: 1, Quantity: 5, Owner: "Delta"},
		{ProductID: 2, Quantity: 3, Owner: "Spirit"},
		{ProductID: 3, Quantity: 7, Owner: "Frontier"},
		{ProductID: 4, Quantity: 9, Owner: "Alaska Airlines"},
		{ProductID: 5, Quantity: 6, Owner: "Southwest Airlines"},
		{ProductID: 6, Quantity: 12, Owner: "JetBlue"},
		{ProductID: 7, Quantity: 3, Owner: "Hawaiian Airlines"},
		{ProductID: 8, Quantity: 3, Owner: "Allegiant Air"},
		{ProductID: 9, Quantity: 7, Owner: "Boeing"},
	}

	for i, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState("PART"+strconv.Itoa(i), assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
	}

	return nil
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, airlinePartNumber string, productID int, quantity int, owner string) error {
	asset := Asset{
		ProductID:   productID,
		Quantity: quantity,
		Owner:  owner,
	}

	airlinePartAsBytes, err := json.Marshal(asset)

	if err != nil {
    	return err
	}

	conn, err := amqp.Dial("amqps://wfsdzxpt:UdYJ3pVxVAEEtnP6RYBzs1fnvbTaocKb@gull.rmq.cloudamqp.com/wfsdzxpt")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.Publish(
		"",     // exchange
		"HF_ML", // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte("{\"message\": \"addProduct\", \"productID\": " + strconv.Itoa(asset.ProductID) + ", \"quantity\": " + strconv.Itoa(asset.Quantity) + "}\n"),
		})
	failOnError(err, "Failed to publish a message")

	return ctx.GetStub().PutState(airlinePartNumber, airlinePartAsBytes)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, airlinePartNumber string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(airlinePartNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", airlinePartNumber)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// TransferAsset updates the owner field of asset with given id in world state.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, airlinePartNumber string, newOwner string) error {
	asset, err := s.ReadAsset(ctx, airlinePartNumber)
	if err != nil {
		return err
	}

	asset.Owner = newOwner
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(airlinePartNumber, assetJSON)
}

func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, airlinePartNumber string, newOwner string) error {
	asset, err := s.ReadAsset(ctx, airlinePartNumber)

	if err != nil {
		return err
	}

	asset.Owner = newOwner

	airlinePartAsBytes, _ := json.Marshal(asset)

	conn, err := amqp.Dial("amqps://wfsdzxpt:UdYJ3pVxVAEEtnP6RYBzs1fnvbTaocKb@gull.rmq.cloudamqp.com/wfsdzxpt")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.Publish(
		"",     // exchange
		"HF_ML", // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte("{\"message\": \"sellProduct\", \"productID\": " + strconv.Itoa(asset.ProductID) + ", \"quantity\": " + strconv.Itoa(asset.Quantity) + ", \"buyer\": " + asset.Owner + "}\n"),
		})
	failOnError(err, "Failed to publish a message")

	return ctx.GetStub().PutState(airlinePartNumber, airlinePartAsBytes)
}

func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	return assetJSON != nil, nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	// See chaincode.env.example
	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	chaincode, err := contractapi.NewChaincode(&SmartContract{})

	if err != nil {
		log.Panicf("error create asset-transfer-basic chaincode: %s", err)
	}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      chaincode,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	if err := server.Start(); err != nil {
		log.Panicf("error starting asset-transfer-basic chaincode: %s", err)
	}
}
