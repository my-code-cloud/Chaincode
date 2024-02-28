package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type OwnerData struct {
	Id					string 	`json:"id"`
	CampaignId			string 	`json:"campaignId"`
	Envelope			string 	`json:"envelope"`
	PrivacyPreference	string 	`json:"privacyPreference"`
	Url					string 	`json:"url"`
}

func (s *OwnerDataSmartContract) ShareData(ctx contractapi.TransactionContextInterface, id, campaignId, envelope, privacyPreference string) error {
    idExists, err := s.dataExists(ctx, id)
    if err != nil {
        return err
    }
    if idExists {
        return fmt.Errorf("Id %s already exists", id)
    }
    campaignExists, err := s.invokeQueryCampaign(ctx, campaignId)
    if err != nil {
        return err
    }
    if campaignExists == false {
        return fmt.Errorf("Campaign %s does not exist", campaignId) 
    }

    owner := OwnerData{
        Id:             	id,
		CampaignId: campaignId,
        Envelope:           envelope,
        PrivacyPreference:  privacyPreference,
    }

    ownerJSON, err := json.Marshal(owner)
    if err != nil {
        return err
    }

	err = ctx.GetStub().PutState(id, ownerJSON)

	if err != nil {
		return err
	}

	return nil
}

func (s *OwnerDataSmartContract) DeleteSharedData(ctx contractapi.TransactionContextInterface, id string) error {
    exists, err := s.dataExists(ctx, id)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("Error while deleting data: the data %s does not exist", id)
    }

    return ctx.GetStub().DelState(id)
}

func (s *OwnerDataSmartContract) dataExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	dataBytes, err := ctx.GetStub().GetState(id)
    if err != nil {
        return false, fmt.Errorf("Failed to read data %s from world state. %v", id, err)
    }
	if dataBytes == nil {
		return false, nil
	}

    return true, nil
}

func (s *OwnerDataSmartContract) invokeQueryCampaign(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
    params := []string{"CampaignExists", id}
	queryArgs := make([][]byte, len(params))
	for i, arg := range params {
		queryArgs[i] = []byte(arg)
	}

	response := ctx.GetStub().InvokeChaincode("campaign", queryArgs, "mychannel")
	if response.Status != shim.OK {
		return false, fmt.Errorf("Failed to query campaign chaincode. message %s status %s Got error: %s", response.Message, string(response.Status), response.Payload)
	}
    
	result, err := strconv.ParseBool(string(response.Payload))
    if err != nil {
        return false, fmt.Errorf("Failed to read campaign %s from world state. %v", id, err)
    }
    if result == false {
        return false, fmt.Errorf("Campaign does not exists, message %s status %s Got error: %s, campaignid: %s", response.Message, string(response.Status), response.Payload, id) 
    }
    return true, nil
}