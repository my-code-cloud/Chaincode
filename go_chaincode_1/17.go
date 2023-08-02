/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type serverConfig struct {
	CCID    string
	Address string
}

// SmartContract provides functions for managing an asset
type SmartContract struct {
	contractapi.Contract
}

type Source struct {
	Type string `json:"type"`
	ID   string `json:"ID"`
}

// Asset describes basic details of what makes up a simple asset
type PointsTransaction struct {
	ID 		   string  `json:"ID"`
	Value     int     `json:"value"`
	// Merchant   string  `json:"merchant"`
	CreatedAt  string  `json:"created_at"`
	Sender     string  `json:"sender"`
	Receiver   string  `json:"receiver"`
	Source     *Source `json:"source"`
}

type MerchantPoints struct {
	ID   string 	`json:"ID"`
	Value int 		`json:"value"`
}

// Customer or Merchant
type Member struct {
	ID					string 				`json:"ID"`
	Merchant   			string  			`json:"merchant"`
	MerchantPoints  	map[string]int      `json:"merchantPoints"`
	Points 				int 				`json:"points"`
	Transaction 		*PointsTransaction 	`json:"transaction"`
}


// InitLedger adds a base set of points transactions to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// transaction1 := PointsTransaction{
	// 	ID: "12738647",
	// 	Value: 1000,
	// 	CreatedAt: "20210929",
	// 	Sender: "zh-CN",
	// 	Receiver: "jin.xiaoming@ekohe.com",
	// 	Source: &Source{Type: "Birthday"},
	// }

	transaction2 := PointsTransaction{
		ID: "12738648",
		Value: 500,
		CreatedAt: "20211009",
		Sender: "zh-TW",
		Receiver: "maxime@ekohe.com",
		Source: &Source{Type: "Order", ID: "737463747"},
	}

	transaction3 := PointsTransaction{
		ID: "12738649",
		Value: 800,
		CreatedAt: "20211011",
		Sender: "jin.xiaoming@ekohe.com",
		Receiver: "zh-TW",
		Source: &Source{Type: "Order", ID: "345342523"},
	}

	
	// Pass []byte{0x00} as null value, as pass a 'nil' value will effectively delete the key from state
	members := []Member{
		Member{ID: "zh-CN", Points: 1000, MerchantPoints: map[string]int{"zh-TW": 800}},
		Member{ID: "zh-TW", Points: 500, MerchantPoints: map[string]int{"jp": 500}},
		Member{ID: "jp", Points: 0, MerchantPoints: map[string]int{}},
		Member{ID: "jin.xiaoming@ekohe.com", Merchant: "zh-CN", Points: 200, Transaction: &transaction3, MerchantPoints: map[string]int{"zh-CN": 1000, "zh-TW": -800}},
		Member{ID: "maxime@ekohe.com", Merchant: "jp", Points: 500, Transaction: &transaction2, MerchantPoints: map[string]int{"zh-TW": 500}},
	}

	for _, member := range members {
		memberAsBytes, _ := json.Marshal(member)
		err := ctx.GetStub().PutState(member.ID, memberAsBytes)

		if err != nil {
			return fmt.Errorf("failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

func (s *SmartContract) GetMember(ctx contractapi.TransactionContextInterface, id string) (*Member, error) {
	bytes, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	if bytes == nil {
		return nil, fmt.Errorf("%s does not exist in world state", id)
	}

	var member Member
	err = json.Unmarshal(bytes, &member)
	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (s *SmartContract) GetAllMerchants(ctx contractapi.TransactionContextInterface) ([]Member, error) {
	return nil, nil
}

func (s *SmartContract) GetCustomersByMerchant(ctx contractapi.TransactionContextInterface, merchant string) ([]Member, error) {
	return nil, nil
}

func (s *SmartContract) GetAllMembers(ctx contractapi.TransactionContextInterface) ([]Member, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []Member{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		member := new(Member)
		err = json.Unmarshal(queryResponse.Value, member)
		if err != nil {
			return nil, err
		}

		// queryResult := QueryResult{Key: queryResponse.Key, Record: pointsTransaction}
		results = append(results, *member)
	}

	return results, nil
}

func (s *SmartContract) CreateMember(ctx contractapi.TransactionContextInterface, id string, merchant string) (*Member) {
	member, _ := s.GetMember(ctx, id)

	if member != nil {
		return member
	}

	if id == merchant {
		merchant = ""
	}

	member = &Member{
		ID: id,
		Merchant: merchant,
		Points: 0,
		Transaction: nil,
		MerchantPoints: map[string]int{},
	}

	memberAsBytes, _ := json.Marshal(member)
	ctx.GetStub().PutState(member.ID, memberAsBytes)

	return member
}

func (s *SmartContract) CreateTransaction(ctx contractapi.TransactionContextInterface, id string, senderKey string, receiverKey string, value int, merchant string, createdAt string, sourceType string, sourceId string) error {
	// exists, err := s.AssetExists(ctx, id)
	transaction := PointsTransaction{
		ID: id,
		Value: value,
		CreatedAt: createdAt,
		Sender: senderKey,
		Receiver: receiverKey,
		Source: &Source{
			Type: sourceType,
			ID: sourceId,
		},
	}

	sender := s.CreateMember(ctx, senderKey, merchant)
	receiver := s.CreateMember(ctx, receiverKey, merchant)

	if sender.Merchant == "" && receiver.Merchant != "" {
		// Case1: Customer get points from a merchant
		receiver.Points += value
		receiver.Transaction = &transaction
		receiver.MerchantPoints[sender.ID] += value

		sender.Points += value
		if sender.ID != receiver.Merchant {
			sender.MerchantPoints[receiver.Merchant] += value
		}
	} else if sender.Merchant != "" && receiver.Merchant == "" {
		// Case2: A merchant receive customer's points by using it in order purchase
		if sender.Points < value {
			// TODO: Alert error about points is not enough to purchase
			return fmt.Errorf("%s does not have enough points", sender.ID)
		}

		sender.Points -= value
		sender.Transaction = &transaction
		sender.MerchantPoints[receiver.ID] -= value

		if sender.Merchant != receiver.ID {
			receiver.MerchantPoints[sender.Merchant] -= value
		} else {
			receiver.Points -= value
		}
	} else if sender.Merchant == "" && receiver.Merchant == "" {
		// Case 3: Transaction between two merchants
		sender.Points += value
		sender.MerchantPoints[receiver.ID] += value
	} else if sender.Merchant != "" && receiver.Merchant != "" {
		// Case 4: Customer give points to others as gift
		if sender.Points < value {
			// TODO: Alert error about points is not enough to purchase
			return fmt.Errorf("%s does not have enough points", sender.ID)
		}

		sender.Points -= value
		sender.Transaction = &transaction
		sender.MerchantPoints[sender.Merchant] -= value

		receiver.Points += value
		receiver.Transaction = &transaction
		receiver.MerchantPoints[sender.Merchant] += value
	}


	senderAsBytes, _ := json.Marshal(sender)
	senderErr := ctx.GetStub().PutState(sender.ID, senderAsBytes)

	if senderErr != nil {
		return senderErr
	}

	receiverAsBytes, _ := json.Marshal(receiver)
	receiverErr := ctx.GetStub().PutState(receiver.ID, receiverAsBytes)

	return receiverErr
}

func main() {
	// See chaincode.env.example
	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	chaincode, err := contractapi.NewChaincode(&SmartContract{})

	if err != nil {
		log.Panicf("error create points-transfer chaincode: %s", err)
	}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      chaincode,
		TLSProps: getTLSProperties(),
	}

	if err := server.Start(); err != nil {
		log.Panicf("error starting points-transfer chaincode: %s", err)
	}
}

func getTLSProperties() shim.TLSProperties {
	// Check if chaincode is TLS enabled
	tlsDisabledStr := getEnvOrDefault("CHAINCODE_TLS_DISABLED", "true")
	key := getEnvOrDefault("CHAINCODE_TLS_KEY", "")
	cert := getEnvOrDefault("CHAINCODE_TLS_CERT", "")
	clientCACert := getEnvOrDefault("CHAINCODE_CLIENT_CA_CERT", "")

	// convert tlsDisabledStr to boolean
	tlsDisabled := getBoolOrDefault(tlsDisabledStr, false)
	var keyBytes, certBytes, clientCACertBytes []byte
	var err error

	if !tlsDisabled {
		keyBytes, err = ioutil.ReadFile(key)
		if err != nil {
			log.Panicf("error while reading the crypto file: %s", err)
		}
		certBytes, err = ioutil.ReadFile(cert)
		if err != nil {
			log.Panicf("error while reading the crypto file: %s", err)
		}
	}
	// Did not request for the peer cert verification
	if clientCACert != "" {
		clientCACertBytes, err = ioutil.ReadFile(clientCACert)
		if err != nil {
			log.Panicf("error while reading the crypto file: %s", err)
		}
	}

	return shim.TLSProperties{
		Disabled: tlsDisabled,
		Key: keyBytes,
		Cert: certBytes,
		ClientCACerts: clientCACertBytes,
	}
}

func getEnvOrDefault(env, defaultVal string) string {
	value, ok := os.LookupEnv(env)
	if !ok {
		value = defaultVal
	}
	return value
}

// Note that the method returns default value if the string
// cannot be parsed!
func getBoolOrDefault(value string, defaultVal bool) bool {
	parsed, err := strconv.ParseBool(value)
	if err!= nil {
		return defaultVal
	}
	return parsed
}
