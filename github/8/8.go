 package main

 import (
	 "bytes"
	 "crypto/sha256"
	 "encoding/json"
	 "fmt"
	 "time"
	 "log"
	//  "github.com/golang/protobuf/ptypes"
	 "github.com/hyperledger/fabric-chaincode-go/pkg/statebased"
	 "github.com/hyperledger/fabric-chaincode-go/shim"
	 "github.com/hyperledger/fabric-contract-api-go/contractapi"
	//  "github.com/jinzhu/copier"

 )
  
 const (
	 typeAssetForSale     = "S"
	 typeAssetBid         = "B"
	 typeAssetSaleReceipt = "SR"
	 typeAssetBuyReceipt  = "BR"
 )
 
 type SmartContract struct {
	 contractapi.Contract
 }

// QueryResult structure used for handling result of query
// type QueryResult struct {
// 	Record    *Asset
// 	TxId      string    `json:"txId"`
// 	Timestamp time.Time `json:"timestamp"`
// }

type Agreement struct {
	ID      string `json:"assetID"`
	Price   int    `json:"price"`
	TradeID string `json:"tradeID"`
	Quantity int   `json:"quantity"`
}

 //---------------------------------
 // Asset struct and properties must be exported (start with capitals) to work with contract api metadata
 type Asset struct {
	ID					string `json:"assetID"`
	Item   				string `json:"item"`
	Owner   			string `json:"owner"`
	PublicDescription 	string `json:"publicDescription"`
	CreationTimestamp   time.Time `json:"creationtimestamp"`
	OwnerOrg          	string `json:"ownerOrg"`
	Parent				string `json:"parent"`
	BatchID				string `json:"batchID"`
}

type Batch struct{
	BatchID 			string `json:"batchID"`
	Item   				string `json:"item"`
	Subtype1   			string `json:"subtype1"`
	Subtype2   			string `json:"subtype2"`
	AssetType   		string `json:"type"`
	Organic   			string `json:"organic"`
}

type privateAsset struct {
	ObjectType			string `json:"object_type"`
	Quality      		string `json:"quality"`
	Quantity   			int `json:"quantity"`
	Salt 		   		string `json:"salt"`
	Unit  				string `json:"unit"`
}

 type receipt struct {
	SellerAssetID			string `json:"sellerAssetID"`
	BuyerAssetID			string `json:"buyerAssetID"`
	SellerName				string `json:"sellerName"`
	BuyerName				string `json:"buyerName"`
	Quantity				int `json:"quantity"`
	Price     				int `json:"price"`
	Timestamp 				time.Time `json:"timestamp"`
 }
 type CompleteAsset struct{
	Assetq Asset
	Batchq Batch
}

 // Init ;  Method for initializing smart contract
func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface) error {
	return nil
}
 
 // CreateAsset creates a asset and sets it as owned by the client's org
 func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, a0 string, a1 string, a2 string, a3 string, a4 string, a5 string, a6 string, ownerName string, batchID string) error {
 
	 transMap, err := ctx.GetStub().GetTransient()
	 if err != nil {
		 return fmt.Errorf("Error getting transient: " + err.Error())
	 }
	 
	 // Asset properties are private, therefore they get passed in transient field
	 immutablePropertiesJSON, ok := transMap["asset_properties"]
	 if !ok {
		 return fmt.Errorf("asset_properties key not found in the transient map")
	 }
	//  var currentPrivate privateAsset

	//  err = json.Unmarshal([]byte(immutablePropertiesJSON), &currentPrivate)
	//  if err != nil {
	// 	 return fmt.Errorf("failed to unmarshal private asset JSON: %s", err.Error())
	//  }
	 
	//  currentPrivate.ID=a0
	//  immutablePropertiesJSON, _ =json.Marshal(currentPrivate)
	 // Get client org id and verify it matches peer org id.
	 // In this scenario, client is only authorized to read/write private data from its own peer.
	 clientOrgID, err := getClientOrgID(ctx, true)
	 if err != nil {
		 return fmt.Errorf("failed to get verified OrgID: %s", err.Error())
	 }

	 x := time.Now()
 

	 // Create and persist asset
	 asset := Asset{
		//  ID:                assetID,
		//  OwnerOrg:          clientOrgID,
		//  PublicDescription: publicDescription,
		 ID:					a0,
		 Item:   				a1,
		//  Subtype1:   			a2,
		//  Subtype2:   			a3,
		 // Quantity   			string `json:"quantity"`
		 // QuantityUnit  		string `json:"quantityunit"`
		 // Quality      		string `json:"quality"`
		//  AssetType:   			a4,
		//  Organic:   			a5,
		 PublicDescription: 	a6,
		 Owner:   				ownerName,
		 CreationTimestamp:  	x,
		 OwnerOrg:          	clientOrgID,
		 Parent:	"",
		 BatchID: batchID,
	 }
	 
	 batch :=Batch{
		BatchID: batchID,
		Item:	a1,			
		Subtype1:  a2, 			
		Subtype2:   a3,			
		AssetType:   	a4,	
		Organic:   a5,
	 }

	 batchJSON, err := json.Marshal(batch)
	 if err != nil {
		 return fmt.Errorf("failed to create batch in JSON: %s", err.Error())
	 }

	 assetJSON, err := json.Marshal(asset)
	 if err != nil {
		 return fmt.Errorf("failed to create asset JSON: %s", err.Error())
	 }

	 err = ctx.GetStub().PutState(batch.BatchID, batchJSON)
	 if err != nil {
		 return fmt.Errorf("failed to put Batch in public data: %s", err.Error())
	 }
 
	 err = ctx.GetStub().PutState(asset.ID, assetJSON)
	 if err != nil {
		 return fmt.Errorf("failed to put Asset in public data: %s", err.Error())
	 }
 
	 // Set the endorsement policy such that an owner org peer is required to endorse future updates
	 err = setAssetStateBasedEndorsement(ctx, asset.ID, clientOrgID)
	 if err != nil {
		 return fmt.Errorf("failed setting state based endorsement for owner: %s", err.Error())
	 }
 
	 // Persist private immutable asset properties to owner's private data collection
	 collection := buildCollectionName(clientOrgID)
	 err = ctx.GetStub().PutPrivateData(collection, asset.ID, []byte(immutablePropertiesJSON))
	 if err != nil {
		 return fmt.Errorf("failed to put Asset private details: %s", err.Error())
	 }
 
	 return nil
 }
 
 // ChangePublicDescription updates the asset public description. Only the current owner can update the public description
 func (s *SmartContract) ChangePublicDescription(ctx contractapi.TransactionContextInterface, assetID string, newDescription string,ownerName string) error {
	 // Get client org id
	 // No need to check client org id matches peer org id, rely on the asset ownership check instead.
	 clientOrgID, err := getClientOrgID(ctx, false)
	 if err != nil {
		 return fmt.Errorf("failed to get verified OrgID: %s", err.Error())
	 }
	 
	 asset, err := s.ReadAsset(ctx, assetID)
	 if err != nil {
		 return fmt.Errorf("failed to get asset: %s", err.Error())
	 }
	 
	 //owner name check 
	 if ownerName != asset.Owner {
		return fmt.Errorf("a client %s cannot update the description of a asset owned by %s", ownerName, asset.Owner)
	}

	 // auth check to ensure that client's org actually owns the asset
	 if clientOrgID != asset.OwnerOrg {
		 return fmt.Errorf("a client from %s cannot update the description of a asset owned by %s", clientOrgID, asset.OwnerOrg)
	 }
 
	 asset.PublicDescription = newDescription
 
	 updatedAssetJSON, err := json.Marshal(asset)
	 if err != nil {
		 return fmt.Errorf("failed to marshal asset: %s", err.Error())
	 }
 
	 return ctx.GetStub().PutState(assetID, updatedAssetJSON)
 }
 
 // AgreeToSell adds seller's asking price to seller's implicit private data collection
 func (s *SmartContract) AgreeToSell(ctx contractapi.TransactionContextInterface, assetID string,ownerName string) error {
	 // Query asset and verify that this clientOrgId actually owns the asset.
	 asset, err := s.ReadAsset(ctx, assetID)
	 if err != nil {
		 return err
	 }

 	//owner name check 
	 if ownerName != asset.Owner {
		return fmt.Errorf("a client %s cannot sell a asset owned by %s", ownerName, asset.Owner)
	}

	 clientOrgID, err := getClientOrgID(ctx, true)
	 if err != nil {
		 return fmt.Errorf("failed to get verified OrgID: %s", err.Error())
	 }
 
	 if clientOrgID != asset.OwnerOrg {
		 return fmt.Errorf("a client from %s cannot sell a asset owned by %s", clientOrgID, asset.OwnerOrg)
	 }
 
	 return agreeToPrice(ctx, assetID, typeAssetForSale,ownerName)
 }
 
 // AgreeToBuy adds buyer's bid price to buyer's implicit private data collection
 func (s *SmartContract) AgreeToBuy(ctx contractapi.TransactionContextInterface, assetID string, buyerName string) error {
	 return agreeToPrice(ctx, assetID, typeAssetBid, buyerName)
 }


 // agreeToPrice adds a bid or ask price to caller's implicit private data collection
 func agreeToPrice(ctx contractapi.TransactionContextInterface, assetID string, priceType string, requesterName string) error {
 
	 // Get client org id and verify it matches peer org id.
	 // In this scenario, client is only authorized to read/write private data from its own peer.
	 clientOrgID, err := getClientOrgID(ctx, true)
	 if err != nil {
		 return fmt.Errorf("failed to get verified OrgID: %s", err.Error())
	 }
 
	 // price is private, therefore it gets passed in transient field
	 transMap, err := ctx.GetStub().GetTransient()
	 if err != nil {
		 return fmt.Errorf("Error getting transient: " + err.Error())
	 }
 
	 // Price hash will get verfied later, therefore always pass and persist the JSON bytes as-is,
	 // so that there is no risk of nondeterministic marshaling.
	 priceJSON, ok := transMap["asset_price"]
	 if !ok {
		 return fmt.Errorf("asset_price key not found in the transient map")
	 }
 
	 collection := buildCollectionName(clientOrgID)
 
	 // Persist the agreed to price in a collection sub-namespace based on priceType key prefix,
	 // to avoid collisions between private asset properties, sell price, and buy price
	 assetPriceKey, err := ctx.GetStub().CreateCompositeKey(priceType, []string{assetID, requesterName})
	 if err != nil {
		 return fmt.Errorf("failed to create composite key: %s", err.Error())
	 }
 
	 err = ctx.GetStub().PutPrivateData(collection, assetPriceKey, priceJSON)
	 if err != nil {
		 return fmt.Errorf("failed to put asset bid: %s", err.Error())
	 }
 
	 return nil
 }
 
 // VerifyAssetProperties implement function to verify asset properties using the hash
 // Allows a buyer to validate the properties of an asset against the owner's implicit private data collection
 func (s *SmartContract) VerifyAssetProperties(ctx contractapi.TransactionContextInterface, assetID string) (bool, error) {
	 transMap, err := ctx.GetStub().GetTransient()
	 if err != nil {
		 return false, fmt.Errorf("Error getting transient: " + err.Error())
	 }
 
	 // Asset properties are private, therefore they get passed in transient field
	 immutablePropertiesJSON, ok := transMap["asset_properties"]
	 if !ok {
		 return false, fmt.Errorf("asset_properties key not found in the transient map")
	 }
 
	 asset, err := s.ReadAsset(ctx, assetID)
	 if err != nil {
		 return false, fmt.Errorf("failed to get asset: %s", err.Error())
	 }
 
	 collectionOwner := buildCollectionName(asset.OwnerOrg)
	 immutablePropertiesOnChainHash, err := ctx.GetStub().GetPrivateDataHash(collectionOwner, assetID)
	 if err != nil {
		 return false, fmt.Errorf("failed to read asset private properties hash from seller's collection: %s", err.Error())
	 }
	 if immutablePropertiesOnChainHash == nil {
		 return false, fmt.Errorf("asset private properties hash does not exist: %s", assetID)
	 }
 
	 // get sha256 hash of passed immutable properties
	 hash := sha256.New()
	 hash.Write(immutablePropertiesJSON)
	 calculatedPropertiesHash := hash.Sum(nil)
 
	 // verify that the hash of the passed immutable properties matches the on-chain hash
	 if !bytes.Equal(immutablePropertiesOnChainHash, calculatedPropertiesHash) {
		 return false, fmt.Errorf("hash %x for passed immutable properties %s does not match on-chain hash %x", calculatedPropertiesHash, immutablePropertiesJSON, immutablePropertiesOnChainHash)
	 }
 
	 return true, nil
 }
 
 // TransferAsset checks transfer conditions and then transfers asset state to buyer.
 // TransferAsset can only be called by current owner
 func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, assetID string, buyerOrgID string, buyerName string, buyQuantity int, ownerName string,splitAssetID string) (*receipt, error) {
 
	 // Get client org id and verify it matches peer org id.
	 // For a transfer, selling client must get endorsement from their own peer and from buyer peer, therefore don't verify client org id matches peer org id
	 clientOrgID, err := getClientOrgID(ctx, false)
	 if err != nil {
		 return nil,fmt.Errorf("failed to get verified OrgID: %s", err.Error())
	 }
	 asset, err := s.ReadAsset(ctx, assetID)
	 if err != nil {
		 return nil,err
	 }
	 //owner name check 
	 if ownerName != asset.Owner {
		return nil,fmt.Errorf("a client %s cannot transfer a asset owned by %s", ownerName, asset.Owner)
	}

	 transMap, err := ctx.GetStub().GetTransient()
	 if err != nil {
		 return nil,fmt.Errorf("Error getting transient: " + err.Error())
	 }
 
	 immutablePropertiesJSON, ok := transMap["asset_properties"]
	 if !ok {
		 return nil,fmt.Errorf("asset_properties key not found in the transient map")
	 }
 
	 priceJSON, ok := transMap["asset_price"]
	 if !ok {
		 return nil,fmt.Errorf("asset_price key not found in the transient map")
	 }
 
	 var agreement Agreement
	 err = json.Unmarshal([]byte(priceJSON), &agreement)
	 if err != nil {
		 return nil,fmt.Errorf("failed to unmarshal price JSON: %s", err.Error())
	 }

	 if(agreement.Quantity != buyQuantity){
		return nil,fmt.Errorf("Quantity in agreement does not match with selling quantity")
	 }
	 var currentPrivate privateAsset

	 err = json.Unmarshal([]byte(immutablePropertiesJSON), &currentPrivate)
	 if err != nil {
		 return nil,fmt.Errorf("failed to unmarshal private asset JSON: %s", err.Error())
	 }
 
	 err = verifyTransferConditions(ctx, asset, immutablePropertiesJSON, clientOrgID, buyerOrgID, priceJSON,ownerName, buyerName, currentPrivate.Quantity, buyQuantity)
	 if err != nil {
		 return nil,fmt.Errorf("failed transfer verification: %s", err.Error())
	 }

	 fmt.Println("in transfer 1")
	 var invoice *receipt 
	 invoice , err = transferAssetState(ctx, asset, immutablePropertiesJSON, clientOrgID, buyerOrgID, agreement.Price, ownerName, buyerName, buyQuantity, splitAssetID)
	 if err != nil {
		 return nil,fmt.Errorf("failed asset transfer: %s", err.Error())
	 }
 
	 return invoice, nil
 
 }
 
 // verifyTransferConditions checks that client org currently owns asset and that both parties have agreed on price
 func verifyTransferConditions(ctx contractapi.TransactionContextInterface, asset *Asset, immutablePropertiesJSON []byte, clientOrgID string, buyerOrgID string, priceJSON []byte, ownerName string, buyerName string,sellerQuantity int,buyerQuanity int) error {
 
	//CHECK 0: owner name check, seller owns the asset 
	if ownerName != asset.Owner {
		return fmt.Errorf("a client %s cannot transfer a asset owned by %s", ownerName, asset.Owner)
	}
	 // CHECK1: auth check to ensure that client's org actually owns the asset
 
	 if clientOrgID != asset.OwnerOrg {
		 return fmt.Errorf("a client from %s cannot transfer a asset owned by %s", clientOrgID, asset.OwnerOrg)
	 }
 
	 // CHECK2: verify that the hash of the passed immutable properties matches the on-chain hash
 
	 // get on chain hash
	 collectionSeller := buildCollectionName(clientOrgID)
	 immutablePropertiesOnChainHash, err := ctx.GetStub().GetPrivateDataHash(collectionSeller, asset.ID)
	 if err != nil {
		 return fmt.Errorf("failed to read asset private properties hash from seller's collection: %s", err.Error())
	 }
	 if immutablePropertiesOnChainHash == nil {
		 return fmt.Errorf("asset private properties hash does not exist: %s", asset.ID)
	 }
 
	 // get sha256 hash of passed immutable properties
	 hash := sha256.New()
	 hash.Write(immutablePropertiesJSON)
	 calculatedPropertiesHash := hash.Sum(nil)
 
	 // verify that the hash of the passed immutable properties matches the on-chain hash
	 if !bytes.Equal(immutablePropertiesOnChainHash, calculatedPropertiesHash) {
		 return fmt.Errorf("hash %x for passed immutable properties %s does not match on-chain hash %x", calculatedPropertiesHash, immutablePropertiesJSON, immutablePropertiesOnChainHash)
	 }
	 
	 // CHECK3: verify that seller and buyer agreed on the same price
 
	 // get seller (current owner) asking price
	 assetForSaleKey, err := ctx.GetStub().CreateCompositeKey(typeAssetForSale, []string{asset.ID,ownerName})
	 if err != nil {
		 return fmt.Errorf("failed to create composite key: %s", err.Error())
	 }
	 sellerPriceHash, err := ctx.GetStub().GetPrivateDataHash(collectionSeller, assetForSaleKey)
	 if err != nil {
		 return fmt.Errorf("failed to get seller price hash: %s", err.Error())
	 }
	 if sellerPriceHash == nil {
		 return fmt.Errorf("seller price for %s does not exist", asset.ID)
	 }
 
	 // get buyer bid price
	 collectionBuyer := buildCollectionName(buyerOrgID)
	 assetBidKey, err := ctx.GetStub().CreateCompositeKey(typeAssetBid, []string{asset.ID,buyerName})
	 if err != nil {
		 return fmt.Errorf("failed to create composite key: %s", err.Error())
	 }
	 buyerPriceHash, err := ctx.GetStub().GetPrivateDataHash(collectionBuyer, assetBidKey)
	 if err != nil {
		 return fmt.Errorf("failed to get buyer price hash: %s", err.Error())
	 }
	 if buyerPriceHash == nil {
		 return fmt.Errorf("buyer price for %s does not exist", asset.ID)
	 }
 
	 // get sha256 hash of passed price
	 hash = sha256.New()
	 hash.Write(priceJSON)
	 calculatedPriceHash := hash.Sum(nil)
 
	 // verify that the hash of the passed price matches the on-chain seller price hash
	 if !bytes.Equal(calculatedPriceHash, sellerPriceHash) {
		 return fmt.Errorf("hash %x for passed price JSON %s does not match on-chain hash %x, seller hasn't agreed to the passed trade id and price", calculatedPriceHash, priceJSON, sellerPriceHash)
	 }
 
	 // verify that the hash of the passed price matches the on-chain buyer price hash
	 if !bytes.Equal(calculatedPriceHash, buyerPriceHash) {
		 return fmt.Errorf("hash %x for passed price JSON %s does not match on-chain hash %x, buyer hasn't agreed to the passed trade id and price", calculatedPriceHash, priceJSON, buyerPriceHash)
	 }

	 //CHECK4: verify whether buying quanity is available or not with seller
	 if(sellerQuantity < buyerQuanity){
		return fmt.Errorf("Not Enough quantity available for sell")
	 }
 
	 // since all checks passed, return without an error
	 return nil
 }

 // transferAssetState makes the public and private state updates for the transferred asset
 func transferAssetState(ctx contractapi.TransactionContextInterface, asset *Asset, immutablePropertiesJSON []byte, clientOrgID string, buyerOrgID string, price int, ownerName string, buyerName string, quantity int, splitAssetID string) (*receipt , error) {
 
	fmt.Println("in transfer 2")
	 // save the asset with the new owner
	 var assetId = asset.ID
	//  var splitAsset *Asset
	 splitAsset := asset
	 splitAsset.ID = splitAssetID	
	 splitAsset.OwnerOrg = buyerOrgID
	 splitAsset.Owner = buyerName
	 splitAsset.Parent=assetId


	 splitAssetJSON, _ := json.Marshal(splitAsset)
		fmt.Println("in transfer 3")
	 err := ctx.GetStub().PutState(splitAsset.ID, splitAssetJSON)
	 if err != nil {
		 return nil,fmt.Errorf("failed to write asset for buyer: %s", err.Error())
	 }
 
	 // Change the endorsement policy to the new owner
	 err = setAssetStateBasedEndorsement(ctx, splitAsset.ID, buyerOrgID)
	 if err != nil {
		 return nil,fmt.Errorf("failed setting state based endorsement for new owner: %s", err.Error())
	 }
	 fmt.Println("in transfer 4")
	 var newprivateasset privateAsset
	 var updatedprivateasset privateAsset

	 err = json.Unmarshal([]byte(immutablePropertiesJSON), &newprivateasset)
	 if err != nil {
		 return nil,fmt.Errorf("failed to unmarshal private asset JSON: %s", err.Error())
	 }
	 fmt.Println("in transfer 5")
	 updatedprivateasset = newprivateasset
	 updatedprivateasset.Quantity = updatedprivateasset.Quantity - quantity

	//  newprivateasset.ID = splitAssetID
	 newprivateasset.Quantity = quantity

	 newprivateassetJSON, _ := json.Marshal(newprivateasset)
	 updatedprivateassetJSON, _ := json.Marshal(updatedprivateasset)
	 fmt.Println("in transfer 6")
	 collectionBuyer := buildCollectionName(buyerOrgID)
	 err = ctx.GetStub().PutPrivateData(collectionBuyer, splitAsset.ID, newprivateassetJSON)
	 if err != nil {
		 return nil,fmt.Errorf("failed to put Asset private properties for buyer: %s", err.Error())
	 }
	 collectionSeller := buildCollectionName(clientOrgID)
	 err = ctx.GetStub().PutPrivateData(collectionSeller, assetId , updatedprivateassetJSON)
	 if err != nil {
		 return nil,fmt.Errorf("failed to update Asset private properties for seller: %s", err.Error())
	 }
	 // Delete the price records for seller
	 assetPriceKey, err := ctx.GetStub().CreateCompositeKey(typeAssetForSale, []string{assetId, ownerName} )
	 if err != nil {
		 return nil,fmt.Errorf("failed to create composite key for seller: %s", err.Error())
	 }
 
	 err = ctx.GetStub().DelPrivateData(collectionSeller, assetPriceKey)
	 if err != nil {
		 return nil,fmt.Errorf("failed to delete asset price from implicit private data collection for seller: %s", err.Error())
	 }
	 fmt.Println("in transfer 7")
	 // Delete the price records for buyer
	 assetPriceKey, err = ctx.GetStub().CreateCompositeKey(typeAssetBid, []string{assetId, buyerName})
	 if err != nil {
		 return nil,fmt.Errorf("failed to create composite key for buyer: %s", err.Error())
	 }
 
	 err = ctx.GetStub().DelPrivateData(collectionBuyer, assetPriceKey)
	 if err != nil {
		 return nil,fmt.Errorf("failed to delete asset price from implicit private data collection for buyer: %s", err.Error())
	 }
	 fmt.Println("in transfer 8")
	 // Keep record for a 'receipt' in both buyer and seller private data collection to record the sales price and date
	 // Persist the agreed to price in a collection sub-namespace based on receipt key prefix
	 receiptBuyKey, err := ctx.GetStub().CreateCompositeKey(typeAssetBuyReceipt, []string{asset.ID, ctx.GetStub().GetTxID()})
	 if err != nil {
		 return nil,fmt.Errorf("failed to create composite key for receipt: %s", err.Error())
	 }
 
	 timestmp, err := ctx.GetStub().GetTxTimestamp()
	 if err != nil {
		 return nil,fmt.Errorf("failed to create timestamp for receipt: %s", err.Error())
	 }
	 fmt.Println("in transfer 9")
	 assetReceipt := receipt{
		 SellerAssetID	:	assetId,
		 BuyerAssetID	:	splitAssetID,
		 SellerName		:	ownerName,
		 BuyerName		:	buyerName,
		 Quantity		:	quantity,
		 Price			:	price,
		 Timestamp		: 	time.Unix(timestmp.Seconds, int64(timestmp.Nanos)),
	 }

	 ret := &assetReceipt
	 receiptJSON, err := json.Marshal(assetReceipt)
	 if err != nil {
		 return nil,fmt.Errorf("failed to marshal receipt: %s", err.Error())
	 }
	 fmt.Println("in transfer 10")
	 err = ctx.GetStub().PutPrivateData(collectionBuyer, receiptBuyKey, receiptJSON)
	 if err != nil {
		 return nil,fmt.Errorf("failed to put private asset receipt for buyer: %s", err.Error())
	 }
 
	 receiptSaleKey, err := ctx.GetStub().CreateCompositeKey(typeAssetSaleReceipt, []string{ctx.GetStub().GetTxID(), assetId})
	 if err != nil {
		 return nil,fmt.Errorf("failed to create composite key for receipt: %s", err.Error())
	 }
	 fmt.Println("in transfer 11")
	 err = ctx.GetStub().PutPrivateData(collectionSeller, receiptSaleKey, receiptJSON)
	 if err != nil {
		 return nil,fmt.Errorf("failed to put private asset receipt for seller: %s", err.Error())
	 }

	 fmt.Println("in transfer 12")

	 return ret, nil
 }
 
 // getClientOrgID gets the client org ID.
 // The client org ID can optionally be verified against the peer org ID, to ensure that a client from another org doesn't attempt to read or write private data from this peer.
 // The only exception in this scenario is for TransferAsset, since the current owner needs to get an endorsement from the buyer's peer.
 func getClientOrgID(ctx contractapi.TransactionContextInterface, verifyOrg bool) (string, error) {
 
	 clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	 if err != nil {
		 return "", fmt.Errorf("failed getting client's orgID: %s", err.Error())
	 }
 
	 if verifyOrg {
		 err = verifyClientOrgMatchesPeerOrg(clientOrgID)
		 if err != nil {
			 return "", err
		 }
	 }
 
	 return clientOrgID, nil
 }
 
 // verify client org id and matches peer org id.
 func verifyClientOrgMatchesPeerOrg(clientOrgID string) error {
	 peerOrgID, err := shim.GetMSPID()
	 if err != nil {
		 return fmt.Errorf("failed getting peer's orgID: %s", err.Error())
	 }
 
	 if clientOrgID != peerOrgID {
		 return fmt.Errorf("client from org %s is not authorized to read or write private data from an org %s peer", clientOrgID, peerOrgID)
	 }
 
	 return nil
 }
 
 // setAssetStateBasedEndorsement adds an endorsement policy to a asset so that only a peer from an owning org can update or transfer the asset.
 func setAssetStateBasedEndorsement(ctx contractapi.TransactionContextInterface, assetID string, orgToEndorse string) error {
 
	 endorsementPolicy, err := statebased.NewStateEP(nil)
 
	 err = endorsementPolicy.AddOrgs(statebased.RoleTypePeer, orgToEndorse)
	 if err != nil {
		 return fmt.Errorf("failed to add org to endorsement policy: %s", err.Error())
	 }
	 epBytes, err := endorsementPolicy.Policy()
	 if err != nil {
		 return fmt.Errorf("failed to create endorsement policy bytes from org: %s", err.Error())
	 }
	 err = ctx.GetStub().SetStateValidationParameter(assetID, epBytes)
	 if err != nil {
		 return fmt.Errorf("failed to set validation parameter on asset: %s", err.Error())
	 }
 
	 return nil
 }
 
 func buildCollectionName(clientOrgID string) string {
	 return fmt.Sprintf("_implicit_org_%s", clientOrgID)
 }
 
 func getClientImplicitCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {
	 clientOrgID, err := getClientOrgID(ctx, true)
	 if err != nil {
		 return "", fmt.Errorf("failed to get verified OrgID: %s", err.Error())
	 }
 
	 err = verifyClientOrgMatchesPeerOrg(clientOrgID)
	 if err != nil {
		 return "", err
	 }
 
	 return buildCollectionName(clientOrgID), nil
 }
 

// ReadAsset returns the public asset data
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Asset, error) {
	// Since only public data is accessed in this function, no access control is required
	

	assetJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to read asset from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("%s does not exist", assetID)
	}

	var asset *Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

func (s *SmartContract) ReadCompleteAsset(ctx contractapi.TransactionContextInterface, assetID string) (*CompleteAsset,error) {
	// Since only public data is accessed in this function, no access control is required
	fmt.Println("pos 1") 
	var asset *Asset
	asset, err := s.ReadAsset(ctx, assetID)
	if err != nil {
		return nil,fmt.Errorf("failed to get asset: %s", err.Error())
	}
	var batchID = asset.BatchID
	fmt.Println("pos 2") 
	batchJSON, err := ctx.GetStub().GetState(batchID)
	if err != nil {
		return nil,fmt.Errorf("failed to get batch: %s", err.Error())
	}
	if batchJSON == nil {
		return nil, nil
	}
	fmt.Println("pos 3") 

	var batch *Batch
	err = json.Unmarshal(batchJSON, &batch)
	if err != nil {
		return nil,err
	}
	fmt.Println("pos 4") 
	completeAsset := CompleteAsset{
		Assetq : *asset,
		Batchq : *batch,
	}
	fmt.Println("pos 5") 
	ret := &completeAsset
	fmt.Println("pos 6") 

	return ret,nil
}


// GetAssetPrivateProperties returns the immutable asset properties from owner's private data collection
func (s *SmartContract) GetAssetPrivateProperties(ctx contractapi.TransactionContextInterface, assetID string, requesterName string) (string, error) {
	// In this scenario, client is only authorized to read/write private data from its own peer.
	collection, err := getClientImplicitCollectionName(ctx)
	if err != nil {
		return "", err
	}

	asset, err := s.ReadAsset(ctx, assetID)
	 
	 //owner name check 
	 if requesterName != asset.Owner {
		return "", fmt.Errorf("a client %s cannot read the private details of a asset owned by %s", requesterName, asset.Owner)
	}

	immutableProperties, err := ctx.GetStub().GetPrivateData(collection, assetID)
	if err != nil {
		return "", fmt.Errorf("failed to read asset private properties from client org's collection: %v", err)
	}
	if immutableProperties == nil {
		return "", fmt.Errorf("asset private details does not exist in client org's collection: %s", assetID)
	}

	return string(immutableProperties), nil
}

// GetAssetSalesPrice returns the sales price
func (s *SmartContract) GetAssetSalesPrice(ctx contractapi.TransactionContextInterface, assetID string, ownerName string) (string, error) {
	asset, _ := s.ReadAsset(ctx, assetID)
	 
	 //owner name check 
	 if ownerName != asset.Owner {
		return "", fmt.Errorf("a client %s cannot read the private details of a asset owned by %s", ownerName, asset.Owner)
	}
	return getAssetPrice(ctx, assetID, typeAssetForSale, ownerName)
}

// GetAssetBidPrice returns the bid price
func (s *SmartContract) GetAssetBidPrice(ctx contractapi.TransactionContextInterface, assetID string,buyerName string) (string, error) {
	return getAssetPrice(ctx, assetID, typeAssetBid, buyerName)
}

// getAssetPrice gets the bid or ask price from caller's implicit private data collection
func getAssetPrice(ctx contractapi.TransactionContextInterface, assetID string, priceType string, requesterName string) (string, error) {
	
	collection, err := getClientImplicitCollectionName(ctx)
	if err != nil {
		return "", err
	}

	assetPriceKey, err := ctx.GetStub().CreateCompositeKey(priceType, []string{assetID,requesterName})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	price, err := ctx.GetStub().GetPrivateData(collection, assetPriceKey)
	if err != nil {
		return "", fmt.Errorf("failed to read asset price from implicit private data collection: %v", err)
	}
	if price == nil {
		return "", fmt.Errorf("asset price does not exist: %s", assetID)
	}

	return string(price), nil
}

// QueryAssetSaleAgreements returns all of an organization's proposed sales
// changed to query agreement reciept
func (s *SmartContract) QueryAssetSaleAgreements(ctx contractapi.TransactionContextInterface) ([]receipt, error) {
	return queryAgreementsByType(ctx,typeAssetSaleReceipt)
}

// QueryAssetBuyAgreements returns all of an organization's proposed bids
func (s *SmartContract) QueryAssetBuyAgreements(ctx contractapi.TransactionContextInterface) ([]receipt, error) {
	return queryAgreementsByType(ctx, typeAssetBuyReceipt)
}

func queryAgreementsByType(ctx contractapi.TransactionContextInterface, agreeType string) ([]receipt, error) {
	collection, err := getClientImplicitCollectionName(ctx)
	if err != nil {
		return nil, err
	}

	// Query for any object type starting with `agreeType`
	receiptsIterator, err := ctx.GetStub().GetPrivateDataByPartialCompositeKey(collection, agreeType, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to read from private data collection: %v", err)
	}
	defer receiptsIterator.Close()

	var receipts []receipt
	for receiptsIterator.HasNext() {
		resp, err := receiptsIterator.Next()
		if err != nil {
			return nil, err
		}

		var receipt receipt
		err = json.Unmarshal(resp.Value, &receipt)
		if err != nil {
			return nil, err
		}

		receipts = append(receipts, receipt)
	}

	return receipts, nil
}

func (s *SmartContract) QueryAssetHistory(ctx contractapi.TransactionContextInterface, assetID string) ([]*Asset, error) {

	var queryID string 
	var asset *Asset

	var results []*Asset
	for queryID = assetID ; queryID!="" ; queryID=asset.Parent {
		asset, _ = s.ReadAsset(ctx,queryID)
		
		results = append(results, asset)
		
	}
   return results, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		log.Panicf("Error create transfer asset chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting asset chaincode: %v", err)
	}
}