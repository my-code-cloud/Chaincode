/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// QueryParcel returns the parcel stored in the world state with given id.
func (s *SmartContract) QueryParcel(ctx contractapi.TransactionContextInterface, parcelID string) (*Parcel, error) {
	// create a composite key using the transaction ID and order id to query the order private details
	parcelKey, err := ctx.GetStub().CreateCompositeKey(parcelKeyType, []string{parcelID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	parcelJSON, err := ctx.GetStub().GetPrivateData(parcelCollection, parcelKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get parcel object: %v", err)
	}
	if parcelJSON == nil {
		return nil, fmt.Errorf("the parcel %s does not exist", parcelKey)
	}

	var parcel *Parcel
	err = json.Unmarshal(parcelJSON, &parcel)
	if err != nil {
		return nil, err
	}

	return parcel, nil
}

// QueryOrder allows all members of the channel to read a public order
func (s *SmartContract) QueryOrder(ctx contractapi.TransactionContextInterface, orderID string) (*ShippingOrder, error) {

	orderJSON, err := ctx.GetStub().GetState(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order object %v: %v", orderID, err)
	}
	if orderJSON == nil {
		return nil, fmt.Errorf("order does not exist")
	}

	var order *ShippingOrder
	err = json.Unmarshal(orderJSON, &order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

//Query all orders by couriers
// func (s *SmartContract) QueryAllOrders(ctx contractapi.TransactionContext) ([]*ShippingOrder, error) {
// 	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resultsIterator.Close()

// 	var orders []*ShippingOrder
// 	for resultsIterator.HasNext() {
// 		queryResponse, err := resultsIterator.Next()
// 		if err != nil {
// 			return nil, err
// 		}
// 		var order ShippingOrder
// 		err = json.Unmarshal(queryResponse.Value, &order)
// 		if err != nil {
// 			return nil, err
// 		}

// 		orders = append(orders, &order)
// 	}
// 	return orders, nil
// }

func (s *SmartContract) QueryOrderByState(ctx contractapi.TransactionContextInterface, state string) ([]*ShippingOrder, error) {
	queryString := fmt.Sprintf(`{"selector":{"objectType":"shippingOrder","bidState":"%s"}}`, state)
	return getQueryResultForQueryString(ctx, queryString)

}

// func (s *SmartContract) GetOrdersForQuery(ctx contractapi.TransactionContextInterface, queryString string) ([]ShippingOrder, error) {

// 	queryResults, err := s.getQueryResultForQueryString(ctx, queryString)

// 	if err != nil {
// 		return nil, fmt.Errorf("Failed to read from ----world state. %s", err.Error())
// 	}

// 	return queryResults, nil

// }

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*ShippingOrder, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var results []*ShippingOrder

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var order ShippingOrder

		err = json.Unmarshal(response.Value, &order)
		if err != nil {
			return nil, err
		}

		results = append(results, &order)
	}
	return results, nil
}

// QueryOrderPrivateProperties returns the immutable parcel properties from owner's private data collection
func (s *SmartContract) QueryOrderPrivateDetails(ctx contractapi.TransactionContextInterface, orderID string, orderTxID string) (*ShippingOrderPrivateDetails, error) {
	// create a composite key using the transaction ID and order id to verify that submitting client is the seller
	orderKey, err := ctx.GetStub().CreateCompositeKey(orderKeyType, []string{orderID, orderTxID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}
	orderPrivateDetailsJSON, err := ctx.GetStub().GetPrivateData(orderCollection, orderKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read order private details: %v", err)
	}
	if orderPrivateDetailsJSON == nil {
		return nil, fmt.Errorf("Order Private Details for %v does not exist in collection %v", orderKey, orderCollection)
	}

	var orderPrivateDetails *ShippingOrderPrivateDetails
	err = json.Unmarshal(orderPrivateDetailsJSON, &orderPrivateDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return orderPrivateDetails, nil
}

// QueryBid allows the courier to read their bid from organization implicit collection
func (s *SmartContract) QueryBid(ctx contractapi.TransactionContextInterface, orderID string, bidTxID string) (*FullBid, error) {

	collection, err := getClientImplicitCollectionNameAndVerifyClientOrg(ctx)
	if err != nil {
		return nil, err
	}
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client identity %v", err)
	}

	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{orderID, bidTxID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	bidJSON, err := ctx.GetStub().GetPrivateData(collection, bidKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get bid %v: %v", bidKey, err)
	}
	if bidJSON == nil {
		return nil, fmt.Errorf("bid %v does not exist", bidKey)
	}

	var bid *FullBid
	err = json.Unmarshal(bidJSON, &bid)
	if err != nil {
		return nil, err
	}

	// check that the client querying the bid is the bid owner
	if bid.Courier != clientID {
		return nil, fmt.Errorf("Permission denied, client id %v is not the owner of the bid", clientID)
	}

	return bid, nil
}

// checkForLowestBid is an internal function that is used to determine if a winning bid has yet to be revealed
func checkForLowestBid(ctx contractapi.TransactionContextInterface, orderPrice int, revealedBids map[string]FullBid, bidders map[string]BidHash) error {
	// Get MSP ID of peer org
	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the peer's MSPID: %v", err)
	}

	var error error
	error = nil

	for bidKey, privateBid := range bidders {

		if _, bidInOrder := revealedBids[bidKey]; bidInOrder {

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

				if bid.Price < orderPrice {
					error = fmt.Errorf("Cannot assign a courier for the order, bidder has a lower price: %v", err)
				}

			} else {

				Hash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
				if err != nil {
					return fmt.Errorf("failed to read bid hash from collection: %v", err)
				}
				if Hash == nil {
					return fmt.Errorf("bid hash does not exist: %s", bidKey)
				}
			}
		}
	}

	return error
}
