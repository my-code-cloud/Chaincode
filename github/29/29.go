package chaincode

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"github.com/golang/protobuf/ptypes"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

/*--------Phase 2 code-------------*/

const assetCollection = "assetCollection"
const transferAgreementObjectType = "transferAgreement"

type AssetPrivateDetails struct {
	ID     string `json:"assetID"`
	Secret int    `json:"secret"`
}

type TransferAgreement struct {
	ID      string `json:"assetID"`
	BuyerID string `json:"buyerID"`
}

func (s *SmartContract) CreatePrivateAsset(ctx contractapi.TransactionContextInterface) error {

	temp := ctx.GetClientIdentity().AssertAttributeValue("retailer", "true")
	if temp == nil {
		return fmt.Errorf("submitting client not authorized to create asset, he is a Retailer")
	}

	err := ctx.GetClientIdentity().AssertAttributeValue("farmer", "true")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to create asset, he is not a Farmer")
	}

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		//log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}

	type assetTransientInput struct {
		//0Type           string `json:"objectType"` //Type is used to distinguish the various types of objects in state database
		ID             string    `json:"assetID"`
		Color          string    `json:"color"`
		Weight         int       `json:"weight"`
		Timestamp      time.Time `json:"timestamp"`
		Creator        string    `json:creator`
		AppraisedValue int       `json:"appraisedValue"`
		TransferedTo   string    `json:transferedTo`
		//these 2 should be private
		Secret int `json:secret`
	}

	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetInput.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}
	if len(assetInput.Color) == 0 {
		return fmt.Errorf("color field must be a non-empty string")
	}
	if assetInput.Weight <= 0 {
		return fmt.Errorf("Weight field must be a positive integer")
	}
	if assetInput.AppraisedValue <= 0 {
		return fmt.Errorf("appraisedValue field must be a positive integer")
	}

	if assetInput.Secret <= 0 {
		return fmt.Errorf("Secret field is needed ")
	}

	//might not be needed
	// Check if asset already exists
	assetAsBytes, err := ctx.GetStub().GetPrivateData(assetCollection, assetInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if assetAsBytes != nil {
		fmt.Println("Asset already exists: " + assetInput.ID)
		return fmt.Errorf("this asset already exists: " + assetInput.ID)
	}

	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	creatorDN, err := s.GetSubmittingClientDN(ctx)
	if err != nil {
		return err
	}
	//Get timestamp of PrivateData creation

	txTimestamp, error := ctx.GetStub().GetTxTimestamp()
	if error != nil {
		return error
	}
	timestamp, erri := ptypes.Timestamp(txTimestamp)
	if erri != nil {
		return erri
	}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	// Make submitting client the owner
	asset := Asset{
		ID:             assetInput.ID,
		Color:          assetInput.Color,
		Weight:         assetInput.Weight,
		Owner:          clientID,
		Timestamp:      timestamp,
		Creator:        creatorDN,
		AppraisedValue: assetInput.AppraisedValue,
		TransferedTo:   ""}

	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}

	err = ctx.GetStub().PutState(assetInput.ID, assetJSONasBytes) //puts data in public
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}

	// Save asset details to collection visible to owning organization
	assetPrivateDetails := AssetPrivateDetails{
		ID:     assetInput.ID,
		Secret: assetInput.Secret,
	}

	assetPrivateDetailsAsBytes, err := json.Marshal(assetPrivateDetails) // marshal asset details to JSON
	if err != nil {
		return fmt.Errorf("failed to marshal into JSON: %v", err)
	}

	// Get collection name for this organization.
	orgCollection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	// Put asset appraised value into owners org specific private data collection
	log.Printf("Put: collection %v, ID %v", orgCollection, assetInput.ID)
	err = ctx.GetStub().PutPrivateData(orgCollection, assetInput.ID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}
	return nil
}

func (s *SmartContract) AgreeToTransfer(ctx contractapi.TransactionContextInterface) error {

	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	// Value is private, therefore it gets passed in transient field
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Persist the JSON bytes as-is so that there is no risk of nondeterministic marshaling.
	valueJSONasBytes, ok := transientMap["asset_value"]
	if !ok {
		return fmt.Errorf("asset_value key not found in the transient map")
	}

	// Unmarshal the tranisent map to get the asset ID.
	var valueJSON AssetPrivateDetails
	err = json.Unmarshal(valueJSONasBytes, &valueJSON)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Do some error checking since we get the chance
	if len(valueJSON.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}
	if valueJSON.Secret <= 0 {
		return fmt.Errorf("appraisedValue field must be a positive integer")
	}

	// Read asset from the private data collection
	asset, err := s.ReadAsset(ctx, valueJSON.ID)
	if err != nil {
		return fmt.Errorf("error reading asset: %v", err)
	}
	if asset == nil {
		return fmt.Errorf("%v does not exist", valueJSON.ID)
	}
	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("AgreeToTransfer cannot be performed: Error %v", err)
	}

	// Get collection name for this organization. Needs to be read by a member of the organization.
	orgCollection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	log.Printf("AgreeToTransfer Put: collection %v, ID %v", orgCollection, valueJSON.ID)
	// Put agreed value in the org specifc private data collection
	err = ctx.GetStub().PutPrivateData(orgCollection, valueJSON.ID, valueJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset bid: %v", err)
	}

	// Create agreeement that indicates which identity has agreed to purchase
	// In a more realistic transfer scenario, a transfer agreement would be secured to ensure that it cannot
	// be overwritten by another channel member
	transferAgreeKey, err := ctx.GetStub().CreateCompositeKey(transferAgreementObjectType, []string{valueJSON.ID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	log.Printf("AgreeToTransfer Put: collection %v, ID %v, Key %v", assetCollection, valueJSON.ID, transferAgreeKey)
	err = ctx.GetStub().PutState(transferAgreeKey, []byte(clientID))
	if err != nil {
		return fmt.Errorf("failed to put asset bid: %v", err)
	}

	return nil
}

func (s *SmartContract) TransferPrivateAsset(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field
	transientTransferJSON, ok := transientMap["asset_owner"]
	if !ok {
		return fmt.Errorf("asset owner not found in the transient map")
	}

	type assetTransferTransientInput struct {
		ID       string `json:"assetID"`
		BuyerMSP string `json:"buyerMSP"`
	}

	var assetTransferInput assetTransferTransientInput
	err = json.Unmarshal(transientTransferJSON, &assetTransferInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetTransferInput.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}
	if len(assetTransferInput.BuyerMSP) == 0 {
		return fmt.Errorf("buyerMSP field must be a non-empty string")
	}
	log.Printf("TransferAsset: verify asset exists ID %v", assetTransferInput.ID)
	// Read asset from world State
	asset, err := s.ReadAsset(ctx, assetTransferInput.ID)
	if err != nil {
		return fmt.Errorf("error reading asset: %v", err)
	}
	if asset == nil {
		return fmt.Errorf("%v does not exist", assetTransferInput.ID)
	}
	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("TransferAsset cannot be performed: Error %v", err)
	}

	// Verify transfer details and transfer owner
	err = s.verifyAgreement(ctx, assetTransferInput.ID, asset.Owner, assetTransferInput.BuyerMSP)
	if err != nil {
		return fmt.Errorf("failed transfer verification: %v", err)
	}

	transferAgreement, err := s.ReadTransferAgreement(ctx, assetTransferInput.ID)
	if err != nil {
		return fmt.Errorf("failed ReadTransferAgreement to find buyerID: %v", err)
	}
	if transferAgreement.BuyerID == "" {
		return fmt.Errorf("BuyerID not found in TransferAgreement for %v", assetTransferInput.ID)
	}

	// Transfer asset in private data collection to new owner
	asset.Owner = transferAgreement.BuyerID

	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed marshalling asset %v: %v", assetTransferInput.ID, err)
	}

	log.Printf("TransferAsset Put: collection %v, ID %v", assetCollection, assetTransferInput.ID)
	err = ctx.GetStub().PutState(assetTransferInput.ID, assetJSONasBytes) //rewrite the asset
	if err != nil {
		return err
	}

	// Get collection name for this organization
	ownersCollection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	// Delete the asset appraised value from this organization's private data collection
	err = ctx.GetStub().DelPrivateData(ownersCollection, assetTransferInput.ID)
	if err != nil {
		return err
	}

	// Delete the transfer agreement from the asset collection
	transferAgreeKey, err := ctx.GetStub().CreateCompositeKey(transferAgreementObjectType, []string{assetTransferInput.ID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().DelState(transferAgreeKey)
	if err != nil {
		return err
	}

	return nil

}

// verifyAgreement is an internal helper function used by TransferAsset to verify
// that the transfer is being initiated by the owner and that the buyer has agreed
// to the same appraisal value as the owner
func (s *SmartContract) verifyAgreement(ctx contractapi.TransactionContextInterface, assetID string, owner string, buyerMSP string) error {

	// Check 1: verify that the transfer is being initiatied by the owner

	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if clientID != owner {
		return fmt.Errorf("error: submitting client identity does not own asset")
	}

	// Check 2: verify that the buyer has agreed to the appraised value

	// Get collection names
	collectionOwner, err := getCollectionName(ctx) // get owner collection from caller identity
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	collectionBuyer := buyerMSP + "PrivateCollection" // get buyers collection

	// Get hash of owners agreed to value
	ownerAppraisedValueHash, err := ctx.GetStub().GetPrivateDataHash(collectionOwner, assetID)
	if err != nil {
		return fmt.Errorf("failed to get hash of appraised value from owners collection %v: %v", collectionOwner, err)
	}
	if ownerAppraisedValueHash == nil {
		return fmt.Errorf("hash of appraised value for %v does not exist in collection %v", assetID, collectionOwner)
	}

	// Get hash of buyers agreed to value
	buyerAppraisedValueHash, err := ctx.GetStub().GetPrivateDataHash(collectionBuyer, assetID)
	if err != nil {
		return fmt.Errorf("failed to get hash of appraised value from buyer collection %v: %v", collectionBuyer, err)
	}
	if buyerAppraisedValueHash == nil {
		return fmt.Errorf("hash of appraised value for %v does not exist in collection %v. AgreeToTransfer must be called by the buyer first", assetID, collectionBuyer)
	}

	// Verify that the two hashes match
	if !bytes.Equal(ownerAppraisedValueHash, buyerAppraisedValueHash) {
		return fmt.Errorf("hash for appraised value for owner %x does not value for seller %x", ownerAppraisedValueHash, buyerAppraisedValueHash)
	}

	return nil
}

// DeleteAsset can be used by the owner of the asset to delete the asset
func (s *SmartContract) DeletePrivateAsset(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("Error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field
	transientDeleteJSON, ok := transientMap["asset_delete"]
	if !ok {
		return fmt.Errorf("asset to delete not found in the transient map")
	}

	type assetDelete struct {
		ID string `json:"assetID"`
	}

	var assetDeleteInput assetDelete
	err = json.Unmarshal(transientDeleteJSON, &assetDeleteInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetDeleteInput.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}

	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("DeleteAsset cannot be performed: Error %v", err)
	}

	log.Printf("Deleting Asset: %v", assetDeleteInput.ID)
	valAsbytes, err := ctx.GetStub().GetPrivateData(assetCollection, assetDeleteInput.ID) //get the asset from chaincode state
	if err != nil {
		return fmt.Errorf("failed to read asset: %v", err)
	}
	if valAsbytes == nil {
		return fmt.Errorf("asset not found: %v", assetDeleteInput.ID)
	}

	ownerCollection, err := getCollectionName(ctx) // Get owners collection
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	//check the asset is in the caller org's private collection
	valAsbytes, err = ctx.GetStub().GetPrivateData(ownerCollection, assetDeleteInput.ID)
	if err != nil {
		return fmt.Errorf("failed to read asset from owner's Collection: %v", err)
	}
	if valAsbytes == nil {
		return fmt.Errorf("asset not found in owner's private Collection %v: %v", ownerCollection, assetDeleteInput.ID)
	}

	// delete the asset from state
	err = ctx.GetStub().DelPrivateData(assetCollection, assetDeleteInput.ID)
	if err != nil {
		return fmt.Errorf("failed to delete state: %v", err)
	}

	// Finally, delete private details of asset
	err = ctx.GetStub().DelPrivateData(ownerCollection, assetDeleteInput.ID)
	if err != nil {
		return err
	}

	return nil

}

// DeleteTranferAgreement can be used by the buyer to withdraw a proposal from
// the asset collection and from his own collection.
func (s *SmartContract) DeleteTranferAgreement(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field
	transientDeleteJSON, ok := transientMap["agreement_delete"]
	if !ok {
		return fmt.Errorf("asset to delete not found in the transient map")
	}

	type assetDelete struct {
		ID string `json:"assetID"`
	}

	var assetDeleteInput assetDelete
	err = json.Unmarshal(transientDeleteJSON, &assetDeleteInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetDeleteInput.ID) == 0 {
		return fmt.Errorf("transient input ID field must be a non-empty string")
	}

	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("DeleteTranferAgreement cannot be performed: Error %v", err)
	}
	// Delete private details of agreement
	orgCollection, err := getCollectionName(ctx) // Get proposers collection.
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}
	tranferAgreeKey, err := ctx.GetStub().CreateCompositeKey(transferAgreementObjectType, []string{assetDeleteInput.
		ID}) // Create composite key
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	valAsbytes, err := ctx.GetStub().GetPrivateData(assetCollection, tranferAgreeKey) //get the transfer_agreement
	if err != nil {
		return fmt.Errorf("failed to read transfer_agreement: %v", err)
	}
	if valAsbytes == nil {
		return fmt.Errorf("asset's transfer_agreement does not exist: %v", assetDeleteInput.ID)
	}

	log.Printf("Deleting TranferAgreement: %v", assetDeleteInput.ID)
	err = ctx.GetStub().DelPrivateData(orgCollection, assetDeleteInput.ID) // Delete the asset
	if err != nil {
		return err
	}

	// Delete transfer agreement record
	err = ctx.GetStub().DelPrivateData(assetCollection, tranferAgreeKey) // remove agreement from state
	if err != nil {
		return err
	}

	return nil

}

// getCollectionName is an internal helper function to get collection of submitting client identity.
func getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get the MSP ID of submitting client identity
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get verified MSPID: %v", err)
	}

	// Create the collection name
	orgCollection := clientMSPID + "PrivateCollection"

	return orgCollection, nil
}

// verifyClientOrgMatchesPeerOrg is an internal function used verify client org id and matches peer org id.
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

/*-------End of phase 2 code -------*/

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Asset struct {
	ID             string    `json:"ID"`
	Color          string    `json:"color"`
	Weight         int       `json:"weight"`
	Owner          string    `json:"owner"`
	AppraisedValue int       `json:"appraisedValue"`
	Timestamp      time.Time `json:"timestamp"`
	Creator        string    `json:creator`
	TransferedTo   string    `json:transferedTo`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	temp := ctx.GetClientIdentity().AssertAttributeValue("retailer", "true")
	if temp == nil {
		return fmt.Errorf("submitting client not authorized to create asset, he is a Retailer")
	}

	err := ctx.GetClientIdentity().AssertAttributeValue("farmer", "true")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to create asset, he is not a Farmer")
	}

	timeS, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}
	timestamp, err := ptypes.Timestamp(timeS)
	if err != nil {
		return err
	}
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	creatorDN, err := s.GetSubmittingClientDN(ctx)
	if err != nil {
		return err
	}

	assets := []Asset{
		{ID: "asset1", Color: "blue", Weight: 5, Owner: clientID, AppraisedValue: 30, Timestamp: timestamp, Creator: creatorDN, TransferedTo: ""},
		{ID: "asset2", Color: "orange", Weight: 5, Owner: clientID, AppraisedValue: 40, Timestamp: timestamp, Creator: creatorDN, TransferedTo: ""},
		{ID: "asset3", Color: "green", Weight: 10, Owner: clientID, AppraisedValue: 50, Timestamp: timestamp, Creator: creatorDN, TransferedTo: ""},
		{ID: "asset4", Color: "yellow", Weight: 10, Owner: clientID, AppraisedValue: 60, Timestamp: timestamp, Creator: creatorDN, TransferedTo: ""},
		{ID: "asset5", Color: "black", Weight: 15, Owner: clientID, AppraisedValue: 70, Timestamp: timestamp, Creator: creatorDN, TransferedTo: ""},
		{ID: "asset6", Color: "pink", Weight: 15, Owner: clientID, AppraisedValue: 80, Timestamp: timestamp, Creator: creatorDN, TransferedTo: ""},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, weight int, appraisedValue int) error {

	txTimestamp, error := ctx.GetStub().GetTxTimestamp()
	if error != nil {
		return error
	}
	timestamp, erri := ptypes.Timestamp(txTimestamp)
	if erri != nil {
		return erri
	}

	temp := ctx.GetClientIdentity().AssertAttributeValue("retailer", "true")
	if temp == nil {
		return fmt.Errorf("submitting client not authorized to create asset, he is a Retailer")
	}

	err := ctx.GetClientIdentity().AssertAttributeValue("farmer", "true")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to create asset, he is not a Farmer")
	}

	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	// Get ID of submitting client identity

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	creatorDN, err := s.GetSubmittingClientDN(ctx)
	if err != nil {
		return err
	}

	asset := Asset{
		ID:             id,
		Color:          color,
		Weight:         weight,
		Owner:          clientID,
		AppraisedValue: appraisedValue,
		Timestamp:      timestamp,
		Creator:        creatorDN,
		TransferedTo:   ""}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, newColor string, newWeight int, newValue int) error {

	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if clientID != asset.Owner {
		return fmt.Errorf("submitting client not authorized to update asset, does not own asset")
	}

	asset.Color = newColor
	asset.Weight = newWeight
	asset.AppraisedValue = newValue

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes a given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {

	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if clientID != asset.Owner {
		return fmt.Errorf("submitting client not authorized to update asset, does not own asset")
	}

	return ctx.GetStub().DelState(id)
}

// TransferAsset updates the owner field of asset with given id in world state.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {

	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if clientID != asset.Owner {
		return fmt.Errorf("submitting client not authorized to update asset, does not own asset")
	}
	asset.TransferedTo = clientID + " is transfering " + id + " to " + newOwner
	asset.Owner = newOwner
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {

	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// GetSubmittingClientIdentity returns the name and issuer of the identity that
// invokes the smart contract. This function base64 decodes the identity string
// before returning the value to the client or smart contract.
//files is located at pkg/cid/cid.go for GetID() on sourcegraph.com
//returns x509::CN=FarmerO,OU=org1+OU=client+OU=department1::CN=ca.org1.example.com,O=org1.example.com,L=Durham,ST=North Carolina,C=US
//on GetId() => ("x509::%s::%s", getDN(&c.cert.Subject), getDN(&c.cert.Issuer)
//DN is distinguished name as defined by RFC 2253
/* https://sourcegraph.com/github.com/hyperledger/fabric-chaincode-go@38d29fabecb9916a8a1ecbd0facb72f2ac32d016/-/blob/pkg/cid/cid.go?L76 */
func (s *SmartContract) GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {

	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	clientName := _between(string(decodeID), "x509::CN=", ",")
	return clientName, nil
}

//GetSubmittingClientDN returns the Distinguished Name as defined by RFC 2253
func (s *SmartContract) GetSubmittingClientDN(ctx contractapi.TransactionContextInterface) (string, error) {

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

//Function to get string between two strings.
func _between(value string, a string, b string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}
func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}

// HistoryQueryResult structure used for returning result of history query
//got it from asset-transfer-ledger-queries
type HistoryQueryResult struct {
	Record    *Asset    `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// GetAssetHistory returns the chain of custody for an asset since issuance.
//got it from asset-transfer-ledger-queries
func (s *SmartContract) GetAssetHistory(ctx contractapi.TransactionContextInterface, assetID string) ([]HistoryQueryResult, error) {
	log.Printf("GetAssetHistory: ID %v", assetID)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(assetID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var prevOwner string = ""
	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, err
			}
		} else {
			asset = Asset{
				ID: assetID,
			}
		}
		if prevOwner == (&asset).Owner {
			continue
		}
		prevOwner = (&asset).Owner

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &asset,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {

	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
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

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// ReadTransferAgreement gets the buyer's identity from the transfer agreement from collection
func (s *SmartContract) ReadTransferAgreement(ctx contractapi.TransactionContextInterface, assetID string) (*TransferAgreement, error) {
	log.Printf("ReadTransferAgreement: collection %v, ID %v", assetCollection, assetID)
	// composite key for TransferAgreement of this asset
	transferAgreeKey, err := ctx.GetStub().CreateCompositeKey(transferAgreementObjectType, []string{assetID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	buyerIdentity, err := ctx.GetStub().GetState(transferAgreeKey) // Get the state from world state
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

// ===== Example: Parameterized rich query =================================================

// QueryAssetByOwner queries for assets based on assetType, owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (s *SmartContract) QueryAssetByOwner(ctx contractapi.TransactionContextInterface, assetType string, owner string) ([]*Asset, error) {

	queryString := fmt.Sprintf("{\"selector\":{\"objectType\":\"%v\",\"owner\":\"%v\"}}", assetType, owner)

	queryResults, err := s.getQueryResultForQueryString(ctx, queryString)
	if err != nil {
		return nil, err
	}
	return queryResults, nil
}

// QueryAssets uses a query string to perform a query for assets.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the QueryAssetByOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
func (s *SmartContract) QueryAssets(ctx contractapi.TransactionContextInterface, queryString string) ([]*Asset, error) {

	queryResults, err := s.getQueryResultForQueryString(ctx, queryString)
	if err != nil {
		return nil, err
	}
	return queryResults, nil
}

// getQueryResultForQueryString executes the passed in query string.
func (s *SmartContract) getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Asset, error) {

	resultsIterator, err := ctx.GetStub().GetPrivateDataQueryResult(assetCollection, queryString)
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
