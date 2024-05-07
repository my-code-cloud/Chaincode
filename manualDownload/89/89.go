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
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
)

type SmartContract struct {
	contractapi.Contract
}

// Auction data
type Auction struct {
	Type         string             `json:"objectType"`
	ItemSold     string             `json:"item"`
	Seller       string             `json:"seller"`
	Quantity     int                `json:"quantity"`
	Orgs         []string           `json:"organizations"`
	PrivateBids  map[string]BidHash `json:"privateBids"`
	RevealedBids map[string]FullBid `json:"revealedBids"`
	Winners      []Winners          `json:"winners"`
	Price        int                `json:"price"`
	Status       string             `json:"status"`
	Auditor      bool               `json:"auditor"`
}

// FullBid is the structure of a revealed bid
type FullBid struct {
	Type     string `json:"objectType"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"`
	Org      string `json:"org"`
	Buyer    string `json:"buyer"`
}

// BidHash is the structure of a private bid
type BidHash struct {
	Org  string `json:"org"`
	Hash string `json:"hash"`
}

// Winners stores the winners of the auction
type Winners struct {
	Buyer    string `json:"buyer"`
	Quantity int    `json:"quantity"`
}

const bidKeyType = "bid"

// SubmitBid is used by the bidder to add the hash of that bid stored in private data to the
// auction. Note that this function alters the auction in private state, and needs
// to meet the auction endorsement policy. Transaction ID is used identify the bid
func (s *SmartContract) SubmitBid(ctx contractapi.TransactionContextInterface, auctionID string, txID string) error {

	// get the MSP ID of the bidder's org
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// get the auction from public state
	auction, err := s.QueryAuction(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("failed to get auction from public state %v", err)
	}

	// the auction needs to be open for users to add their bid
	status := auction.Status
	if status != "open" {
		return fmt.Errorf("cannot join closed or ended auction")
	}

	// get the inplicit collection name of bidder's org
	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// use the transaction ID passed as a parameter to create composite bid key
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{auctionID, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// get the hash of the bid if found in private collection
	bidHash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
	if err != nil {
		return fmt.Errorf("failed to read bid bash from collection: %v", err)
	}
	if bidHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", bidKey)
	}

	// store the hash along with the bidder's organization
	newHash := BidHash{
		Org:  clientOrgID,
		Hash: fmt.Sprintf("%x", bidHash),
	}

	bidders := make(map[string]BidHash)
	bidders = auction.PrivateBids
	bidders[bidKey] = newHash
	auction.PrivateBids = bidders

	// Add the bidding organization to the list of participating organization's if it is not already
	orgs := auction.Orgs
	if !(contains(orgs, clientOrgID)) {
		newOrgs := append(orgs, clientOrgID)
		auction.Orgs = newOrgs

		err = setAssetStateBasedEndorsement(ctx, auctionID, newOrgs, auction.Auditor)
		if err != nil {
			return fmt.Errorf("failed setting state based endorsement for new organization: %v", err)
		}
	}

	newAuctionJSON, _ := json.Marshal(auction)

	err = ctx.GetStub().PutState(auctionID, newAuctionJSON)
	if err != nil {
		return fmt.Errorf("failed to update auction: %v", err)
	}

	return nil
}

// RevealBid is used by a bidder to reveal their bid after the auction is closed
func (s *SmartContract) RevealBid(ctx contractapi.TransactionContextInterface, auctionID string, txID string) error {

	// get the MSP ID of the bidder's org
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	transientBidJSON, ok := transientMap["bid"]
	if !ok {
		return fmt.Errorf("bid key not found in the transient map")
	}

	// get implicit collection name of organization ID
	collection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to get implicit collection name: %v", err)
	}

	// use transaction ID to create composit bid key
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{auctionID, txID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// get bid hash of bid if private bid on the public ledger
	bidHash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
	if err != nil {
		return fmt.Errorf("failed to read bid bash from collection: %v", err)
	}
	if bidHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", bidKey)
	}

	// get auction from public state
	auction, err := s.QueryAuction(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("failed to get auction from public state %v", err)
	}

	// check that the bidders org is a participant in the auction
	orgs := auction.Orgs
	if !(contains(orgs, clientOrgID)) {
		return fmt.Errorf("Particiant is not a member of the auction", err)
	}

	// Complete a series of three checks before we add the bid to the auction

	// check 1: check that the auction is closed. We cannot reveal an
	// bid to an open auction
	status := auction.Status
	if status != "closed" {
		return fmt.Errorf("cannot reveal bid for open or ended auction")
	}

	// check 2: check that hash of revealed bid matches hash of private bid
	// on the public ledger. This checks that the bidder is telling the truth
	// about the value of their bid

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

	// check 3; check hash of relealed bid matches hash of private bid that was
	// added earlier. This ensures that the bid has not changed since it
	// was added to the auction

	bidders := auction.PrivateBids
	privateBidHashString := bidders[bidKey].Hash

	onChainBidHashString := fmt.Sprintf("%x", bidHash)
	if privateBidHashString != onChainBidHashString {
		return fmt.Errorf("hash %s for bid JSON %s does not match hash in auction: %s, bidder must have changed bid",
			privateBidHashString,
			transientBidJSON,
			onChainBidHashString,
		)
	}

	// we can add the bid to the auction if all checks have passed
	type transientBidInput struct {
		Quantity int    `json:"quantity"`
		Price    int    `json:"price"`
		Org      string `json:"org"`
		Buyer    string `json:"buyer"`
	}

	// unmarshal bid imput
	var bidInput transientBidInput
	err = json.Unmarshal(transientBidJSON, &bidInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// get ID of submitting client
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	// marshal transient parameters and ID and MSPID into bid object
	newBid := FullBid{
		Type:     bidKeyType,
		Quantity: bidInput.Quantity,
		Price:    bidInput.Price,
		Org:      bidInput.Org,
		Buyer:    bidInput.Buyer,
	}

	// check 4: make sure that the transaction is being submitted is the bidder
	if bidInput.Buyer != clientID {
		return fmt.Errorf("Permission denied, client id %v is not the owner of the bid", clientID)
	}

	revealedBids := make(map[string]FullBid)
	revealedBids = auction.RevealedBids
	revealedBids[bidKey] = newBid
	auction.RevealedBids = revealedBids

	auctionJSON, _ := json.Marshal(auction)

	// put auction with bid added back into state
	err = ctx.GetStub().PutState(auctionID, auctionJSON)
	if err != nil {
		return fmt.Errorf("failed to update auction: %v", err)
	}

	return nil
}

// CloseAuction can be used by the seller to close the auction. This prevents
// bids from being added to the auction, and allows users to reveal their bid
func (s *SmartContract) CloseAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {

	// get the MSP ID of the bidder's org
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// get auction from public state
	auction, err := s.QueryAuction(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("failed to get auction from public state %v", err)
	}

	// check that the bidders org is a participant in the auction
	orgs := auction.Orgs
	if !(contains(orgs, clientOrgID)) {
		return fmt.Errorf("Particiant is not a member of the auction", err)
	}

	// the auction can only be closed by the seller

	// get ID of submitting client
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	seller := auction.Seller
	if seller != clientID {
		return fmt.Errorf("auction can only be closed by seller: %v", err)
	}

	status := auction.Status
	if status != "open" {
		return fmt.Errorf("cannot close auction that is not open")
	}

	auction.Status = string("closed")

	closedAuctionJSON, _ := json.Marshal(auction)

	err = ctx.GetStub().PutState(auctionID, closedAuctionJSON)
	if err != nil {
		return fmt.Errorf("failed to close auction: %v", err)
	}

	return nil
}

// EndAuction both changes the auction status to closed and calculates the winners
// of the auction
func (s *SmartContract) EndAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {

	// get the MSP ID of the bidder's org
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// get auction from public state
	auction, err := s.QueryAuction(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("failed to get auction from public state %v", err)
	}

	// check that the bidders org is a participant in the auction
	orgs := auction.Orgs
	if !(contains(orgs, clientOrgID)) {
		return fmt.Errorf("Particiant is not a member of the auction", err)
	}

	// Check that the auction is being ended by the seller

	// get ID of submitting client
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	seller := auction.Seller
	if seller != clientID {
		return fmt.Errorf("auction can only be ended by seller: %v", err)
	}

	status := auction.Status
	if status != "closed" {
		return fmt.Errorf("Can only end a closed auction")
	}

	// get the list of revealed bids

	revealedBidMap := auction.RevealedBids
	if len(auction.RevealedBids) == 0 {
		return fmt.Errorf("No bids have been revealed, cannot end auction: %v", err)
	}

	// sort the map of revealed bids to make it easier to calculate winners
	// if bids are tied, fill smaller bids first
	var bidders []FullBid

	for _, bid := range revealedBidMap {
		bidders = append(bidders, bid)
	}

	sort.Slice(bidders, func(p, q int) bool {
		if bidders[p].Price > bidders[q].Price {
			return true
		}
		if bidders[p].Price < bidders[q].Price {
			return false
		}
		return bidders[p].Quantity < bidders[q].Quantity
	})

	i := 0
	remainingQuantity := auction.Quantity

	// calculate the winners
	for remainingQuantity > 0 {

		// create the next winning bid
		winner := Winners{
			Buyer:    bidders[i].Buyer,
			Quantity: bidders[i].Quantity,
		}

		// add them to the list of winners and change the winning price
		auction.Winners = append(auction.Winners, winner)
		auction.Price = bidders[i].Price

		// Calculate the quantity that goes to the winner
		// if there is sufficient quantity to give them the full bid
		if remainingQuantity > bidders[i].Quantity {
			remainingQuantity = remainingQuantity - bidders[i].Quantity

			// if there is not, give the remainder
		} else {
			auction.Winners[i].Quantity = remainingQuantity
			remainingQuantity = 0
		}
		i++
		if i == len(bidders) {
			remainingQuantity = 0
		}
	}

	// check if there is a winning bid that has yet to be revealed
	err = checkForHigherBid(ctx, auction.Price, auction.RevealedBids, auction.PrivateBids)
	if err != nil {
		return fmt.Errorf("Cannot end auction: %v", err)
	}

	auction.Status = string("ended")

	endedAuctionJSON, _ := json.Marshal(auction)

	err = ctx.GetStub().PutState(auctionID, endedAuctionJSON)
	if err != nil {
		return fmt.Errorf("failed to end auction: %v", err)
	}
	return nil
}

// QueryAuction allows all members of the channel to read a public auction
func (s *SmartContract) QueryAuction(ctx contractapi.TransactionContextInterface, auctionID string) (*Auction, error) {

	auctionJSON, err := ctx.GetStub().GetState(auctionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction object %v: %v", auctionID, err)
	}
	if auctionJSON == nil {
		return nil, fmt.Errorf("auction does not exist")
	}

	var auction *Auction
	err = json.Unmarshal(auctionJSON, &auction)
	if err != nil {
		return nil, err
	}

	return auction, nil
}

// checkForHigherBid is an internal function that is used to determine if a winning bid has yet to be revealed
func checkForHigherBid(ctx contractapi.TransactionContextInterface, auctionPrice int, revealedBidders map[string]FullBid, bidders map[string]BidHash) error {

	// Get MSP ID of peer org
	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the peer's MSPID: %v", err)
	}

	var error error
	error = nil

	for bidKey, privateBid := range bidders {

		if _, bidInAuction := revealedBidders[bidKey]; bidInAuction {

			//bid is already revealed, no action to take

		} else {

			collection := "_implicit_org_" + privateBid.Org

			if privateBid.Org == peerMSPID {

				bidJSON, err := ctx.GetStub().GetPrivateData(collection, bidKey)
				if err != nil {
					return fmt.Errorf("failed to get bid %v: %v", bidKey, err)
				}
				if bidJSON == nil {
					return fmt.Errorf("bid %v does not exist", bidKey)
				}

				var bid *FullBid
				err = json.Unmarshal(bidJSON, &bid)
				if err != nil {
					return err
				}

				if bid.Price > auctionPrice {
					error = fmt.Errorf("Cannot close auction, bidder has a higher price: %v", err)
				}

			} else {

				hash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
				if err != nil {
					return fmt.Errorf("failed to read bid hash from collection: %v", err)
				}
				if hash == nil {
					return fmt.Errorf("bid hash does not exist: %s", bidKey)
				}
			}
		}
	}

	return error
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

func contains(sli []string, str string) bool {
	for _, a := range sli {
		if a == str {
			return true
		}
	}
	return false
}

func setAssetStateBasedEndorsement(ctx contractapi.TransactionContextInterface, assetId string, mspids []string, auditor bool) error {

	principals := make([]*msp.MSPPrincipal, len(mspids))
	participantSigsPolicy := make([]*common.SignaturePolicy, len(mspids))

	for i, id := range mspids {
		principal, err := proto.Marshal(
			&msp.MSPRole{
				Role:          msp.MSPRole_PEER,
				MspIdentifier: id,
			},
		)
		if err != nil {
			return err
		}
		principals[i] = &msp.MSPPrincipal{
			PrincipalClassification: msp.MSPPrincipal_ROLE,
			Principal:               principal,
		}
		participantSigsPolicy[i] = &common.SignaturePolicy{
			Type: &common.SignaturePolicy_SignedBy{
				SignedBy: int32(i),
			},
		}
	}

	if auditor == false {
		// create the defalt policy for an auction without an auditor

		policy := &common.SignaturePolicyEnvelope{
			Version: 0,
			Rule: &common.SignaturePolicy{
				Type: &common.SignaturePolicy_NOutOf_{
					NOutOf: &common.SignaturePolicy_NOutOf{
						N:     int32(len(mspids)),
						Rules: participantSigsPolicy,
					},
				},
			},
			Identities: principals,
		}

		spBytes, err := proto.Marshal(policy)
		if err != nil {
			return err
		}
		err = ctx.GetStub().SetStateValidationParameter(assetId, spBytes)
		if err != nil {
			return fmt.Errorf("failed to set validation parameter on auction: %v", err)
		}
	} else {

		// create the defalt policy for an auction with an auditor

		// create the auditor identity and signature policy
		auditorMSP, err := proto.Marshal(
			&msp.MSPRole{
				Role:          msp.MSPRole_PEER,
				MspIdentifier: "Org3MSP",
			},
		)
		if err != nil {
			return err
		}
		principals = append(principals, &msp.MSPPrincipal{
			PrincipalClassification: msp.MSPPrincipal_ROLE,
			Principal:               auditorMSP,
		},
		)
		// Create the policies in case the auditor is needed. In this case, an
		// auditor and 1 participant can update the auction.
		auditorPolicies := make([]*common.SignaturePolicy, 2)
		auditorPolicies[0] = &common.SignaturePolicy{
			Type: &common.SignaturePolicy_SignedBy{
				SignedBy: int32(len(principals) - 1),
			},
		}
		auditorPolicies[1] = &common.SignaturePolicy{
			Type: &common.SignaturePolicy_NOutOf_{
				NOutOf: &common.SignaturePolicy_NOutOf{
					N:     1,
					Rules: participantSigsPolicy,
				},
			},
		}

		// The auditor policy below is equivilent to AND(auditor, OR(participants))
		policies := make([]*common.SignaturePolicy, 2)
		policies[0] = &common.SignaturePolicy{
			Type: &common.SignaturePolicy_NOutOf_{
				NOutOf: &common.SignaturePolicy_NOutOf{
					N:     2,
					Rules: auditorPolicies,
				},
			},
		}
		// Participants can also update the auction without an auditor
		policies[1] = &common.SignaturePolicy{
			Type: &common.SignaturePolicy_NOutOf_{
				NOutOf: &common.SignaturePolicy_NOutOf{
					N:     int32(len(mspids)),
					Rules: participantSigsPolicy,
				},
			},
		}
		// Either the auditor policy or the participant policy can update
		// the auction
		policy := &common.SignaturePolicyEnvelope{
			Version: 0,
			Rule: &common.SignaturePolicy{
				Type: &common.SignaturePolicy_NOutOf_{
					NOutOf: &common.SignaturePolicy_NOutOf{
						N:     1,
						Rules: policies,
					},
				},
			},
			Identities: principals,
		}
		spBytes, err := proto.Marshal(policy)
		if err != nil {
			return err
		}
		err = ctx.GetStub().SetStateValidationParameter(assetId, spBytes)
		if err != nil {
			return fmt.Errorf("failed to set validation parameter on auction: %v", err)
		}
	}
	return nil
}
