package golang

import (
	"fmt"
	"runtime/debug"

	. "github.com/davidkhala/goutils"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type CommonChaincode struct {
	Mock    bool
	Debug   bool
	Name    string
	Channel string
	CCAPI   shim.ChaincodeStubInterface // chaincode API
}

type KeyModification struct {
	TxId      string
	Value     []byte
	Timestamp TimeLong // as unix nano
	IsDelete  bool
}

func ParseHistory(iterator shim.HistoryQueryIteratorInterface, filter func(KeyModification) bool) []KeyModification {
	defer PanicError(iterator.Close())
	var result []KeyModification
	for iterator.HasNext() {
		keyModification, err := iterator.Next()
		PanicError(err)
		var timeStamp = keyModification.Timestamp
		var t TimeLong
		t = t.FromTimeStamp(*timeStamp)
		var translated = KeyModification{
			keyModification.TxId,
			keyModification.Value,
			t,
			keyModification.IsDelete}
		if filter == nil || filter(translated) {
			result = append(result, translated)
		}
	}
	return result
}

type StateKV struct {
	Namespace string
	Key       string
	Value     string
}
type QueryResponseMetadata struct {
	FetchedRecordsCount int
	Bookmark            string
}

func ParseStates(iterator shim.StateQueryIteratorInterface, filter func(StateKV) bool) []StateKV {
	defer PanicError(iterator.Close())
	var kvs []StateKV
	for iterator.HasNext() {
		kv, err := iterator.Next()
		PanicError(err)
		var stateKV = StateKV{kv.Namespace, kv.Key, string(kv.Value)}
		if filter == nil || filter(stateKV) {
			kvs = append(kvs, stateKV)
		}
	}
	return kvs
}

type Args struct {
	buff [][]byte
}

func ArgsBuilder(fcn string) Args {
	return Args{[][]byte{[]byte(fcn)}}
}

func (t *Args) AppendBytes(bytes []byte) *Args {
	t.buff = append(t.buff, bytes)
	return t
}
func (t *Args) AppendArg(str string) *Args {
	t.buff = append(t.buff, []byte(str))
	return t
}
func (t Args) Get() [][]byte {
	return t.buff
}

// a readable structure of peer.response
type PeerResponse struct {
	// A status code that should follow the HTTP status codes.
	Status int32 `json:"status,omitempty"`
	// A message associated with the response code.
	Message string `json:"message,omitempty"`
	// A payload that can be used to include metadata with this response.
	Payload string `json:"payload,omitempty"`
}

func (cc *CommonChaincode) Prepare(ccAPI shim.ChaincodeStubInterface) {
	cc.CCAPI = ccAPI
	cc.Channel = ccAPI.GetChannelID()
}

// return empty for if no record.
func (cc CommonChaincode) GetChaincodeID() string {
	var iterator, _ = cc.GetStateByRangeWithPagination("", "", 1, "")
	if !iterator.HasNext() {
		return ""
	}
	var kv, err = iterator.Next()
	PanicError(err)
	return kv.GetNamespace()
}

func (cc CommonChaincode) GetPrivateData(collection, key string) []byte {
	var r, err = cc.CCAPI.GetPrivateData(collection, key)
	PanicError(err)
	return r
}
func (cc CommonChaincode) GetPrivateObj(collection, key string, v interface{}) bool {
	var r, err = cc.CCAPI.GetPrivateData(collection, key)
	PanicError(err)
	if r == nil {
		return false
	}
	FromJson(r, v)
	return true
}
func (cc CommonChaincode) PutPrivateObj(collection, key string, v interface{}) {
	var err = cc.CCAPI.PutPrivateData(collection, key, ToJson(v))
	PanicError(err)
}
func (cc CommonChaincode) PutPrivateData(collection, key string, value []byte) {
	var err = cc.CCAPI.PutPrivateData(collection, key, value)
	PanicError(err)
}

func (cc CommonChaincode) GetPrivateDataByPartialCompositeKey(collection, objectType string, keys []string) shim.StateQueryIteratorInterface {
	var r, err = cc.CCAPI.GetPrivateDataByPartialCompositeKey(collection, objectType, keys)
	PanicError(err)
	return r
}
func (cc CommonChaincode) GetPrivateDataByRange(collection, startKey, endKey string) shim.StateQueryIteratorInterface {
	var r, err = cc.CCAPI.GetPrivateDataByRange(collection, startKey, endKey)

	PanicError(err)
	return r
}
func (cc CommonChaincode) GetPrivateDataQueryResult(collection, query string) shim.StateQueryIteratorInterface {
	var r, err = cc.CCAPI.GetPrivateDataQueryResult(collection, query)
	PanicError(err)
	return r
}
func (cc CommonChaincode) DelPrivateData(collection, key string) {
	var err = cc.CCAPI.DelPrivateData(collection, key)
	PanicError(err)
}

// TODO is it used as getAll state starting with prefix?
func (cc CommonChaincode) GetStateRange(collection, prefix string) shim.StateQueryIteratorInterface {
	return cc.GetPrivateDataByRange(collection, prefix, prefix+"\x7f")
}

func ImplicitCollection(mspid string) string {
	return "_implicit_org_" + mspid
}

func PanicPeerResponse(resp peer.Response) {
	if resp.Status >= shim.ERRORTHRESHOLD {
		var errorPB = PeerResponse{
			resp.Status,
			resp.Message,
			string(resp.Payload),
		}
		PanicString(string(ToJson(errorPB)))
	}
}

func (cc CommonChaincode) InvokeChaincode(chaincodeName string, args [][]byte, channel string) peer.Response {
	var resp = cc.CCAPI.InvokeChaincode(chaincodeName, args, channel)
	PanicPeerResponse(resp)
	return resp
}

func (cc CommonChaincode) SplitCompositeKey(compositeKey string) (string, []string) {
	objectType, attributes, err := cc.CCAPI.SplitCompositeKey(compositeKey)
	PanicError(err)
	return objectType, attributes
}
func (cc CommonChaincode) CreateCompositeKey(objectType string, attributes []string) string {
	var key, err = cc.CCAPI.CreateCompositeKey(objectType, attributes)
	PanicError(err)
	return key
}
func (cc CommonChaincode) GetBinding() []byte {
	var result, err = cc.CCAPI.GetBinding()
	PanicError(err)
	return result
}
func (cc CommonChaincode) GetState(key string) []byte {
	var bytes, err = cc.CCAPI.GetState(key)
	PanicError(err)
	return bytes
}
func (cc CommonChaincode) GetStateObj(key string, v interface{}) bool {
	var bytes = cc.GetState(key)
	if bytes == nil {
		return false
	}
	FromJson(bytes, v)
	return true
}
func (cc CommonChaincode) GetTransient() map[string][]byte {
	transient, err := cc.CCAPI.GetTransient()
	PanicError(err)
	return transient
}
func (cc CommonChaincode) PutStateObj(key string, v interface{}) {
	var bytes = ToJson(v)
	cc.PutState(key, bytes)
}
func (cc CommonChaincode) PutState(key string, value []byte) {
	var err = cc.CCAPI.PutState(key, value)
	PanicError(err)
}
func (cc CommonChaincode) DelState(key string) {
	var err = cc.CCAPI.DelState(key)
	PanicError(err)
}
func (cc CommonChaincode) GetTxTimestamp() timestamp.Timestamp {
	ts, err := cc.CCAPI.GetTxTimestamp()
	PanicError(err)
	return *ts
}

func (cc CommonChaincode) GetHistoryForKey(key string) shim.HistoryQueryIteratorInterface {
	var r, err = cc.CCAPI.GetHistoryForKey(key)
	PanicError(err)
	return r
}
func (cc CommonChaincode) GetStateByPartialCompositeKey(objectType string, keys []string) shim.StateQueryIteratorInterface {
	var r, err = cc.CCAPI.GetStateByPartialCompositeKey(objectType, keys)
	PanicError(err)
	return r
}
func (cc CommonChaincode) GetStateByRange(startKey string, endKey string) shim.StateQueryIteratorInterface {
	var r, err = cc.CCAPI.GetStateByRange(startKey, endKey)
	PanicError(err)
	return r
}

// GetStateByPartialCompositeKeyWithPagination This call is only supported in a read only transaction.
func (cc CommonChaincode) GetStateByPartialCompositeKeyWithPagination(objectType string, keys []string, pageSize int, bookmark string) (shim.StateQueryIteratorInterface, QueryResponseMetadata) {
	var iterator, r, err = cc.CCAPI.GetStateByPartialCompositeKeyWithPagination(objectType, keys, int32(pageSize), bookmark)
	PanicError(err)
	return iterator, QueryResponseMetadata{int(r.FetchedRecordsCount), r.Bookmark}
}

// GetStateByRangeWithPagination This call is only supported in a read only transaction.
func (cc CommonChaincode) GetStateByRangeWithPagination(startKey, endKey string, pageSize int, bookmark string) (shim.StateQueryIteratorInterface, QueryResponseMetadata) {
	var iterator, r, err = cc.CCAPI.GetStateByRangeWithPagination(startKey, endKey, int32(pageSize), bookmark)
	PanicError(err)
	return iterator, QueryResponseMetadata{int(r.FetchedRecordsCount), r.Bookmark}
}

func (cc CommonChaincode) SetEvent(name string, payload []byte) {
	var err = cc.CCAPI.SetEvent(name, payload)
	PanicError(err)
}

var DeferHandlerPeerResponse = func(errString string, params ...interface{}) bool {
	var response = params[0].(*peer.Response)
	response.Status = shim.ERROR
	response.Message = errString
	response.Payload = []byte(errString)
	fmt.Println("DeferHandlerPeerResponse", errString)
	debug.PrintStack()
	return true
}

// GetMSPID From https://github.com/hyperledger/fabric-chaincode-go/commit/2d899240a7ed642a381ba9df2f6b0c303cb149dc
func GetMSPID() string {
	var mspId, err = shim.GetMSPID()
	PanicError(err)
	return mspId
}
func (cc CommonChaincode) GetFunctionAndArgs() (string, [][]byte) {
	var allArgs = cc.CCAPI.GetArgs()
	var fcn = ""
	var args = [][]byte{}
	if len(allArgs) >= 1 {
		fcn = string(allArgs[0])
		args = allArgs[1:]
	}
	return fcn, args
}
func ChaincodeStart(cc shim.Chaincode) {
	var err = shim.Start(cc)
	PanicError(err)
}
