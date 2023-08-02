package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Token struct {
	Name   string `json:"name"`
	Type  string `json:"type"`
	Rate string `json:"rate"`
	Owner  string `json:"owner"`
}

type QueryResult struct {
	Key    string `json:"Key"`
	Record *Token
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	tokens := []Token{
		Token{Type: "chickenPoints", Name: "Sysco-Chicken-points", Rate: "10 points = 1 sysco token", Owner: "Tomoko"},
		Token{Type: "applePoints", Name: "Sysco-Apple-points", Rate: "20 points = 1 sysco token", Owner: "Brad"},
		Token{Type: "pizzaPoints", Name: "PizzaHut-Pizza-points", Rate: "5 points = 1 sysco token", Owner: "Jin Soo"},
		Token{Type: "mangoPoints", Name: "mangoSupplier-mango-points", Rate: "40 points = 1 sysco token", Owner: "Max"},
		Token{Type: "coffeePoints", Name: "Starbucks-Coffee-points", Rate: "50 points = 1 sysco token", Owner: "Adriana"},
		Tokenr{Type: "teaPoints", Name: "Sysco-tea-points", Rate: "40 points = 1 sysco token", Owner: "Michel"},
		Token{Type: "utensilPoints", Name: "Sysco-Utensil-points", Rate: "7 points = 1 sysco token", Owner: "Aarav"},
		Token{Type: "bananaPoints", Name: "bananaFarm-banana-points", Rate: "60 points = 1 sysco token", Owner: "Pari"},
		
	}

	for i, token := range tokens {
		tokenAsBytes, _ := json.Marshal(token)
		err := ctx.GetStub().PutState("TOKEN"+strconv.Itoa(i), tokenAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

func (s *SmartContract) CreateToken(ctx contractapi.TransactionContextInterface, tokenNumber string, name string, type string, rate string, owner string) error {
	token := Token{
		Name   name,
		Type:  type,
		Rate: rate,
		Owner:  owner,
	}

	tokenAsBytes, _ := json.Marshal(token)

	return ctx.GetStub().PutState(tokenNumber, tokenAsBytes)
}

func (s *SmartContract) QueryToken(ctx contractapi.TransactionContextInterface, tokenNumber string) (*Token, error) {
	tokenAsBytes, err := ctx.GetStub().GetState(tokenNumber)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if tokenAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", tokenNumber)
	}

	token := new(Token)
	_ = json.Unmarshal(tokenAsBytes, token)

	return token, nil
}

func (s *SmartContract) QueryAllTokens(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := "TOKEN0"
	endKey := "TOKEN9"

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		token := new(token)
		_ = json.Unmarshal(queryResponse.Value, token)

		queryResult := QueryResult{Key: queryResponse.Key, Record: token}
		results = append(results, queryResult)
	}

	return results, nil
}

func (s *SmartContract) ChangeTokenOwner(ctx contractapi.TransactionContextInterface, tokenNumber string, newOwner string) error {
	token, err := s.QueryToken(ctx, tokenNumber)

	if err != nil {
		return err
	}

	token.Owner = newOwner

	tokenAsBytes, _ := json.Marshal(token)

	return ctx.GetStub().PutState(tokenNumber, tokenAsBytes)
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}
