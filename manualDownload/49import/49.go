package chaincode

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"unsafe"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/pkg/statebased"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

const syntax = "proto3"

// GetName returns the name of the contract
func (s *SmartContract) GetName() string {
	return "Practice_SmartContract"
}

// Asset describes basic details of what makes up a simple asset
type Asset struct {
	ID             string `json:"ID"`
	Color          string `json:"color"`
	Size           int    `json:"size"`
	Owner          string `json:"owner"`
	AppraisedValue int    `json:"appraisedValue"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "asset1", Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300},
		{ID: "asset2", Color: "red", Size: 5, Owner: "Brad", AppraisedValue: 400},
		{ID: "asset3", Color: "green", Size: 10, Owner: "Jin Soo", AppraisedValue: 500},
		{ID: "asset4", Color: "yellow", Size: 10, Owner: "Max", AppraisedValue: 600},
		{ID: "asset5", Color: "black", Size: 15, Owner: "Adriana", AppraisedValue: 700},
		{ID: "asset6", Color: "white", Size: 15, Owner: "Michel", AppraisedValue: 800},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	asset.Owner = newOwner
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// SomeStubMethod stub其他的无法通过mock方式测试的方法练习
func (s *SmartContract) SomeStubMethod(ctx contractapi.TransactionContextInterface, assetID string) error {
	stub := ctx.GetStub()
	// stub.GetArgs()和stub.GetStringArgs()都是获取调用链码时的入参，第一个参数时方法名，后面的参数是这个方法的参数的信息,如下：
	// 2021/01/25 08:06:32 stub.GetArgs(),i=0, arg=Practice_SmartContract:SomeStubMethod
	//2021/01/25 08:06:32 stub.GetArgs(),i=1, arg=asset1
	for i, arg := range stub.GetArgs() {
		log.Printf("stub.GetArgs(),i=%d, arg=%s", i, byteToString(arg))
	}
	for i, arg := range stub.GetStringArgs() {
		log.Printf("stub.GetStringArgs(),i=%d, arg=%s", i, arg)
	}
	binding, err := stub.GetBinding()
	if err != nil {
		return err
	}
	log.Printf("stub.GetBinding()=%s", byteToString(binding))
	for k, v := range stub.GetDecorations() {
		log.Printf("stub.GetDecorations(), k=%s, v=%s", k, byteToString(v))
	}
	// stub.GetCreator()返回的是证书，如过是组织s2.supply.com的管理员发起的交易，则此处获得的是：Admin@s2.supply.com-cert.pem
	creator, err := stub.GetCreator()
	if err != nil {
		return err
	}
	log.Printf("stub.GetCreator()=%s", byteToString(creator))
	// 已经签名的提议，包含以下内容：
	// 1.通道名称
	// 2.链码名称
	// 3.发起交易的组织名称
	// 4.发起交易的人的证书
	// 5.调用链码时的入参：方法名，参数等
	// stub.GetSignedProposal().GetProposalBytes()的信息如下：
	//2021/01/25 08:06:32 stub.GetSignedProposal().GetProposalBytes()=
	//�
	//v��������"alljoinchannel*@252b6bbd22eeaf2193cdbc86fe7bd9fa257e33a6209a5da7d81dcc41b8bb1b9d:secured_supply�
	//�
	//GylSOrg2MSP�-----BEGIN CERTIFICATE-----
	//MIICETCCAbegAwIBAgIRAJw2YUKkmyKusGHm33D7LhkwCgYIKoZIzj0EAwIwbTEL
	//MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
	//cmFuY2lzY28xFjAUBgNVBAoTDXMyLnN1cHBseS5jb20xGTAXBgNVBAMTEGNhLnMy
	//LnN1cHBseS5jb20wHhcNMjEwMTA3MDgzMTAwWhcNMzEwMTA1MDgzMTAwWjBYMQsw
	//CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
	//YW5jaXNjbzEcMBoGA1UEAwwTQWRtaW5AczIuc3VwcGx5LmNvbTBZMBMGByqGSM49
	//AgEGCCqGSM49AwEHA0IABJ6An5vHmug1YBIUXKuD50ZJ79TiwDkW5uEr2ZkXU5Em
	//XwVlxwCOKpfqKOr1Xdk0DWMlAQPQIxeXktdVBJxFc4KjTTBLMA4GA1UdDwEB/wQE
	//AwIHgDAMBgNVHRMBAf8EAjAAMCsGA1UdIwQkMCKAIGO9q5qcp089i7bDqwyxRYdg
	//aX65Bvs4X5wCsXWbxj37MAoGCCqGSM49BAMCA0gAMEUCIQCRBC/uF8ooaLQzSDo6
	//e5+4UbBqjSi5MUy3IYfVrM5tHQIgaGHKXcKZY7q0Txs6LsbtayW6kWPOAee6Z1W8
	//top2VDc=
	//-----END CERTIFICATE-----
	//�w�}dȧC>�v�@�El�S����I
	//G
	//Esecured_supply/
	//%Practice_SmartContract:SomeStubMethod
	//asset1
	proposal, err := stub.GetSignedProposal()
	if err != nil {
		return err
	}
	log.Printf("stub.GetSignedProposal()=%#v", proposal)
	bytes := proposal.GetProposalBytes()
	log.Printf("stub.GetSignedProposal().GetProposalBytes()=%s", byteToString(bytes))
	p := &peer.Proposal{}
	err = proto.Unmarshal(bytes, p)
	if err != nil {
		return err
	}
	log.Printf("stub.GetSignedProposal().GetProposalBytes(),proto.Unmarshal=%#v", p)
	//headerBytes:= p.GetHeader()
	//header := &peer.ChaincodeHeaderExtension{}
	//err = proto.Unmarshal(headerBytes, header)
	//if err != nil {
	//	return err
	//}
	//log.Printf("stub.GetSignedProposal().GetProposalBytes()-Proposal-GetHeader()=%#v", header)
	//payloadBytes := p.GetPayload()
	//payload := &peer.ChaincodeProposalPayload{}
	//err = proto.Unmarshal(payloadBytes, payload)
	//if err != nil {
	//	return err
	//}
	//log.Printf("stub.GetSignedProposal().GetProposalBytes()-Proposal-GetPayload()=%#v", payload)
	log.Printf("stub.GetSignedProposal().GetSignature()=%s", byteToString(proposal.GetSignature()))

	// 设置一个Event
	if err := stub.SetEvent("hello event", []byte("hello")); err != nil {
		return err
	}
	//2021/01/25 10:22:57 stub.GetHistoryForKey(asset1), next=&queryresult.KeyModification{
	//TxId:"f251ce5352e294cd628fc0b5d09271ebe8253b41d66069c164195fe2783c3adc",
	//Value:[]uint8{0x7b, 0x22, 0x49, 0x44, 0x22, 0x3a, 0x22, 0x61, 0x73, 0x73, 0x65, 0x74, 0x31, 0x22, 0x2c
	//, 0x22, 0x63, 0x6f, 0x6c, 0x6f, 0x72, 0x22, 0x3a, 0x22, 0x62, 0x6c, 0x75, 0x65, 0x22, 0x2c, 0x22, 0x73, 0x69, 0x7a, 0x65, 0x22, 0x3a, 0x35, 0x2c, 0x22, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x22, 0x3a, 0x22, 0x54, 0x6f, 0x6d, 0x6f, 0x6b, 0x6f, 0x22, 0x2c, 0x22, 0x61, 0x70, 0x70, 0x72, 0x61, 0x69, 0x73, 0x65, 0x64, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x33, 0x30, 0x30, 0x7d},
	//Timestamp:(*timestamp.Timestamp)(0xc00043d1a0),
	//IsDelete:false, XXX_NoUnkeyedLiteral:struct {}{},
	//XXX_unrecognized:[]uint8(nil),
	//XXX_sizecache:0}
	assetHistory, err := stub.GetHistoryForKey(assetID)
	if err != nil {
		return err
	}
	defer assetHistory.Close()
	for assetHistory.HasNext() {
		next, err := assetHistory.Next()
		if err != nil {
			return err
		}
		log.Printf("stub.GetHistoryForKey(%s), next=%#v", assetID, next)
	}

	return nil
}

func byteToString(data []byte) string {
	str := (*string)(unsafe.Pointer(&data))
	return *str
}

func (s *SmartContract) ContractPractice(ctx contractapi.TransactionContextInterface) {
	// s.GetName()是当前智能合约的名称，一个链码包中有多个智能合约，每个智能合约的名称必须不同，因此最好每个智能合约都实现这个方法来定义自己的名称
	log.Printf("s.GetName()=%s", s.GetName())
	log.Printf("s.GetInfo()=%#v", s.GetInfo())
	log.Printf("s.GetTransactionContextHandler()=%#v", s.GetTransactionContextHandler())
}

// ClientIdentityPractice ClientIdentity接口提供的方法练习
func (s *SmartContract) ClientIdentityPractice(ctx contractapi.TransactionContextInterface) error {
	log.Println("ClientIdentityPractice==================start=====================")
	clientIdentity := ctx.GetClientIdentity()
	id, err := clientIdentity.GetID()
	if err != nil {
		return err
	}
	log.Printf("clientIdentity.GetID()=%s", id)
	mspid, err := clientIdentity.GetMSPID()
	if err != nil {
		return err
	}
	log.Printf("clientIdentity.GetMSPID()=%s", mspid)
	certificate, err := clientIdentity.GetX509Certificate()
	if err != nil {
		return err
	}
	log.Printf("clientIdentity.GetX509Certificate()=%#v", certificate)
	value, found, err := clientIdentity.GetAttributeValue("test")
	if err != nil {
		return err
	}
	if found {
		log.Printf("clientIdentity.GetAttributeValue(\"test\")=%s", value)
	}

	if err := clientIdentity.AssertAttributeValue("test", "hello"); err != nil {
		log.Printf("clientIdentity.AssertAttributeValue(\"test\", \"hello\") error!")
		return err
	}

	log.Println("ClientIdentityPractice===================end======================")
	return nil
}

// GetUnknownTransaction returns the current set unknownTransaction, may be nil
func (s *SmartContract) GetUnknownTransaction() interface{} {
	return s.UnknownTransaction
}

// Default 如果不指定方法名称时指定的默认方法
func (s *SmartContract) UnknownTransaction(ctx contractapi.TransactionContextInterface) string {
	log.Printf("hello, i'm Default func！")
	return "Bye!"
}

// GetBeforeTransaction returns the current set beforeTransaction, may be nil
func (s *SmartContract) GetBeforeTransaction() interface{} {
	return s.BeforeTransaction
}

func (s *SmartContract) BeforeTransaction(ctx contractapi.TransactionContextInterface) {
	log.Printf("i'm BeforeTransaction")
}

// GetAfterTransaction returns the current set afterTransaction, may be nil
func (s *SmartContract) GetAfterTransaction() interface{} {
	return s.AfterTransaction
}

func (s *SmartContract) AfterTransaction(ctx contractapi.TransactionContextInterface) {
	log.Printf("i'm AfterTransaction")
}

func (s *SmartContract) IgnoredMe(ctx contractapi.TransactionContextInterface) {
	log.Printf("Ignored Me!")
}

func (s *SmartContract) GetIgnoredFunctions() []string {
	return []string{"IgnoredMe"}
}

func mockInitLedger(t *testing.T, stub *shimtest.MockStub) {
	assets := []Asset{
		{ID: AssetId, Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300},
		{ID: "asset2", Color: "red", Size: 5, Owner: "Brad", AppraisedValue: 400},
		{ID: "asset3", Color: "green", Size: 10, Owner: "Jin Soo", AppraisedValue: 500},
		{ID: "asset4", Color: "yellow", Size: 10, Owner: "Max", AppraisedValue: 600},
		{ID: "asset5", Color: "black", Size: 15, Owner: "Adriana", AppraisedValue: 700},
		{ID: "asset6", Color: "white", Size: 15, Owner: "Michel", AppraisedValue: 800},
	}
	stub.MockTransactionStart("test")
	putState(t, stub, assets...)
	id := stub.GetTxID()
	timestamp, err := stub.GetTxTimestamp()
	channelID := stub.GetChannelID()
	require.NoError(t, err)
	require.NotNil(t, timestamp)
	log.Printf("GetTxID()=%s, GetTxTimestamp()=%s, GetChannelID()=%s", id, timestamp, channelID)
	stub.MockTransactionEnd("test")
}

func marshal(asset Asset, t *testing.T) []byte {
	assetJSON, err := json.Marshal(asset)
	require.NoError(t, err)
	return assetJSON
}

// ChaincodeStubInterface#PutState()
func putState(t *testing.T, stub *shimtest.MockStub, assets ...Asset) {
	for _, asset := range assets {
		log.Printf("putState=%v", asset)
		require.NoError(t, stub.PutState(asset.ID, marshal(asset, t)))
	}
}

// ChaincodeStubInterface#GetState()
// ChaincodeStubInterface#PutState()
// ChaincodeStubInterface#DelState()
func getState(assetId string, t *testing.T, stub *shimtest.MockStub) {
	// 获取指定key的资产的世界状态
	state, err := stub.GetState(assetId)
	require.NoError(t, err)
	printAsset(t, state)
	newAssetID := "temp001"
	newAsset := Asset{ID: newAssetID, Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300}
	// put一个新的资产
	putState(t, stub, newAsset)
	// 查询新的资产
	newState, err := stub.GetState(newAssetID)
	require.NoError(t, err)
	require.NotNil(t, newState)
	printAsset(t, newState)
	// 指定资产ID删除资产
	require.NoError(t, stub.DelState(newAssetID))
	// 删除之后重新查询新的资产
	newStateAgain, err := stub.GetState(newAssetID)
	require.NoError(t, err)
	require.Nil(t, newStateAgain)
}

func getHistoryForKey(assetId string, t *testing.T, stub *shimtest.MockStub) {
	// 获取key的历史数据，目前mock还未实现
	history, err := stub.GetHistoryForKey(assetId)
	require.NoError(t, err)
	if history != nil {
		if history.HasNext() {
			next, err := history.Next()
			require.NoError(t, err)
			marshal, err := json.Marshal(next)
			require.NoError(t, err)
			log.Printf("asset=%s history=%s", assetId, marshal)
		}
		history.Close()
	}
}

func printAsset(t *testing.T, state []byte) {
	var a Asset
	require.NoError(t, json.Unmarshal(state, &a))
	marshal, err := json.Marshal(a)
	require.NoError(t, err)
	log.Printf("result state json value = %s", marshal)
}

// ChaincodeStubInterface#GetArgs()
// ChaincodeStubInterface#GetStringArgs()
func getArgs(t *testing.T, stub *shimtest.MockStub) {
	args := stub.GetArgs()
	for _, arg := range args {
		log.Printf("stub.GetArgs(), %s", byteToString(arg))
	}

	stringArgs := stub.GetStringArgs()
	for _, argString := range stringArgs {
		log.Print(argString)
	}
}

// ChaincodeStubInterface#GetStateByRange(startKey, endKey string) (StateQueryIteratorInterface, error)
// ChaincodeStubInterface#GetStateByRangeWithPagination(startKey, endKey string, pageSize int32, bookmark string) (StateQueryIteratorInterface, *pb.QueryResponseMetadata, error)
func getStateByRange(t *testing.T, stub *shimtest.MockStub) {
	// GetStateByRange不指定startKey和endKey，会返回全部的资产；谨慎使用
	states, err := stub.GetStateByRange("", "")
	require.NoError(t, err)
	printStateQueryIteratorInterface(t, states)
	// GetStateByRangeWithPagination 因为mockStub直接返回三个nil，所以无法在mock环境测试
	pagination, metadata, err := stub.GetStateByRangeWithPagination("", "", 5, "")
	require.NoError(t, err)
	log.Print("==========================================================================================")
	log.Printf("GetStateByRangeWithPagination metadata=%v", metadata)
	printStateQueryIteratorInterface(t, pagination)
}

func printStateQueryIteratorInterface(t *testing.T, states shim.StateQueryIteratorInterface) {
	if states != nil {
		for states.HasNext() {
			next, err := states.Next()
			require.NoError(t, err)
			log.Print(next)
		}
		states.Close()
	}
}

// ChaincodeStubInterface#CreateCompositeKey(objectType string, attributes []string) (string, error)
// ChaincodeStubInterface#SplitCompositeKey(compositeKey string) (string, []string, error)
func createCompositeKey(t *testing.T, stub *shimtest.MockStub) {
	objectType := "test"
	attributes := []string{"param1", "param2", "param3", "end"}
	// 创建组合key，拼接了一下
	key, err := stub.CreateCompositeKey(objectType, attributes)
	require.NoError(t, err)
	log.Printf("key=%s", key)
	// 分割组合key，CreateCompositeKey的逆运算
	compositeKey, strings, err := stub.SplitCompositeKey(key)
	require.Equal(t, objectType, compositeKey)
	require.Equal(t, attributes, strings)
	newAsset := Asset{ID: key, Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300}
	putState(t, stub, newAsset)
	empty := []string{}
	// 根据创建组合key的参数查询，后面的参数可以是空，这样会全部匹配出来
	states, err := stub.GetStateByPartialCompositeKey(objectType, empty)
	require.NoError(t, err)
	require.NotNil(t, states)
	printStateQueryIteratorInterface(t, states)
}

const (
	AssetId        string = "asset1"
	TestMSP        string = "TestMSP"
	TestCollection string = "private_TestMSP"
	Blank          string = ""
)

// ChaincodeStubInterface#SetStateValidationParameter(key string, ep []byte) error
// ChaincodeStubInterface#GetStateValidationParameter(key string) ([]byte, error)
func setStateValidationParameter(t *testing.T, stub *shimtest.MockStub) {
	// 新建一个基于状态的背书策略
	endorsementPolicy, err := statebased.NewStateEP(nil)
	require.NoError(t, err)
	// 向背书策略添加需要背书的公司
	require.NoError(t, endorsementPolicy.AddOrgs(statebased.RoleTypeMember, TestMSP))
	policy, err := endorsementPolicy.Policy()
	require.NoError(t, err)
	// SetStateValidationParameter设置基于状态的背书策略
	require.NoError(t, stub.SetStateValidationParameter(AssetId, policy))
	// GetStateValidationParameter获取基于状态的背书策略
	parameter, err := stub.GetStateValidationParameter(AssetId)
	require.NoError(t, err)
	str := byteToString(parameter)
	// 打印出来的StateValidationParameter有特殊字符，所以使用包含传入的字符的方式断言
	log.Printf("ID=%s, StateValidationParameter=%s", AssetId, str)
	require.Contains(t, str, TestMSP)
}

// ChaincodeStubInterface#GetPrivateData(collection, key string) ([]byte, error)
// ChaincodeStubInterface#GetPrivateDataHash(collection, key string) ([]byte, error) 获取私有数据的hash值，这个方法就算不是私有数据的所有者也可以调用，mock版本没有实现；
// ChaincodeStubInterface#DelPrivateData(collection, key string) error 删除私有数据，mock版本没有实现；
// ChaincodeStubInterface#SetPrivateDataValidationParameter(collection, key string, ep []byte) error 设置私有数据的
// ChaincodeStubInterface#GetPrivateDataValidationParameter(collection, key string) ([]byte, error)
// ChaincodeStubInterface#GetPrivateDataByRange(collection, startKey, endKey string) (StateQueryIteratorInterface, error) 根据范围查询私有数据
// ChaincodeStubInterface#GetPrivateDataByPartialCompositeKey(collection, objectType string, keys []string) (StateQueryIteratorInterface, error)
func getPrivateData(t *testing.T, stub *shimtest.MockStub) {
	key := "private001"
	privateAsset := Asset{ID: key, Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300}
	bytes, err := json.Marshal(privateAsset)
	require.NoError(t, err)
	// 添加私有数据
	require.NoError(t, stub.PutPrivateData(TestCollection, key, bytes))
	// 获取私有资产
	data, err := stub.GetPrivateData(TestCollection, key)
	require.NoError(t, err)
	require.NotNil(t, data)
	printAsset(t, data)
	// 使用不存在的其他的collection获取私有资产，不会返回error，会返回nil数据
	data, err = stub.GetPrivateData("test_collections", key)
	require.NoError(t, err)
	require.Nil(t, data)
	// 使用其他的key获取不存在私有资产
	data, err = stub.GetPrivateData(TestCollection, AssetId)
	require.NoError(t, err)
	require.Nil(t, data)
	// 查询公共资产数据,断言没有这个资产
	state, err := stub.GetState(key)
	require.NoError(t, err)
	require.Nil(t, state)

	endorsementPolicy, err := statebased.NewStateEP(nil)
	require.NoError(t, err)
	require.NoError(t, endorsementPolicy.AddOrgs(statebased.RoleTypeMember, TestMSP))
	policy, err := endorsementPolicy.Policy()
	require.NoError(t, err)
	require.NoError(t, stub.SetPrivateDataValidationParameter(TestCollection, key, policy))
	parameter, err := stub.GetPrivateDataValidationParameter(TestCollection, key)
	require.NoError(t, err)
	str := byteToString(parameter)
	// 打印出来的StateValidationParameter有特殊字符，所以使用包含传入的字符的方式断言
	log.Printf("ID=%s, StateValidationParameter=%s", AssetId, str)
	require.Contains(t, str, TestMSP)
	// GetPrivateDataHash(collection, key string) ([]byte, error) 获取私有数据的hash值，这个方法就算不是私有数据的所有者也可以调用，mock版本没有实现；
	// DelPrivateData(collection, key string) error 删除私有数据，mock版本没有实现；
	//require.NoError(t, stub.DelPrivateData(TestCollection, key))
	//// 删除之后再次查询，断言已经没有此资产
	//data, err = stub.GetPrivateData(TestCollection, key)
	//require.NoError(t, err)
	//require.Nil(t, state)
	// GetPrivateDataByRange没有实现
	//byRange, err := stub.GetPrivateDataByRange(TestCollection, Blank, Blank)
	//require.NoError(t, err)
	//require.NotNil(t, byRange)
	//for byRange.HasNext() {
	//	next, err := byRange.Next()
	//	require.NotNil(t, err)
	//	log.Print(next)
	//}
}

// ChaincodeStubInterface#ChaincodeStubInterface#GetCreator() ([]byte, error) 获取签约交易提议的人，签约提议的人也是这个交易的创建者; mockstub未实现
// ChaincodeStubInterface#GetTransient() (map[string][]byte, error) 获取临时数据，这个方法只有设置了临时数据的peer才能查到数据，主要是为了做隐私保护的，详情参考隐秘的交易资产
// ChaincodeStubInterface#GetBinding() ([]byte, error) TODO 待理解
// ChaincodeStubInterface#GetDecorations() ([]byte, error) TODO 待理解,目前看是为了传递更多关于提议的的额外数据
// ChaincodeStubInterface#GetSignedProposal() ([]byte, error) 获取提议; mockstub未实现
// ChaincodeStubInterface#SetEvent(name string, payload []byte) error  允许链码在提议的response设置一个事件。无论交易的有效性如何，事件都将在已提交的块中的交易内可用。一个交易只能包含一个事件，并且如果是链码调用另一个链码的情况，事件只能在最外层。
func stubOthers(t *testing.T, stub *shimtest.MockStub) {
	m := make(map[string][]byte)
	tempAsset := Asset{ID: "temp001", Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300}
	m["temp_asset"], _ = json.Marshal(tempAsset)
	require.NoError(t, stub.SetTransient(m))
	transient, err := stub.GetTransient()
	require.NoError(t, err)
	require.NotNil(t, transient)
	for k, v := range transient {
		log.Printf("k=%s, v=%s", k, v)
	}
	decorations := stub.GetDecorations()
	for k, v := range decorations {
		log.Printf("k=%s, v=%s", k, v)
	}
}

// 测试shim.ChaincodeStubInterface接口
func stubTest(t *testing.T, stub *shimtest.MockStub) {
	assetId := AssetId
	mockInitLedger(t, stub)
	stub.MockTransactionStart("test1")
	getState(assetId, t, stub)
	//getHistoryForKey(assetId, t, stub)
	getArgs(t, stub)
	stub.MockTransactionStart("test1")
	getStateByRange(t, stub)
	createCompositeKey(t, stub)
	setStateValidationParameter(t, stub)
	getPrivateData(t, stub)
	stubOthers(t, stub)
}

// 测试contractapi.Contract的方法
func contractTest(t *testing.T, ccc *contractapi.ContractChaincode, stub *shimtest.MockStub) {
	log.Printf("DefaultContract=%s", ccc.DefaultContract)
	info := ccc.Info
	log.Printf("info=%v", info)
	stub.MockTransactionStart("contract_test")
	// 如果调用一个不存在的方法，如果实现了GetUnknownTransaction接口，则会执行此接口返回的方法；否则不执行，并且也不会报错，但是如果有before方法是会执行的
	response := stub.MockInvoke("uuid_002", [][]byte{[]byte("Unknow")})
	log.Printf("response=%#v, response.Status=%d, response.Payload=%s", response, response.Status, byteToString(response.Payload))
	// 调用一个被忽略的方法, 虽然IgnoredMe方法在智能合约中存在，但是因为合约满足IgnoreContractInterface接口然后把这个方法加入到了忽略列表中，所以最后还是调用的默认方法
	response = stub.MockInvoke("uuid_002", [][]byte{[]byte("IgnoredMe")})
	log.Printf("response=%#v, response.Status=%d, response.Payload=%s", response, response.Status, byteToString(response.Payload))
	// 指定某个指定合约，调用一个不存在的方法，冒号前面的部分是智能合约名称，后面是方法名称
	response = stub.MockInvoke("uuid_002", [][]byte{[]byte("TestSmartContract:Unknow")})
	log.Printf("response=%#v, response.Status=%d, response.Payload=%s", response, response.Status, byteToString(response.Payload))
	//invoke := ccc.Invoke(stub)
	//log.Printf("response=%v", invoke)
	stub.MockTransactionEnd("uuid_001")
	transactionSerializer := ccc.TransactionSerializer
	log.Printf("transactionSerializer=%v", transactionSerializer)
}

// 测试入口
func TestStart(t *testing.T) {
	// 一个链码包中可以有多个智能合约
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{}, &TestSmartContract{})
	require.NoError(t, err)
	// NewMockStub
	stub := shimtest.NewMockStub("mockSub", assetChaincode)
	stubTest(t, stub)
	contractTest(t, assetChaincode, stub)
}

type TestSmartContract struct {
	contractapi.Contract
}

// GetUnknownTransaction returns the current set unknownTransaction, may be nil
func (t *TestSmartContract) GetUnknownTransaction() interface{} {
	return t.UnknownTransaction
}

// Default 如果不指定方法名称时指定的默认方法
func (t *TestSmartContract) UnknownTransaction(ctx contractapi.TransactionContextInterface) string {
	log.Printf("hello, i'm Default func in TestSmartContract！")
	return "i'm TestSmartContract, Bye!"
}
