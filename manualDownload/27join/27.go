/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	// "bytes"
	// "encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-chaincode-go/pkg/statebased"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const assetCollection = "publicView"
const transferAgreementObjectType = "transferAgreement"

const (
	typeAsset = "A"
)

// SmartContract of this fabric sample
type SmartContract struct {
	contractapi.Contract
}

type Asset struct {
	Type     string `json:"objectType"`
	ID       string `json:"assetID"`
	Owner    string `json:"owner"`
	Lender   string `json:"lender"`
	Borrower string `json:"borrower"`

	Amount    int `json:"amount"`
	StartDate int `json:"startDate"`
	EndDate   int `json:"endDate"`

	BorrowerAddress string   `json:"senderAddress"`
	InvestorAddress string   `json:"investorAddress"`
	OwnerAddress    string   `json:"receiverAddress"`
	PaymentHashes   []string `json:"paymentHashes"`
}

type AssetPrivate struct {
	SecretMessage string `json:"secretMessage"`
}

type TransferAgreement struct {
	ID      string `json:"assetID"`
	BuyerID string `json:"buyerID"`
}

func (s *SmartContract) IssueAsset(ctx contractapi.TransactionContextInterface, assetID string, amount int, start int, end int) error {

	clientID, orgID, err := getClientOrgID(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to get verified OrgID: %v", err)
	}

	resJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if resJSON != nil {
		return fmt.Errorf("asset with id: %s already exist", assetID)
	}

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	transientData, ok := transientMap[assetID]
	if !ok {
		//log error to stdout
		return fmt.Errorf("data for asset %v not found in the transient map", assetID)
	}

	var transientInput AssetPrivate
	err = json.Unmarshal(transientData, &transientInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(transientInput.SecretMessage) == 0 {
		return fmt.Errorf("message field must be a non-empty string")
	}

	asset := Asset{
		Type:      "loan-asset",
		ID:        assetID,
		Owner:     clientID,
		Amount:    amount,
		StartDate: start,
		EndDate:   end,
	}

	if len(asset.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}
	if asset.StartDate <= 0 {
		return fmt.Errorf("start date must be a positive integer")
	}
	if asset.EndDate <= 0 {
		return fmt.Errorf("end date must be a positive integer")
	}
	if asset.Amount <= 0 {
		return fmt.Errorf("amount field must be a positive integer")
	}

	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to create asset JSON: %v", err)
	}

	transientBytes, err := json.Marshal(transientInput)
	if err != nil {
		return fmt.Errorf("failed to create asset JSON: %v", err)
	}

	compositeKey, err := ctx.GetStub().CreateCompositeKey(typeAsset, []string{assetID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	log.Printf("IssueAsset Put: collection %v, ID %v, owner %v", "general", assetID, orgID)
	err = ctx.GetStub().PutState(compositeKey, assetBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset in public data: %v", err)
	}

	// Set the endorsement policy such that an owner org peer is required to endorse future updates
	err = setAssetStateBasedEndorsement(ctx, compositeKey, orgID)
	if err != nil {
		return fmt.Errorf("failed setting state based endorsement for owner: %v", err)
	}

	collectionPriv, _ := getCollectionName(ctx)
	log.Printf("CreateAsset Put: collection %v, ID %v, owner %v", collectionPriv, assetID, orgID)
	err = ctx.GetStub().PutPrivateData(collectionPriv, assetID, transientBytes)
	if err != nil {
		return fmt.Errorf("failed to put Asset private details: %v", err)
	}

	return nil
}

// AgreeToTransfer is used by the potential buyer of the asset to agree to the
// asset value. The agreed to appraisal value is stored in the buying orgs
// org specifc collection, while the the buyer client ID is stored in the asset collection
// using a composite key
func (s *SmartContract) AgreeToTransfer(ctx contractapi.TransactionContextInterface) error {

	clientID, orgID, err := getClientOrgID(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to get verified OrgID: %v", err)
	}
	// Value is private, therefore it gets passed in transient field
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Persist the JSON bytes as-is so that there is no risk of nondeterministic marshaling.
	valueJSONasBytes, ok := transientMap["message"]
	if !ok {
		return fmt.Errorf("message key not found in the transient map")
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
	if valueJSON.AppraisedValue <= 0 {
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
	err = ctx.GetStub().PutPrivateData(assetCollection, transferAgreeKey, []byte(clientID))
	if err != nil {
		return fmt.Errorf("failed to put asset bid: %v", err)
	}

	return nil
}

// // TransferAsset transfers the asset to the new owner by setting a new owner ID
// func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface) error {

// 	transientMap, err := ctx.GetStub().GetTransient()
// 	if err != nil {
// 		return fmt.Errorf("error getting transient %v", err)
// 	}

// 	// Asset properties are private, therefore they get passed in transient field
// 	transientTransferJSON, ok := transientMap["asset_owner"]
// 	if !ok {
// 		return fmt.Errorf("asset owner not found in the transient map")
// 	}

// 	type assetTransferTransientInput struct {
// 		ID       string `json:"assetID"`
// 		BuyerMSP string `json:"buyerMSP"`
// 	}

// 	var assetTransferInput assetTransferTransientInput
// 	err = json.Unmarshal(transientTransferJSON, &assetTransferInput)
// 	if err != nil {
// 		return fmt.Errorf("failed to unmarshal JSON: %v", err)
// 	}

// 	if len(assetTransferInput.ID) == 0 {
// 		return fmt.Errorf("assetID field must be a non-empty string")
// 	}
// 	if len(assetTransferInput.BuyerMSP) == 0 {
// 		return fmt.Errorf("buyerMSP field must be a non-empty string")
// 	}
// 	log.Printf("TransferAsset: verify asset exists ID %v", assetTransferInput.ID)
// 	// Read asset from the private data collection
// 	asset, err := s.ReadAsset(ctx, assetTransferInput.ID)
// 	if err != nil {
// 		return fmt.Errorf("error reading asset: %v", err)
// 	}
// 	if asset == nil {
// 		return fmt.Errorf("%v does not exist", assetTransferInput.ID)
// 	}
// 	// Verify that the client is submitting request to peer in their organization
// 	err = verifyClientOrgMatchesPeerOrg(ctx)
// 	if err != nil {
// 		return fmt.Errorf("TransferAsset cannot be performed: Error %v", err)
// 	}

// 	// Verify transfer details and transfer owner
// 	err = s.verifyAgreement(ctx, assetTransferInput.ID, asset.Owner, assetTransferInput.BuyerMSP)
// 	if err != nil {
// 		return fmt.Errorf("failed transfer verification: %v", err)
// 	}

// 	transferAgreement, err := s.ReadTransferAgreement(ctx, assetTransferInput.ID)
// 	if err != nil {
// 		return fmt.Errorf("failed ReadTransferAgreement to find buyerID: %v", err)
// 	}
// 	if transferAgreement.BuyerID == "" {
// 		return fmt.Errorf("BuyerID not found in TransferAgreement for %v", assetTransferInput.ID)
// 	}

// 	// Transfer asset in private data collection to new owner
// 	asset.Owner = transferAgreement.BuyerID

// 	assetJSONasBytes, err := json.Marshal(asset)
// 	if err != nil {
// 		return fmt.Errorf("failed marshalling asset %v: %v", assetTransferInput.ID, err)
// 	}

// 	log.Printf("TransferAsset Put: collection %v, ID %v", assetCollection, assetTransferInput.ID)
// 	err = ctx.GetStub().PutPrivateData(assetCollection, assetTransferInput.ID, assetJSONasBytes) //rewrite the asset
// 	if err != nil {
// 		return err
// 	}

// 	// Get collection name for this organization
// 	ownersCollection, err := getCollectionName(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
// 	}

// 	// Delete the asset appraised value from this organization's private data collection
// 	err = ctx.GetStub().DelPrivateData(ownersCollection, assetTransferInput.ID)
// 	if err != nil {
// 		return err
// 	}

// 	// Delete the transfer agreement from the asset collection
// 	transferAgreeKey, err := ctx.GetStub().CreateCompositeKey(transferAgreementObjectType, []string{assetTransferInput.ID})
// 	if err != nil {
// 		return fmt.Errorf("failed to create composite key: %v", err)
// 	}

// 	err = ctx.GetStub().DelPrivateData(assetCollection, transferAgreeKey)
// 	if err != nil {
// 		return err
// 	}

// 	return nil

// }

// // verifyAgreement is an internal helper function used by TransferAsset to verify
// // that the transfer is being initiated by the owner and that the buyer has agreed
// // to the same appraisal value as the owner
// func (s *SmartContract) verifyAgreement(ctx contractapi.TransactionContextInterface, assetID string, owner string, buyerMSP string) error {

// 	// Check 1: verify that the transfer is being initiatied by the owner

// 	// Get ID of submitting client identity
// 	clientID, err := submittingClientIdentity(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	if clientID != owner {
// 		return fmt.Errorf("error: submitting client identity does not own asset")
// 	}

// 	// Check 2: verify that the buyer has agreed to the appraised value

// 	// Get collection names
// 	collectionOwner, err := getCollectionName(ctx) // get owner collection from caller identity
// 	if err != nil {
// 		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
// 	}

// 	collectionBuyer := buyerMSP + "PrivateCollection" // get buyers collection

// 	// Get hash of owners agreed to value
// 	ownerAppraisedValueHash, err := ctx.GetStub().GetPrivateDataHash(collectionOwner, assetID)
// 	if err != nil {
// 		return fmt.Errorf("failed to get hash of appraised value from owners collection %v: %v", collectionOwner, err)
// 	}
// 	if ownerAppraisedValueHash == nil {
// 		return fmt.Errorf("hash of appraised value for %v does not exist in collection %v", assetID, collectionOwner)
// 	}

// 	// Get hash of buyers agreed to value
// 	buyerAppraisedValueHash, err := ctx.GetStub().GetPrivateDataHash(collectionBuyer, assetID)
// 	if err != nil {
// 		return fmt.Errorf("failed to get hash of appraised value from buyer collection %v: %v", collectionBuyer, err)
// 	}
// 	if buyerAppraisedValueHash == nil {
// 		return fmt.Errorf("hash of appraised value for %v does not exist in collection %v. AgreeToTransfer must be called by the buyer first", assetID, collectionBuyer)
// 	}

// 	// Verify that the two hashes match
// 	if !bytes.Equal(ownerAppraisedValueHash, buyerAppraisedValueHash) {
// 		return fmt.Errorf("hash for appraised value for owner %x does not value for seller %x", ownerAppraisedValueHash, buyerAppraisedValueHash)
// 	}

// 	return nil
// }

// // DeleteAsset can be used by the owner of the asset to delete the asset
// func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface) error {

// 	transientMap, err := ctx.GetStub().GetTransient()
// 	if err != nil {
// 		return fmt.Errorf("Error getting transient: %v", err)
// 	}

// 	// Asset properties are private, therefore they get passed in transient field
// 	transientDeleteJSON, ok := transientMap["asset_delete"]
// 	if !ok {
// 		return fmt.Errorf("asset to delete not found in the transient map")
// 	}

// 	type assetDelete struct {
// 		ID string `json:"assetID"`
// 	}

// 	var assetDeleteInput assetDelete
// 	err = json.Unmarshal(transientDeleteJSON, &assetDeleteInput)
// 	if err != nil {
// 		return fmt.Errorf("failed to unmarshal JSON: %v", err)
// 	}

// 	if len(assetDeleteInput.ID) == 0 {
// 		return fmt.Errorf("assetID field must be a non-empty string")
// 	}

// 	// Verify that the client is submitting request to peer in their organization
// 	err = verifyClientOrgMatchesPeerOrg(ctx)
// 	if err != nil {
// 		return fmt.Errorf("DeleteAsset cannot be performed: Error %v", err)
// 	}

// 	log.Printf("Deleting Asset: %v", assetDeleteInput.ID)
// 	valAsbytes, err := ctx.GetStub().GetPrivateData(assetCollection, assetDeleteInput.ID) //get the asset from chaincode state
// 	if err != nil {
// 		return fmt.Errorf("failed to read asset: %v", err)
// 	}
// 	if valAsbytes == nil {
// 		return fmt.Errorf("asset not found: %v", assetDeleteInput.ID)
// 	}

// 	ownerCollection, err := getCollectionName(ctx) // Get owners collection
// 	if err != nil {
// 		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
// 	}

// 	//check the asset is in the caller org's private collection
// 	valAsbytes, err = ctx.GetStub().GetPrivateData(ownerCollection, assetDeleteInput.ID)
// 	if err != nil {
// 		return fmt.Errorf("failed to read asset from owner's Collection: %v", err)
// 	}
// 	if valAsbytes == nil {
// 		return fmt.Errorf("asset not found in owner's private Collection %v: %v", ownerCollection, assetDeleteInput.ID)
// 	}

// 	// delete the asset from state
// 	err = ctx.GetStub().DelPrivateData(assetCollection, assetDeleteInput.ID)
// 	if err != nil {
// 		return fmt.Errorf("failed to delete state: %v", err)
// 	}

// 	// Finally, delete private details of asset
// 	err = ctx.GetStub().DelPrivateData(ownerCollection, assetDeleteInput.ID)
// 	if err != nil {
// 		return err
// 	}

// 	return nil

// }

// // DeleteTranferAgreement can be used by the buyer to withdraw a proposal from
// // the asset collection and from his own collection.
// func (s *SmartContract) DeleteTranferAgreement(ctx contractapi.TransactionContextInterface) error {

// 	transientMap, err := ctx.GetStub().GetTransient()
// 	if err != nil {
// 		return fmt.Errorf("error getting transient: %v", err)
// 	}

// 	// Asset properties are private, therefore they get passed in transient field
// 	transientDeleteJSON, ok := transientMap["agreement_delete"]
// 	if !ok {
// 		return fmt.Errorf("asset to delete not found in the transient map")
// 	}

// 	type assetDelete struct {
// 		ID string `json:"assetID"`
// 	}

// 	var assetDeleteInput assetDelete
// 	err = json.Unmarshal(transientDeleteJSON, &assetDeleteInput)
// 	if err != nil {
// 		return fmt.Errorf("failed to unmarshal JSON: %v", err)
// 	}

// 	if len(assetDeleteInput.ID) == 0 {
// 		return fmt.Errorf("transient input ID field must be a non-empty string")
// 	}

// 	// Verify that the client is submitting request to peer in their organization
// 	err = verifyClientOrgMatchesPeerOrg(ctx)
// 	if err != nil {
// 		return fmt.Errorf("DeleteTranferAgreement cannot be performed: Error %v", err)
// 	}
// 	// Delete private details of agreement
// 	orgCollection, err := getCollectionName(ctx) // Get proposers collection.
// 	if err != nil {
// 		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
// 	}
// 	tranferAgreeKey, err := ctx.GetStub().CreateCompositeKey(transferAgreementObjectType, []string{assetDeleteInput.
// 		ID}) // Create composite key
// 	if err != nil {
// 		return fmt.Errorf("failed to create composite key: %v", err)
// 	}

// 	valAsbytes, err := ctx.GetStub().GetPrivateData(assetCollection, tranferAgreeKey) //get the transfer_agreement
// 	if err != nil {
// 		return fmt.Errorf("failed to read transfer_agreement: %v", err)
// 	}
// 	if valAsbytes == nil {
// 		return fmt.Errorf("asset's transfer_agreement does not exist: %v", assetDeleteInput.ID)
// 	}

// 	log.Printf("Deleting TranferAgreement: %v", assetDeleteInput.ID)
// 	err = ctx.GetStub().DelPrivateData(orgCollection, assetDeleteInput.ID) // Delete the asset
// 	if err != nil {
// 		return err
// 	}

// 	// Delete transfer agreement record
// 	err = ctx.GetStub().DelPrivateData(assetCollection, tranferAgreeKey) // remove agreement from state
// 	if err != nil {
// 		return err
// 	}

// 	return nil

// }

// getCollectionName is an internal helper function to get collection of submitting client identity.
func getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get the MSP ID of submitting client identity
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get verified MSPID: %v", err)
	}

	// Create the collection name
	orgCollection := clientMSPID + "_view"

	return orgCollection, nil
}

func getClientOrgID(ctx contractapi.TransactionContextInterface, verifyOrg bool) (string, string, error) {
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", "", fmt.Errorf("failed getting client's orgID: %v", err)
	}

	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", "", fmt.Errorf("failed getting client's orgID: %v", err)
	}

	if verifyOrg {
		err = verifyClientOrgMatchesPeerOrg(clientOrgID)
		if err != nil {
			return "", "", err
		}
	}

	return clientID, clientOrgID, nil
}

// verifyClientOrgMatchesPeerOrg checks the client org id matches the peer org id.
func verifyClientOrgMatchesPeerOrg(clientOrgID string) error {
	peerOrgID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting peer's orgID: %v", err)
	}

	if clientOrgID != peerOrgID {
		return fmt.Errorf("client from org %s is not authorized to read or write private data from an org %s peer",
			clientOrgID,
			peerOrgID,
		)
	}

	return nil
}

// setAssetStateBasedEndorsement adds an endorsement policy to a asset so that only a peer from an owning org
// can update or transfer the asset.
func setAssetStateBasedEndorsement(ctx contractapi.TransactionContextInterface, assetID string, orgToEndorse string) error {
	endorsementPolicy, err := statebased.NewStateEP(nil)
	if err != nil {
		return err
	}
	err = endorsementPolicy.AddOrgs(statebased.RoleTypePeer, orgToEndorse)
	if err != nil {
		return fmt.Errorf("failed to add org to endorsement policy: %v", err)
	}
	policy, err := endorsementPolicy.Policy()
	if err != nil {
		return fmt.Errorf("failed to create endorsement policy bytes from org: %v", err)
	}
	err = ctx.GetStub().SetStateValidationParameter(assetID, policy)
	if err != nil {
		return fmt.Errorf("failed to set validation parameter on asset: %v", err)
	}

	return nil
}

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
