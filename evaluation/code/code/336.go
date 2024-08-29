/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

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

// GetAllAssetPrivateCollection performs a range query based on the start and end keys provided. Range
// queries can be used to read data from private data collections, but can not be used in
// a transaction that also writes to private data.
func (s *SmartContract) GetAllAssetPrivateCollection(ctx contractapi.TransactionContextInterface, collectionName string) ([]*AssetPrivateDetails, error) {

	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(collectionName, "", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []*AssetPrivateDetails{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset *AssetPrivateDetails
		err = json.Unmarshal(response.Value, &asset)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
		}

		results = append(results, asset)
	}

	return results, nil

}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// QueryAssetByINE queries for assets based on ine.
// =========================================================================================
// The result set is built and returned as a byte array containing the JSON results.
func (s *SmartContract) QueryAssetByINE(ctx contractapi.TransactionContextInterface, ine string) ([]*AssetPublicDetails, error) {
	queryString := fmt.Sprintf(`{"selector":{"ineClv":"%s"}}`, ine)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// constructQueryResponseFromIterator constructs a slice of assets from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*AssetPublicDetails, error) {
	var assets []*AssetPublicDetails
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset AssetPublicDetails
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// PrivateQueryAssetByINE queries for assets based on ine.
// =========================================================================================
// The result set is built and returned as a byte array containing the JSON results.
func (s *SmartContract) PrivateQueryAssetByINE(ctx contractapi.TransactionContextInterface, collectionName string, ine string) ([]*AssetPrivateDetails, error) {
	queryString := fmt.Sprintf(`{"selector":{"ineClv":"%s"}}`, ine)
	resultsIterator, err := ctx.GetStub().GetPrivateDataQueryResult(collectionName, queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructPrivateQueryResponseFromIterator(resultsIterator)
}

// PrivateExistAssetByStringField queries for assets based on ine.
// =========================================================================================
// The result set is built and returned as a byte array containing the JSON results.
func (s *SmartContract) PrivateExistAssetByStringField(ctx contractapi.TransactionContextInterface, collectionName string, assetID string, key string, value string) (bool, error) {
	queryString := fmt.Sprintf(`{"selector":{"%s":"%s"}}`, key, value)
	resultsIterator, err := ctx.GetStub().GetPrivateDataQueryResult(collectionName, queryString)
	if err != nil {
		return false, fmt.Errorf("failed to read asset: %v", err)
	}

	if resultsIterator.HasNext() != false {
		queryResult, err := resultsIterator.Next()
		defer resultsIterator.Close()
		if err != nil {
			return false, fmt.Errorf("failed to get asset: %v", err)
		}
		var asset AssetPrivateDetails
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return false, fmt.Errorf("failed to mix asset: %v", err)
		}
		if asset.ID != assetID {
			log.Printf("Asset exist in collection")
			return true, nil
		}
	}

	log.Printf("Asset does not exist in collection")
	return false, nil
}

// constructQueryResponseFromIterator constructs a slice of assets from the resultsIterator
func constructPrivateQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*AssetPrivateDetails, error) {
	var assets []*AssetPrivateDetails
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset AssetPrivateDetails
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
