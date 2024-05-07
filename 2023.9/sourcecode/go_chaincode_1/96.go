package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

type SmartContract struct {
	contractapi.Contract
}

// Response 统一消息返回
type Response struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Commodity 商品信息
type Commodity struct {
	//作物ID
	CropsID string `json:"crops_id"`
	//商品名称
	CommodityName string `json:"commodity_name"`
	//销售商ID
	SellerID string `json:"seller_id"`
	//价格
	Price string `json:"price"`
	//销售状态
	State string `json:"state"`
}

// CreateCommodity 上架商品
func (s *SmartContract) CreateCommodity(ctx contractapi.TransactionContextInterface, CropsID, CommodityName, SellerID, Price, State string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	ok, err := IsExist(ctx, CropsID)
	if err != nil {
		resp.Code = "106"
		resp.Msg = err.Error()
		return &resp
	}
	if ok {
		resp.Code = "300"
		resp.Msg = fmt.Sprintf("crops_id already exist")
		return &resp
	}
	commodity := Commodity{CropsID, CommodityName, SellerID, Price, State}
	err = WriteLedger(ctx, commodity, CropsID)
	if err != nil {
		resp.Code = "101"
		resp.Msg = err.Error()
		return &resp
	}
	return &resp
}

// ChangeCommodity 修改状态
func (s *SmartContract) ChangeCommodity(ctx contractapi.TransactionContextInterface, CropsID, CommodityName, SellerID, Price, State string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	commodity := Commodity{CropsID, CommodityName, SellerID, Price, State}
	_, err := GetState0(ctx, CropsID)
	if err != nil {
		resp.Code = "102"
		resp.Msg = err.Error()
		return &resp
	}
	err = WriteLedger(ctx, commodity, CropsID)
	if err != nil {
		resp.Code = "101"
		resp.Msg = err.Error()
		return &resp
	}
	return &resp
}

// DelCommodity 删除商品
func (s *SmartContract) DelCommodity(ctx contractapi.TransactionContextInterface, CropsID string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	_, err := GetState0(ctx, CropsID)
	if err != nil {
		resp.Code = "102"
		resp.Msg = err.Error()
		return &resp
	}
	err = ctx.GetStub().DelState(CropsID)
	if err != nil {
		resp.Code = "104"
		resp.Msg = err.Error()
		return &resp
	}
	return &resp
}

// QueryAllBySellerID 根据销售商ID查询所有商品
func (s *SmartContract) QueryAllBySellerID(ctx contractapi.TransactionContextInterface, SellerID string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	var err error
	resp.Data, err = GetState1(ctx, "seller_id", SellerID)
	if err != nil {
		resp.Code = "103"
		resp.Msg = err.Error()
		return &resp
	}
	return &resp
}

// QueryAllByName 根据商品名称查询商品
func (s *SmartContract) QueryAllByName(ctx contractapi.TransactionContextInterface, CommodityName string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	var err error
	resp.Data, err = GetState1(ctx, "commodity_name", CommodityName)
	if err != nil {
		resp.Code = "103"
		resp.Msg = err.Error()
		return &resp
	}
	return &resp
}

func main() {
	Chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating chaincode: %v", err)
	}
	if err = Chaincode.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}

// WriteLedger 写入账本	传入 ctx, 结构体, 键值
func WriteLedger(ctx contractapi.TransactionContextInterface, obj interface{}, key string) error {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return errors.New(fmt.Sprintf("json error: %s", err))
	}
	//写入区块链账本
	if err = ctx.GetStub().PutState(key, jsonBytes); err != nil {
		return errors.New(fmt.Sprintf("Error writing to blockchain ledger: %s", err))
	}
	return nil
}

// GetState0 根据键查询,一般查询	传入 ctx, 键值
func GetState0(ctx contractapi.TransactionContextInterface, key string) ([]byte, error) {
	result, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Data acquisition error: %s", err))
	}
	if result == nil {
		return nil, errors.New(fmt.Sprintf("The Key of the query does not exist: %s", key))
	}
	return result, nil
}

// GetState1 富查询数据,针对单一键查询	传入 ctx, 查询json格式, 键值
func GetState1(ctx contractapi.TransactionContextInterface, keyName string, key string) ([]byte, error) {
	queryString := fmt.Sprintf(`{"selector":{"%s":"%s"}}`, keyName, key)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Data acquisition error: %s", err))
	}
	if !resultsIterator.HasNext() {
		return nil, errors.New(fmt.Sprintf("The Key of the query does not exist: %s", key))
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	bArrayMemberAlreadyWritten := false
	buffer.WriteString(`{"result":[`)

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next() //获取迭代器中的每一个值
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Data acquisition from iterator error: %s", err))
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString(string(queryResponse.Value)) //将查询结果放入Buffer中
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString(`]}`)

	return buffer.Bytes(), nil
}

// IsExist 根据键查询,一般查询	传入 ctx, 键值
func IsExist(ctx contractapi.TransactionContextInterface, key string) (bool, error) {
	result, err := ctx.GetStub().GetState(key)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Data acquisition error: %s", err))
	}
	if result == nil {
		return false, nil
	}
	return true, nil
}
