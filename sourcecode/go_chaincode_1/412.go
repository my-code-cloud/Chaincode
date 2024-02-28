package main

import (
	// "github.com/hyperledger/fabric-contract-api-go/contractapi"
	// "encoding/json"
	// "fmt"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Record struct {
	RecordId     string `json:"record_id,omitempty"  bson:"record_id"  form:"record_id"  binding:"record_id"`
	OwnerName    string `json:"owner_name,omitempty"  bson:"owner_name"  form:"owner_name"  binding:"owner_name"`
	Email        string `json:"email,omitempty"  bson:"email"  form:"email"  binding:"email"`
	NoShm        string `json:"no_shm,omitempty"  bson:"no_shm"  form:"no_shm"  binding:"no_shm"`
	Provinsi     string `json:"provinsi,omitempty"  bson:"provinsi"  form:"provinsi"  binding:"provinsi"`
	Kabupaten    string `json:"kabupaten,omitempty"  bson:"kabupaten"  form:"kabupaten"  binding:"kabupaten"`
	Kelurahan    string `json:"kelurahan,omitempty"  bson:"kelurahan"  form:"kelurahan"  binding:"kelurahan"`
	Penerbitan   string `json:"penerbitan,omitempty"  bson:"penerbitan"  form:"penerbitan"  binding:"penerbitan"`
	Luas         string `json:"luas,omitempty"  bson:"luas"  form:"luas"  binding:"luas"`
	CertFilename string `json:"cert_filename,omitempty"  bson:"cert_filename"  form:"cert_filename"  binding:"cert_filename"`
	CertCid      string `json:"cert_cid,omitempty"  bson:"cert_cid"  form:"cert_cid"  binding:"cert_cid"`
}

const index = "email~recordid"
const index2 = "owner~recordid"
const index3 = "noshm~recordid"

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}

func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, recordId string) (bool, error) {
	assetJson, err := ctx.GetStub().GetState(recordId)
	if err != nil {
		return false, fmt.Errorf("failed to read from state database: %v", err)
	}
	return assetJson != nil, nil
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface,
	recordId string,
	ownerName string,
	email string,
	noShm string,
	provinsi string,
	kabupaten string,
	kelurahan string,
	penerbitan string,
	luas string,
	certFilename string,
	certCid string,
	) error {
	exists, err := s.AssetExists(ctx, recordId)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the record %s is already exist", recordId)
	}

	record := Record{
		RecordId:   recordId,
		OwnerName:  ownerName,
		Email:      email,
		NoShm:      noShm,
		Provinsi:   provinsi,
		Kabupaten:  kabupaten,
		Kelurahan:  kelurahan,
		Penerbitan: penerbitan,
		Luas:       luas,
		CertFilename: certFilename,
		CertCid: certCid,
	}

	assetJson, err := json.Marshal(record)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(recordId, assetJson)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	value := []byte{0x00}
	indexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{email, recordId})
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().PutState(indexKey, value)
	if err != nil {
		fmt.Printf("index : %s", err.Error())
	}

	indexKey2, err := ctx.GetStub().CreateCompositeKey(index2, []string{ownerName, recordId})
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().PutState(indexKey2, value)
	if err != nil {
		fmt.Printf("index : %s", err.Error())
	}
	indexKey3, err := ctx.GetStub().CreateCompositeKey(index3, []string{noShm, recordId})
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return ctx.GetStub().PutState(indexKey3, value)
	// return ctx.GetStub().PutState(userUid, assetJson)
}

func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Record, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Record
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Record
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, recordId string) (*Record, error) {
	assetJson, err := ctx.GetStub().GetState(recordId)
	if err != nil {
		return nil, fmt.Errorf("faield to read from state database: %v", err)
	}
	if assetJson == nil {
		return nil, fmt.Errorf("the record %s does not exist", recordId)
	}

	var record Record
	err = json.Unmarshal(assetJson, &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface,
	recordId string,
	ownerName string,
	email string,
	noShm string,
	provinsi string,
	kabupaten string,
	kelurahan string,
	penerbitan string,
	luas string,
	certFilename string,
	certCid string,
	) error {
	exists, err := s.AssetExists(ctx, recordId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the record %s is does not exist", recordId)
	}

	record := Record{
		RecordId:   recordId,
		OwnerName:  ownerName,
		Email:      email,
		NoShm:      noShm,
		Provinsi:   provinsi,
		Kabupaten:  kabupaten,
		Kelurahan:  kelurahan,
		Penerbitan: penerbitan,
		Luas:       luas,
		CertFilename: certFilename,
		CertCid: certCid,
	}

	assetJson, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(recordId, assetJson)
}

func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, recordId string) error {
	exists, err := s.AssetExists(ctx, recordId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the record %s does not exist", recordId)
	}
	
	record, err := s.ReadAsset(ctx, recordId)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	indexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{record.Email, recordId})
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().DelState(indexKey)
	if err != nil {
		return fmt.Errorf(err.Error())
	}


	indexKey2, err := ctx.GetStub().CreateCompositeKey(index2, []string{record.OwnerName, recordId})
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().DelState(indexKey2)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	indexKey3, err := ctx.GetStub().CreateCompositeKey(index3, []string{record.NoShm, recordId})
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().DelState(indexKey3)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	
	return ctx.GetStub().DelState(recordId)
}

func (s *SmartContract) ReadAssetByEmail(ctx contractapi.TransactionContextInterface, email string) ([]*Record, error) {
	indexIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(index, []string{email})
	if err != nil {
		return nil, err
	}
	defer indexIterator.Close()

	var records []*Record
	// var parts []string
	for indexIterator.HasNext() {
		responseRange, err := indexIterator.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		if len(compositeKeyParts) > 1 {
			recordId := compositeKeyParts[1]
			record, err := s.ReadAsset(ctx, recordId)
			if err != nil {
				return nil, err
			}
			records = append(records, record)
		}

		// parts = append(parts, compositeKeyParts[0])
		// parts = append(parts, compositeKeyParts[1])

	}
	return records, nil
}

func (s *SmartContract) ReadAssetByNoShm(ctx contractapi.TransactionContextInterface, noShm string) ([]*Record, error) {
	indexIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(index3, []string{noShm})
	if err != nil {
		return nil, err
	}
	defer indexIterator.Close()

	var records []*Record
	// var parts []string
	for indexIterator.HasNext() {
		responseRange, err := indexIterator.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		if len(compositeKeyParts) > 1 {
			recordId := compositeKeyParts[1]
			record, err := s.ReadAsset(ctx, recordId)
			if err != nil {
				return nil, err
			}
			records = append(records, record)
		}

		// parts = append(parts, compositeKeyParts[0])
		// parts = append(parts, compositeKeyParts[1])

	}
	return records, nil
}

func (s *SmartContract) ReadAssetByOwner(ctx contractapi.TransactionContextInterface, ownerName string) ([]*Record, error) {
	indexIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(index2, []string{ownerName})
	if err != nil {
		return nil, err
	}
	defer indexIterator.Close()

	var records []*Record
	// var parts []string
	for indexIterator.HasNext() {
		responseRange, err := indexIterator.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		if len(compositeKeyParts) > 1 {
			recordId := compositeKeyParts[1]
			record, err := s.ReadAsset(ctx, recordId)
			if err != nil {
				return nil, err
			}
			records = append(records, record)
		}

		// parts = append(parts, compositeKeyParts[0])
		// parts = append(parts, compositeKeyParts[1])

	}
	return records, nil
}
