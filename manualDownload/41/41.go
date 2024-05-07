/*
SPDX-License-Identifier: Apache-2.0
*/

package auction

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// BidReturn is the data type returned to an auction admin
type BidReturn struct {
	ID  string      `json:"id"`
	Bid *PrivateBid `json:"bid"`
}

// AskReturn is the data type returned to an auction admin
type AskReturn struct {
	ID  string      `json:"id"`
	Ask *PrivateAsk `json:"ask"`
}

const privateBidKeyType = "privateBid"
const publicBidKeyType = "publicBid"

type SmartContract struct {
	contractapi.Contract
}

// Auction Round is the structure of a bid in public state
type AuctionRound struct {
	Type     string            `json:"objectType"`
	ID       string            `json:"id"`
	Round    int               `json:"round"`
	Status   string            `json:"status"`
	ItemSold string            `json:"item"`
	Price    int               `json:"price"`
	Quantity int               `json:"quantity"`
	Sold     int               `json:"sold"`
	Demand   int               `json:"demand"`
	Sellers  map[string]Seller `json:"sellers"`
	Bidders  map[string]Bidder `json:"bidders"`
}

// PrivateBid is the structure of a bid in private state
type PrivateBid struct {
	Type     string `json:"objectType"`
	Quantity int    `json:"quantity"`
	Org      string `json:"org"`
	Buyer    string `json:"buyer"`
	Price    int    `json:"price"`
}

// Bid is the structure of a bid that will be made public
type PublicBid struct {
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

// PublicAsk is the structure of a bid in public state
type PublicAsk struct {
	Type     string `json:"objectType"`
	Quantity int    `json:"quantity"`
	Org      string `json:"org"`
	Seller   string `json:"seller"`
}

// BidAskHash is the structure of a private bid or ask in the public order book
type BidAskHash struct {
	Org  string `json:"org"`
	Hash []byte `json:"hash"`
}

// Bidder is the structure that lives on the auction
type Bidder struct {
	Buyer    string `json:"buyer"`
	Org      string `json:"org"`
	Quantity int    `json:"quantityBid"`
	Won      int    `json:"quantityWon"`
}

// Seller is the structure that lives on the auction
type Seller struct {
	Seller   string `json:"seller"`
	Org      string `json:"org"`
	Quantity int    `json:"quantity"`
	Sold     int    `json:"sold"`
	Unsold   int    `json:"unsold"`
}

// incrementAmount is the price increase of each new round of the auction
const incrementAmount = 5

const privateAskKeyType = "privateAsk"
const publicAskKeyType = "publicAsk"

// Ask is used to sell a certain item. The ask is stored in private data
// of the sellers organization, and identified by the item and transaction id
func (s *SmartContract) Ask(ctx contractapi.TransactionContextInterface, item string) (string, error) {

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}

	privateAskJSON, ok := transientMap["privateAsk"]
	if !ok {
		return "", fmt.Errorf("bid key not found in the transient map")
	}

	publicAskJSON, ok := transientMap["publicAsk"]
	if !ok {
		return "", fmt.Errorf("bid key not found in the transient map")
	}

	// get the implicit collection name using the bidder's organization ID
	collection, err := getCollectionName(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// the bidder has to target their peer to store the bid
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return "", fmt.Errorf("Cannot store bid on this peer, not a member of this org: Error %v", err)
	}

	// the transaction ID is used as a unique index for the bid
	txID := ctx.GetStub().GetTxID()

	// create a composite key using the item and transaction ID
	privateAskKey, err := ctx.GetStub().CreateCompositeKey(privateAskKeyType, []string{item, txID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the bid into the organization's implicit data collection
	err = ctx.GetStub().PutPrivateData(collection, privateAskKey, privateAskJSON)
	if err != nil {
		return "", fmt.Errorf("failed to input price into collection: %v", err)
	}

	// create a composite key using the item and transaction ID
	publicAskKey, err := ctx.GetStub().CreateCompositeKey(publicAskKeyType, []string{item, txID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the bid into the organization's implicit data collection
	err = ctx.GetStub().PutPrivateData(collection, publicAskKey, publicAskJSON)
	if err != nil {
		return "", fmt.Errorf("failed to input price into collection: %v", err)
	}

	// return the trannsaction ID so that the uset can identify their bid
	return txID, nil
}

// SubmitAsk is used to add an ask to an active auction round
func (s *SmartContract) SubmitAsk(ctx contractapi.TransactionContextInterface, auctionID string, round int, txID string) error {

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	transientAskJSON, ok := transientMap["publicAsk"]
	if !ok {
		return fmt.Errorf("bid key not found in the transient map")
	}

	auction, err := s.QueryAuctionRound(ctx, auctionID, round)
	if err != nil {
		return fmt.Errorf("Error getting auction round from state")
	}

	// create a composite key for bid using the transaction ID
	publicAskKey, err := ctx.GetStub().CreateCompositeKey(publicAskKeyType, []string{auction.ItemSold, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// Check 1: the auction needs to be open for users to add their bid
	status := auction.Status
	if status != "open" {
		return fmt.Errorf("cannot join closed or ended auction")
	}

	// check 3: check that bid has not changed on the public book
	publicAsk, err := s.QueryPublic(ctx, auction.ItemSold, publicAskKeyType, txID)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from public order book: %v", err)
	}

	collection := "_implicit_org_" + publicAsk.Org

	askHash, err := ctx.GetStub().GetPrivateDataHash(collection, publicAskKey)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from collection: %v", err)
	}
	if askHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", askHash)
	}

	hash := sha256.New()
	hash.Write(transientAskJSON)
	calculatedAskJSONHash := hash.Sum(nil)

	// verify that the hash of the passed immutable properties matches the on-chain hash
	if !bytes.Equal(calculatedAskJSONHash, askHash) {
		return fmt.Errorf("hash %x for bid JSON %s does not match hash in auction: %x",
			calculatedAskJSONHash,
			transientAskJSON,
			askHash,
		)
	}

	if !bytes.Equal(publicAsk.Hash, askHash) {
		return fmt.Errorf("Bidder has changed their bid")
	}

	var ask *PublicAsk
	err = json.Unmarshal(transientAskJSON, &ask)
	if err != nil {
		return err
	}

	// store the hash along with the sellers's organization

	newSeller := Seller{
		Seller:   ask.Seller,
		Org:      ask.Org,
		Quantity: ask.Quantity,
		Unsold:   ask.Quantity,
	}

	// add to the list of sellers
	sellers := make(map[string]Seller)
	sellers = auction.Sellers
	sellers[publicAskKey] = newSeller

	newQuantity := 0
	for _, seller := range sellers {
		newQuantity = newQuantity + seller.Quantity
	}
	auction.Quantity = newQuantity
	auction.Sellers = sellers

	// create a composite key for auction round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	newAuctionJSON, _ := json.Marshal(auction)

	// put update auction in state
	err = ctx.GetStub().PutState(auctionKey, newAuctionJSON)
	if err != nil {
		return fmt.Errorf("failed to update auction: %v", err)
	}

	return nil
}

// DeleteAsk allows the seller of the bid to delete their bid from private data
func (s *SmartContract) DeleteAsk(ctx contractapi.TransactionContextInterface, item string, txID string) error {

	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// create a composite key using the item and transaction ID
	privateAskKey, err := ctx.GetStub().CreateCompositeKey(privateAskKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// check that the owner is being deleted by the ask owner
	err = s.checkAskOwner(ctx, collection, privateAskKey)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelPrivateData(collection, privateAskKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", privateAskKey, err)
	}

	// create a composite key using the item and transaction ID
	publicAskKey, err := ctx.GetStub().CreateCompositeKey(publicAskKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().DelPrivateData(collection, publicAskKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", publicAskKey, err)
	}

	return nil
}

// NewPublicAsk adds an ask to the public order book. This ensures
// that sellers cannot change their ask during an active auction
func (s *SmartContract) NewPublicAsk(ctx contractapi.TransactionContextInterface, item string, txID string) error {

	// get the implicit collection name using the seller's organization ID
	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// create a composite key using the item and transaction ID
	askKey, err := ctx.GetStub().CreateCompositeKey(publicAskKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	hash, err := ctx.GetStub().GetPrivateDataHash(collection, askKey)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from collection: %v", err)
	}
	if hash == nil {
		return fmt.Errorf("bid hash does not exist: %s", askKey)
	}

	// get the org of the subitting bidder
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// store the hash along with the seller's organization in the public order book
	publicAsk := BidAskHash{
		Org:  clientOrgID,
		Hash: hash,
	}

	publicAskJSON, _ := json.Marshal(publicAsk)

	// put the ask hash of the bid in the public order book
	err = ctx.GetStub().PutState(askKey, publicAskJSON)
	if err != nil {
		return fmt.Errorf("failed to input bid into state: %v", err)
	}

	return nil
}

// CreateAuction creates a new auction on the public channel.
// Each round of teh auction is stored as a seperate key in the world state
func (s *SmartContract) CreateAuction(ctx contractapi.TransactionContextInterface, auctionID string, itemSold string, reservePrice int) error {

	existingAuction, err := s.QueryAuction(ctx, auctionID)
	if existingAuction != nil {
		return fmt.Errorf("Cannot create new auction: auction already exists")
	}

	sellers := make(map[string]Seller)
	bidders := make(map[string]Bidder)

	// Check if there is an ask from your org that is lower
	// than the reserve price of the first round before creating the auction
	err = checkForLowerAsk(ctx, reservePrice, itemSold, sellers)
	if err != nil {
		return fmt.Errorf("seller has lower ask, cannot open a new auction at this price: %v", err)
	}

	// Create the first round of the auction
	auction := AuctionRound{
		Type:     "auction",
		ID:       auctionID,
		Round:    0,
		Status:   "open",
		ItemSold: itemSold,
		Quantity: 0,
		Demand:   0,
		Price:    reservePrice,
		Sellers:  sellers,
		Bidders:  bidders,
	}

	auctionJSON, err := json.Marshal(auction)
	if err != nil {
		return err
	}

	// create a composite key for the auction round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(auction.Round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// put auction round into state
	err = ctx.GetStub().PutState(auctionKey, auctionJSON)
	if err != nil {
		return fmt.Errorf("failed to put auction in public data: %v", err)
	}

	// create an event to notify buyers and sellers of a new auction
	err = ctx.GetStub().SetEvent("CreateAuction", []byte(auctionID))
	if err != nil {
		return fmt.Errorf("event failed to register: %v", err)
	}

	return nil
}

// CreateNewRound creates a new round of the auction. The new round has a seperate key
// in world state. Bidders and sellers have the abiltiy to join the round at the
// new price
func (s *SmartContract) CreateNewRound(ctx contractapi.TransactionContextInterface, auctionID string, newRound int) error {

	// checks before creatin a new round

	// check 1: the round has not already been created
	auction, err := s.QueryAuctionRound(ctx, auctionID, newRound)
	if auction != nil {
		return fmt.Errorf("Cannot create new round: round already exists")
	}

	// check 2: there was there a previous round
	previousRound := newRound - 1

	auction, err = s.QueryAuctionRound(ctx, auctionID, previousRound)
	if err != nil {
		return fmt.Errorf("Cannot create round until previous round is created")
	}

	// check 3: check if the round is still active
	err = s.activeAuctionChecks(ctx, auction)
	if err != nil {
		return fmt.Errorf("Cannot close round, round and auction is still active")
	}

	// Allocate quantity sold to bids and quantity won to asks
	auction, err = s.allocateSold(ctx, auction)
	if err != nil {
		return fmt.Errorf("Error allocated quanitity sold")
	}

	// check 4: confirm that Demand >= Supply for the previous round before creating a new round
	if auction.Sold == auction.Demand {
		return fmt.Errorf("Cannot create new round: demand is not yet greater than supply")
	}

	// If all four checks have passed, create a new round

	bidders := make(map[string]Bidder)

	auction.Round = newRound
	auction.Price = auction.Price + incrementAmount
	auction.Bidders = bidders
	auction.Demand = 0

	newAuctionRoundJSON, err := json.Marshal(auction)
	if err != nil {
		return err
	}
	// create a composite key for the new round
	newAuctionRoundKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(auction.Round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().PutState(newAuctionRoundKey, newAuctionRoundJSON)
	if err != nil {
		return fmt.Errorf("failed to create new auction round: %v", err)
	}

	// create an event to notify buyers and sellers of a new round
	err = ctx.GetStub().SetEvent("CreateNewRound", []byte(auctionID))
	if err != nil {
		return fmt.Errorf("event failed to register: %v", err)
	}

	return nil
}

// CloseAuctionRound closes a given round of the auction. This prevents
// bids from being added to the auction round, signaling that auction has
// reached a steady state.
func (s *SmartContract) CloseAuctionRound(ctx contractapi.TransactionContextInterface, auctionID string, round int) error {

	auction, err := s.QueryAuctionRound(ctx, auctionID, round)
	if err != nil {
		return fmt.Errorf("Error getting auction round from state")
	}

	status := auction.Status
	if status != "open" {
		return fmt.Errorf("Can only close an open auction")
	}

	// checks confirms if the auction is still active before it can
	// be closed
	err = s.activeAuctionChecks(ctx, auction)
	if err != nil {
		return fmt.Errorf("Cannot close round, round and auction is still active")
	}

	// allocate quantity sold to bids and quantity won to asks
	auction, err = s.allocateSold(ctx, auction)
	if err != nil {
		return fmt.Errorf("Error allocated quanitity sold")
	}

	// confirm that Supply = Demand before closing the auction
	if auction.Demand > auction.Sold {
		return fmt.Errorf("Cannot create new round: demand is not equal to supply")
	}

	auction.Status = string("closed")

	closedAuction, _ := json.Marshal(auction)

	// create a composite key for the new round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the updated auction round in state
	err = ctx.GetStub().PutState(auctionKey, closedAuction)
	if err != nil {
		return fmt.Errorf("failed to close auction: %v", err)
	}

	// create an event that a round has closed
	err = ctx.GetStub().SetEvent("CloseRound", []byte(auctionID))
	if err != nil {
		return fmt.Errorf("event failed to register: %v", err)
	}

	return nil
}

// EndAuction defines the closed round as final stage of the auction.
// all other auction rounds are deleted from state.
func (s *SmartContract) EndAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {

	auction, err := s.QueryAuction(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("Error getting auction round from state")
	}

	// find if a round has been closed. If a round is closed, declare round final.
	closedRound := false
	for _, auctionRound := range auction {
		if auctionRound.Status == "closed" {
			closedRound = true
			auctionRound.Status = "final"
		}
	}

	// error if no round has been closed
	if closedRound == false {
		return fmt.Errorf("Cannot end auction. No rounds have been closed.")
	}

	// remove all open rounds
	for _, auctionRound := range auction {

		auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(auctionRound.Round)})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}

		if auctionRound.Status != "final" {
			err = ctx.GetStub().DelState(auctionKey)
			if err != nil {
				return fmt.Errorf("failed to delete auction round %v: %v", auctionKey, err)
			}
		}
	}

	// create an event that the auction has ended.
	err = ctx.GetStub().SetEvent("EndAuction", []byte(auctionID))
	if err != nil {
		return fmt.Errorf("event failed to register: %v", err)
	}

	return nil
}

//activeAuctionChecks completes a series of checks to see if the auction is still active before
// closing a round.
func (s *SmartContract) activeAuctionChecks(ctx contractapi.TransactionContextInterface, auction *AuctionRound) error {

	// check 1: check that all bids have been added to the round
	err := checkForHigherBid(ctx, auction.Price, auction.ItemSold, auction.Bidders)
	if err != nil {
		return fmt.Errorf("Cannot close auction: %v", err)
	}

	// check 2: check that all asks have been added to the round
	err = checkForLowerAsk(ctx, auction.Price, auction.ItemSold, auction.Sellers)
	if err != nil {
		return fmt.Errorf("Cannot close auction: %v", err)
	}

	return nil
}

// allocateSold allocates excess demand to sellers when new rounds are created
// or a round is sold

func (s *SmartContract) allocateSold(ctx contractapi.TransactionContextInterface, auction *AuctionRound) (*AuctionRound, error) {

	sellers := make(map[string]Seller)
	sellers = auction.Sellers

	bidders := make(map[string]Bidder)
	bidders = auction.Bidders

	previousSold := auction.Sold
	newSold := 0
	if auction.Quantity > auction.Demand {
		newSold = auction.Demand
		remainingSold := newSold - previousSold
		for bid, bidder := range bidders {
			bidder.Won = bidder.Quantity
			bidders[bid] = bidder
		}
		totalUnsold := 0
		for _, seller := range sellers {
			totalUnsold = totalUnsold + seller.Unsold
		}
		if totalUnsold > 0 && remainingSold > 0 {
			for ask, seller := range sellers {
				seller.Sold = seller.Sold + (seller.Unsold*remainingSold)/totalUnsold
				seller.Unsold = seller.Quantity - seller.Sold
				sellers[ask] = seller
			}
		}
	} else {
		for ask, seller := range sellers {
			seller.Sold = seller.Quantity
			seller.Unsold = 0
			sellers[ask] = seller
		}
		newSold = auction.Quantity
		if auction.Demand > 0 && auction.Sold > 0 {
			for bid, bidder := range bidders {
				bidder.Won = (bidder.Quantity * auction.Sold) / auction.Demand
				bidders[bid] = bidder
			}
		}
	}
	auction.Sold = newSold
	auction.Bidders = bidders
	auction.Sellers = sellers

	return auction, nil
}

// Bid is used to create a bid for a certain item. The bid is stored in the private
// data collection on the peer of the bidder's organization. The function returns
// the transaction ID so that users can identify and query their bid
func (s *SmartContract) Bid(ctx contractapi.TransactionContextInterface, item string) (string, error) {

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}

	privateBidJSON, ok := transientMap["privateBid"]
	if !ok {
		return "", fmt.Errorf("bid key not found in the transient map")
	}

	publicBidJSON, ok := transientMap["publicBid"]
	if !ok {
		return "", fmt.Errorf("bid key not found in the transient map")
	}

	// get the implicit collection name using the bidder's organization ID
	collection, err := getCollectionName(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// the bidder has to target their peer to store the bid
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return "", fmt.Errorf("Cannot store bid on this peer, not a member of this org: Error %v", err)
	}

	// the transaction ID is used as a unique index for the bid
	txID := ctx.GetStub().GetTxID()

	// create composite key for the private bid using the item and the txid
	privateBidKey, err := ctx.GetStub().CreateCompositeKey(privateBidKeyType, []string{item, txID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the bid into the organization's implicit data collection
	err = ctx.GetStub().PutPrivateData(collection, privateBidKey, privateBidJSON)
	if err != nil {
		return "", fmt.Errorf("failed to input bid into collection: %v", err)
	}

	// create composite key for the public bid using the item and the txid
	publicBidKey, err := ctx.GetStub().CreateCompositeKey(publicBidKeyType, []string{item, txID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}

	// put the bid into the organization's implicit data collection
	err = ctx.GetStub().PutPrivateData(collection, publicBidKey, publicBidJSON)
	if err != nil {
		return "", fmt.Errorf("failed to input bid into collection: %v", err)
	}

	// return the trannsaction ID so that the uset can identify their bid
	return txID, nil
}

// SubmitBid adds a bid to an auction round. If successful, updates the
// quantity demanded and the quantity won by each bid
func (s *SmartContract) SubmitBid(ctx contractapi.TransactionContextInterface, auctionID string, round int, txID string) error {

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	transientBidJSON, ok := transientMap["publicBid"]
	if !ok {
		return fmt.Errorf("bid key not found in the transient map")
	}

	auction, err := s.QueryAuctionRound(ctx, auctionID, round)
	if err != nil {
		return fmt.Errorf("Error getting auction round from state")
	}

	// create a composite key for bid using the transaction ID
	publicBidKey, err := ctx.GetStub().CreateCompositeKey(publicBidKeyType, []string{auction.ItemSold, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// Check 1: the auction needs to be open for users to add their bid
	status := auction.Status
	if status != "open" {
		return fmt.Errorf("cannot join closed or ended auction")
	}

	// Check 2: the user needs to have joined the previous auction in order to
	// add their bid
	previousRound := round - 1

	if previousRound >= 0 {

		auctionLastRound, err := s.QueryAuctionRound(ctx, auctionID, previousRound)
		if err != nil {
			return fmt.Errorf("cannot pull previous auction round from state")
		}

		previousBidders := make(map[string]Bidder)
		previousBidders = auctionLastRound.Bidders

		if _, previousBid := previousBidders[publicBidKey]; previousBid {

			//bid is in the previous auction, no action to take
		} else {
			return fmt.Errorf("bidder needs to have joined previous round")
		}
	}

	// check 3: check that bid has not changed on the public book
	publicBid, err := s.QueryPublic(ctx, auction.ItemSold, publicBidKeyType, txID)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from public order book: %v", err)
	}

	collection := "_implicit_org_" + publicBid.Org

	bidHash, err := ctx.GetStub().GetPrivateDataHash(collection, publicBidKey)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from collection: %v", err)
	}
	if bidHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", bidHash)
	}

	hash := sha256.New()
	hash.Write(transientBidJSON)
	calculatedBidJSONHash := hash.Sum(nil)

	// verify that the hash of the passed immutable properties matches the on-chain hash
	if !bytes.Equal(calculatedBidJSONHash, bidHash) {
		return fmt.Errorf("hash %x for bid JSON %s does not match hash in auction: %x",
			calculatedBidJSONHash,
			transientBidJSON,
			bidHash,
		)
	}

	if !bytes.Equal(publicBid.Hash, bidHash) {
		return fmt.Errorf("Bidder has changed their bid")
	}

	var bid *PublicBid
	err = json.Unmarshal(transientBidJSON, &bid)
	if err != nil {
		return err
	}

	// now that all checks have passed, create new bid
	newBidder := Bidder{
		Buyer:    bid.Buyer,
		Org:      bid.Org,
		Quantity: bid.Quantity,
		Won:      0,
	}

	// add the bid to the new list of bidders
	bidders := make(map[string]Bidder)
	bidders = auction.Bidders
	bidders[publicBidKey] = newBidder

	newDemand := 0
	for _, bidder := range bidders {
		newDemand = newDemand + bidder.Quantity
	}
	auction.Demand = newDemand
	auction.Bidders = bidders

	// create a composite for the auction round
	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	newAuctionJSON, _ := json.Marshal(auction)

	// put the updated auction round in state
	err = ctx.GetStub().PutState(auctionKey, newAuctionJSON)
	if err != nil {
		return fmt.Errorf("failed to update auction: %v", err)
	}

	return nil
}

// DeleteBid allows the submitter of the bid to delete their bid from the private data
// collection and from private state
func (s *SmartContract) DeleteBid(ctx contractapi.TransactionContextInterface, item string, txID string) error {

	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	publicBidKey, err := ctx.GetStub().CreateCompositeKey(publicBidKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().DelPrivateData(collection, publicBidKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", publicBidKey, err)
	}

	privateBidKey, err := ctx.GetStub().CreateCompositeKey(privateBidKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = s.checkBidOwner(ctx, collection, privateBidKey)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelPrivateData(collection, privateBidKey)
	if err != nil {
		return fmt.Errorf("failed to get bid %v: %v", privateBidKey, err)
	}

	return nil
}

// NewPublicBid adds a bid to the public order book. This ensures
// that bidders cannot change their bid during an active auction
func (s *SmartContract) NewPublicBid(ctx contractapi.TransactionContextInterface, item string, txID string) error {

	// get the implicit collection name using the bidder's organization ID
	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// create composite key for the bid using the item and the txid
	publicBidKey, err := ctx.GetStub().CreateCompositeKey(publicBidKeyType, []string{item, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	hash, err := ctx.GetStub().GetPrivateDataHash(collection, publicBidKey)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from collection: %v", err)
	}
	if hash == nil {
		return fmt.Errorf("bid hash does not exist: %s", publicBidKey)
	}

	// get the org of the subitting bidder
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// store the hash along with the seller's organization in the public order book
	publicBid := BidAskHash{
		Org:  clientOrgID,
		Hash: hash,
	}

	publicBidJSON, _ := json.Marshal(publicBid)

	// put the ask hash of the bid in the public order book
	err = ctx.GetStub().PutState(publicBidKey, publicBidJSON)
	if err != nil {
		return fmt.Errorf("failed to input bid into state: %v", err)
	}

	return nil
}

// QueryAuction allows all members of the channel to read all rounds of a public auction
func (s *SmartContract) QueryAuction(ctx contractapi.TransactionContextInterface, auctionID string) ([]*AuctionRound, error) {

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("auction", []string{auctionID})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var auctionRounds []*AuctionRound
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var auctionRound AuctionRound
		err = json.Unmarshal(queryResponse.Value, &auctionRound)
		if err != nil {
			return nil, err
		}

		auctionRounds = append(auctionRounds, &auctionRound)
	}

	return auctionRounds, nil
}

// QueryAuctionRound allows all members of the channel to read a public auction round
func (s *SmartContract) QueryAuctionRound(ctx contractapi.TransactionContextInterface, auctionID string, round int) (*AuctionRound, error) {

	auctionKey, err := ctx.GetStub().CreateCompositeKey("auction", []string{auctionID, "Round", strconv.Itoa(round)})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	auctionJSON, err := ctx.GetStub().GetState(auctionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction object %v: %v", auctionID, err)
	}
	if auctionJSON == nil {
		return nil, fmt.Errorf("auction does not exist")
	}

	var auctionRound *AuctionRound
	err = json.Unmarshal(auctionJSON, &auctionRound)
	if err != nil {
		return nil, err
	}

	return auctionRound, nil
}

// QueryBid allows the submitter of the bid or an auction admin to read their bid from private state
func (s *SmartContract) QueryBid(ctx contractapi.TransactionContextInterface, item string, txID string) (*PrivateBid, error) {

	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	bidKey, err := ctx.GetStub().CreateCompositeKey(privateBidKeyType, []string{item, txID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	// only the bid owner or the auction admin can read a bid
	err = s.checkBidOwner(ctx, collection, bidKey)
	if err != nil {
		err = ctx.GetClientIdentity().AssertAttributeValue("role", "auctionAdmin")
		if err != nil {
			return nil, fmt.Errorf("submitting client needs to be the bid owner or an auction admin")
		}
	}

	bidJSON, err := ctx.GetStub().GetPrivateData(collection, bidKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid %v: %v", bidKey, err)
	}
	if bidJSON == nil {
		return nil, fmt.Errorf("bid %v does not exist", bidKey)
	}

	var bid *PrivateBid
	err = json.Unmarshal(bidJSON, &bid)
	if err != nil {
		return nil, err
	}

	return bid, nil
}

// QueryAsk allows a seller or an auction admin to read their bid from private state
func (s *SmartContract) QueryAsk(ctx contractapi.TransactionContextInterface, item string, txID string) (*PrivateAsk, error) {

	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	askKey, err := ctx.GetStub().CreateCompositeKey(privateAskKeyType, []string{item, txID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	err = s.checkAskOwner(ctx, collection, askKey)
	if err != nil {
		err = ctx.GetClientIdentity().AssertAttributeValue("role", "auctionAdmin")
		if err != nil {
			return nil, fmt.Errorf("submitting client needs to be the ask owner or an auction admin")
		}
	}

	askJSON, err := ctx.GetStub().GetPrivateData(collection, askKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid %v: %v", askKey, err)
	}
	if askJSON == nil {
		return nil, fmt.Errorf("ask %v does not exist", askKey)
	}

	var ask *PrivateAsk
	err = json.Unmarshal(askJSON, &ask)
	if err != nil {
		return nil, err
	}

	return ask, nil
}

// QueryBids returns all bids from a private data collection for a certain item.
// this function is used by auction admins to add bids to a open auction
func (s *SmartContract) QueryBids(ctx contractapi.TransactionContextInterface, item string) ([]BidReturn, error) {

	// the function can only be used by an auction admin
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "auctionAdmin")
	if err != nil {
		return nil, fmt.Errorf("submitting client needs to be an auction admin")
	}

	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// return bids using the item
	resultsIterator, err := ctx.GetStub().GetPrivateDataByPartialCompositeKey(collection, privateBidKeyType, []string{item})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// return the bid and the transaction id, so that the bid can be submitted
	var bidReturns []BidReturn
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		_, keyParts, Err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
		if Err != nil {
			return nil, err
		}

		txID := keyParts[1]

		var bid *PrivateBid
		err = json.Unmarshal(queryResponse.Value, &bid)
		if err != nil {
			return nil, err
		}

		bidReturn := BidReturn{
			ID:  txID,
			Bid: bid,
		}

		bidReturns = append(bidReturns, bidReturn)
	}

	return bidReturns, nil
}

// QueryAsks returns all asks from a private data collection for a certain item.
// this function is used by auction admins to add asks to a open auction
func (s *SmartContract) QueryAsks(ctx contractapi.TransactionContextInterface, item string) ([]AskReturn, error) {

	err := ctx.GetClientIdentity().AssertAttributeValue("role", "auctionAdmin")
	if err != nil {
		return nil, fmt.Errorf("submitting client needs to be an auction admin")
	}

	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	collection, err := getCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// return ask using the item
	resultsIterator, err := ctx.GetStub().GetPrivateDataByPartialCompositeKey(collection, privateAskKeyType, []string{item})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// return the ask and the transaction id, so that the bid can be submitted
	var askReturns []AskReturn
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		_, keyParts, Err := ctx.GetStub().SplitCompositeKey(queryResponse.Key)
		if Err != nil {
			return nil, Err
		}

		txID := keyParts[1]

		var ask *PrivateAsk
		err = json.Unmarshal(queryResponse.Value, &ask)
		if err != nil {
			return nil, err
		}

		askReturn := AskReturn{
			ID:  txID,
			Ask: ask,
		}

		askReturns = append(askReturns, askReturn)

	}

	return askReturns, nil
}

// checkForHigherBid is an internal function that is used to determine if
// there is a higher bid that has yet to be added to an auction round
func checkForHigherBid(ctx contractapi.TransactionContextInterface, auctionPrice int, item string, bidders map[string]Bidder) error {

	// Get MSP ID of peer org
	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the peer's MSPID: %v", err)
	}

	var error error
	error = nil

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(publicBidKeyType, []string{item})
	if err != nil {
		return err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		var publicBid BidAskHash
		err = json.Unmarshal(queryResponse.Value, &publicBid)
		if err != nil {
			return err
		}

		publicBidKey := queryResponse.Key

		if _, bidInAuction := bidders[publicBidKey]; bidInAuction {

			//bid is already in the auction, no action to take

		} else {

			_, keyParts, err := ctx.GetStub().SplitCompositeKey(publicBidKey)
			if err != nil {
				return fmt.Errorf("failed to split composite key: %v", err)
			}

			privateBidKey, err := ctx.GetStub().CreateCompositeKey(privateBidKeyType, keyParts)
			if err != nil {
				return fmt.Errorf("failed to create composite key: %v", err)
			}

			collection := "_implicit_org_" + publicBid.Org

			if publicBid.Org == peerMSPID {

				bidJSON, err := ctx.GetStub().GetPrivateData(collection, privateBidKey)
				if err != nil {
					return fmt.Errorf("failed to get bid %v: %v", privateBidKey, err)
				}
				if bidJSON == nil {
					return fmt.Errorf("bid %v does not exist", privateBidKey)
				}

				var bid *PrivateBid
				err = json.Unmarshal(bidJSON, &bid)
				if err != nil {
					return err
				}

				if bid.Price >= auctionPrice {
					error = fmt.Errorf("Cannot close auction round, bidder has a higher price: %v", err)
				}

			} else {

				hash, err := ctx.GetStub().GetPrivateDataHash(collection, privateBidKey)
				if err != nil {
					return fmt.Errorf("failed to read bid hash from collection: %v", err)
				}
				if hash == nil {
					return fmt.Errorf("bid hash does not exist: %s", privateBidKey)
				}
			}
		}
	}

	return error
}

// checkForLowerAsk is an internal function that is used to determine
// is there is a lower ask that has not yet been added to the round
func checkForLowerAsk(ctx contractapi.TransactionContextInterface, auctionPrice int, item string, sellers map[string]Seller) error {

	// Get MSP ID of peer org
	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the peer's MSPID: %v", err)
	}

	var error error
	error = nil

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(publicAskKeyType, []string{item})
	if err != nil {
		return err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		var publicAsk BidAskHash
		err = json.Unmarshal(queryResponse.Value, &publicAsk)
		if err != nil {
			return err
		}

		publicAskKey := queryResponse.Key

		if _, askInAuction := sellers[publicAskKey]; askInAuction {

			//ask is already in the auction, no action to take

		} else {

			_, keyParts, err := ctx.GetStub().SplitCompositeKey(publicAskKey)
			if err != nil {
				return fmt.Errorf("failed to split composite key: %v", err)
			}

			privateAskKey, err := ctx.GetStub().CreateCompositeKey(privateAskKeyType, keyParts)
			if err != nil {
				return fmt.Errorf("failed to create composite key: %v", err)
			}

			collection := "_implicit_org_" + publicAsk.Org

			if publicAsk.Org == peerMSPID {

				askJSON, err := ctx.GetStub().GetPrivateData(collection, privateAskKey)
				if err != nil {
					return fmt.Errorf("failed to get bid %v: %v", publicAskKey, err)
				}
				if askJSON == nil {
					return fmt.Errorf("ask %v does not exist", privateAskKey)
				}

				var ask *PrivateAsk
				err = json.Unmarshal(askJSON, &ask)
				if err != nil {
					return err
				}

				if ask.Price <= auctionPrice {
					error = fmt.Errorf("Cannot close auction round, seller has a lower price: %v", err)
				}

			} else {

				hash, err := ctx.GetStub().GetPrivateDataHash(collection, privateAskKey)
				if err != nil {
					return fmt.Errorf("failed to read bid hash from collection: %v", err)
				}
				if hash == nil {
					return fmt.Errorf("bid hash does not exist: %s", privateAskKey)
				}
			}
		}
	}

	return error
}

// QueryPublic allows you to read the public hash on the order book
func (s *SmartContract) QueryPublic(ctx contractapi.TransactionContextInterface, item string, askSell string, txID string) (*BidAskHash, error) {

	bidAskKey, err := ctx.GetStub().CreateCompositeKey(askSell, []string{item, txID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	bidAskJSON, err := ctx.GetStub().GetState(bidAskKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid %v: %v", bidAskKey, err)
	}
	if bidAskJSON == nil {
		return nil, fmt.Errorf("bid or ask %v does not exist", bidAskKey)
	}

	var hash *BidAskHash
	err = json.Unmarshal(bidAskJSON, &hash)
	if err != nil {
		return nil, err
	}

	return hash, nil
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
