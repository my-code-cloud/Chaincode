// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package shim

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"unicode/utf8"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	peerpb "github.com/hyperledger/fabric-protos-go/peer"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
)

// ChaincodeStub is an object passed to chaincode for shim side handling of
// APIs.
type ChaincodeStub struct {
	TxID                       string
	ChannelID                  string
	chaincodeEvent             *pb.ChaincodeEvent
	args                       [][]byte
	handler                    *Handler
	signedProposal             *pb.SignedProposal
	proposal                   *pb.Proposal
	validationParameterMetakey string

	// Additional fields extracted from the signedProposal
	creator   []byte
	transient map[string][]byte
	binding   []byte

	decorations map[string][]byte
}

// ReadAsset reads the information from collection
func (c *Contract) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Asset, error) {

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

// ChaincodeInvocation functionality

func newChaincodeStub(handler *Handler, channelID, txid string, input *pb.ChaincodeInput, signedProposal *pb.SignedProposal) (*ChaincodeStub, error) {
	stub := &ChaincodeStub{
		TxID:                       txid,
		ChannelID:                  channelID,
		args:                       input.Args,
		handler:                    handler,
		signedProposal:             signedProposal,
		decorations:                input.Decorations,
		validationParameterMetakey: pb.MetaDataKeys_VALIDATION_PARAMETER.String(),
	}

	// TODO: sanity check: verify that every call to init with a nil
	// signedProposal is a legitimate one, meaning it is an internal call
	// to system chaincodes.
	if signedProposal != nil {
		var err error

		stub.proposal = &pb.Proposal{}
		err = proto.Unmarshal(signedProposal.ProposalBytes, stub.proposal)
		if err != nil {

			return nil, fmt.Errorf("failed to extract Proposal from SignedProposal: %s", err)
		}

		// check for header
		if len(stub.proposal.GetHeader()) == 0 {
			return nil, errors.New("failed to extract Proposal fields: proposal header is nil")
		}

		// Extract creator, transient, binding...
		hdr := &common.Header{}
		if err := proto.Unmarshal(stub.proposal.GetHeader(), hdr); err != nil {
			return nil, fmt.Errorf("failed to extract proposal header: %s", err)
		}

		// extract and validate channel header
		chdr := &common.ChannelHeader{}
		if err := proto.Unmarshal(hdr.ChannelHeader, chdr); err != nil {
			return nil, fmt.Errorf("failed to extract channel header: %s", err)
		}
		validTypes := map[common.HeaderType]bool{
			common.HeaderType_ENDORSER_TRANSACTION: true,
			common.HeaderType_CONFIG:               true,
		}
		if !validTypes[common.HeaderType(chdr.GetType())] {
			return nil, fmt.Errorf(
				"invalid channel header type. Expected %s or %s, received %s",
				common.HeaderType_ENDORSER_TRANSACTION,
				common.HeaderType_CONFIG,
				common.HeaderType(chdr.GetType()),
			)
		}

		// extract creator from signature header
		shdr := &common.SignatureHeader{}
		if err := proto.Unmarshal(hdr.GetSignatureHeader(), shdr); err != nil {
			return nil, fmt.Errorf("failed to extract signature header: %s", err)
		}
		stub.creator = shdr.GetCreator()

		// extract trasient data from proposal payload
		payload := &pb.ChaincodeProposalPayload{}
		if err := proto.Unmarshal(stub.proposal.GetPayload(), payload); err != nil {
			return nil, fmt.Errorf("failed to extract proposal payload: %s", err)
		}
		stub.transient = payload.GetTransientMap()

		// compute the proposal binding from the nonce, creator and epoch
		epoch := make([]byte, 8)
		binary.LittleEndian.PutUint64(epoch, chdr.GetEpoch())
		digest := sha256.Sum256(append(append(shdr.GetNonce(), stub.creator...), epoch...))
		stub.binding = digest[:]

	}

	return stub, nil
}

// GetTxID returns the transaction ID for the proposal
func (s *ChaincodeStub) GetTxID() string {
	return s.TxID
}

// GetChannelID returns the channel for the proposal
func (s *ChaincodeStub) GetChannelID() string {
	return s.ChannelID
}

// GetDecorations ...
func (s *ChaincodeStub) GetDecorations() map[string][]byte {
	return s.decorations
}

// GetMSPID returns the local mspid of the peer by checking the CORE_PEER_LOCALMSPID
// env var and returns an error if the env var is not set
func GetMSPID() (string, error) {
	mspid := os.Getenv("CORE_PEER_LOCALMSPID")

	if mspid == "" {
		return "", errors.New("'CORE_PEER_LOCALMSPID' is not set")
	}

	return mspid, nil
}

// ------------- Call Chaincode functions ---------------

// InvokeChaincode documentation can be found in interfaces.go
func (s *ChaincodeStub) InvokeChaincode(chaincodeName string, args [][]byte, channel string) pb.Response {
	// Internally we handle chaincode name as a composite name
	if channel != "" {
		chaincodeName = chaincodeName + "/" + channel
	}
	return s.handler.handleInvokeChaincode(chaincodeName, args, s.ChannelID, s.TxID)
}

// --------- State functions ----------

// GetState documentation can be found in interfaces.go
func (s *ChaincodeStub) GetState(key string) ([]byte, error) {
	// Access public data by setting the collection to empty string
	collection := ""
	return s.handler.handleGetState(collection, key, s.ChannelID, s.TxID)
}

// SetStateValidationParameter documentation can be found in interfaces.go
func (s *ChaincodeStub) SetStateValidationParameter(key string, ep []byte) error {
	return s.handler.handlePutStateMetadataEntry("", key, s.validationParameterMetakey, ep, s.ChannelID, s.TxID)
}

// GetStateValidationParameter documentation can be found in interfaces.go
func (s *ChaincodeStub) GetStateValidationParameter(key string) ([]byte, error) {
	md, err := s.handler.handleGetStateMetadata("", key, s.ChannelID, s.TxID)
	if err != nil {
		return nil, err
	}
	if ep, ok := md[s.validationParameterMetakey]; ok {
		return ep, nil
	}
	return nil, nil
}

// PutState documentation can be found in interfaces.go
func (s *ChaincodeStub) PutState(key string, value []byte) error {
	if key == "" {
		return errors.New("key must not be an empty string")
	}
	// Access public data by setting the collection to empty string
	collection := ""
	return s.handler.handlePutState(collection, key, value, s.ChannelID, s.TxID)
}

func (s *ChaincodeStub) createStateQueryIterator(response *pb.QueryResponse) *StateQueryIterator {
	return &StateQueryIterator{
		CommonIterator: &CommonIterator{
			handler:    s.handler,
			channelID:  s.ChannelID,
			txid:       s.TxID,
			response:   response,
			currentLoc: 0,
		},
	}
}

// GetQueryResult documentation can be found in interfaces.go
func (s *ChaincodeStub) GetQueryResult(query string) (StateQueryIteratorInterface, error) {
	// Access public data by setting the collection to empty string
	collection := ""
	// ignore QueryResponseMetadata as it is not applicable for a rich query without pagination
	iterator, _, err := s.handleGetQueryResult(collection, query, nil)

	return iterator, err
}

// DelState documentation can be found in interfaces.go
func (s *ChaincodeStub) DelState(key string) error {
	// Access public data by setting the collection to empty string
	collection := ""
	return s.handler.handleDelState(collection, key, s.ChannelID, s.TxID)
}

//  ---------  private state functions  ---------

// GetPrivateData documentation can be found in interfaces.go
func (s *ChaincodeStub) GetPrivateData(collection string, key string) ([]byte, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection must not be an empty string")
	}
	return s.handler.handleGetState(collection, key, s.ChannelID, s.TxID)
}

// GetPrivateDataHash documentation can be found in interfaces.go
func (s *ChaincodeStub) GetPrivateDataHash(collection string, key string) ([]byte, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection must not be an empty string")
	}
	return s.handler.handleGetPrivateDataHash(collection, key, s.ChannelID, s.TxID)
}

// PutPrivateData documentation can be found in interfaces.go
func (s *ChaincodeStub) PutPrivateData(collection string, key string, value []byte) error {
	if collection == "" {
		return fmt.Errorf("collection must not be an empty string")
	}
	if key == "" {
		return fmt.Errorf("key must not be an empty string")
	}
	return s.handler.handlePutState(collection, key, value, s.ChannelID, s.TxID)
}

// DelPrivateData documentation can be found in interfaces.go
func (s *ChaincodeStub) DelPrivateData(collection string, key string) error {
	if collection == "" {
		return fmt.Errorf("collection must not be an empty string")
	}
	return s.handler.handleDelState(collection, key, s.ChannelID, s.TxID)
}

// GetPrivateDataByRange documentation can be found in interfaces.go
func (s *ChaincodeStub) GetPrivateDataByRange(collection, startKey, endKey string) (StateQueryIteratorInterface, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection must not be an empty string")
	}
	if startKey == "" {
		startKey = emptyKeySubstitute
	}
	if err := validateSimpleKeys(startKey, endKey); err != nil {
		return nil, err
	}
	// ignore QueryResponseMetadata as it is not applicable for a range query without pagination
	iterator, _, err := s.handleGetStateByRange(collection, startKey, endKey, nil)

	return iterator, err
}

func (s *ChaincodeStub) createRangeKeysForPartialCompositeKey(objectType string, attributes []string) (string, string, error) {
	partialCompositeKey, err := s.CreateCompositeKey(objectType, attributes)
	if err != nil {
		return "", "", err
	}
	startKey := partialCompositeKey
	endKey := partialCompositeKey + string(maxUnicodeRuneValue)

	return startKey, endKey, nil
}

// GetPrivateDataByPartialCompositeKey documentation can be found in interfaces.go
func (s *ChaincodeStub) GetPrivateDataByPartialCompositeKey(collection, objectType string, attributes []string) (StateQueryIteratorInterface, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection must not be an empty string")
	}

	startKey, endKey, err := s.createRangeKeysForPartialCompositeKey(objectType, attributes)
	if err != nil {
		return nil, err
	}
	// ignore QueryResponseMetadata as it is not applicable for a partial composite key query without pagination
	iterator, _, err := s.handleGetStateByRange(collection, startKey, endKey, nil)

	return iterator, err
}

// GetPrivateDataQueryResult documentation can be found in interfaces.go
func (s *ChaincodeStub) GetPrivateDataQueryResult(collection, query string) (StateQueryIteratorInterface, error) {
	if collection == "" {
		return nil, fmt.Errorf("collection must not be an empty string")
	}
	// ignore QueryResponseMetadata as it is not applicable for a range query without pagination
	iterator, _, err := s.handleGetQueryResult(collection, query, nil)

	return iterator, err
}

// GetPrivateDataValidationParameter documentation can be found in interfaces.go
func (s *ChaincodeStub) GetPrivateDataValidationParameter(collection, key string) ([]byte, error) {
	md, err := s.handler.handleGetStateMetadata(collection, key, s.ChannelID, s.TxID)
	if err != nil {
		return nil, err
	}
	if ep, ok := md[s.validationParameterMetakey]; ok {
		return ep, nil
	}
	return nil, nil
}

// SetPrivateDataValidationParameter documentation can be found in interfaces.go
func (s *ChaincodeStub) SetPrivateDataValidationParameter(collection, key string, ep []byte) error {
	return s.handler.handlePutStateMetadataEntry(collection, key, s.validationParameterMetakey, ep, s.ChannelID, s.TxID)
}

// CommonIterator documentation can be found in interfaces.go
type CommonIterator struct {
	handler    *Handler
	channelID  string
	txid       string
	response   *pb.QueryResponse
	currentLoc int
}

// StateQueryIterator documentation can be found in interfaces.go
type StateQueryIterator struct {
	*CommonIterator
}

// HistoryQueryIterator documentation can be found in interfaces.go
type HistoryQueryIterator struct {
	*CommonIterator
}

// General interface for supporting different types of query results.
// Actual types differ for different queries
type queryResult interface{}

type resultType uint8

// TODO: Document constants
/*
	Constants ...
*/
const (
	StateQueryResult resultType = iota + 1
	HistoryQueryResult
)

func createQueryResponseMetadata(metadataBytes []byte) (*pb.QueryResponseMetadata, error) {
	metadata := &pb.QueryResponseMetadata{}
	err := proto.Unmarshal(metadataBytes, metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (s *ChaincodeStub) handleGetStateByRange(collection, startKey, endKey string,
	metadata []byte) (StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {

	response, err := s.handler.handleGetStateByRange(collection, startKey, endKey, metadata, s.ChannelID, s.TxID)
	if err != nil {
		return nil, nil, err
	}

	iterator := s.createStateQueryIterator(response)
	responseMetadata, err := createQueryResponseMetadata(response.Metadata)
	if err != nil {
		return nil, nil, err
	}

	return iterator, responseMetadata, nil
}

func (s *ChaincodeStub) handleGetQueryResult(collection, query string,
	metadata []byte) (StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {

	response, err := s.handler.handleGetQueryResult(collection, query, metadata, s.ChannelID, s.TxID)
	if err != nil {
		return nil, nil, err
	}

	iterator := s.createStateQueryIterator(response)
	responseMetadata, err := createQueryResponseMetadata(response.Metadata)
	if err != nil {
		return nil, nil, err
	}

	return iterator, responseMetadata, nil
}

// GetStateByRange documentation can be found in interfaces.go
func (s *ChaincodeStub) GetStateByRange(startKey, endKey string) (StateQueryIteratorInterface, error) {
	if startKey == "" {
		startKey = emptyKeySubstitute
	}
	if err := validateSimpleKeys(startKey, endKey); err != nil {
		return nil, err
	}
	collection := ""

	// ignore QueryResponseMetadata as it is not applicable for a range query without pagination
	iterator, _, err := s.handleGetStateByRange(collection, startKey, endKey, nil)

	return iterator, err
}

// GetHistoryForKey documentation can be found in interfaces.go
func (s *ChaincodeStub) GetHistoryForKey(key string) (HistoryQueryIteratorInterface, error) {
	response, err := s.handler.handleGetHistoryForKey(key, s.ChannelID, s.TxID)
	if err != nil {
		return nil, err
	}
	return &HistoryQueryIterator{CommonIterator: &CommonIterator{s.handler, s.ChannelID, s.TxID, response, 0}}, nil
}

//CreateCompositeKey documentation can be found in interfaces.go
func (s *ChaincodeStub) CreateCompositeKey(objectType string, attributes []string) (string, error) {
	return CreateCompositeKey(objectType, attributes)
}

//SplitCompositeKey documentation can be found in interfaces.go
func (s *ChaincodeStub) SplitCompositeKey(compositeKey string) (string, []string, error) {
	return splitCompositeKey(compositeKey)
}

// CreateCompositeKey ...
func CreateCompositeKey(objectType string, attributes []string) (string, error) {
	if err := validateCompositeKeyAttribute(objectType); err != nil {
		return "", err
	}
	ck := compositeKeyNamespace + objectType + string(minUnicodeRuneValue)
	for _, att := range attributes {
		if err := validateCompositeKeyAttribute(att); err != nil {
			return "", err
		}
		ck += att + string(minUnicodeRuneValue)
	}
	return ck, nil
}

func splitCompositeKey(compositeKey string) (string, []string, error) {
	componentIndex := 1
	components := []string{}
	for i := 1; i < len(compositeKey); i++ {
		if compositeKey[i] == minUnicodeRuneValue {
			components = append(components, compositeKey[componentIndex:i])
			componentIndex = i + 1
		}
	}
	return components[0], components[1:], nil
}

func validateCompositeKeyAttribute(str string) error {
	if !utf8.ValidString(str) {
		return fmt.Errorf("not a valid utf8 string: [%x]", str)
	}
	for index, runeValue := range str {
		if runeValue == minUnicodeRuneValue || runeValue == maxUnicodeRuneValue {
			return fmt.Errorf(`input contains unicode %#U starting at position [%d]. %#U and %#U are not allowed in the input attribute of a composite key`,
				runeValue, index, minUnicodeRuneValue, maxUnicodeRuneValue)
		}
	}
	return nil
}

//To ensure that simple keys do not go into composite key namespace,
//we validate simplekey to check whether the key starts with 0x00 (which
//is the namespace for compositeKey). This helps in avoding simple/composite
//key collisions.
func validateSimpleKeys(simpleKeys ...string) error {
	for _, key := range simpleKeys {
		if len(key) > 0 && key[0] == compositeKeyNamespace[0] {
			return fmt.Errorf(`first character of the key [%s] contains a null character which is not allowed`, key)
		}
	}
	return nil
}

//GetStateByPartialCompositeKey function can be invoked by a chaincode to query the
//state based on a given partial composite key. This function returns an
//iterator which can be used to iterate over all composite keys whose prefix
//matches the given partial composite key. This function should be used only for
//a partial composite key. For a full composite key, an iter with empty response
//would be returned.
func (s *ChaincodeStub) GetStateByPartialCompositeKey(objectType string, attributes []string) (StateQueryIteratorInterface, error) {
	collection := ""
	startKey, endKey, err := s.createRangeKeysForPartialCompositeKey(objectType, attributes)
	if err != nil {
		return nil, err
	}
	// ignore QueryResponseMetadata as it is not applicable for a partial composite key query without pagination
	iterator, _, err := s.handleGetStateByRange(collection, startKey, endKey, nil)

	return iterator, err
}

func createQueryMetadata(pageSize int32, bookmark string) ([]byte, error) {
	// Construct the QueryMetadata with a page size and a bookmark needed for pagination
	metadata := &pb.QueryMetadata{PageSize: pageSize, Bookmark: bookmark}
	metadataBytes, err := proto.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	return metadataBytes, nil
}

// GetStateByRangeWithPagination ...
func (s *ChaincodeStub) GetStateByRangeWithPagination(startKey, endKey string, pageSize int32,
	bookmark string) (StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {

	if startKey == "" {
		startKey = emptyKeySubstitute
	}
	if err := validateSimpleKeys(startKey, endKey); err != nil {
		return nil, nil, err
	}

	collection := ""

	metadata, err := createQueryMetadata(pageSize, bookmark)
	if err != nil {
		return nil, nil, err
	}

	return s.handleGetStateByRange(collection, startKey, endKey, metadata)
}

// GetStateByPartialCompositeKeyWithPagination ...
func (s *ChaincodeStub) GetStateByPartialCompositeKeyWithPagination(objectType string, keys []string,
	pageSize int32, bookmark string) (StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {

	collection := ""

	metadata, err := createQueryMetadata(pageSize, bookmark)
	if err != nil {
		return nil, nil, err
	}

	startKey, endKey, err := s.createRangeKeysForPartialCompositeKey(objectType, keys)
	if err != nil {
		return nil, nil, err
	}
	return s.handleGetStateByRange(collection, startKey, endKey, metadata)
}

// GetQueryResultWithPagination ...
func (s *ChaincodeStub) GetQueryResultWithPagination(query string, pageSize int32,
	bookmark string) (StateQueryIteratorInterface, *pb.QueryResponseMetadata, error) {
	// Access public data by setting the collection to empty string
	collection := ""

	metadata, err := createQueryMetadata(pageSize, bookmark)
	if err != nil {
		return nil, nil, err
	}
	return s.handleGetQueryResult(collection, query, metadata)
}

// Next ...
func (iter *StateQueryIterator) Next() (*queryresult.KV, error) {
	result, err := iter.nextResult(StateQueryResult)
	if err != nil {
		return nil, err
	}
	return result.(*queryresult.KV), err
}

// Next ...
func (iter *HistoryQueryIterator) Next() (*queryresult.KeyModification, error) {
	result, err := iter.nextResult(HistoryQueryResult)
	if err != nil {
		return nil, err
	}
	return result.(*queryresult.KeyModification), err
}

// HasNext documentation can be found in interfaces.go
func (iter *CommonIterator) HasNext() bool {
	if iter.currentLoc < len(iter.response.Results) || iter.response.HasMore {
		return true
	}
	return false
}

// getResultsFromBytes deserializes QueryResult and return either a KV struct
// or KeyModification depending on the result type (i.e., state (range/execute)
// query, history query). Note that queryResult is an empty golang
// interface that can hold values of any type.
func (iter *CommonIterator) getResultFromBytes(queryResultBytes *pb.QueryResultBytes,
	rType resultType) (queryResult, error) {

	if rType == StateQueryResult {
		stateQueryResult := &queryresult.KV{}
		if err := proto.Unmarshal(queryResultBytes.ResultBytes, stateQueryResult); err != nil {
			return nil, fmt.Errorf("error unmarshaling result from bytes: %s", err)
		}
		return stateQueryResult, nil

	} else if rType == HistoryQueryResult {
		historyQueryResult := &queryresult.KeyModification{}
		if err := proto.Unmarshal(queryResultBytes.ResultBytes, historyQueryResult); err != nil {
			return nil, err
		}
		return historyQueryResult, nil
	}
	return nil, errors.New("wrong result type")
}

func (iter *CommonIterator) fetchNextQueryResult() error {
	response, err := iter.handler.handleQueryStateNext(iter.response.Id, iter.channelID, iter.txid)
	if err != nil {
		return err
	}
	iter.currentLoc = 0
	iter.response = response
	return nil
}

// nextResult returns the next QueryResult (i.e., either a KV struct or KeyModification)
// from the state or history query iterator. Note that queryResult is an
// empty golang interface that can hold values of any type.
func (iter *CommonIterator) nextResult(rType resultType) (queryResult, error) {
	if iter.currentLoc < len(iter.response.Results) {
		// On valid access of an element from cached results
		queryResult, err := iter.getResultFromBytes(iter.response.Results[iter.currentLoc], rType)
		if err != nil {
			return nil, err
		}
		iter.currentLoc++

		if iter.currentLoc == len(iter.response.Results) && iter.response.HasMore {
			// On access of last item, pre-fetch to update HasMore flag
			if err = iter.fetchNextQueryResult(); err != nil {
				return nil, err
			}
		}

		return queryResult, err
	} else if !iter.response.HasMore {
		// On call to Next() without check of HasMore
		return nil, errors.New("no such key")
	}

	// should not fall through here
	// case: no cached results but HasMore is true.
	return nil, errors.New("invalid iterator state")
}

// Close documentation can be found in interfaces.go
func (iter *CommonIterator) Close() error {
	_, err := iter.handler.handleQueryStateClose(iter.response.Id, iter.channelID, iter.txid)
	return err
}

// GetArgs documentation can be found in interfaces.go
func (s *ChaincodeStub) GetArgs() [][]byte {
	return s.args
}

// GetStringArgs documentation can be found in interfaces.go
func (s *ChaincodeStub) GetStringArgs() []string {
	args := s.GetArgs()
	strargs := make([]string, 0, len(args))
	for _, barg := range args {
		strargs = append(strargs, string(barg))
	}
	return strargs
}

// GetFunctionAndParameters documentation can be found in interfaces.go
func (s *ChaincodeStub) GetFunctionAndParameters() (function string, params []string) {
	allargs := s.GetStringArgs()
	function = ""
	params = []string{}
	if len(allargs) >= 1 {
		function = allargs[0]
		params = allargs[1:]
	}
	return
}

// GetCreator documentation can be found in interfaces.go
func (s *ChaincodeStub) GetCreator() ([]byte, error) {
	return s.creator, nil
}

// GetTransient documentation can be found in interfaces.go
func (s *ChaincodeStub) GetTransient() (map[string][]byte, error) {
	return s.transient, nil
}

// GetBinding documentation can be found in interfaces.go
func (s *ChaincodeStub) GetBinding() ([]byte, error) {
	return s.binding, nil
}

// GetSignedProposal documentation can be found in interfaces.go
func (s *ChaincodeStub) GetSignedProposal() (*pb.SignedProposal, error) {
	return s.signedProposal, nil
}

// GetArgsSlice documentation can be found in interfaces.go
func (s *ChaincodeStub) GetArgsSlice() ([]byte, error) {
	args := s.GetArgs()
	res := []byte{}
	for _, barg := range args {
		res = append(res, barg...)
	}
	return res, nil
}

// GetTxTimestamp documentation can be found in interfaces.go
func (s *ChaincodeStub) GetTxTimestamp() (*timestamp.Timestamp, error) {
	hdr := &common.Header{}
	if err := proto.Unmarshal(s.proposal.Header, hdr); err != nil {
		return nil, fmt.Errorf("error unmarshaling Header: %s", err)
	}

	chdr := &common.ChannelHeader{}
	if err := proto.Unmarshal(hdr.ChannelHeader, chdr); err != nil {
		return nil, fmt.Errorf("error unmarshaling ChannelHeader: %s", err)
	}

	return chdr.GetTimestamp(), nil
}

// ------------- ChaincodeEvent API ----------------------

// SetEvent documentation can be found in interfaces.go
func (s *ChaincodeStub) SetEvent(name string, payload []byte) error {
	if name == "" {
		return errors.New("event name can not be empty string")
	}
	s.chaincodeEvent = &pb.ChaincodeEvent{EventName: name, Payload: payload}
	return nil
}

func toChaincodeArgs(args ...string) [][]byte {
	ccArgs := make([][]byte, len(args))
	for i, a := range args {
		ccArgs[i] = []byte(a)
	}
	return ccArgs
}

func TestNewChaincodeStub(t *testing.T) {
	expectedArgs := toChaincodeArgs("function", "arg1", "arg2")
	expectedDecorations := map[string][]byte{"decoration-key": []byte("decoration-value")}
	expectedCreator := []byte("signature-header-creator")
	expectedTransient := map[string][]byte{"key": []byte("value")}
	expectedEpoch := uint64(999)

	validSignedProposal := &peerpb.SignedProposal{
		ProposalBytes: marshalOrPanic(&peerpb.Proposal{
			Header: marshalOrPanic(&common.Header{
				ChannelHeader: marshalOrPanic(&common.ChannelHeader{
					Type:  int32(common.HeaderType_ENDORSER_TRANSACTION),
					Epoch: expectedEpoch,
				}),
				SignatureHeader: marshalOrPanic(&common.SignatureHeader{
					Creator: expectedCreator,
				}),
			}),
			Payload: marshalOrPanic(&peerpb.ChaincodeProposalPayload{
				Input:        []byte("chaincode-proposal-input"),
				TransientMap: expectedTransient,
			}),
		}),
	}

	tests := []struct {
		signedProposal *peerpb.SignedProposal
		expectedErr    string
	}{
		{signedProposal: nil},
		{signedProposal: proto.Clone(validSignedProposal).(*peerpb.SignedProposal)},
		{
			signedProposal: &peerpb.SignedProposal{ProposalBytes: []byte("garbage")},
			expectedErr:    "failed to extract Proposal from SignedProposal: proto: can't skip unknown wire type 7",
		},
		{
			signedProposal: &peerpb.SignedProposal{},
			expectedErr:    "failed to extract Proposal fields: proposal header is nil",
		},
		{
			signedProposal: &peerpb.SignedProposal{},
			expectedErr:    "failed to extract Proposal fields: proposal header is nil",
		},
		{
			signedProposal: &peerpb.SignedProposal{
				ProposalBytes: marshalOrPanic(&peerpb.Proposal{
					Header: marshalOrPanic(&common.Header{
						ChannelHeader: marshalOrPanic(&common.ChannelHeader{
							Type:  int32(common.HeaderType_CONFIG_UPDATE),
							Epoch: expectedEpoch,
						}),
					}),
				}),
			},
			expectedErr: "invalid channel header type. Expected ENDORSER_TRANSACTION or CONFIG, received CONFIG_UPDATE",
		},
	}

	for _, tt := range tests {
		stub, err := newChaincodeStub(
			&Handler{},
			"channel-id",
			"transaction-id",
			&peerpb.ChaincodeInput{Args: expectedArgs[:], Decorations: expectedDecorations},
			tt.signedProposal,
		)
		if tt.expectedErr != "" {
			assert.Error(t, err)
			assert.EqualError(t, err, tt.expectedErr)
			continue
		}
		assert.NoError(t, err)
		assert.NotNil(t, stub)

		assert.Equal(t, &Handler{}, stub.handler, "expected empty handler")
		assert.Equal(t, "channel-id", stub.ChannelID)
		assert.Equal(t, "transaction-id", stub.TxID)
		assert.Equal(t, expectedArgs, stub.args)
		assert.Equal(t, expectedDecorations, stub.decorations)
		assert.Equal(t, "VALIDATION_PARAMETER", stub.validationParameterMetakey)
		if tt.signedProposal == nil {
			assert.Nil(t, stub.proposal, "expected nil proposal")
			assert.Nil(t, stub.creator, "expected nil creator")
			assert.Nil(t, stub.transient, "expected nil transient")
			assert.Nil(t, stub.binding, "expected nil binding")
			continue
		}

		prop := &peerpb.Proposal{}
		err = proto.Unmarshal(tt.signedProposal.ProposalBytes, prop)
		assert.NoError(t, err)
		assert.Equal(t, prop, stub.proposal)

		assert.Equal(t, expectedCreator, stub.creator)
		assert.Equal(t, expectedTransient, stub.transient)

		epoch := make([]byte, 8)
		binary.LittleEndian.PutUint64(epoch, expectedEpoch)
		shdr := &common.SignatureHeader{}
		digest := sha256.Sum256(append(append(shdr.GetNonce(), expectedCreator...), epoch...))
		assert.Equal(t, digest[:], stub.binding)
	}
}

func TestChaincodeStubSetEvent(t *testing.T) {
	stub := &ChaincodeStub{}
	err := stub.SetEvent("", []byte("event payload"))
	assert.EqualError(t, err, "event name can not be empty string")
	assert.Nil(t, stub.chaincodeEvent)

	stub = &ChaincodeStub{}
	err = stub.SetEvent("name", []byte("payload"))
	assert.NoError(t, err)
	assert.Equal(t, &peerpb.ChaincodeEvent{EventName: "name", Payload: []byte("payload")}, stub.chaincodeEvent)
}

func TestChaincodeStubAccessors(t *testing.T) {
	stub := &ChaincodeStub{TxID: "transaction-id"}
	assert.Equal(t, "transaction-id", stub.GetTxID())

	stub = &ChaincodeStub{ChannelID: "channel-id"}
	assert.Equal(t, "channel-id", stub.GetChannelID())

	stub = &ChaincodeStub{decorations: map[string][]byte{"key": []byte("value")}}
	assert.Equal(t, map[string][]byte{"key": []byte("value")}, stub.GetDecorations())

	stub = &ChaincodeStub{args: [][]byte{[]byte("function"), []byte("arg1"), []byte("arg2")}}
	assert.Equal(t, [][]byte{[]byte("function"), []byte("arg1"), []byte("arg2")}, stub.GetArgs())
	assert.Equal(t, []string{"function", "arg1", "arg2"}, stub.GetStringArgs())

	f, a := stub.GetFunctionAndParameters()
	assert.Equal(t, "function", f)
	assert.Equal(t, []string{"arg1", "arg2"}, a)

	as, err := stub.GetArgsSlice()
	assert.NoError(t, err)
	assert.Equal(t, []byte("functionarg1arg2"), as)

	stub = &ChaincodeStub{}
	f, a = stub.GetFunctionAndParameters()
	assert.Equal(t, "", f)
	assert.Empty(t, a)

	stub = &ChaincodeStub{creator: []byte("creator")}
	creator, err := stub.GetCreator()
	assert.NoError(t, err)
	assert.Equal(t, []byte("creator"), creator)

	stub = &ChaincodeStub{transient: map[string][]byte{"key": []byte("value")}}
	transient, err := stub.GetTransient()
	assert.NoError(t, err)
	assert.Equal(t, map[string][]byte{"key": []byte("value")}, transient)

	stub = &ChaincodeStub{binding: []byte("binding")}
	binding, err := stub.GetBinding()
	assert.NoError(t, err)
	assert.Equal(t, []byte("binding"), binding)

	stub = &ChaincodeStub{signedProposal: &peerpb.SignedProposal{ProposalBytes: []byte("proposal-bytes")}}
	sp, err := stub.GetSignedProposal()
	assert.NoError(t, err)
	assert.Equal(t, &peerpb.SignedProposal{ProposalBytes: []byte("proposal-bytes")}, sp)
}

func TestChaincodeStubGetTxTimestamp(t *testing.T) {
	now := ptypes.TimestampNow()
	tests := []struct {
		proposal    *peerpb.Proposal
		ts          *timestamp.Timestamp
		expectedErr string
	}{
		{
			ts: now,
			proposal: &peerpb.Proposal{
				Header: marshalOrPanic(&common.Header{
					ChannelHeader: marshalOrPanic(&common.ChannelHeader{
						Timestamp: now,
					}),
				}),
			},
		},
		{
			proposal: &peerpb.Proposal{
				Header: marshalOrPanic(&common.Header{
					ChannelHeader: []byte("garbage-channel-header"),
				}),
			},
			expectedErr: "error unmarshaling ChannelHeader: proto: can't skip unknown wire type 7",
		},
		{
			proposal:    &peerpb.Proposal{Header: []byte("garbage-header")},
			expectedErr: "error unmarshaling Header: proto: can't skip unknown wire type 7",
		},
	}

	for _, tt := range tests {
		stub := &ChaincodeStub{proposal: tt.proposal}
		ts, err := stub.GetTxTimestamp()
		if tt.expectedErr != "" {
			assert.EqualError(t, err, tt.expectedErr)
			continue
		}

		assert.NoError(t, err)
		assert.True(t, proto.Equal(ts, tt.ts))
	}
}

func TestGetMSPID(t *testing.T) {
	_, err := GetMSPID()
	assert.EqualError(t, err, "'CORE_PEER_LOCALMSPID' is not set")

	os.Setenv("CORE_PEER_LOCALMSPID", "mspid")

	mspid, err := GetMSPID()
	assert.NoError(t, err)
	assert.Equal(t, "mspid", mspid)

	os.Unsetenv("CORE_PEER_LOCALMSPID")
}

func TestChaincodeStubHandlers(t *testing.T) {
	var tests = []struct {
		name     string
		resType  peerpb.ChaincodeMessage_Type
		payload  []byte
		testFunc func(*ChaincodeStub, *Handler, *testing.T, []byte)
	}{
		{
			name:    "Simple Response",
			resType: peerpb.ChaincodeMessage_RESPONSE,
			payload: []byte("myvalue"),
			testFunc: func(s *ChaincodeStub, h *Handler, t *testing.T, payload []byte) {
				resp, err := s.GetState("key")
				if err != nil {
					t.Fatalf("Unexpected error for GetState: %s", err)
				}
				assert.Equal(t, payload, resp)

				resp, err = s.GetPrivateData("col", "key")
				if err != nil {
					t.Fatalf("Unexpected error for GetState: %s", err)
				}
				assert.Equal(t, payload, resp)
				_, err = s.GetPrivateData("", "key")
				assert.EqualError(t, err, "collection must not be an empty string")

				resp, err = s.GetPrivateDataHash("col", "key")
				if err != nil {
					t.Fatalf("Unexpected error for GetPrivateDataHash: %s", err)
				}
				assert.Equal(t, payload, resp)
				_, err = s.GetPrivateDataHash("", "key")
				assert.EqualError(t, err, "collection must not be an empty string")

				err = s.PutState("key", payload)
				assert.NoError(t, err)

				err = s.PutPrivateData("col", "key", payload)
				assert.NoError(t, err)
				err = s.PutPrivateData("", "key", payload)
				assert.EqualError(t, err, "collection must not be an empty string")
				err = s.PutPrivateData("col", "", payload)
				assert.EqualError(t, err, "key must not be an empty string")

				err = s.SetStateValidationParameter("key", payload)
				assert.NoError(t, err)

				err = s.SetPrivateDataValidationParameter("col", "key", payload)
				assert.NoError(t, err)

				err = s.DelState("key")
				assert.NoError(t, err)

				err = s.DelPrivateData("col", "key")
				assert.NoError(t, err)
				err = s.DelPrivateData("", "key")
				assert.EqualError(t, err, "collection must not be an empty string")
			},
		},
		{
			name:    "ValidationParameter",
			resType: peerpb.ChaincodeMessage_RESPONSE,
			payload: marshalOrPanic(
				&peerpb.StateMetadataResult{
					Entries: []*peerpb.StateMetadata{
						{
							Metakey: "mkey",
							Value:   []byte("metavalue"),
						},
					},
				},
			),
			testFunc: func(s *ChaincodeStub, h *Handler, t *testing.T, payload []byte) {
				resp, err := s.GetStateValidationParameter("key")
				if err != nil {
					t.Fatalf("Unexpected error for GetStateValidationParameter: %s", err)
				}
				assert.Equal(t, []byte("metavalue"), resp)

				resp, err = s.GetPrivateDataValidationParameter("col", "key")
				if err != nil {
					t.Fatalf("Unexpected error for GetPrivateDataValidationParameter: %s", err)
				}
				assert.Equal(t, []byte("metavalue"), resp)
			},
		},
		{
			name:    "InvokeChaincode",
			resType: peerpb.ChaincodeMessage_RESPONSE,
			payload: marshalOrPanic(
				&peerpb.ChaincodeMessage{
					Type: peerpb.ChaincodeMessage_COMPLETED,
					Payload: marshalOrPanic(
						&peerpb.Response{
							Status:  OK,
							Payload: []byte("invokechaincode"),
						},
					),
				},
			),
			testFunc: func(s *ChaincodeStub, h *Handler, t *testing.T, payload []byte) {
				resp := s.InvokeChaincode("cc", [][]byte{}, "channel")
				assert.Equal(t, resp.Payload, []byte("invokechaincode"))
			},
		},
		{
			name:    "QueryResponse",
			resType: peerpb.ChaincodeMessage_RESPONSE,
			payload: marshalOrPanic(
				&peerpb.QueryResponse{
					Results: []*peerpb.QueryResultBytes{
						{
							ResultBytes: marshalOrPanic(
								&queryresult.KV{
									Key:   "querykey",
									Value: []byte("queryvalue"),
								},
							),
						},
					},
					Metadata: marshalOrPanic(
						&peerpb.QueryResponseMetadata{
							Bookmark:            "book",
							FetchedRecordsCount: 1,
						},
					),
					HasMore: true,
				},
			),
			testFunc: func(s *ChaincodeStub, h *Handler, t *testing.T, payload []byte) {
				expectedResult := &queryresult.KV{
					Key:   "querykey",
					Value: []byte("queryvalue"),
				}

				// stub stuff
				sqi, err := s.GetQueryResult("query")
				if err != nil {
					t.Fatalf("Unexpected error for GetQueryResult: %s", err)
				}
				kv, err := sqi.Next()
				if err != nil {
					t.Fatalf("Unexpected error for GetQueryResult: %s", err)
				}
				assert.Equal(t, expectedResult, kv)

				sqi, err = s.GetPrivateDataQueryResult("col", "query")
				if err != nil {
					t.Fatalf("Unexpected error for GetPrivateDataQueryResult: %s", err)
				}
				kv, err = sqi.Next()
				if err != nil {
					t.Fatalf("Unexpected error for GetPrivateDataQueryResult: %s", err)
				}
				assert.Equal(t, expectedResult, kv)

				_, err = s.GetPrivateDataQueryResult("", "query")
				assert.EqualError(t, err, "collection must not be an empty string")

				sqi, err = s.GetStateByRange("", "end")
				if err != nil {
					t.Fatalf("Unexpected error for GetStateByRange: %s", err)
				}
				// first result
				kv, err = sqi.Next()
				if err != nil {
					t.Fatalf("Unexpected error for GetStateByRange: %s", err)
				}
				assert.Equal(t, expectedResult, kv)
				// second result
				assert.True(t, sqi.HasNext())
				kv, err = sqi.Next()
				if err != nil {
					t.Fatalf("Unexpected error for GetStateByRange: %s", err)
				}
				assert.Equal(t, expectedResult, kv)
				err = sqi.Close()
				assert.NoError(t, err)

				sqi, qrm, err := s.GetStateByRangeWithPagination("", "end", 1, "book")
				kv, err = sqi.Next()
				if err != nil {
					t.Fatalf("Unexpected error for GetStateByRangeWithPagination: %s", err)
				}
				assert.Equal(t, expectedResult, kv)
				assert.Equal(t, "book", qrm.GetBookmark())
				assert.Equal(t, int32(1), qrm.GetFetchedRecordsCount())

				sqi, err = s.GetPrivateDataByRange("col", "", "end")
				if err != nil {
					t.Fatalf("Unexpected error for GetPrivateDataByRange: %s", err)
				}
				kv, err = sqi.Next()
				if err != nil {
					t.Fatalf("Unexpected error for GetPrivateDataByRange: %s", err)
				}
				assert.Equal(t, expectedResult, kv)

				_, err = s.GetPrivateDataByRange("", "", "end")
				assert.EqualError(t, err, "collection must not be an empty string")

				sqi, err = s.GetStateByPartialCompositeKey("object", []string{"attr1", "attr2"})
				kv, err = sqi.Next()
				if err != nil {
					t.Fatalf("Unexpected error for GetStateByPartialCompositeKey: %s", err)
				}
				assert.Equal(t, expectedResult, kv)

				sqi, err = s.GetPrivateDataByPartialCompositeKey("col", "object", []string{"attr1", "attr2"})
				kv, err = sqi.Next()
				if err != nil {
					t.Fatalf("Unexpected error for GetPrivateDataByPartialCompositeKey: %s", err)
				}
				assert.Equal(t, expectedResult, kv)

				_, err = s.GetPrivateDataByPartialCompositeKey("", "object", []string{"attr1", "attr2"})
				assert.EqualError(t, err, "collection must not be an empty string")

				sqi, qrm, err = s.GetStateByPartialCompositeKeyWithPagination(
					"object",
					[]string{"key1", "key2"},
					1,
					"book",
				)
				kv, err = sqi.Next()
				if err != nil {
					t.Fatalf("Unexpected error for GetStateByPartialCompositeKeyWithPagination: %s", err)
				}
				assert.Equal(t, expectedResult, kv)
				assert.Equal(t, "book", qrm.GetBookmark())
				assert.Equal(t, int32(1), qrm.GetFetchedRecordsCount())

				sqi, qrm, err = s.GetQueryResultWithPagination("query", 1, "book")
				kv, err = sqi.Next()
				if err != nil {
					t.Fatalf("Unexpected error forGetQueryResultWithPagination: %s", err)
				}
				assert.Equal(t, expectedResult, kv)
				assert.Equal(t, "book", qrm.GetBookmark())
				assert.Equal(t, int32(1), qrm.GetFetchedRecordsCount())
			},
		},
		{
			name:    "GetHistoryForKey",
			resType: peerpb.ChaincodeMessage_RESPONSE,
			payload: marshalOrPanic(
				&peerpb.QueryResponse{
					Results: []*peerpb.QueryResultBytes{
						{
							ResultBytes: marshalOrPanic(
								&queryresult.KeyModification{
									TxId:  "txid",
									Value: []byte("historyforkey"),
								},
							),
						},
					},
					HasMore: false,
				},
			),
			testFunc: func(s *ChaincodeStub, h *Handler, t *testing.T, payload []byte) {
				expectedResult := &queryresult.KeyModification{
					TxId:  "txid",
					Value: []byte("historyforkey"),
				}
				hqi, err := s.GetHistoryForKey("key")
				if err != nil {
					t.Fatalf("Unexpected error for GetHistoryForKey: %s", err)
				}
				km, err := hqi.Next()
				if err != nil {
					t.Fatalf("Unexpected error for GetPrivateDataByRangee: %s", err)
				}
				assert.Equal(t, expectedResult, km)
				assert.False(t, hqi.HasNext())
			},
		},
		{
			name:    "Error Conditions",
			resType: peerpb.ChaincodeMessage_ERROR,
			payload: []byte("error"),
			testFunc: func(s *ChaincodeStub, h *Handler, t *testing.T, payload []byte) {
				_, err := s.GetState("key")
				assert.EqualError(t, err, string(payload))

				_, err = s.GetPrivateDataHash("col", "key")
				assert.EqualError(t, err, string(payload))

				_, err = s.GetStateValidationParameter("key")
				assert.EqualError(t, err, string(payload))

				err = s.PutState("key", payload)
				assert.EqualError(t, err, string(payload))

				err = s.SetPrivateDataValidationParameter("col", "key", payload)
				assert.EqualError(t, err, string(payload))

				err = s.DelState("key")
				assert.EqualError(t, err, string(payload))

				_, err = s.GetStateByRange("start", "end")
				assert.EqualError(t, err, string(payload))

				_, err = s.GetQueryResult("query")
				assert.EqualError(t, err, string(payload))

				_, err = s.GetHistoryForKey("key")
				assert.EqualError(t, err, string(payload))

				resp := s.InvokeChaincode("cc", [][]byte{}, "channel")
				assert.Equal(t, payload, resp.GetPayload())

			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{
				cc:               &mockChaincode{},
				responseChannels: map[string]chan peerpb.ChaincodeMessage{},
				state:            ready,
			}
			stub := &ChaincodeStub{
				ChannelID:                  "channel",
				TxID:                       "txid",
				handler:                    handler,
				validationParameterMetakey: "mkey",
			}
			chatStream := &mock.PeerChaincodeStream{}
			chatStream.SendStub = func(msg *peerpb.ChaincodeMessage) error {
				go func() {
					handler.handleResponse(
						&peerpb.ChaincodeMessage{
							Type:      test.resType,
							ChannelId: msg.GetChannelId(),
							Txid:      msg.GetTxid(),
							Payload:   test.payload,
						},
					)
				}()
				return nil
			}
			handler.chatStream = chatStream
			test.testFunc(stub, handler, t, test.payload)
		})
	}
}
