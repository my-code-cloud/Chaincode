/*
 *  Copyright 2017 - 2019 KB Kontrakt LLC - All Rights Reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */
package tests

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	gomock "github.com/golang/mock/gomock"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	queryresult "github.com/hyperledger/fabric-protos-go/ledger/queryresult"
)

// MockStateQueryIteratorInterface is a mock of StateQueryIteratorInterface interface
type MockStateQueryIteratorInterface struct {
	ctrl     *gomock.Controller
	recorder *MockStateQueryIteratorInterfaceMockRecorder
}

// MockChaincodeStubInterface is a mock of shim.ChaincodeStubInterface interface
type MockChaincodeStubInterface struct {
	ctrl     *gomock.Controller
	recorder *MockChaincodeStubInterfaceMockRecorder
}

// MockChaincodeStubInterfaceMockRecorder is the mock recorder for MockChaincodeStubInterface
type MockChaincodeStubInterfaceMockRecorder struct {
	mock *MockChaincodeStubInterface
}

// MockStateQueryIteratorInterfaceMockRecorder is the mock recorder for MockStateQueryIteratorInterface
type MockStateQueryIteratorInterfaceMockRecorder struct {
	mock *MockStateQueryIteratorInterface
}

// MockCommonIteratorInterface is a mock of CommonIteratorInterface interface
type MockCommonIteratorInterface struct {
	ctrl     *gomock.Controller
	recorder *MockCommonIteratorInterfaceMockRecorder
}

// MockCommonIteratorInterfaceMockRecorder is the mock recorder for MockCommonIteratorInterface
type MockCommonIteratorInterfaceMockRecorder struct {
	mock *MockCommonIteratorInterface
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStateQueryIteratorInterface) EXPECT() *MockStateQueryIteratorInterfaceMockRecorder {
	return m.recorder
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCommonIteratorInterface) EXPECT() *MockCommonIteratorInterfaceMockRecorder {
	return m.recorder
}

// NewSMC .
func NewSMC(t *testing.T, ccname string) (*MockChaincodeStubInterface, *shimtest.MockStub, *gomock.Controller) {
	ctrl := gomock.NewController(t)

	mock := shimtest.NewMockStub(ccname, nil)
	mock.MockTransactionStart("TX1")

	return NewMockChaincodeStubInterface(ctrl), mock, ctrl
}

// GomockRexp .
func GomockRexp(rexp string) gomock.Matcher {
	return gomockRexpMatcher{regexp.MustCompile(rexp)}
}

type gomockRexpMatcher struct {
	rexp *regexp.Regexp
}

func (e gomockRexpMatcher) Matches(x interface{}) bool {
	return e.rexp.MatchString(fmt.Sprintf("%+v", x))
}

func (e gomockRexpMatcher) String() string {
	return fmt.Sprintf("is %v regexp matched", e.rexp)
}

func getSerializedValueObject(val interface{}) []byte {
	var err error
	var data []byte

	switch val := val.(type) {
	case []byte:
		data = val
	default:
		data, err = json.Marshal(&val)
		if err != nil {
			panic(err)
		}
	}

	return data
}

func jsonPreFillByKeyValPairs(callback func(string, []byte), keyValPairs ...interface{}) {
	for inx, l := 0, len(keyValPairs); inx < l; inx += 2 {
		keyName := keyValPairs[inx].(string)
		data := getSerializedValueObject(keyValPairs[inx+1])
		callback(keyName, data)
	}
}

// JSONFillMockState .
func JSONFillMockState(mock *shimtest.MockStub, keyValPairs ...interface{}) {
	jsonPreFillByKeyValPairs(func(key string, data []byte) {
		mock.PutState(key, data)
	}, keyValPairs...)
}

// JSONFillMockPrvState .
func JSONFillMockPrvState(mock *shimtest.MockStub, collectionName string, keyValPairs ...interface{}) {
	jsonPreFillByKeyValPairs(func(key string, data []byte) {
		mock.PutPrivateData(collectionName, key, data)
	}, keyValPairs...)
}

// WrapShimMockGetPrivState .
func WrapShimMockGetPrivState(mock *shimtest.MockStub) interface{} {
	return func(collection, key string) ([]byte, error) {
		return mock.GetPrivateData(collection, key)
	}
}

// WrapShimMockPutPrivState .
func WrapShimMockPutPrivState(mock *shimtest.MockStub) interface{} {
	return func(collection, key string, data []byte) error {
		return mock.PutPrivateData(collection, key, data)
	}
}

// WrapShimMockDelPrivState .
func WrapShimMockDelPrivState(mock *shimtest.MockStub) interface{} {
	return func(collection, key string, data []byte) error {
		return mock.DelPrivateData(collection, key)
	}
}

// WrapShimMockGetState .
func WrapShimMockGetState(mock *shimtest.MockStub) interface{} {
	return func(key string) ([]byte, error) {
		return mock.GetState(key)
	}
}

// WrapShimMockPutState .
func WrapShimMockPutState(mock *shimtest.MockStub) interface{} {
	return func(key string, data []byte) error {
		return mock.PutState(key, data)
	}
}

// WrapShimMockDelState .
func WrapShimMockDelState(mock *shimtest.MockStub) interface{} {
	return func(key string, data []byte) error {
		return mock.DelState(key)
	}
}

// WrapShimMockTxStamp .
func WrapShimMockTxStamp(mock *shimtest.MockStub) interface{} {
	return func() (*timestamp.Timestamp, error) {
		return mock.GetTxTimestamp()
	}
}

// WrapShimMockTxID .
func WrapShimMockTxID(mock *shimtest.MockStub) interface{} {
	return func() string {
		return mock.GetTxID()
	}
}

// MakeJSONStateIteratorFuncs .
func MakeJSONStateIteratorFuncs(namespace string, keyValPairs ...interface{}) (interface{}, interface{}) {
	inx, l := 0, len(keyValPairs)

	return func() (*queryresult.KV, error) {
			inx += 2
			if inx <= l {
				keyName := keyValPairs[inx-2].(string)
				return &queryresult.KV{
					Namespace: namespace,
					Key:       keyName,
					Value:     getSerializedValueObject(keyValPairs[inx-1]),
				}, nil
			}
			return nil, nil
		}, func() bool {
			if inx >= l {
				return false
			}
			return true
		}
}

// MakeJSONStateIterator .
func MakeJSONStateIterator(ctrl *gomock.Controller, namespace string,
	keyValPairs ...interface{}) *MockStateQueryIteratorInterface {

	nextTimes := 1
	if len(keyValPairs) == 0 {
		nextTimes = 0
	}

	iter, hasNext := MakeJSONStateIteratorFuncs(namespace, keyValPairs...)

	queryMock := NewMockStateQueryIteratorInterface(ctrl)
	queryMock.EXPECT().Next().DoAndReturn(iter).MinTimes(nextTimes)
	queryMock.EXPECT().HasNext().DoAndReturn(hasNext).MinTimes(1)
	queryMock.EXPECT().Close().MinTimes(1)

	return queryMock
}

// NewMockStateQueryIteratorInterface creates a new mock instance
func NewMockStateQueryIteratorInterface(ctrl *gomock.Controller) *MockStateQueryIteratorInterface {
	mock := &MockStateQueryIteratorInterface{ctrl: ctrl}
	mock.recorder = &MockStateQueryIteratorInterfaceMockRecorder{mock}
	return mock
}

// NewMockChaincodeStubInterface creates a new mock instance
func NewMockChaincodeStubInterface(ctrl *gomock.Controller) *MockChaincodeStubInterface {
	mock := &MockChaincodeStubInterface{ctrl: ctrl}
	mock.recorder = &MockChaincodeStubInterfaceMockRecorder{mock}
	return mock
}

// Close indicates an expected call of Close
func (mr *MockStateQueryIteratorInterfaceMockRecorder) Close() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStateQueryIteratorInterface)(nil).Close))
}

// HasNext indicates an expected call of HasNext
func (mr *MockStateQueryIteratorInterfaceMockRecorder) HasNext() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasNext", reflect.TypeOf((*MockStateQueryIteratorInterface)(nil).HasNext))
}

// Next indicates an expected call of Next
func (mr *MockStateQueryIteratorInterfaceMockRecorder) Next() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Next", reflect.TypeOf((*MockStateQueryIteratorInterface)(nil).Next))
}

// HasNext mocks base method
func (m *MockStateQueryIteratorInterface) HasNext() bool {
	ret := m.ctrl.Call(m, "HasNext")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Close mocks base method
func (m *MockStateQueryIteratorInterface) Close() error {
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Next mocks base method
func (m *MockStateQueryIteratorInterface) Next() (*queryresult.KV, error) {
	ret := m.ctrl.Call(m, "Next")
	ret0, _ := ret[0].(*queryresult.KV)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}
