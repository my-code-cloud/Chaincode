package chaincode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/pkg/statebased"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

/*
1. QueryInterestTokenFromTradeId(tradeId, string)
2.
*/

type SmartContract struct {
	contractapi.Contract
}

// DevicePublicDetails ...
type DevicePublicDetails struct {
	Owner       string `json:"owner"`
	ID          string `json:"deviceId"` // uniqueId = DEVICE_{ID} on collection_Marketplace
	Data        string `json:"dataDescription"`
	Description string `json:"description"`
	OnSale      bool   `json:"onSale"`
}

type DevicePrivateDetails struct { // Device Meta data
	ID     string `json:"deviceId"` // uniqueId on collection_devicePrivatedetails
	Secret string `json:"deviceSecret"`
}

// Agreement
type TradeAgreement struct { // the hash of respective trade agreements should match
	ID         string    `json:"tradeId"`  // unique key on collection_TradeAgreement
	DeviceId   string    `json:"deviceId"` // search all trades for this device
	Price      int       `json:"tradePrice"`
	RevokeTime time.Time `json:"revoke_time"`
}

type InterestToken struct { // token of interest passed by the bidder
	ID                       string `json:"tradeId"`  // search all biddings for this device
	DeviceId                 string `json:"deviceId"` // unique key as TRADE_{deviceID} on Collection_Marketplace
	BidderID                 string `json:"bidderId"`
	SellerId                 string `json:"seller_id"`
	TradeAgreementCollection string `json:"dealsCollection"` // required to generate private-data hash for the bidder's agreement collection:tradeID
}

// to be returned via event
type Receipt struct {
	TimeStamp     time.Time `json:"time_stamp"`
	Seller        string    `json:"seller"`
	Buyer         string    `json:"buyer"`
	TransactionId string    `json:"trade_confirmation_transaction_id"`
	TradeId       string    `json:"trade_id"`
	Type          string    `json:"type"`
	RevokeTime    time.Time `json:"revoke_time"`
}

// on the blockchain
type TradeConfirmation struct {
	Type                string    `json:"type"`
	SellerAgreementHash string    `json:"seller_agreement_hash"`
	BuyerAgreementHash  string    `json:"buyer_agreement_hash"`
	RevokeTime          time.Time `json:"revoke_time"`
}

// temp object to be returned from verifyTradeAgreements
type AgreementDetails struct {
	TradeId             string
	BuyerID             string
	RevokeTime          time.Time
	SellerAgreementHash string
	BuyerAgreementHash  string
}

type DeviceDataObject struct {
	Timestamp     time.Time `json:"timestamp"`
	Data          string    `json:"dataJSON"` // JSON Data -> string
	TransactionId string    `json:"transactionId"`
}

// Data
type DeviceData struct {
	// DeviceId
	// Data - JSON object
	ID   string             `json:"deviceId"`
	Data []DeviceDataObject `json:"dataJSON"` // JSON Data -> string
}

// ACL
type ACLObject struct {
	BuyerId    string    `json:"buyerId"`
	TradeID    string    `json:"tradeId"`
	RevokeTime time.Time `json:"revoke_time"`
}

type DeviceACL struct {
	// Device ID
	// TradeID
	ID   string      `json:"deviceId"`
	List []ACLObject `json:"acl"`
}

func (s *SmartContract) GetAndVerifyTradeAgreements(ctx contractapi.TransactionContextInterface, tradeId string) (AgreementDetails, error) {
	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
	}

	bidderIntrestToken, err := s.QueryInterestTokenFromTradeId(ctx, tradeId)
	if err != nil {
		return AgreementDetails{}, fmt.Errorf("Cannot get BidderInterestToken, %v", err.Error())
	}
	bidderId := bidderIntrestToken[0].BidderID
	deviceId := bidderIntrestToken[0].DeviceId
	bidderTradeAgreementCollection := bidderIntrestToken[0].TradeAgreementCollection
	fmt.Println(bidderIntrestToken[0])

	err = verifyClientOrgMatchesOwner(ctx, deviceId)
	if err != nil {
	}

	ownerTradeAgreementCollection, err := getTradeAgreementCollection(ctx)
	if err != nil {
	}

	sellerAgreementHash, err := getAgreementHash(ctx, ownerTradeAgreementCollection, tradeId)
	buyerAgreementHash, err := getAgreementHash(ctx, bidderTradeAgreementCollection, tradeId)
	if !bytes.Equal(sellerAgreementHash, buyerAgreementHash) {
		return AgreementDetails{}, fmt.Errorf("Agreements do not match")
	}

	agreementDetails := AgreementDetails{
		TradeId:             tradeId,
		BuyerID:             bidderId,
		SellerAgreementHash: string(sellerAgreementHash),
		BuyerAgreementHash:  string(buyerAgreementHash),
	}
	return agreementDetails, nil
}

func (s *SmartContract) AddToACL(ctx contractapi.TransactionContextInterface, bidderId string, tradeId string, deviceId string) error {
	revokeTime, err := s.GetRevokeTime(ctx, tradeId)
	newACLObject := ACLObject{
		TradeID:    tradeId,
		BuyerId:    bidderId,
		RevokeTime: revokeTime,
	}
	fmt.Println("newAClObject\n")
	fmt.Println(newACLObject)

	aclCollection, err := getACLCollection(ctx)
	fmt.Println(aclCollection)
	fmt.Printf("%s %s %s \n\n", bidderId, tradeId, deviceId)
	aclAsBytes, err := ctx.GetStub().GetPrivateData(aclCollection, deviceId)
	if err != nil {
		fmt.Println(err)
	}

	var acl DeviceACL
	err = json.Unmarshal(aclAsBytes, &acl)
	fmt.Println("acl\n")
	fmt.Println(acl)

	acl.ID = deviceId
	acl.List = append(acl.List, newACLObject)
	fmt.Println("acl\n")
	fmt.Println(acl)

	aclAsBytes, err = json.Marshal(acl)
	if err != nil {
		return fmt.Errorf("Marshalling Error %v", err.Error())
	}

	err = ctx.GetStub().PutPrivateData(aclCollection, deviceId, aclAsBytes)
	if err != nil {
		fmt.Println("Error while putting private data")
		return fmt.Errorf("Error Putting in ACL %v", err.Error())
	}
	return nil
}

func (s *SmartContract) GenerateReceipt(ctx contractapi.TransactionContextInterface, ad AgreementDetails) error {

	revokeTime, err := s.GetRevokeTime(ctx, ad.TradeId)
	tradeConfirmation := TradeConfirmation{
		Type:                "TRADE_CONFIRMATION",
		SellerAgreementHash: ad.SellerAgreementHash,
		BuyerAgreementHash:  ad.BuyerAgreementHash,
		RevokeTime:          revokeTime,
	}

	tradeConfirmationAsBytes, err := ctx.GetStub().GetState(ad.TradeId)
	if err != nil {
	}
	if tradeConfirmationAsBytes != nil {
		return fmt.Errorf("TradeId Already Exists on Blockchain")
	}

	tradeConfirmationAsBytes, err = json.Marshal(tradeConfirmation)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(ad.TradeId, tradeConfirmationAsBytes)
	// check transactionid in database
	transactionId := ctx.GetStub().GetTxID()
	sellerId, err := shim.GetMSPID()
	tradeEventPayload := Receipt{
		Type:          "Trade Receipt",
		Buyer:         ad.BuyerID,
		Seller:        sellerId,
		TransactionId: transactionId,
		TimeStamp:     time.Now(),
		TradeId:       ad.TradeId,
		RevokeTime:    revokeTime,
	}
	tradeEventPayloadAsBytes, err := json.Marshal(tradeEventPayload)
	fmt.Println("INSIDE RECEIPT CONTRACT")
	return ctx.GetStub().SetEvent("RECEIPT-EVENT", tradeEventPayloadAsBytes)
}

func getAgreementHash(ctx contractapi.TransactionContextInterface, collection string, tradeId string) ([]byte, error) {
	agreementHashAsBytes, err := ctx.GetStub().GetPrivateDataHash(collection, tradeId)
	if err != nil {
		return nil, err
	}
	return agreementHashAsBytes, nil
}

func verifyClientOrgMatchesOwner(ctx contractapi.TransactionContextInterface, deviceId string) error {
	marketplaceCollection, err := getMarketplaceCollection()
	if err != nil {
	}
	deviceKey := generateKeyForDevice(deviceId)
	device := readDevicePublicDetails(ctx, marketplaceCollection, deviceKey) // returns marshalled data

	clientOrg, err := ctx.GetClientIdentity().GetMSPID()
	if device.Owner != clientOrg {
		return fmt.Errorf("clientOrg %v doesnot match Owner %v ", clientOrg, device.Owner)
	}
	return nil
}

func readDevicePublicDetails(ctx contractapi.TransactionContextInterface, collection string, key string) DevicePublicDetails {
	deviceAsBytes, err := ctx.GetStub().GetPrivateData(collection, key)
	if err != nil {
	}
	var device DevicePublicDetails
	err = json.Unmarshal(deviceAsBytes, &device)
	if err != nil {
	}
	return device
}

// todo
func (s *SmartContract) revokeDataDistribution(tradeId string) error {
	return nil
}

// TODO
// 1. chain of custody
// 2. what if one of the orgs changes contract details later
//      - can we prevent any updates on existing trade contract

// ---------------------------keys for collection -------------------------

func generateKeyForInterestToken(tradeId string) string {
	return "TRADE_" + tradeId
}

func generateKeyForDevice(deviceId string) string {
	return "DEVICE_" + deviceId
}

func generateKeyForDevicedata(deviceID string) string {
	return "DATA_" + deviceID
}

// ----------------------Collection names---------------------------

func getMarketplaceCollection() (string, error) {
	return "collection_Marketplace", nil
}

//func getDealsCollection() (string, error) {
//    msp, err := shim.GetMSPID()
//    if err != nil {return "", err}
//
//    return msp + "_dealsCollection", nil
//}

func getTradeAgreementCollection(ctx contractapi.TransactionContextInterface) (string, error) {
	msp, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", err
	}

	return msp + "_tradeAgreementCollection", nil
}

func getPrivateDetailsCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {
	msp, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", err
	}

	return msp + "_privateDetailsCollection", nil
}
func getACLCollection(ctx contractapi.TransactionContextInterface) (string, error) {
	msp, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", err
	}

	return msp + "_aclCollection", nil
}
func getSharingCollection(seller string, buyer string) (string, error) {
	var temparr []string
	temparr = append(temparr, seller)
	temparr = append(temparr, buyer)

	sort.Strings(temparr)
	return temparr[0] + "_" + temparr[1] + "_shareCollection", nil
}

// ------------------------------------------------------------------------

func verifyClientOrgMatchesPeerOrg(ctx contractapi.TransactionContextInterface) error {
	clientMSP, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
	}

	peerMSP, err := shim.GetMSPID()
	if err != nil {
	}

	if clientMSP != peerMSP {
		return fmt.Errorf("client MSP %v does not match PeerMSP %v", clientMSP, peerMSP)
	}
	return nil
}

func setDeviceStateBasedEndorsement(ctx contractapi.TransactionContextInterface, deviceKey string, orgId string, collection string) error {
	// create a new state based policy for key = deviceId
	ep, err := statebased.NewStateEP(nil)
	if err != nil {
	}

	// issue roles, here the owner org for a device
	err = ep.AddOrgs(statebased.RoleTypePeer, orgId)
	if err != nil {
	}

	policy, err := ep.Policy()
	if err != nil {
	}

	err = ctx.GetStub().SetPrivateDataValidationParameter(collection, deviceKey, policy)
	return nil
}

const assetCollection = "assetCollection"

//import "github.com/hyperledger/fabric-contract-api-go/contractapi"
//

//collection = MArketplace

//// queryOnSaleDataMarketplace -> list of DevicePublicDetails onSale
func (s *SmartContract) QueryOnSaleDataMarketplace(ctx contractapi.TransactionContextInterface) ([]*DevicePublicDetails, error) {
	marketplaceCollection, _ := getMarketplaceCollection()

	queryString := fmt.Sprintf(`{"selector":{"onSale":true, "_id":{"$regex":"DEVICE*"}}}`)
	resultsIterator, err := getQueryResultForQueryString(ctx, marketplaceCollection, queryString)
	if err != nil {
		return nil, err
	}
	return constructPublicDevicesQueryResponseFromIterator(resultsIterator)
}

func (t *SmartContract) QueryDevices(ctx contractapi.TransactionContextInterface, queryString string) ([]*DevicePublicDetails, error) {
	marketplaceCollection, _ := getMarketplaceCollection()
	resultsIterator, err := getQueryResultForQueryString(ctx, marketplaceCollection, queryString)

	if err != nil {
		return nil, err
	}
	return constructPublicDevicesQueryResponseFromIterator(resultsIterator)
}

func (s *SmartContract) QuerySharedDevices(ctx contractapi.TransactionContextInterface, ownerOrg string) ([]string, error) {
	selfMsp, mspErr := ctx.GetClientIdentity().GetMSPID()
	if mspErr != nil {

		return nil, mspErr
	}
	sharedDevicesDetailsCollection, _ := getSharingCollection(ownerOrg, selfMsp)

	queryString := fmt.Sprintf(`{"selector":{"_id":{"$regex":"DATA*"}}}`)

	resultsIterator, err := getQueryResultForQueryString(ctx, sharedDevicesDetailsCollection, queryString)
	if err != nil {
		return nil, err
	}

	fullData, err := constructDevicesDataQueryResponseFromIterator(resultsIterator)

	var devicesList []string

	for d := range fullData {
		devicesList = append(devicesList, fullData[d].ID)
	}

	return devicesList, nil
}

//collection = MArketplace
// key
// queryBidders -> InterestToken for a tradeId
func (s *SmartContract) QueryInterestTokenFromTradeId(ctx contractapi.TransactionContextInterface, tradeId string) ([]*InterestToken, error) {
	marketplaceCollection, _ := getMarketplaceCollection()

	queryString := fmt.Sprintf(`{"selector":{"tradeId":"%s", "_id":{"$regex":"TRADE*"}}}`, tradeId)
	resultsIterator, err := getQueryResultForQueryString(ctx, marketplaceCollection, queryString)
	if err != nil {
		return nil, err
	}
	return constructInterestTokensQueryResponseFromIterator(resultsIterator)
}

// queryBidders -> list of all InterestTokens for a tradeId

func (s *SmartContract) QueryInterestTokensForDevice(ctx contractapi.TransactionContextInterface, deviceId string) ([]*InterestToken, error) {
	marketplaceCollection, _ := getMarketplaceCollection()

	queryString := fmt.Sprintf(`{"selector":{"deviceId":"%s", "_id":{"$regex":"TRADE*"}}}`, deviceId)

	resultsIterator, err := getQueryResultForQueryString(ctx, marketplaceCollection, queryString)
	if err != nil {
		return nil, fmt.Errorf("No Interest Tokens for such device")
	}
	return constructInterestTokensQueryResponseFromIterator(resultsIterator)
}

func (s *SmartContract) QueryTradeAgreementsForDevice(ctx contractapi.TransactionContextInterface, deviceId string) ([]*TradeAgreement, error) {
	tradeAgreementCollection, _ := getTradeAgreementCollection(ctx)

	queryString := fmt.Sprintf(`{"selector":{"deviceId":"%s"}}`, deviceId)

	resultsIterator, err := getQueryResultForQueryString(ctx, tradeAgreementCollection, queryString)
	if err != nil {
		return nil, err
	}
	return constructTradeAgreementsQueryResponseFromIterator(resultsIterator)
}

func (s *SmartContract) QueryACLForDevice(ctx contractapi.TransactionContextInterface, deviceId string) (*DeviceACL, error) {
	aclCollection, _ := getACLCollection(ctx)

	aclAsBytes, err := ctx.GetStub().GetPrivateData(aclCollection, deviceId)
	if err != nil {
		fmt.Println("No ACL for the Device")
		fmt.Println(err.Error())
		return nil, fmt.Errorf("No ACL for the device", err.Error())
	}
	if aclAsBytes == nil {
		fmt.Println("Empty ACL")
		return nil, fmt.Errorf("Empty ACL")
	}
	var acl DeviceACL
	err = json.Unmarshal(aclAsBytes, &acl)
	if err != nil {
	}
	fmt.Println(acl)
	return &acl, nil
}

func (t *SmartContract) QueryInterestTokens(ctx contractapi.TransactionContextInterface, queryString string) ([]*InterestToken, error) {
	marketplaceCollection, _ := getMarketplaceCollection()
	resultsIterator, err := getQueryResultForQueryString(ctx, marketplaceCollection, queryString)

	if err != nil {
		return nil, err
	}
	return constructInterestTokensQueryResponseFromIterator(resultsIterator)
}

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, collectionName string, queryString string) (shim.StateQueryIteratorInterface, error) {

	resultsIterator, err := ctx.GetStub().GetPrivateDataQueryResult(collectionName, queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return resultsIterator, nil
}

func constructPublicDevicesQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*DevicePublicDetails, error) {
	var assets []*DevicePublicDetails
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset DevicePublicDetails
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func constructInterestTokensQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*InterestToken, error) {
	var assets []*InterestToken
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset InterestToken
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func constructTradeAgreementsQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*TradeAgreement, error) {
	var assets []*TradeAgreement
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset TradeAgreement
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func constructDevicesDataQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]DeviceData, error) {
	var assets []DeviceData
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset DeviceData
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}

	return assets, nil
}

func (s *SmartContract) GetRevokeTime(ctx contractapi.TransactionContextInterface, tradeId string) (time.Time, error) {
	tradeAgreementCollection, err := getTradeAgreementCollection(ctx)
	if err != nil {
	}
	tradeAgreementAsBytes, err := ctx.GetStub().GetPrivateData(tradeAgreementCollection, tradeId)
	if err != nil {
	}
	var tradeAgreement TradeAgreement
	err = json.Unmarshal(tradeAgreementAsBytes, &tradeAgreement)
	fmt.Println("\nTrade Agreement")
	fmt.Println(tradeAgreement)
	if err != nil {
	}
	return tradeAgreement.RevokeTime, nil
}
