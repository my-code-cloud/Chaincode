/*
SPDX-License-Identifier: Apache-2.0
*/

package auction

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// PrivateBid is the structure of a bid in private state
type PrivateBid struct {
	Type     string `json:"objectType"`
	Quantity int    `json:"quantity"`
	Org      string `json:"org"`
	Buyer    string `json:"buyer"`
	Price    int    `json:"price"`
}

// PrivateAsk is the structure of a bid in private state
type PrivateAsk struct {
	Type     string `json:"objectType"`
	Quantity int    `json:"quantity"`
	Org      string `json:"org"`
	Seller   string `json:"seller"`
	Price    int    `json:"price"`
}

func (s *SmartContract) GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {

	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	return string(decodeID), nil
}

// getCollectionName is an internal helper function to get collection of submitting client identity.
func getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get the MSP ID of submitting client identity
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get verified MSPID: %v", err)
	}

	// Create the collection name
	orgCollection := "_implicit_org_" + clientMSPID

	return orgCollection, nil
}

// verifyClientOrgMatchesPeerOrg is an internal function used to verify that client org id matches peer org id.
func verifyClientOrgMatchesPeerOrg(ctx contractapi.TransactionContextInterface) error {

	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the client's MSPID: %v", err)
	}
	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the peer's MSPID: %v", err)
	}

	if clientMSPID != peerMSPID {
		return fmt.Errorf("client from org %v is not authorized to read or write private data from an org %v peer", clientMSPID, peerMSPID)
	}

	return nil
}

// checkBidOwner returns an error if a client who is not the bid owner
// tries to query a bid
func (s *SmartContract) checkBidOwner(ctx contractapi.TransactionContextInterface, collection string, bidKey string) error {

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	bidJSON, err := ctx.GetStub().GetPrivateData(collection, bidKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", bidKey, err)
	}
	if bidJSON == nil {
		return fmt.Errorf("bid %v does not exist", bidKey)
	}

	var bid *PrivateBid
	err = json.Unmarshal(bidJSON, &bid)
	if err != nil {
		return err
	}

	// check that the client querying the bid is the bid owner
	if bid.Buyer != clientID {
		return fmt.Errorf("Permission denied, client id %v is not the owner of the bid", clientID)
	}

	return nil
}

// checkAskOwner returns an error if a client who is not the bid owner
// tries to query a bid
func (s *SmartContract) checkAskOwner(ctx contractapi.TransactionContextInterface, collection string, askKey string) error {

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	askJSON, err := ctx.GetStub().GetPrivateData(collection, askKey)
	if err != nil {
		return fmt.Errorf("failed to get ask %v: %v", askKey, err)
	}
	if askJSON == nil {
		return fmt.Errorf("ask %v does not exist", askKey)
	}

	var ask *PrivateAsk
	err = json.Unmarshal(askJSON, &ask)
	if err != nil {
		return err
	}

	// check that the client querying the bid is the bid owner
	if ask.Seller != clientID {
		return fmt.Errorf("Permission denied, client id %v is not the owner of the ask", clientID)
	}

	return nil
}
