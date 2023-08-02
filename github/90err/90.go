/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ipfsdatastore

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

// ChaincodeStubWrapper implements the github.com/ipfs/go-datastore.Batching interface.
// This implementation uses a chaincode stub to store and retrieve data.
type ChaincodeStubWrapper struct {
	*dataStore
	stub       shim.ChaincodeStubInterface
	collection string
}

var logger = flogging.MustGetLogger("ext_offledger")

type dataStore struct {
}

// Has returns whether the `key` is mapped to a `value`.
// This function is only called just before a Put is called.
// It is cheaper to always return false and let the 'Put' go through
// than to ask for the data from other peers before each Put.
func (s *dataStore) Has(key datastore.Key) (bool, error) {
	return false, nil
}

// GetSize returns the size of the `value` named by `key`.
// Note: This function is never called for the DCAS so we'll leave
// it unimplemented. If it is ever called in the future then
// an implementation will need to be provided.
func (s *dataStore) GetSize(key datastore.Key) (int, error) {
	panic("not implemented")
}

// Query searches the datastore and returns a query result.
// Note: This function is never called for the DCAS so we'll leave
// it unimplemented. If it is ever called in the future then
// an implementation will need to be provided.
func (s *dataStore) Query(q query.Query) (query.Results, error) {
	panic("not implemented")
}

// Sync does nothing.
func (s *dataStore) Sync(prefix datastore.Key) error {
	// No-op
	return nil
}

// Close does nothing
func (s *dataStore) Close() error {
	// No-op
	return nil
}

// Batch is not supported
func (s *dataStore) Batch() (datastore.Batch, error) {
	// Not supported
	return nil, datastore.ErrBatchUnsupported
}

// Delete removes the value for given `key`.
// Note: This function is never called for the DCAS so we'll leave
// it unimplemented. If it is ever called in the future then
// an implementation will need to be provided.
func (s *dataStore) Delete(key datastore.Key) error {
	panic("not implemented")
}

// NewStubWrapper returns a stub-wrapping IPFS data store
func NewStubWrapper(coll string, stub shim.ChaincodeStubInterface) *ChaincodeStubWrapper {
	return &ChaincodeStubWrapper{
		stub:       stub,
		collection: coll,
	}
}

// Get retrieves the object `value` named by `key`.
// Get will return ErrNotFound if the key is not mapped to a value.
func (s *ChaincodeStubWrapper) Get(key datastore.Key) ([]byte, error) {
	logger.Debugf("Getting key %s", key)

	v, err := s.stub.GetPrivateData(s.collection, key.String())
	if err != nil {
		return nil, err
	}

	if len(v) == 0 {
		return nil, datastore.ErrNotFound
	}

	return v, nil
}

// Put stores the object `value` named by `key`.
func (s *ChaincodeStubWrapper) Put(key datastore.Key, value []byte) error {
	logger.Debugf("Putting key %s, Value: %s", key, value)

	return s.stub.PutPrivateData(s.collection, key.String(), value)
}
