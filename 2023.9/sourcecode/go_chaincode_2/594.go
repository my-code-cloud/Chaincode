/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2019-12-14 15:48 
# @File : config_service.go
# @Description : 业务service的基类
# @Attention : 
*/
package cc

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"myLibrary/go-library/blockchain/constants"
	error3 "myLibrary/go-library/chaincode/error"
	"myLibrary/go-library/common/blockchain/base"
	error2 "myLibrary/go-library/common/error"
	"myLibrary/go-library/go/base/service"
	"myLibrary/go-library/go/converters"
	"net/http"
	"strconv"
)

type IChainCodeLogicServiceHelper interface {
	Encrypt(bytes []byte, version base.Version) ([]byte, error2.IBaseError)
	Decrypt(bytes []byte, version base.Version) ([]byte, error2.IBaseError)
	GetConcreteKey(stub shim.ChaincodeStubInterface, key base.ObjectType, args ...interface{}) (string, error2.IBaseError)
}

type BaseChainCodeBaseLogicServiceImpl struct {
	Stub shim.ChaincodeStubInterface
	// 版本信息
	Version uint64
	// 交易的主旨信息
	BaseTransactionType base.TransBaseTypeV2

	ChainCodeHelper IChainCodeLogicServiceHelper
	// IBaseChainCode
	service.IBaseService
}

// // BeforeStart :方法开始调用
// func (receiver *BaseServiceImpl) BeforeStart(method string) {
// 	receiver.MethodName = method
// 	methodName := receiver.BaseInitConifg.GetLogger().GetPrefix() + " -> " + method
// 	receiver.BaseInitConifg.GetLogger().SetPrefix(methodName)
// 	receiver.BaseInitConifg.GetLogger().Info("开始调用:" + methodName)
// }


func NewBaseChainCodeBaseLogicServiceImpl(logger service.IBaseService, s shim.ChaincodeStubInterface, version uint64, BaseTransactionType base.TransBaseTypeV2) *BaseChainCodeBaseLogicServiceImpl {
	b := new(BaseChainCodeBaseLogicServiceImpl)
	// b.IBaseChainCode = init
	b.Stub = s
	b.Version = version
	b.BaseTransactionType = BaseTransactionType
	b.IBaseService = logger
	return b
}

// ////////////////////////////////////////////////////
// ///////   业务辅助方法
// ////////////////////////////////////////////////////
// ////////////////////////////////////////////////////
func (b *BaseChainCodeBaseLogicServiceImpl) CheckExist(objectType base.ObjectType, req ...interface{}) (bool, error2.IBaseError) {
	key, baseError := b.ChainCodeHelper.GetConcreteKey(b.Stub, objectType, req...)
	if nil != baseError {
		b.Error("获取ot=[%s]的key失败:%s", objectType, baseError.Error())
		return false, error2.ErrorsWithMessage(baseError, "获取key失败")
	}

	_, bytes, baseError := b.GetByKey(base.Key(key))
	if nil != baseError {
		b.Error("链上获取数据失败:%s", baseError.Error())
		return false, error2.ErrorsWithMessage(baseError, "链上查询数据失败")
	} else if nil == bytes || len(bytes) == 0 {
		return false, nil
	}

	return true, nil
}
func (b *BaseChainCodeBaseLogicServiceImpl) BuildCompositeKey(k base.ObjectType, args ...interface{}) (base.Key, error2.IBaseError) {
	b.Debug("开始创建组合键,ot=[%+v],args=[%+v]", k, args)
	s, e := b.ChainCodeHelper.GetConcreteKey(b.Stub, k, args...)
	// s, e := b.Stub.CreateCompositeKey(string(k), attributes)
	if nil != e {
		return "", error3.NewChainCodeError(e, "创建组合键失败")
	}
	b.Debug("创建的组合键为:[%+v]", s)
	return base.Key(s), nil
}

// putKey的同时 重新修饰值 使得业务区分,现用于区分 是 作品版权还是章节版权
func (b *BaseChainCodeBaseLogicServiceImpl) PutByKeyWithDecorate(configReq base.BCPutStateReq, data interface{}, decorater func([]byte) []byte) error2.IBaseError {
	bytes, e := b.buildBytes(configReq, data)
	if nil != e {
		b.Error("组合bytes失败:%s", e.Error())
		return e
	}

	if decorater != nil {
		bytes = decorater(bytes)
		b.Debug("遗留字段中 copyrightTypeIndex为:[%d],是否为chapter:[%v]", bytes[constants.LEFT_BYTE_BGEIN+constants.LEFT_BYTE_COPYRIGHT_TYPE_INDEX], bytes[constants.LEFT_BYTE_BGEIN+constants.LEFT_BYTE_COPYRIGHT_TYPE_INDEX] == 2)
	}

	// b.Debug("appendBytes: length=[%d],values=[%s]", len(leftIndexBytes), string(leftIndexBytes))
	b.Debug("bytes中最后一个byte的值为:[%v],是否是chapter:[%v]", bytes[len(bytes)-1], bytes[len(bytes)-1] == 2)
	if e := b.putByKey(configReq.Key, bytes); nil != e {
		b.Error("上传数据失败:%s", e.Error())
		return error2.ErrorsWithMessage(e, "上传数据失败")
	}
	b.Debug("成功上传数据,key={%s},总length为:{%d}", configReq.Key, len(bytes))

	return nil
}

func (b *BaseChainCodeBaseLogicServiceImpl) buildBytes(configReq base.BCPutStateReq, data interface{}) ([]byte, error2.IBaseError) {
	if configReq.Key == "" {
		return nil, error2.NewArguError(nil, "参数key不可为空")
	}
	b.Debug("开始上传,上传信息为: [%v]  ,上传数据为:{%v}", configReq, data)
	tranTypeLength := byte(len(b.BaseTransactionType))
	bytes := make([]byte, 0)
	// bytes := converter.BigEndianInt642Bytes(int64(b.BaseTransactionType))
	if configReq.From != "" {
		decodeBytes := []byte(configReq.From)
		bytes = append(bytes, decodeBytes...)
	} else {
		decodeBytes := make([]byte, constants.FROM_WALLET_ADDRESS_BYTE_LENGTH)
		bytes = append(bytes, decodeBytes...)
	}

	if configReq.To != "" {
		decodeBytes := []byte(configReq.To)
		bytes = append(bytes, decodeBytes...)
	} else {
		decodeBytes := make([]byte, constants.TO_WALLET_ADDRESS_BYTE_LENGTH)
		bytes = append(bytes, decodeBytes...)
	}

	amountBytes := converter.BigEndianFloat64ToByte(float64(configReq.Token))
	bytes = append(bytes, amountBytes...)

	values := make([]byte, 0)
	switch data.(type) {
	case []byte:
		values = append(values, data.([]byte)...)
	default:
		marshal, e := json.Marshal(data)
		if nil != e {
			return nil, error2.NewJSONSerializeError(e, "序列化上链参数失败")
		}
		values = append(values, marshal...)
	}

	v := b.Version
	bytes = append(bytes, converter.BigEndianInt642Bytes(int64(b.Version))...)
	leftBytes := make([]byte, constants.LEFT_BYTE_LENGTH)
	// 设置基本类型长度
	leftBytes[constants.LEFT_BYTE_TYPE_LENGTH_INDEX] = tranTypeLength
	if configReq.NeedEncrypt {
		leftBytes[constants.LEFT_BYTE_CRYPT_INDEX] = 1
		b.Debug("key=[%s]的数据需要进行加密,版本号位:[%d]", configReq.Key, v)
		encrypt, baseError := b.ChainCodeHelper.Encrypt(values, base.Version(v))
		if nil != baseError {
			b.Error("加密失败:%s", baseError.Error())
			return nil, error2.ErrorsWithMessage(baseError, "加密失败")
		}
		values = encrypt
		b.Debug("[PutByKey] [加密] 开始上传 {key=%s ; configType=%d ,fromWalletAddress=%s,toWalletAddress=%s,token=%v,version=%v,value=%s} 至区块链", configReq.Key, b.BaseTransactionType, configReq.From, configReq.To, configReq.Token, v, string(values))
	} else {
		b.Debug("[PutByKey] [非加密] 开始上传 {key=%s ; configType=%d ,fromWalletAddress=%s,toWalletAddress=%s,token=%v,version=%v,value=%s} 至区块链", configReq.Key, b.BaseTransactionType, configReq.From, configReq.To, configReq.Token, v, string(values))
	}
	bytes = append(bytes, leftBytes...)

	// 添加空余字节
	idleBytes := make([]byte, constants.IDLE_BYTE_LENGTH)
	bytes = append(bytes, idleBytes...)

	bytes = append(bytes, values...)

	// 添加基本类型字节
	configTypeBytes := b.BaseTransactionType.BigEndianConvtBytes()
	bytes = append(bytes, configTypeBytes...)

	return bytes, nil
}
func (b *BaseChainCodeBaseLogicServiceImpl) PutByKey(configReq base.BCPutStateReq, data interface{}) error2.IBaseError {
	if configReq.Key == "" {
		return error2.NewArguError(nil, "参数key不可为空")
	}
	b.Debug("开始上传,上传信息为: [%v]  ,上传数据为:{%v}", configReq, data)
	tranTypeLength := byte(len(b.BaseTransactionType))
	bytes := make([]byte, 0)
	// bytes := converter.BigEndianInt642Bytes(int64(b.BaseTransactionType))
	if configReq.From != "" {
		decodeBytes := []byte(configReq.From)
		bytes = append(bytes, decodeBytes...)
	} else {
		decodeBytes := make([]byte, constants.FROM_WALLET_ADDRESS_BYTE_LENGTH)
		bytes = append(bytes, decodeBytes...)
	}

	if configReq.To != "" {
		decodeBytes := []byte(configReq.To)
		bytes = append(bytes, decodeBytes...)
	} else {
		decodeBytes := make([]byte, constants.TO_WALLET_ADDRESS_BYTE_LENGTH)
		bytes = append(bytes, decodeBytes...)
	}

	amountBytes := converter.BigEndianFloat64ToByte(float64(configReq.Token))
	bytes = append(bytes, amountBytes...)

	values := make([]byte, 0)
	switch data.(type) {
	case []byte:
		values = append(values, data.([]byte)...)
	default:
		marshal, e := json.Marshal(data)
		if nil != e {
			return error2.NewJSONSerializeError(e, "序列化上链参数失败")
		}
		values = append(values, marshal...)
	}

	v := b.Version
	bytes = append(bytes, converter.BigEndianInt642Bytes(int64(b.Version))...)
	leftBytes := make([]byte, constants.LEFT_BYTE_LENGTH)
	// 设置基本类型长度
	leftBytes[constants.LEFT_BYTE_TYPE_LENGTH_INDEX] = tranTypeLength
	if configReq.NeedEncrypt {
		leftBytes[constants.LEFT_BYTE_CRYPT_INDEX] = 1
		b.Debug("key=[%s]的数据需要进行加密,版本号位:[%d]", configReq.Key, v)
		encrypt, baseError := b.ChainCodeHelper.Encrypt(values, base.Version(v))
		if nil != baseError {
			b.Error("加密失败:%s", baseError.Error())
			return error2.ErrorsWithMessage(baseError, "加密失败")
		}
		values = encrypt
		b.Debug("[PutByKey] [加密] 开始上传 {key=%s ; configType=%d ,fromWalletAddress=%s,toWalletAddress=%s,token=%v,version=%v,value=%s} 至区块链", configReq.Key, b.BaseTransactionType, configReq.From, configReq.To, configReq.Token, v, string(values))
	} else {
		b.Debug("[PutByKey] [非加密] 开始上传 {key=%s ; configType=%d ,fromWalletAddress=%s,toWalletAddress=%s,token=%v,version=%v,value=%s} 至区块链", configReq.Key, b.BaseTransactionType, configReq.From, configReq.To, configReq.Token, v, string(values))
	}
	bytes = append(bytes, leftBytes...)

	// 添加空余字节
	idleBytes := make([]byte, constants.IDLE_BYTE_LENGTH)
	bytes = append(bytes, idleBytes...)

	bytes = append(bytes, values...)

	// 添加基本类型字节
	configTypeBytes := b.BaseTransactionType.BigEndianConvtBytes()
	bytes = append(bytes, configTypeBytes...)

	if e := b.putByKey(configReq.Key, bytes); nil != e {
		b.Error("上传数据失败:%s", e.Error())
		return error2.ErrorsWithMessage(e, "上传数据失败")
	}
	b.Debug("成功上传数据,key={%s},总length为:{%d}", configReq.Key, len(bytes))

	return nil
}
func (b *BaseChainCodeBaseLogicServiceImpl) putByKey(key base.Key, bytes []byte) error2.IBaseError {
	if e := b.Stub.PutState(string(key), bytes); nil != e {
		return error3.NewChainCodeError(e, "插入数据失败")
	}

	return nil
}
func (b *BaseChainCodeBaseLogicServiceImpl) GetByKey(key base.Key) (base.BCBaseNodeInfo, []byte, error2.IBaseError) {
	return b.getByKey(string(key))
}
func (b *BaseChainCodeBaseLogicServiceImpl) GetDecryptDataByKey(key string) (base.BCBaseNodeInfo, []byte, error2.IBaseError) {
	info, bytes, baseError := b.getByKey(key)
	if nil != baseError {
		return info, bytes, baseError
	}
	if info.Encrypted {
		b.Debug("该数据被加密,开始解密,版本号为:[%v]", info.Version)
		bytes, baseError = b.ChainCodeHelper.Decrypt(bytes, info.Version)
		if nil != baseError {
			b.Error("解密失败:%s", baseError.Error())
			return info, nil, error3.NewChainCodeError(baseError, "解密失败")
		}
		b.Debug("解密成功")
	}
	return info, bytes, nil
}
func (this *BaseChainCodeBaseLogicServiceImpl) getByKey(key string) (base.BCBaseNodeInfo, []byte, error2.IBaseError) {
	var (
		result base.BCBaseNodeInfo
	)

	this.Debug("[GetByKey] 开始往区块链中获取数据 key=%s", key)
	bytes, e := this.Stub.GetState(key)
	if nil != e {
		this.Error("[GetByKey] 从区块链上获取数据 {key=%v} 的时候发生了错误:%s", key, e.Error())
		return result, nil, error3.NewChainCodeError(e, "获取数据失败")
	}
	if nil != bytes && len(bytes) >= constants.VLINK_COMMON_INDEX_END {
		node, modelWallets := base.GetRegularInfoV2(bytes)
		result = node
		this.Debug("[getByKey] 的leftBytes长度为:%d,modelWallets的长度为:%d", len(result.LeftBytes), len(modelWallets))
		if node.Encrypted {
			this.Debug("该key 对应的值被加密,需要进行解密")
			modelWallets, e = this.ChainCodeHelper.Decrypt(modelWallets, node.Version)
			if nil != e {
				this.Error("解密数据失败:{%s}", e.Error())
				return result, modelWallets, error2.NewCryptError(e, "解密数据失败")
			}
			this.Debug("[GetState] [解密后的数据为] 成功从链上获取{key=%s的信息} ,解析得到的 基本类型为:[%d] from钱包地址为:[%s],to钱包地址为:[%s],交易金额为:[%v],版本为:[%v],遗留字段为:[%v],模型对象的json为:[%s]", key, node.TxBaseType, node.From, node.To, node.Token, node.Version, node.LeftBytes, string(modelWallets))
		} else {
			this.Debug("[GetState] [非加密] 成功从链上获取{key=%s的信息} ,解析得到的 基本类型为:[%d] from钱包地址为:[%s],to钱包地址为:[%s],交易金额为:[%v],版本为:[%v],遗留字段为:[%v],模型对象的json为:[%s]", key, node.TxBaseType, node.From, node.To, node.Token, node.Version, node.LeftBytes, node.LeftBytes, string(modelWallets))
		}
		// if node.Version != 0 {
		// 	this.Debug("[GetState] 成功从链上获取{key=%s的信息} ,解析得到的 基本类型为:[%d],类型为:[%d] ,from钱包地址为:[%s],to钱包地址为:[%s],交易金额为:[%v],版本为:[%v],模型对象的json为:[%s]", key, node.TxBaseType, node.From, node.Version, node.To, node.Token, hex.EncodeToString(modelWallets))
		// } else {
		// 	this.Debug("[GetState] 成功从链上获取{key=%s的信息} ,解析得到的 基本类型为:[%d],类型为:[%d] ,from钱包地址为:[%s],to钱包地址为:[%s],交易金额为:[%v],版本为:[%v],模型对象的json为:[%s]", key, node.TxBaseType, node.From, node.Version, node.To, node.Token, modelWallets)
		// }
		return result, modelWallets, nil
	} else {
		this.Debug("[GetState] 链上{key=%s}的数据为空", key)
		return result, nil, nil
	}

}
func (b *BaseChainCodeBaseLogicServiceImpl) getKey(stub shim.ChaincodeStubInterface, key base.ObjectType, args ...interface{}) (string, error2.IBaseError) {
	return b.ChainCodeHelper.GetConcreteKey(stub, key, args...)
}

// 跨链调用
func (b *BaseChainCodeBaseLogicServiceImpl) InvokeOtherCC(req base.InvokeBaseReq) (base.BaseFabricResp, error2.IBaseError) {
	var args []string
	args = append(args, string(req.MethodName))

	bytes, e := json.Marshal(req.Data)
	if e != nil {
		return base.BaseFabricResp{}, error2.NewJSONSerializeError(e, fmt.Sprintf("序列化data=[%v]", req.Data))
	}
	args = append(args, string(bytes))

	invokeResp := b.Stub.InvokeChaincode(req.ChaincodeName, [][]byte{[]byte(args[0]), []byte(args[1]), []byte(strconv.Itoa(int(b.Version)))}, req.ChannelName)
	b.Debug("invoke 的返回值为:状态码:{%d},返回值:{%v},msg:{%s}", invokeResp.Status, string(invokeResp.Payload), invokeResp.Message)

	if invokeResp.Status != http.StatusOK {
		b.Error("调用链码=[%s],method=[%v],channel=[%v]失败:%s", req.ChaincodeName, req.MethodName, req.ChannelName, invokeResp.GetMessage())
		return base.BaseFabricResp{}, error3.NewChainCodeError(nil, "区块链调用失败:"+invokeResp.Message)
	}
	resp, err := handleResponse(invokeResp.Payload)
	if nil != err {
		b.Error("处理返回值失败:%s", e.Error())
		return base.BaseFabricResp{}, error2.ErrorsWithMessage(err, "处理返回值失败")
	}

	return resp, nil
}

func handleResponse(bytes []byte) (base.BaseFabricResp, error2.IBaseError) {
	// bytes := response.Payload
	var resp base.BaseFabricResp

	e := json.Unmarshal(bytes, &resp)
	if nil != e {
		return resp, error2.NewJSONSerializeError(e, "反序列化为 BaseFabricResp 失败")
	}

	return resp, nil
}
