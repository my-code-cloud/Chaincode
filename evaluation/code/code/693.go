package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hlf-mipt/saac-v2-core/pkg/model"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
)

var NotExistsErr = errors.New("the asset does not exist")

type assetsRepository struct {
}

func newAssetsRepository() *assetsRepository {
	return &assetsRepository{}
}

func (svc *assetsRepository) SetPrivate(ctx contractapi.TransactionContextInterface, mspId string, item *model.Asset) error {
	assetJSON, err := json.Marshal(*item)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutPrivateData(mspId, strconv.Itoa(item.ID), assetJSON)
	if err != nil {
		return err
	}

	return nil
}

func (svc *assetsRepository) GetPrivate(ctx contractapi.TransactionContextInterface, mspId string, id string) (*model.Asset, error) {
	assetJSON, err := ctx.GetStub().GetPrivateData(mspId, id)
	if err != nil {
		err = fmt.Errorf("failed to read from world state: %v", err)
		return nil, err
	}
	if assetJSON == nil {
		err = NotExistsErr
		return nil, err
	}

	var asset model.Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}
