/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ReadAsset reads the information from collection
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Asset, error) {

	log.Printf("ReadAsset: collection %v, ID %v", assetCollection, assetID)
	assetJSON, err := ctx.GetStub().GetPrivateData(assetCollection, assetID) //get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %v", err)
	}

	//No Asset found, return empty response
	if assetJSON == nil {
		log.Printf("%v does not exist in collection %v", assetID, assetCollection)
		return nil, nil
	}

	var asset *Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return asset, nil

}

// ReadAssetPrivateDetails reads the asset private details in organization specific collection
func (s *SmartContract) ReadAssetPrivateDetails(ctx contractapi.TransactionContextInterface, collection string, assetID string) (*AssetPrivateDetails, error) {
	log.Printf("ReadAssetPrivateDetails: collection %v, ID %v", collection, assetID)
	assetDetailsJSON, err := ctx.GetStub().GetPrivateData(collection, assetID) // Get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset details: %v", err)
	}
	if assetDetailsJSON == nil {
		log.Printf("AssetPrivateDetails for %v does not exist in collection %v", assetID, collection)
		return nil, nil
	}

	var assetDetails *AssetPrivateDetails
	err = json.Unmarshal(assetDetailsJSON, &assetDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return assetDetails, nil
}

// ReadTransferAgreement gets the buyer's identity from the transfer agreement from collection
func (s *SmartContract) ReadTransferAgreement(ctx contractapi.TransactionContextInterface, assetID string) (*TransferAgreement, error) {
	log.Printf("ReadTransferAgreement: collection %v, ID %v", assetCollection, assetID)
	// composite key for TransferAgreement of this asset
	transferAgreeKey, err := ctx.GetStub().CreateCompositeKey(transferAgreementObjectType, []string{assetID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	buyerIdentity, err := ctx.GetStub().GetPrivateData(assetCollection, transferAgreeKey) // Get the identity from collection
	if err != nil {
		return nil, fmt.Errorf("failed to read TransferAgreement: %v", err)
	}
	if buyerIdentity == nil {
		log.Printf("TransferAgreement for %v does not exist", assetID)
		return nil, nil
	}
	agreement := &TransferAgreement{
		ID:      assetID,
		BuyerID: string(buyerIdentity),
	}
	return agreement, nil
}

// GetAssetByRange performs a range query based on the start and end keys provided. Range
// queries can be used to read data from private data collections, but can not be used in
// a transaction that also writes to private data.
func (s *SmartContract) GetAssetByRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string) ([]*Asset, error) {

	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(assetCollection, startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []*Asset{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset *Asset
		err = json.Unmarshal(response.Value, &asset)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
		}

		results = append(results, asset)
	}

	return results, nil

}
