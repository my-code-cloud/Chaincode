// Test Chaincode for Thesis Research

package chaincode

import (
	//"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	//"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// STRUCTS  (Note: Fields must start with a capital letter to be exported to JSON)

type Client struct {
	// contains filtered or unexported fields
}
type ClientOption func(*Client) error

// Asset: Generic type used for queries on dB which contains multiple types
type Asset struct {
	// Fields from Part Struct:
	ObjectType   string       `json:"docType"`
	PartID       string       `json:"partID,omitempty" metadata:",optional"`
	PartName     string       `json:"partName,omitempty" metadata:",optional"`
	ParentPart   string       `json:"parentPart,omitempty" metadata:",optional"`
	Customer     string       `json:"customer,omitempty" metadata:",optional"`
	NumOfDims    int          `json:"numOfDims,omitempty" metadata:",optional"`
	Requirements []DesignData `json:"requirements,omitempty" metadata:",optional"`
	// Unique fields from Order struct:
	OrderID      string `json:"orderID,omitempty" metadata:",optional"`
	OrderDate    string `json:"orderDate,omitempty" metadata:",optional"`
	DueDate      string `json:"dueDate,omitempty" metadata:",optional"`
	QtyRequired  int    `json:"qtyRequired,omitempty" metadata:",optional"`
	QtyRemaining int    `json:"qtyRemaining,omitempty" metadata:",optional"`
	Supplier     string `json:"supplier,omitempty" metadata:",optional"`
	Status       string `json:"status,omitempty" metadata:",optional"`
	Comments     string `json:"comments,omitempty" metadata:",optional"`
	// Unique fields from PartRecord struct:
	LotID          string           `json:"lotID,omitempty" metadata:",optional"`
	Quantity       int              `json:"quantity,omitempty" metadata:",optional"`
	QualityStatus  string           `json:"qualityStatus,omitempty" metadata:",optional"`
	Operation      string           `json:"operation,omitempty" metadata:",optional"`
	Subcomponents  []string         `json:"subcomponents,omitempty" metadata:",optional"`
	InspectionData []InspectionData `json:"inspectionData,omitempty" metadata:",optional"`
	Location       string           `json:"location,omitempty" metadata:",optional"`
	UpdateBy       string           `json:"updateBy,omitempty" metadata:",optional"`
}

// PublicAsset: struct for network-wide referencing (both parts and orders)
type PublicAsset struct {
	ObjectType  string   `json:"docType"`
	AssetID     string   `json:"assetID"`
	ParentAsset string   `json:"ParentAsset,omitempty" metadata:",optional"`
	ChildAssets []string `json:"childAssets,omitempty" metadata:",optional"`
	Issuer      string   `json:"issuer,omitempty" metadata:",optional"`
	Collection  string   `json:"collection"`                            // provide location for more information
	Status      string   `json:"status,omitempty" metadata:",optional"` // use for quality alerts and actions
}

// Order: struct for Product Orders
type Order struct {
	ObjectType   string `json:"docType"` // used to distinguish different types of objects in state database
	OrderID      string `json:"orderID"` // primary key
	PartID       string `json:"PartID"`  // foreign key: reference part ID of existing part for order
	OrderDate    string `json:"orderDate"`
	DueDate      string `json:"dueDate"`
	QtyRequired  int    `json:"qtyRequired"`  // quantity of parts required for order
	QtyRemaining int    `json:"qtyRemaining"` // quantity of parts remaining to complete order
	Customer     string `json:"customer"`     // submitting organization
	Supplier     string `json:"supplier"`     // fulfilling organization
	Status       string `json:"status"`       // used to show supplier acceptance of order
	Comments     string `json:"comments,omitempty" metadata:",optional"`
}

// Part: struct for products (private data)
type Part struct {
	ObjectType   string       `json:"docType"` //used to distinguish different types of objects in state database
	PartID       string       `json:"partID"`
	PartName     string       `json:"partName"`
	ParentPart   string       `json:"parentPart"` //part (if any) that this part goes into (0-1)
	Customer     string       `json:"customer"`
	NumOfDims    int          `json:"numOfDims"`
	Requirements []DesignData `json:"requirements"` //embedded struct
}

// DesignData: define customer dimension requirements (ucl/lcl/target for each dimension)
type DesignData struct {
	DimensionID int     `json:"dimID"`
	UCL         float32 `json:"ucl"`
	LCL         float32 `json:"lcl"`
	Target      float32 `json:"target"`
	CPK         float32 `json:"cpk"`
}

// PartRecord: struct for production data
// Note: added 'metadata:",optional"' to prevent error when querying record with null attribute values
type PartRecord struct {
	ObjectType     string           `json:"docType"`                                       // used to differentiate parts, orders, records, etc.
	LotID          string           `json:"lotID"`                                         //unique identifier for product record
	OrderID        string           `json:"orderID"`                                       //link part data to supply chain database
	PartID         string           `json:"partID"`                                        //link part data to supply chain database
	Quantity       int              `json:"quantity"`                                      //number of parts in a lot
	QualityStatus  string           `json:"qualityStatus"`                                 //results of inspections given as attribute (pass/fail) values
	Operation      string           `json:"operation"`                                     // current operation or stage of parts in the lifecycle
	Subcomponents  []string         `json:"subcomponents,omitempty" metadata:",optional"`  // Lot Numbers of child parts used in assembly
	InspectionData []InspectionData `json:"inspectionData,omitempty" metadata:",optional"` //dimensional measurements (array of structs)
	Location       string           `json:"location"`                                      //physical location of parts
	UpdateBy       string           `json:"updateBy"`                                      // record person or entity making update
}

// InspectionData: struct for recording measured dimensions
type InspectionData struct {
	DimensionID   int       `json:"dimID"` //foreign key (primary key defined elsewhere)
	Measurements  []float32 `json:"measurements,omitempty" metadata:",optional"`
	Inspector     string    `json:"inspector"` // person or entity carrying out the inspection
	QualityResult string    `json:"qualityResult"`
}

// HistoryQueryResult structure used for returning result of history query:
type HistoryQueryResult struct {
	Record    *PublicAsset `json:"record"`
	TxId      string       `json:"txId"`
	Timestamp time.Time    `json:"timestamp"`
	IsDelete  bool         `json:"isDelete"`
}

// SmartContract: contract for handling writing and reading from the world state
type SmartContract struct {
	contractapi.Contract
}

//========================================================================================
// 															list of functions to create
//========================================================================================
//
// add part
// add dimensions to part
// create order for part
//
// accept order
// add subcomponent
// add dimensions to subcomponent
// manufacture part(s)
// inspect and approve part(s)
// ship part(s)
//
// receive shipment
// update order
//
// ~~ helper functions ~~
// update order
// update part status
// add measurements
// evaluate measurements
//

//========================================================================================
// 																   Function Definitions
//========================================================================================

// *** PART DEFINITIONS ***
// FUNCTION AddPart: create part and update parent part (using private data collections)
func (s *SmartContract) AddPart(ctx contractapi.TransactionContextInterface) error {

	// get input from transient field:
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		// log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}

	// temporary struct to hold input:
	type transientInput struct {
		PartID     string    `json:"partID"`
		PartName   string    `json:"partName"`
		ParentPart string    `json:"parentPart"`
		NumOfDims  int       `json:"numOfDims"`
		Dimensions []float32 `json:"dimensions"`
		Collection string    `json:"collection"`
	}

	// save input to struct
	var assetInput transientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	// input validation:
	if len(assetInput.PartID) == 0 {
		return fmt.Errorf("PartID field must be a non-empty string")
	}
	if len(assetInput.PartName) == 0 {
		return fmt.Errorf("Part Name field must be a non-empty string")
	}
	if assetInput.NumOfDims <= 0 {
		return fmt.Errorf("Number of Dimensions field must be a positive integer")
	}
	if len(assetInput.Collection) == 0 {
		return fmt.Errorf("Collection name must be provided")
	}
	if len(assetInput.Dimensions) > 0 {
		// check that dimension input matches expected number of dimensions (5 values per dimension)
		if len(assetInput.Dimensions) != assetInput.NumOfDims*5 {
			return fmt.Errorf("Unexpected number of dimension arguments provided. Expected %v values, but received %v values", assetInput.NumOfDims*5, len(assetInput.Dimensions))
		}
		// check if dimension input is correctly formatted (consecutive dimension numbers at each i*5 position)
		for i := 0; i < assetInput.NumOfDims; i++ {
			if int(assetInput.Dimensions[i*5]) != i+1 {
				return fmt.Errorf("Dimension number mismatch at position %v: expected %v, but got %v.", i*5, i+1, assetInput.Dimensions[i*5])
			}
		}
	}

	// Check if part already exists
	assetAsBytes, err := ctx.GetStub().GetState(assetInput.PartID)
	if err != nil {
		return fmt.Errorf("While checking if part already exists, failed to get asset: %v", err)
	} else if assetAsBytes != nil {
		fmt.Println("Part already exists: " + assetInput.PartID)
		return fmt.Errorf("this part already exists: " + assetInput.PartID)
	}

	// If Parent Part is provided, ensure Parent Part exists
	if assetInput.ParentPart == "" {
		fmt.Println("No parent part provided") // not a fatal error - if not provided, then skip error-checking
	} else {
		assetAsBytes, err = ctx.GetStub().GetState(assetInput.ParentPart)
		if err != nil {
			return fmt.Errorf("While checking for parent part, failed to get asset: %v", err)
		} else if assetAsBytes == nil {
			fmt.Println("Problem retrieving parent part: " + assetInput.ParentPart)
			return fmt.Errorf("this asset does not exist: " + assetInput.ParentPart)
		}
	}

	// get client MSPID
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the client's MSPID: %v", err)
	}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("AddPart cannot be performed: Error %v", err)
	}

	// ** Channel Record **

	publicPart := PublicAsset{
		ObjectType:  "part",
		AssetID:     assetInput.PartID,
		Issuer:      strings.TrimSuffix(clientMSPID, "MSP"),
		ParentAsset: assetInput.ParentPart,
		ChildAssets: make([]string, 0),
		Status:      "active",
		Collection:  assetInput.Collection,
	}

	// convert to JSON
	assetJSONasBytes, err := json.Marshal(publicPart)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}

	// Save asset to channel
	log.Printf("AddPart Put: channel, ID: %v, owner: %v", assetInput.PartID, clientMSPID)
	err = ctx.GetStub().PutState(assetInput.PartID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put part %v into world state: %v", assetInput.PartID, err)
	}

	// ** Private Collection **

	partPrivateDetails := Part{
		ObjectType: "part",
		PartID:     assetInput.PartID,
		PartName:   assetInput.PartName,
		ParentPart: assetInput.ParentPart,
		//ChildParts:   make([]string, 0),
		Customer:     strings.TrimSuffix(clientMSPID, "MSP"),
		NumOfDims:    assetInput.NumOfDims,
		Requirements: make([]DesignData, assetInput.NumOfDims),
	}

	// update dimension data if provided:
	if len(assetInput.Dimensions) > 0 {
		for i := 0; i < len(assetInput.Dimensions); i += 5 {
			partPrivateDetails.Requirements[i/5] = DesignData{
				DimensionID: int(assetInput.Dimensions[i]),
				UCL:         assetInput.Dimensions[i+1],
				LCL:         assetInput.Dimensions[i+2],
				Target:      assetInput.Dimensions[i+3],
				CPK:         assetInput.Dimensions[i+4],
			}
		}
	}

	// marshal asset details to JSON
	assetPrivateDetailsAsBytes, err := json.Marshal(partPrivateDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal into JSON: %v", err)
	}

	// Put asset value into owners org specific private data collection
	log.Printf("Put: collection %v, ID %v", assetInput.Collection, assetInput.PartID)
	err = ctx.GetStub().PutPrivateData(assetInput.Collection, assetInput.PartID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put part private details: %v", err)
	}

	// update parent part on channel
	err = s.addChild(ctx, assetInput.PartID, assetInput.ParentPart)
	if err != nil {
		return fmt.Errorf("failure while updating parent part: %v", err)
	}

	return nil

}

// FUNCTION UpdateDimensions: update part dimensions and save to private data collection
// Warning: this is an old function and should be reviewed for updates before using
func (s *SmartContract) UpdateDimensions(ctx contractapi.TransactionContextInterface) error {

	//TODO: error-checking
	// make sure that part.Requirements[dimID] exists, and
	// check to see if part.Requirements[dimID] is empty to avoid overwriting

	// get input from transient field
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}
	// extract dimensions slice from transient
	transientDimsJSON, ok := transientMap["dimensions"]
	if !ok {
		return fmt.Errorf("dimensional data not found in the transient map input")
	}

	// temporary struct to hold input
	type transientInput struct {
		PartID     string    `json:"partID"`
		Dimensions []float32 `json:"dimensions"`
		Collection string    `json:"collection"`
	}
	// save input to struct
	var assetInput transientInput
	err = json.Unmarshal(transientDimsJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// verify client is submitting to own peer
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	// validate input
	if len(assetInput.PartID) == 0 {
		return fmt.Errorf("PartID field must be a non-empty string")
	}
	if len(assetInput.Dimensions) == 0 {
		return fmt.Errorf("No dimensional values provided")
	}
	if len(assetInput.Dimensions)%5 != 0 {
		return fmt.Errorf("Unexpected number of dimension arguments provided. Expected a multiple of 5, but received %v", len(assetInput.Dimensions))
	}
	if len(assetInput.Collection) == 0 {
		return fmt.Errorf("Collection name must be provided")
	}

	// get part & unmarshal
	asset, err := s.ReadPartPrivateDetails(ctx, assetInput.Collection, assetInput.PartID)
	if err != nil {
		return fmt.Errorf("Error reading asset: %v.", err)
	} else if asset == nil {
		return fmt.Errorf("%v does not exist", assetInput.PartID)
	}
	// check for ill-formed part entry or mismatched input to avoid hard crash of chaincode
	if len(asset.Requirements) != len(assetInput.Dimensions)/5 {
		log.Printf("Number of input dimensions provided do not match those of the retrieved part:")
		log.Printf("Part %v read from ledger has %v dimensions", asset.PartID, len(asset.Requirements))
		log.Printf("Input dimensions for Part %v has %v dimensions", assetInput.PartID, len(assetInput.Dimensions)/5)
		return fmt.Errorf("Number of input dimensions do not match those of part retrieved from ledger")
	}

	// use loop and update part with new dimension data
	for i := 0; i < len(assetInput.Dimensions); i += 5 {
		asset.Requirements[i/5] = DesignData{
			DimensionID: int(assetInput.Dimensions[i]),
			UCL:         assetInput.Dimensions[i+1],
			LCL:         assetInput.Dimensions[i+2],
			Target:      assetInput.Dimensions[i+3],
			CPK:         assetInput.Dimensions[i+4],
		}
	}

	// marshal part to JSON
	assetPrivateDetailsAsBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("Failed to marshal into JSON: %v", err)
	}

	// write to private collection
	log.Printf("Put: collection %v, ID %v", assetInput.Collection, assetInput.PartID)
	err = ctx.GetStub().PutPrivateData(assetInput.Collection, assetInput.PartID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("Failed to put part private details: %v", err)
	}

	return nil
}

// *** ORDER DEFINITION ***
// FUNCTION AddOrder: create order for part - writes full record to private collection, writes "header" to public collection
func (s *SmartContract) AddOrder(ctx contractapi.TransactionContextInterface) error {

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}
	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		return fmt.Errorf("asset not found in the transient map input")
	}

	// temp struct for input
	type assetTransientInput struct {
		OrderID string `json:"orderID"` //primary key
		PartID  string `json:"PartID"`  //foreign key: reference part ID of existing part for order
		//OrderDate  string `json:"orderDate"`  //convert to auto-set
		DueDate    string `json:"dueDate"`
		Quantity   int    `json:"quantity"`
		Supplier   string `json:"supplier"`
		Collection string `json:"collection"`
	}

	// unmarshal and save input to temporary struct
	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	// error checking:
	if len(assetInput.PartID) == 0 {
		return fmt.Errorf("PartID field must be a non-empty string")
	}
	if len(assetInput.DueDate) == 0 {
		return fmt.Errorf("Due Date field must be a non-empty string")
	}
	if assetInput.Quantity <= 0 {
		return fmt.Errorf("Quantity field must be a positive integer")
	}
	if len(assetInput.Supplier) == 0 {
		return fmt.Errorf("Supplier field must be a non-empty string")
	}
	if len(assetInput.Collection) == 0 {
		return fmt.Errorf("Collection name must be provided")
	}

	// Check if asset (order) already exists
	assetAsBytes, err := ctx.GetStub().GetState(assetInput.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order asset: %v", err)
	} else if assetAsBytes != nil {
		fmt.Println("Order already exists: " + assetInput.OrderID)
		return fmt.Errorf("Order %v already exists.", assetInput.OrderID)
	}

	// Check if part already exists in collection
	assetAsBytes, err = ctx.GetStub().GetPrivateData(assetInput.Collection, assetInput.PartID)
	if err != nil {
		return fmt.Errorf("failed to get part asset: %v", err)
	} else if assetAsBytes == nil {
		fmt.Println("Listed part does not exist: " + assetInput.PartID)
		return fmt.Errorf("Referenced part %v does not exist in this collection. Please ensure you are using the correct collection and create part before issuing order.", assetInput.PartID)
	}

	// get client MSPID
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the client's MSPID: %v", err)
	}

	// Verify that client is submitting request to own organization's peer
	// This is to prevent another organization from reading or writing private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	// get time for order
	loc, _ := time.LoadLocation("UTC")
	currentDate := time.Now().In(loc).Format("2006-01-02")

	// build order with input values
	order := Order{
		ObjectType:   "order",
		OrderID:      assetInput.OrderID,
		PartID:       assetInput.PartID,
		OrderDate:    currentDate,
		DueDate:      assetInput.DueDate,
		QtyRequired:  assetInput.Quantity,
		QtyRemaining: assetInput.Quantity,
		Customer:     strings.TrimSuffix(clientMSPID, "MSP"),
		Supplier:     strings.TrimSuffix(assetInput.Supplier, "MSP"),
		Status:       "new order",
	}

	// marshal to JSON
	assetPrivateDetailsAsBytes, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}

	// Put asset appraised value into owners org specific private data collection
	log.Printf("Put: collection %v, ID %v", assetInput.Collection, assetInput.OrderID)
	err = ctx.GetStub().PutPrivateData(assetInput.Collection, assetInput.OrderID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}

	// ------------------------------------------------------------------------------------------
	// save order to "public" collection
	// build order with input values
	publicOrder := PublicAsset{
		ObjectType: "order",
		AssetID:    assetInput.OrderID,
		Issuer:     strings.TrimSuffix(clientMSPID, "MSP"),
		Status:     "new order",
		Collection: assetInput.Collection,
	}

	// marshal to JSON
	assetJSONasBytes, err := json.Marshal(publicOrder)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}

	// Save asset to public collection
	log.Printf("CreateAsset Put: channel, ID %v, owner %v", assetInput.OrderID, clientMSPID)
	err = ctx.GetStub().PutState(assetInput.OrderID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset %v into world state: %v", assetInput.OrderID, err)
	}

	return nil
}

// FUNCTION AcceptRejectOrder: accept or reject order - can only be called by supplier listed in order
// --consider modifying function to allow for any approved status value, not just accept/reject
func (s *SmartContract) AcceptRejectOrder(ctx contractapi.TransactionContextInterface) error {

	// Optional TODO: change comments from string to []string and append comments for history

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}
	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		return fmt.Errorf("asset not found in the transient map input")
	}

	// temp struct for input
	type assetTransientInput struct {
		OrderID    string `json:"orderID"`
		Decision   string `json:"decision"`
		Comments   string `json:"comments"`
		Collection string `json:"collection"`
	}

	// unmarshal and save input to temporary struct
	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	// error checking:
	if len(assetInput.OrderID) == 0 {
		return fmt.Errorf("OrderID field must be a non-empty string")
	}
	if len(assetInput.Decision) == 0 {
		return fmt.Errorf("Decision field must be a non-empty string")
	}
	if assetInput.Decision != "accept" && assetInput.Decision != "reject" {
		return fmt.Errorf("Valid decisions are 'accept' or 'reject'")
	}
	if len(assetInput.Collection) == 0 {
		return fmt.Errorf("Collection name must be provided")
	}

	// Check if asset (order) already exists, get order from private collection & unmarshal
	privateOrder, err := s.ReadOrderPrivateDetails(ctx, assetInput.Collection, assetInput.OrderID)
	if err != nil {
		return fmt.Errorf("Error reading order: %v.", err)
	} else if privateOrder == nil {
		return fmt.Errorf("%v does not exist", assetInput.OrderID)
	}

	// Check if order status is valid for accepting
	switch {
	case privateOrder.Status == "new order": // new order can be accepted/rejected by supplier
	case privateOrder.Status == "accepted by supplier": // allow supplier to change their mind after acccepting?
	case privateOrder.Status == "rejected by supplier": // allow supplier to change their mind after rejecting
	default: // any other order status from above indicates that order was accepted and processed
		return fmt.Errorf("Order was already accepted by supplier and can no longer be updated.")
	}

	// get client MSPID
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the client's MSPID: %v", err)
	}

	// Verify that client is submitting request to own organization's peer
	// This is to prevent another organization from reading or writing private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	// ** Private Collection **

	// Check if ClientMSPID matches order Supplier field
	if privateOrder.Supplier != strings.TrimSuffix(clientMSPID, "MSP") {
		fmt.Printf("Client MSPID %v does not match Supplier MSPID %v", privateOrder.Supplier, clientMSPID)
		return fmt.Errorf("Client MSPID %v does not match Supplier MSPID %v", clientMSPID, privateOrder.Supplier)
	}

	// change decison to past-tense for update
	var decision string
	if assetInput.Decision == "accept" {
		decision = "accepted by supplier"
	} else if assetInput.Decision == "reject" {
		decision = "rejected by supplier"
	} else {
		return fmt.Errorf("Unexpected Decision Format")
	}

	// update order
	privateOrder.Status = decision
	privateOrder.Comments = assetInput.Comments

	// marshal order to JSON
	assetPrivateDetailsAsBytes, err := json.Marshal(privateOrder)
	if err != nil {
		return fmt.Errorf("Failed to marshal into JSON: %v", err)
	}

	// write to private collection
	log.Printf("Put: collection %v, ID %v", assetInput.Collection, assetInput.OrderID)
	err = ctx.GetStub().PutPrivateData(assetInput.Collection, assetInput.OrderID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("Failed to put order private details: %v", err)
	}

	// ** Public Collection **

	// get order from channel & unmarshal
	publicOrder, err := s.ReadAsset(ctx, assetInput.OrderID)
	if err != nil {
		return fmt.Errorf("Error reading order: %v.", err)
	} else if publicOrder == nil {
		return fmt.Errorf("%v does not exist", assetInput.OrderID)
	}

	// update order
	publicOrder.Status = assetInput.Decision

	// marshal to JSON
	assetJSONasBytes, err := json.Marshal(publicOrder)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}

	// Save asset to channel
	log.Printf("CreateAsset Put: channel, ID %v, owner %v", assetInput.OrderID, clientMSPID)
	err = ctx.GetStub().PutState(assetInput.OrderID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset %v into world state: %v", assetInput.OrderID, err)
	}

	return nil

}

// *** MANUFACTURE SECTION ***

// FUNCTION MakePart: create new lot record
func (s *SmartContract) MakePart(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		return "", fmt.Errorf("asset not found in the transient map input")
	}

	// temp struct for input
	type assetTransientInput struct {
		PartID        string   `json:"PartID"`
		OrderID       string   `json:"orderID"`
		Quantity      int      `json:"quantity"`
		Subcomponents []string `json:"subcomponents"` // Lot Numbers of child parts used in assembly
		Collection    string   `json:"collection"`
	}

	// unmarshal and save input to temporary struct
	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	// error checking:
	if len(assetInput.PartID) == 0 {
		return "", fmt.Errorf("PartID field is required. (It must be a non-empty string.)")
	}
	if len(assetInput.OrderID) == 0 {
		return "", fmt.Errorf("OrderID is required. (It must be a non-empty string.)")
	}
	if assetInput.Quantity <= 0 {
		return "", fmt.Errorf("The Lot quantity is required.  (It must be an integer > 0.)")
	}
	if len(assetInput.Collection) == 0 {
		return "", fmt.Errorf("Collection name must be provided")
	}

	log.Printf("beginning manufacture validation process for part %v", assetInput.PartID)

	// Check if order exists and unmarshal
	privateOrder, err := s.ReadOrderPrivateDetails(ctx, assetInput.Collection, assetInput.OrderID)
	if err != nil {
		return "", fmt.Errorf("Error reading order: %v.", err)
	} else if privateOrder == nil {
		return "", fmt.Errorf("%v does not exist", assetInput.OrderID)
	}

	// Check if order is already accepted
	if privateOrder.Status == "new order" {
		return "", fmt.Errorf("Order must be accepted before parts can be manufactured.")
	}

	// get part from public channel & unmarshal
	channelPart, err := s.ReadAsset(ctx, assetInput.PartID)
	if err != nil {
		return "", fmt.Errorf("Error reading part from channel: %v.", err)
	} else if channelPart == nil {
		return "", fmt.Errorf("%v does not exist", assetInput.PartID)
	}

	// get part from private collection and unmarshal
	part, err := s.ReadPartPrivateDetails(ctx, assetInput.Collection, assetInput.PartID)
	if err != nil {
		return "", fmt.Errorf("Error reading asset: %v.", err)
	} else if part == nil {
		return "", fmt.Errorf("%v does not exist", assetInput.PartID)
	}

	// Verify that client is submitting request to own organization's peer
	// This is to prevent another organization from reading or writing private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return "", fmt.Errorf("Invalid credentials. Cannot perform transaction: Error %v", err)
	}

	// get client MSPID
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting the client's MSPID: %v", err)
	}

	// Check if ClientMSPID matches order Supplier field
	if privateOrder.Supplier != strings.TrimSuffix(clientMSPID, "MSP") {
		fmt.Printf("Client MSPID %v does not match Supplier MSPID %v", privateOrder.Supplier, clientMSPID)
		return "", fmt.Errorf("Client MSPID %v does not match Supplier MSPID %v", clientMSPID, privateOrder.Supplier)
	}

	// get client user ID attribute from client x509 certificate
	clientID, found, err := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	if err != nil {
		return "", fmt.Errorf("error getting attribute EnrollmentID: %v", err)
	}
	if found == false {
		log.Println("Client attribute EnrollmentID not found")
	}
	log.Println("client id value is:", clientID)
	clientOrg := strings.TrimSuffix(clientMSPID, "MSP")

	// construct user identity
	userID := clientOrg + "_" + clientID

	// set lot number:
	loc, _ := time.LoadLocation("UTC")
	lotID := assetInput.PartID + time.Now().In(loc).Format("20060102150405") + clientOrg

	// if part contains child parts, require the child part LotIDs for assembly:
	if len(channelPart.ChildAssets) > 0 {
		log.Printf("Child Parts exist for Part %v", part.PartID)
		if len(channelPart.ChildAssets) > len(assetInput.Subcomponents) {
			log.Printf("Number of provided Subassembly Lot Numbers is less than the number of required Child Parts")
			return "", fmt.Errorf("Incorrect number of subassembly Lot Numbers Provided.  Expected minimum of %v subcomponent lot numbers, but received %v subcomponent lot numbers.", len(channelPart.ChildAssets), len(assetInput.Subcomponents))
		}
	}
	log.Printf("Finished subcomponent number validation")
	//loop through subcomponents and ensure valid lot number is supplied:
	for _, v := range assetInput.Subcomponents {
		log.Printf("Begin loop through subcomponents: %v", assetInput.Subcomponents)
		//check component and all subcomponents for quality alerts:
		qualityCheck, err := s.CheckForQualityAlerts(ctx, v)
		if err != nil {
			return "", fmt.Errorf("error: %v", err)
		} else if qualityCheck != "pass" {
			log.Printf("Quality alert detected for component ID %v", v)
			return "", fmt.Errorf("Issue found when checking component %v: %v", v, qualityCheck)
		}
		// do not need to check lot location as expect to receive lot # by scanning physical barcode with parts
	}

	log.Printf("Finished subcomponent quantity validation")

	//loop through subcomponents and ensure valid lot number is supplied:
	for _, v := range assetInput.Subcomponents {
		log.Printf("Begin loop through subcomponents: %v", assetInput.Subcomponents)
		//check component and all subcomponents for quality alerts:
		qualityCheck, err := s.CheckForQualityAlerts(ctx, v)
		if err != nil {
			return "", fmt.Errorf("error: %v", err)
		} else if qualityCheck != "pass" {
			log.Printf("Quality alert detected for component ID %v", v)
			return "", fmt.Errorf("Issue found when checking component %v: %v", v, qualityCheck)
		}
		// do not need to check lot location as expect to receive lot # by scanning physical barcode with parts
	}
	log.Printf("Finished subcomponent validation loop")

	// build record with input values
	record := PartRecord{
		ObjectType:     "lot record",
		LotID:          lotID,
		PartID:         assetInput.PartID,
		OrderID:        assetInput.OrderID,
		Quantity:       assetInput.Quantity,
		QualityStatus:  "awaiting inspection",
		Operation:      "manufacturing",
		Subcomponents:  assetInput.Subcomponents,
		InspectionData: make([]InspectionData, part.NumOfDims), // initialize empty slice
		Location:       clientOrg,
		UpdateBy:       userID,
	}

	// marshal to JSON
	assetPrivateDetailsAsBytes, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("failed to marshal part into JSON: %v", err)
	}

	// Put asset appraised value into owners org specific private data collection
	log.Printf("Put: collection %v, ID %v", assetInput.Collection, lotID)
	err = ctx.GetStub().PutPrivateData(assetInput.Collection, lotID, assetPrivateDetailsAsBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put asset private details: %v", err)
	}

	// update Order only if required:
	if privateOrder.Status != "manufacturing" {

		// update Order Status
		privateOrder.Status = "manufacturing"

		// marshal order to JSON
		assetPrivateDetailsAsBytes, err = json.Marshal(privateOrder)
		if err != nil {
			return "", fmt.Errorf("Failed to marshal order into JSON: %v", err)
		}

		// write to private collection
		log.Printf("Put: collection %v, ID %v", assetInput.Collection, assetInput.OrderID)
		err = ctx.GetStub().PutPrivateData(assetInput.Collection, assetInput.OrderID, assetPrivateDetailsAsBytes)
		if err != nil {
			return "", fmt.Errorf("Failed to put order private details: %v", err)
		}
	}

	// update publicRecord for lookups
	publicRecord := PublicAsset{
		ObjectType:  "lot record",
		AssetID:     lotID,
		ChildAssets: assetInput.Subcomponents,
		Collection:  assetInput.Collection,
	}

	// convert to JSON
	assetJSONasBytes, err := json.Marshal(publicRecord)
	if err != nil {
		return "", fmt.Errorf("failed to marshal record %v into JSON: %v", lotID, err)
	}
	// Save to channel
	log.Printf("Save Record %v to Channel", lotID)
	err = ctx.GetStub().PutState(lotID, assetJSONasBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put record %v into word state: %v", lotID, err)
	}

	return lotID, nil
}

// FUNCTION InspectPart: inspect and approve or reject partRecord(s)
func (s *SmartContract) InspectPart(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		return "", fmt.Errorf("asset not found in the transient map input")
	}

	// temp struct for input
	type assetTransientInput struct {
		PartID       string    `json:"partID"`
		LotID        string    `json:"lotID"`
		DimensionID  int       `json:"dimensionID"`
		Measurements []float32 `json:"measurements"`
		Collection   string    `json:"collection"`
	}

	// unmarshal and save input to temporary struct
	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	// error checking:
	if len(assetInput.PartID) == 0 {
		return "", fmt.Errorf("PartID field is required. (It must be a non-empty string.)")
	}
	if assetInput.DimensionID == 0 {
		return "", fmt.Errorf("DimensionID is required. (It must be a non-empty string.)")
	}
	if len(assetInput.Measurements) == 0 {
		return "", fmt.Errorf("One or more dimension measurements are required.  (Please provide a non-empty slice.)")
	}
	if len(assetInput.Collection) == 0 {
		return "", fmt.Errorf("Collection name must be provided")
	}

	// Check if partRecord already exists
	partRecord, err := s.ReadPartRecord(ctx, assetInput.Collection, assetInput.LotID)
	if err != nil {
		return "", fmt.Errorf("Error reading partRecord: %v.", err)
	} else if partRecord == nil {
		return "", fmt.Errorf("%v does not exist", assetInput.LotID)
	}

	// Verify that client is submitting request to own organization's peer
	// This is to prevent another organization from reading or writing private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return "", fmt.Errorf("Invalid credentials. Cannot perform transaction: Error %v", err)
	}

	// get client MSPID
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting the client's MSPID: %v", err)
	}

	// get client user ID attribute from client x509 certificate
	clientID, found, err := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	if err != nil {
		return "", fmt.Errorf("error getting attribute EnrollmentID: %v", err)
	}
	if found == false {
		log.Println("Client attribute EnrollmentID not found")
	}
	log.Println("client id value is:", clientID)

	// construct user identity
	org := strings.TrimSuffix(clientMSPID, "MSP")
	userID := org + "_" + clientID

	// confirm lot number matches part ID provided
	if partRecord.PartID != assetInput.PartID {
		log.Printf("Input error: Provided Part ID %v does not match Part ID %v on provided lot %v", assetInput.PartID, partRecord.PartID, assetInput.LotID)
		return "", fmt.Errorf("Input error: Provided Part ID %v does not match Part ID %v on provided lot %v", assetInput.PartID, partRecord.PartID, assetInput.LotID)
	}

	// get part & unmarshal
	part, err := s.ReadPartPrivateDetails(ctx, assetInput.Collection, assetInput.PartID)
	if err != nil {
		return "", fmt.Errorf("Error reading part: %v.", err)
	} else if part == nil {
		return "", fmt.Errorf("%v does not exist", assetInput.PartID)
	}
	// get part requirements for dimension in question
	Requirements := part.Requirements[assetInput.DimensionID-1]

	// for each measurement, compare against part requirements
	var qResult string
	var qMessage string
	for i, v := range assetInput.Measurements {
		if v > Requirements.UCL || v < Requirements.LCL {
			qResult = "fail"
			qMessage = fmt.Sprintf("Dimension ID %v value of %v is outside the acceptable range of %v-%v.", i, v, Requirements.LCL, Requirements.UCL)
			break
		} else {
			qResult = "pass"
		}
	}
	partRecord.InspectionData[assetInput.DimensionID-1] = InspectionData{
		DimensionID:   assetInput.DimensionID,
		Measurements:  assetInput.Measurements,
		Inspector:     userID,
		QualityResult: qResult,
	}

	// if any dimension is failing, set part quality status to "rejected"
	if qResult == "fail" {
		partRecord.QualityStatus = "rejected"
	} else { // check if all dimensions pass and if so update part quality status to "accepted"
		s.PartQualityAutoUpdate(ctx, partRecord)
	}

	// update partRecord operation
	partRecord.Operation = "inspection"

	// marshal partRecord to JSON
	assetPrivateDetailsAsBytes, err := json.Marshal(partRecord)
	if err != nil {
		return "", fmt.Errorf("Failed to marshal into JSON: %v", err)
	}

	// write to private collection
	log.Printf("Put: collection %v, ID %v", assetInput.Collection, partRecord.LotID)
	err = ctx.GetStub().PutPrivateData(assetInput.Collection, partRecord.LotID, assetPrivateDetailsAsBytes)
	if err != nil {
		return "", fmt.Errorf("Failed to put partRecord private details: %v", err)
	}

	// update publicRecord for lookups
	publicRecord := PublicAsset{
		AssetID:     partRecord.LotID,
		ObjectType:  "lot record",
		ChildAssets: partRecord.Subcomponents,
		Collection:  assetInput.Collection,
	}
	// if quality is fail, then update accordingly
	// updating the public record with rejects is used for global defect tracking & recall identification
	// (customer can always get complete status through private record)
	if partRecord.QualityStatus == "rejected" {
		publicRecord.Status = partRecord.QualityStatus
	}
	// convert to JSON
	assetJSONasBytes, err := json.Marshal(publicRecord)
	if err != nil {
		return "", fmt.Errorf("failed to marshal record %v into JSON: %v", partRecord.LotID, err)
	}
	// Save to channel
	log.Printf("Save Record %v to Channel", partRecord.LotID)
	err = ctx.GetStub().PutState(partRecord.LotID, assetJSONasBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put record %v into world state: %v", partRecord.LotID, err)
	}

	// give useful return value
	if partRecord.QualityStatus == "rejected" {
		return "Failed inspection: " + qMessage, nil
	} else {
		return "Passed inspection", nil
	}
}

// *** LOGISTICS ***
// FUNCTION: ship part(s)
func (s *SmartContract) ShipParts(ctx contractapi.TransactionContextInterface) error {
	// logic:
	// load PartRecord
	// check if QualityStatus == accepted
	// set PartRecord.Location = in-transit from CLIENT_ORG to Destination_Org
	// set Order.Status = shipped

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		return fmt.Errorf("asset not found in the transient map input")
	}

	// temp struct for input
	type assetTransientInput struct {
		LotID          string `json:"lotID"`
		DestinationOrg string `json:"destinationOrg"`
		Collection     string `json:"collection"`
	}

	// unmarshal and save input to temporary struct
	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	// error checking:
	if len(assetInput.LotID) == 0 {
		return fmt.Errorf("LotID field is required. (It must be a non-empty string.)")
	}
	if len(assetInput.DestinationOrg) == 0 {
		return fmt.Errorf("DestinationOrg field is required. (It must be a non-empty string.)")
	}
	if len(assetInput.Collection) == 0 {
		return fmt.Errorf("Collection name must be provided")
	}

	// Check if partRecord already exists and unmarshal to memory
	partRecord, err := s.ReadPartRecord(ctx, assetInput.Collection, assetInput.LotID)
	if err != nil {
		return fmt.Errorf("Error reading partRecord: %v.", err)
	} else if partRecord == nil {
		return fmt.Errorf("%v does not exist", assetInput.LotID)
	}
	// check if QualityStatus = accepted
	if partRecord.QualityStatus != "accepted" {
		return fmt.Errorf("Unable to ship because Lot %v does not satisfy Quality Requirements.", assetInput.LotID)
	}
	// check all subcomponents for quality alerts
	qualityCheck, err := s.CheckForQualityAlerts(ctx, assetInput.LotID)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	} else if qualityCheck != "pass" {
		return fmt.Errorf("Quality issue found when checking component %v: %v", assetInput.LotID, qualityCheck)
	}

	// get client MSPID
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the client's MSPID: %v", err)
	}
	// get organizations
	fromOrg := strings.TrimSuffix(clientMSPID, "MSP")
	toOrg := strings.TrimSuffix(assetInput.DestinationOrg, "MSP")

	// Update Lot Location and operation
	partRecord.Location = "in transit from " + fromOrg + " to " + toOrg
	partRecord.Operation = "Shipping"

	// Verify that client is submitting request to own organization's peer
	// This is to prevent another organization from reading or writing private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("Invalid credentials. Cannot perform transaction: Error %v", err)
	}

	// get order from private collection & unmarshal
	privateOrder, err := s.ReadOrderPrivateDetails(ctx, assetInput.Collection, partRecord.OrderID)
	if err != nil {
		return fmt.Errorf("Error reading order: %v.", err)
	} else if privateOrder == nil {
		return fmt.Errorf("%v does not exist", partRecord.OrderID)
	}

	// Check to make sure Order is still open
	if privateOrder.Status == "completed" {
		return fmt.Errorf("Unable to process Shipment.  The listed Order %v is already completed.", privateOrder.OrderID)
	} else {
		// Check to see if this shipment will complete the order:
		if partRecord.Quantity >= privateOrder.QtyRemaining {
			privateOrder.Status = "Shipped to customer"
		}
	}

	// marshal order to JSON
	assetPrivateDetailsAsBytes, err := json.Marshal(privateOrder)
	if err != nil {
		return fmt.Errorf("Failed to marshal order into JSON: %v", err)
	}

	// write order to private collection
	log.Printf("Put: collection %v, ID %v", assetInput.Collection, partRecord.OrderID)
	err = ctx.GetStub().PutPrivateData(assetInput.Collection, partRecord.OrderID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("Failed to put order private details: %v", err)
	}

	// marshal partRecord to JSON
	assetPrivateDetailsAsBytes, err = json.Marshal(partRecord)
	if err != nil {
		return fmt.Errorf("Failed to marshal partRecord into JSON: %v", err)
	}

	// write partRecord to private collection
	log.Printf("Put: collection %v, ID %v", assetInput.Collection, partRecord.LotID)
	err = ctx.GetStub().PutPrivateData(assetInput.Collection, partRecord.LotID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("Failed to put partRecord private details: %v", err)
	}

	return nil
}

// FUNCTION: receive shipment (no inspection)
func (s *SmartContract) ReceiveParts(ctx contractapi.TransactionContextInterface) error {
	// load data: LotNumber, Quantity
	// ensure LotID is valid
	// update location
	// update Order

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		return fmt.Errorf("asset not found in the transient map input")
	}

	// temp struct for input
	type assetTransientInput struct {
		LotID      string `json:"lotID"`
		Quantity   int    `json:"quantity"`
		Collection string `json:"collection"`
	}

	// unmarshal and save input to temporary struct
	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	// error checking:
	if len(assetInput.LotID) == 0 {
		return fmt.Errorf("LotID field is required. (It must be a non-empty string.)")
	}
	if assetInput.Quantity == 0 {
		return fmt.Errorf("Quantity is required. (It must be a positive integer.)")
	}
	if len(assetInput.Collection) == 0 {
		return fmt.Errorf("Collection name must be provided")
	}

	// Check if partRecord already exists and unmarshal to memory
	partRecord, err := s.ReadPartRecord(ctx, assetInput.Collection, assetInput.LotID)
	if err != nil {
		return fmt.Errorf("Error reading partRecord: %v.", err)
	} else if partRecord == nil {
		return fmt.Errorf("%v does not exist", assetInput.LotID)
	}
	// check if QualityStatus = accepted
	if partRecord.QualityStatus != "accepted" {
		return fmt.Errorf("Unable to receive shipment because Lot %v does not satisfy Quality Requirements.", assetInput.LotID)
	}
	// check all subcomponents for quality alerts
	qualityCheck, err := s.CheckForQualityAlerts(ctx, assetInput.LotID)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	} else if qualityCheck != "pass" {
		return fmt.Errorf("Quality issue found when checking component %v: %v", assetInput.LotID, qualityCheck)
	}

	// get client MSPID
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the client's MSPID: %v", err)
	}
	// get organizations
	Org := strings.TrimSuffix(clientMSPID, "MSP")

	// Update Lot Location and operation
	partRecord.Location = Org + " Receiving"
	partRecord.Operation = "Receiving"

	// Verify that client is submitting request to own organization's peer
	// This is to prevent another organization from reading or writing private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("Invalid credentials. Cannot perform transaction: Error %v", err)
	}

	// get order from private collection & unmarshal
	privateOrder, err := s.ReadOrderPrivateDetails(ctx, assetInput.Collection, partRecord.OrderID)
	if err != nil {
		return fmt.Errorf("Error reading order: %v.", err)
	} else if privateOrder == nil {
		return fmt.Errorf("%v does not exist", partRecord.OrderID)
	}

	// Check to make sure that Order is still open
	// if not return error ("This order is closed")
	// Else subtract quantity of parts from privateOrder.QtyRemaining
	// Check if privateOrder.QtyRemaining <= 0
	// if so, set privateOrder.QtyRemaining = 0
	// and set privateOrder.Status = "completed"

	if privateOrder.Status != "completed" {
		privateOrder.QtyRemaining = privateOrder.QtyRemaining - assetInput.Quantity
		if privateOrder.QtyRemaining <= 0 {
			privateOrder.QtyRemaining = 0
			privateOrder.Status = "completed"
		}
	} else {
		return fmt.Errorf("Unable to Receive Shipment.  The listed Order %v is already completed.", privateOrder.OrderID)
	}

	// marshal order to JSON
	assetPrivateDetailsAsBytes, err := json.Marshal(privateOrder)
	if err != nil {
		return fmt.Errorf("Failed to marshal order into JSON: %v", err)
	}

	// write order to private collection
	log.Printf("Put: collection %v, ID %v", assetInput.Collection, partRecord.OrderID)
	err = ctx.GetStub().PutPrivateData(assetInput.Collection, partRecord.OrderID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("Failed to put order private details: %v", err)
	}

	// marshal partRecord to JSON
	assetPrivateDetailsAsBytes, err = json.Marshal(partRecord)
	if err != nil {
		return fmt.Errorf("Failed to marshal partRecord into JSON: %v", err)
	}

	// write partRecord to private collection
	log.Printf("Put: collection %v, ID %v", assetInput.Collection, partRecord.LotID)
	err = ctx.GetStub().PutPrivateData(assetInput.Collection, partRecord.LotID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("Failed to put partRecord private details: %v", err)
	}

	return nil
}

// FUNCTION: receive & inspect shipment TODO: not implemented
func (s *SmartContract) ReceivingInspection(ctx contractapi.TransactionContextInterface) error {
	// load data: LotNumber, Quantity, Dimensions to inspect, Inspection Values
	// ensure LotID is valid
	// update location
	// for each Dimension to inspect, inspect against values provided
	// update Order
	return nil
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< Query Functions >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// QueryAssets: an ad-hoc query (rich query) on Channel data
func (s *SmartContract) QueryAssets(ctx contractapi.TransactionContextInterface, queryString string) ([]*PublicAsset, error) {
	return getQueryResultForQueryString(ctx, queryString)
}

// QueryPrivateAssets: ad-hoc query (rich query) on Private Data
func (s *SmartContract) QueryPrivateAssets(ctx contractapi.TransactionContextInterface, collection string, queryString string) ([]*Asset, error) {

	queryResults, err := s.getQueryResultForQueryStringPrivate(ctx, collection, queryString)
	if err != nil {
		return nil, err
	}
	return queryResults, nil
}

// QueryByRange: query by range of asset ID on channel TODO: test
func (s *SmartContract) QueryByRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string) ([]*PublicAsset, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// PrivateQueryByRange: return range of asset IDs from private data collection TODO: test
func (s *SmartContract) PrivateQueryByRange(ctx contractapi.TransactionContextInterface, collection string, startKey string, endKey string) ([]*Asset, error) {

	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(collection, startKey, endKey)
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

// LotTracePrivate: get list of lot numbers where provided lot is used
func (s *SmartContract) LotTracePrivate(ctx contractapi.TransactionContextInterface, collection string, queryLotID string) ([]string, error) {

	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(collection, "", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	resultList := []string{} //declare empty slice to hold return values

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

		if asset.ObjectType == "lot record" {
			for i := 0; i < len(asset.Subcomponents); i++ { //loop through each subcomponent to see if any match queryLotID
				if asset.Subcomponents[i] == queryLotID {
					resultList = append(resultList, asset.LotID) // if so, add asset.LotID to array
					i = len(asset.Subcomponents)                 // stop inner loop on match
				}
			}
		}
	}

	return resultList, nil

}

// LotTraceChannel: get list of lot numbers where provided lot is used in Channel
func (s *SmartContract) LotTraceChannel(ctx contractapi.TransactionContextInterface, queryLotID string) ([]string, error) {

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	resultList := []string{} //declare empty slice to hold return values

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset *PublicAsset
		err = json.Unmarshal(response.Value, &asset)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
		}

		if asset.ObjectType == "lot record" {
			for i := 0; i < len(asset.ChildAssets); i++ { //loop through each subcomponent to see if any match queryLotID
				if asset.ChildAssets[i] == queryLotID {
					resultList = append(resultList, asset.AssetID) // if so, add asset.LotID to array
					i = len(asset.ChildAssets)                     // stop inner loop on match
				}
			}
		}
	}

	return resultList, nil

}

// GetHistory: given a TX ID, return corresponding state from Channel
func (s *SmartContract) GetHistory(ctx contractapi.TransactionContextInterface, recordID string) (*Asset, error) {

	var r Asset
	// get record (byte slice)
	rawInput, err := ctx.GetStub().GetState(recordID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %v", err)
	}
	if rawInput == nil {
		return nil, fmt.Errorf("Asset does not exist")
	}
	// unmarshal to JSON
	err = json.Unmarshal(rawInput, &r)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	// return value
	return &r, nil
}

// GetPublicHistory returns the list of transaction IDs for a given asset on Channel
func (t *SmartContract) GetPublicHistory(ctx contractapi.TransactionContextInterface, assetID string) ([]HistoryQueryResult, error) {
	log.Printf("GetPublicHistory: ID %v", assetID)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(assetID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var publicRecord PublicAsset
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &publicRecord)
			if err != nil {
				return nil, err
			}
		} else {
			publicRecord = PublicAsset{
				AssetID: assetID,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &publicRecord,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< Helper Functions >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// DeleteAsset: delete asset from channel and collection world states
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, collection string) error {

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

	valAsbytes, err := ctx.GetStub().GetState(assetDeleteInput.ID) //get the asset from chaincode state
	if err != nil {
		return fmt.Errorf("failed to read asset: %v", err)
	}
	if valAsbytes == nil {
		return fmt.Errorf("asset not found: %v", assetDeleteInput.ID)
	}

	//unmarshal asset to check properties:
	var asset *PublicAsset
	err = json.Unmarshal(valAsbytes, &asset)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	// get client org name
	clientOrg, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the client's MSPID: %v", err)
	}
	clientOrg = strings.TrimSuffix(clientOrg, "MSP")

	//Check if calling org matches issuing org for asset:
	if asset.Issuer != clientOrg {
		log.Printf("Attempt by %v to delete asset %v owned by %v", clientOrg, assetDeleteInput.ID, asset.Issuer)
		return fmt.Errorf("Unable to delete asset that you do not own.")
	}

	//check the asset is in the caller org's private collection
	valAsbytes, err = ctx.GetStub().GetPrivateData(collection, assetDeleteInput.ID)
	if err != nil {
		return fmt.Errorf("failed to read asset from owner's Collection: %v", err)
	}
	if valAsbytes == nil {
		return fmt.Errorf("asset %v does not exist in the current collection %v: ", assetDeleteInput.ID, collection)
	}

	log.Printf("Deleting Asset: %v", assetDeleteInput.ID)
	// delete the asset from channel:
	err = ctx.GetStub().DelState(assetDeleteInput.ID)
	if err != nil {
		return fmt.Errorf("failed to delete asset from channel: %v", err)
	}

	// Finally, delete private details of asset from collection:
	err = ctx.GetStub().DelPrivateData(collection, assetDeleteInput.ID)
	if err != nil {
		return fmt.Errorf("failed to delete asset from priate collection: %v", err)
	}

	return nil

}

// RecordExists returns true when part with given partID exists in world state
func (s *SmartContract) RecordExists(ctx contractapi.TransactionContextInterface, partID string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(partID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// verifyClientOrgMatchesPeerOrg is an internal function used to verify client org id matches peer org id.
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

// submittingClientIdentity
func submittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
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

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*PublicAsset, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// constructQueryResponseFromIterator constructs a slice of assets from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*PublicAsset, error) {
	var assets []*PublicAsset
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset PublicAsset
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// getQueryResultForQueryStringPrivate executes the passed in query string.
func (s *SmartContract) getQueryResultForQueryStringPrivate(ctx contractapi.TransactionContextInterface, collection string, queryString string) ([]*Asset, error) {

	resultsIterator, err := ctx.GetStub().GetPrivateDataQueryResult(collection, queryString)
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

// ReadAsset reads public assets (parts and orders) from channel
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*PublicAsset, error) {

	log.Printf("ReadAsset: ID %v", assetID)
	assetJSON, err := ctx.GetStub().GetState(assetID) //get the asset from channel world state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %v", err)
	}

	//No Asset found, return empty response
	if assetJSON == nil {
		log.Printf("%v does not exist on this channel", assetID)
		return nil, nil
	}

	var asset *PublicAsset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return asset, nil

}

// ReadPartPrivateDetails reads the part private details in organization specific collection
func (s *SmartContract) ReadPartPrivateDetails(ctx contractapi.TransactionContextInterface, collection string, assetID string) (*Part, error) {
	log.Printf("ReadAssetPrivateDetails: collection %v, ID %v", collection, assetID)
	assetDetailsJSON, err := ctx.GetStub().GetPrivateData(collection, assetID) // Get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset details: %v", err)
	}
	if assetDetailsJSON == nil {
		log.Printf("AssetPrivateDetails for %v does not exist in collection %v", assetID, collection)
		return nil, nil
	}

	var assetDetails *Part
	err = json.Unmarshal(assetDetailsJSON, &assetDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// check to make sure asset is type "part"
	if assetDetails.ObjectType != "part" {
		return nil, fmt.Errorf("ID %v is type \"%v,\" not type \"part\"", assetID, assetDetails.ObjectType)
	}

	return assetDetails, nil
}

// ReadOrderPrivateDetails reads the order private details in organization specific collection
func (s *SmartContract) ReadOrderPrivateDetails(ctx contractapi.TransactionContextInterface, collection string, assetID string) (*Order, error) {
	log.Printf("ReadAssetPrivateDetails: collection %v, ID %v", collection, assetID)
	assetDetailsJSON, err := ctx.GetStub().GetPrivateData(collection, assetID) // Get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset details: %v", err)
	}
	if assetDetailsJSON == nil {
		log.Printf("AssetPrivateDetails for %v does not exist in collection %v", assetID, collection)
		return nil, nil
	}

	var assetDetails *Order
	err = json.Unmarshal(assetDetailsJSON, &assetDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// check to make sure asset is "order"
	if assetDetails.ObjectType != "order" {
		return nil, fmt.Errorf("ID %v is type \"%v,\" not type \"order\"", assetID, assetDetails.ObjectType)
	}

	return assetDetails, nil
}

// ReadPartRecord: reads the partRecord from provided collection
func (s *SmartContract) ReadPartRecord(ctx contractapi.TransactionContextInterface, collection string, assetID string) (*PartRecord, error) {
	log.Printf("ReadPartRecord: collection %v, ID %v", collection, assetID)
	assetDetailsJSON, err := ctx.GetStub().GetPrivateData(collection, assetID) // Get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset details: %v", err)
	}
	if assetDetailsJSON == nil {
		log.Printf("PartRecord for %v does not exist in collection %v", assetID, collection)
		return nil, nil
	}

	var assetDetails *PartRecord
	err = json.Unmarshal(assetDetailsJSON, &assetDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// check to make sure asset is "lot record"
	if assetDetails.ObjectType != "lot record" {
		return nil, fmt.Errorf("ID %v is type \"%v,\" not type \"lot record\"", assetID, assetDetails.ObjectType)
	}

	return assetDetails, nil
}

// ReadChannelPartRecord: reads the partRecord from the Channel (as PublicAsset)

// CheckForAsset: check if public asset exists and if so returns type
func (s *SmartContract) CheckForAsset(ctx contractapi.TransactionContextInterface, assetID string) (string, error) {
	assetJSON, err := ctx.GetStub().GetState(assetID) //get the asset from channel world state
	if err != nil {
		return "", fmt.Errorf("failed to read asset: %v", err)
	}

	// If no asset found, return "none"
	if assetJSON == nil {
		log.Printf("%v does not exist on this channel", assetID)
		return "none", nil
	}

	// else, return asset type
	var asset *PublicAsset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return asset.ObjectType, nil
}

// FUNCTION: addChild - adds child part to parent part record
// takes currentPartID (child), parentPartID, collectionName
func (s *SmartContract) addChild(ctx contractapi.TransactionContextInterface, childPartID string, parentPartID string) error {

	// check if parent part is provided or is empty
	if parentPartID == "" {
		fmt.Println("No parent part provided. Skipping update of parent part.")
		return nil
	}

	// ** Channel update **

	// Read parent part from channel and unmarshal
	publicAsset, err := s.ReadAsset(ctx, parentPartID)
	if err != nil {
		return fmt.Errorf("error reading asset: %v", err)
	}
	if publicAsset == nil {
		return fmt.Errorf("%v does not exist", parentPartID)
	}
	// update parent part (append child to slice)
	publicAsset.ChildAssets = append(publicAsset.ChildAssets, childPartID)
	// marshal back to JSON
	assetDetailsAsBytes, err := json.Marshal(publicAsset) // marshal asset details to JSON
	if err != nil {
		return fmt.Errorf("failed to marshal into JSON: %v", err)
	}
	// write updated parent part to channel
	log.Printf("Put: channel, ID %v", parentPartID)
	err = ctx.GetStub().PutState(parentPartID, assetDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put updated Parent part record to channel: %v", err)
	}

	return nil
}

// Check if all dimensions are OK and if so, set part to accepted
// FUNCTION: PartQualityAutoUpdate
func (s *SmartContract) PartQualityAutoUpdate(ctx contractapi.TransactionContextInterface, part *PartRecord) *PartRecord {
	var lastQualityValue string
	for i, v := range part.InspectionData {
		if v.QualityResult == "pass" {
			lastQualityValue = v.QualityResult
			continue
		} else if v.QualityResult == "fail" {
			log.Printf("Part %v failed quality test due to dimension %v which has a value of %v.", part.PartID, i, v)
			lastQualityValue = v.QualityResult
			break
		} else if v.QualityResult == "" {
			log.Println("One or more dimensions not inspected")
			log.Printf("Part %v failed quality test due to dimension %v which has a value of %v.", part.PartID, i, v)
			lastQualityValue = v.QualityResult
			break
		} else {
			log.Println("Unexpected Error when checking Part Quality")
			log.Printf("Part %v failed quality test due to dimension %v which has a value of %v.", part.PartID, i, v)
			lastQualityValue = v.QualityResult
			break
		}
	}
	if lastQualityValue == "pass" {
		part.QualityStatus = "accepted"
	}
	return part
}

// CheckForQualityAlerts: check given lot and all component lots to ensure that none have a quality alert status
func (s *SmartContract) CheckForQualityAlerts(ctx contractapi.TransactionContextInterface, ID string) (string, error) {
	// ---- check partID public record status -----
	log.Printf("running function CheckForQualityAlerts to check ID %v", ID)

	// Read parent part from channel and unmarshal
	publicAsset, err := s.ReadAsset(ctx, ID)
	if err != nil {
		return "", fmt.Errorf("error reading asset: %v", err)
	}
	// check lot ID status:
	if publicAsset.Status == "rejected" {
		log.Printf("ID %v has a status of rejected", ID)
		return fmt.Sprintf("Quality issue detected at lot ID %v", publicAsset.AssetID), nil
		//return "fail", nil
	} else {
		//iterate over each entry in children slice recursively:
		for i, v := range publicAsset.ChildAssets {
			log.Printf("iterating over child asset %v at index %v", v, i)
			result, err := s.CheckForQualityAlerts(ctx, v)
			if err != nil {
				return "", fmt.Errorf("error: %v", err)
			}
			if result != "pass" {
				return fmt.Sprintf("Quality issue detected at lot ID %v", v), nil
				//return "fail", nil
			}
		}
	}
	return "pass", nil
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< Experimental/Unfinished Functions >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// More testing
// FUNCTION GetHistory2: given a key, return corresponding states from history
func (s *SmartContract) GetHistory2(ctx contractapi.TransactionContextInterface, recordID string) error {

	historyIer, err := ctx.GetStub().GetHistoryForKey(recordID)
	if err != nil {
		return fmt.Errorf("failed to get record: %v", err)
	}
	if historyIer == nil {
		return fmt.Errorf("Record does not exist")
	}

	//
	for historyIer.HasNext() {
		mod, err := historyIer.Next()
		if err != nil {
			return fmt.Errorf("Problem retrieving history: %v", err)
		}
		log.Printf("Transaction Record %v with Data: %v, status: %v", string(mod.TxId), string(mod.Value), mod.IsDelete)
	}
	return nil
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< Unused Functions >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// InitLedger: create sample part data
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	var newPart Part
	newPart.PartID = "A01"       //set PartID
	newPart.PartName = "bracket" //set PartName
	newPart.Customer = "GM"      //set Customer
	newPart.NumOfDims = 5        //set NumOfDims
	newPart.Requirements = make([]DesignData, newPart.NumOfDims)
	for i := 0; i < newPart.NumOfDims; i++ {
		target := float32(1.5)
		ucl := target + 0.02
		lcl := target - 0.02
		cpk := float32(1.67)
		newPart.Requirements[i] = DesignData{i + 1, target, ucl, lcl, cpk}
	}
	//marshal to JSON and write to worldstate
	assetJSON, err := json.Marshal(newPart)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(newPart.PartID, assetJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	return nil
}

// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<< Planned Functions >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// EditAsset: overwrite existing asset with new values
