package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type AnonymizedKG struct {
	Id					string 	`json:"id"`
	CampaignId			string 	`json:"campaignId"`
	RecipientId			string 	`json:"recipientId"`
	RollupEnvelope		string 	`json:"rollupEnvelope"`
	RecipientEnvelope	string 	`json:"recipientEnvelope"`
	Signature			string 	`json:"signature"`
	Verified			bool	`json:"verified"`
    Shared			    bool	`json:"shared"`
}

func (s *AnonymizedKGSmartContract) StoreAnonymizedKG(ctx contractapi.TransactionContextInterface, id, campaignId, recipientId, rollupEnvelope, signature string) error {
    idExists, err := s.anonymizedKGExists(ctx, id)
    if err != nil {
        return err
    }
    if idExists {
        return fmt.Errorf("Id %s already exists", id)
    }

    exists, err := s.invokeQueryCampaign(ctx, campaignId)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("Campaign %s does not exist", campaignId)
    }

    anonymizedKG := AnonymizedKG{
        Id:             	id,
		CampaignId: 		campaignId,
		RecipientId: 		recipientId,
        RollupEnvelope:     rollupEnvelope,
		RecipientEnvelope: 	"",
        Signature:  		signature,
		Verified: 			false,
        Shared: 			false,
    }

    anonymizedKGJSON, err := json.Marshal(anonymizedKG)
    if err != nil {
        return err
    }

	err = ctx.GetStub().PutState(id, anonymizedKGJSON)

	if err != nil {
		return err
	}

	return nil
}

func (s *AnonymizedKGSmartContract) StoreProof(ctx contractapi.TransactionContextInterface, KGId, userCommit, rollupCommit string) (bool, error) {
    idExists, err := s.anonymizedKGExists(ctx, KGId)
    if err != nil {
        return false, err
    }
    if !idExists {
        return false, fmt.Errorf("Id %s does not exist", KGId)
    }

    anonymizedKG, err := s.getAnonymizedKG(ctx, KGId)
	if err != nil {
		return false, fmt.Errorf("anonynimized KG %s does not exist", KGId)
	}
	anonymizedKG.Verified = rollupCommit == userCommit

    anonymizedKGJSON, err := json.Marshal(anonymizedKG)
    if err != nil {
        return rollupCommit == userCommit, err
    }

	return rollupCommit == userCommit, ctx.GetStub().PutState(KGId, anonymizedKGJSON)
}

func (s *AnonymizedKGSmartContract) ShareAnonymizedKGWithRecipient(ctx contractapi.TransactionContextInterface, KGId, campaignId, recipientId, recipientEnvelope string) error {
    campaignExists, _ := s.invokeQueryCampaign(ctx, campaignId)
    if campaignExists == false {
        return fmt.Errorf("Campaign %s does not exist", campaignId)
    }

    anonymizedKG, err := s.getAnonymizedKG(ctx, KGId)
	if err != nil {
		return fmt.Errorf("Anonymized KG %s does not exist", KGId)
	}
    if anonymizedKG.Verified != true {
        return fmt.Errorf("Anonymized KG %s is not verified by the rollup server", KGId)
    }
    if anonymizedKG.Shared != true {
        return fmt.Errorf("Anonymized KG %s has been already shared", KGId)
    }
    if recipientId != anonymizedKG.RecipientId {
        return fmt.Errorf("Anonymized KG %s has the wrong recipient id %s", KGId, recipientId)
    }

	anonymizedKG.RecipientEnvelope = recipientEnvelope
    anonymizedKG.Shared = true

    anonymizedKGJSON, err := json.Marshal(anonymizedKG)
    if err != nil {
        return err
    }

    ctx.GetStub().PutState(KGId, anonymizedKGJSON)

	return nil
}

func (s *AnonymizedKGSmartContract) DeleteAnonymizedKG(ctx contractapi.TransactionContextInterface, id string) error {
    exists, err := s.anonymizedKGExists(ctx, id)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("Error while deleting data: the data %s does not exist", id)
    }

    return ctx.GetStub().DelState(id)
}

func (s *AnonymizedKGSmartContract) anonymizedKGExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	KGBytes, err := ctx.GetStub().GetState(id)
    if err != nil {
        return false, fmt.Errorf("Failed to read KG %s from world state. %v", id, err)
    }
	if KGBytes == nil {
		return false, nil
	}
    return true, nil
}

func (s *AnonymizedKGSmartContract) getAnonymizedKG(ctx contractapi.TransactionContextInterface, id string) (*AnonymizedKG, error) {
	anonymizedKGBytes, err := ctx.GetStub().GetState(id)
    if err != nil {
        return nil, fmt.Errorf("Failed to read anonymized KG %s from world state. %v", id, err)
    }
	if anonymizedKGBytes == nil {
		return nil, fmt.Errorf("anonymized KG %s does not exist", id)
	}

	anonymizedKG := new(AnonymizedKG)
	_ = json.Unmarshal(anonymizedKGBytes, anonymizedKG)

    return anonymizedKG, nil
}

func (s *AnonymizedKGSmartContract) invokeQueryCampaign(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
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



/*----------------------------------------------------------------------------------------------------------------------------
// Only for caliper testing                                                                                                   |
// When updating an asset caliper gives MVCC conflict for race condition, because in the tests I always update the same asset |
// For an equal comparison I will do an query and then write on a new dummy asset id                                         |
//----------------------------------------------------------------------------------------------------------------------------*/

func (s *AnonymizedKGSmartContract) CaliperStoreProof(ctx contractapi.TransactionContextInterface, KGId, dummyId, userCommit, rollupCommit string) (bool, error) {
    idExists, err := s.anonymizedKGExists(ctx, KGId)
    if err != nil {
        return false, err
    }
    if !idExists {
        return false, fmt.Errorf("Id %s does not exist", KGId)
    }

    dummyAnonymizedKG := AnonymizedKG{
        Id:             	dummyId,
		CampaignId: 		"",
		RecipientId: 		"",
        RollupEnvelope:     "",
		RecipientEnvelope: 	"",
        Signature:  		"",
		Verified: 			rollupCommit == userCommit,
    }

    dummyAnonymizedKGJSON, err := json.Marshal(dummyAnonymizedKG)
    if err != nil {
        return rollupCommit == userCommit, err
    }

	return rollupCommit == userCommit, ctx.GetStub().PutState(dummyId, dummyAnonymizedKGJSON)
}

func (s *AnonymizedKGSmartContract) CaliperShareAnonymizedKGWithRecipient(ctx contractapi.TransactionContextInterface, KGId, dummyId1, dummyId2, campaignId, recipientId, recipientEnvelope string) error {
    idExists, err := s.anonymizedKGExists(ctx, KGId)
    if err != nil {
        return err
    }
    if !idExists {
        return fmt.Errorf("Id %s does not exist", KGId)
    }

    campaignExists, err := s.invokeQueryCampaign(ctx, campaignId)
    if err != nil {
        return err
    }
    if !campaignExists {
        return fmt.Errorf("Campaign %s does not exist", campaignId)
    }

    dummyAnonymizedKG := AnonymizedKG{
        Id:             	dummyId1,
		CampaignId: 		"",
		RecipientId: 		"",
        RollupEnvelope:     "",
		RecipientEnvelope: 	recipientEnvelope,
        Signature:  		"",
		Verified: 			true,
    }

    dummyAnonymizedKGJSON, err := json.Marshal(dummyAnonymizedKG)
    if err != nil {
        return err
    }

    ctx.GetStub().PutState(dummyAnonymizedKG.Id, dummyAnonymizedKGJSON)

	return nil
}