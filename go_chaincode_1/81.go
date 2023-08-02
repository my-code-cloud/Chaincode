/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
	"github.com/hyperledger/fabric-samples/token-erc-721/chaincode-go/chaincode"
)

type serverConfig struct {
	CCID    string
	Address string
}

func main() {

	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	print("Start chaincode erc-721")
	nftContract := new(chaincode.TokenERC721Contract)
	nftContract.Info.Version = "0.0.1"
	nftContract.Info.Description = "ERC-721 fabric port"
	nftContract.Info.License = new(metadata.LicenseMetadata)
	nftContract.Info.License.Name = "Apache-2.0"
	nftContract.Info.Contact = new(metadata.ContactMetadata)
	nftContract.Info.Contact.Name = "Matias Salimbene"

	chaincode, err := contractapi.NewChaincode(nftContract)

	if err != nil {
		panic("Could not create chaincode from TokenERC721Contract." + err.Error())
	}

	chaincode.Info.Title = "ERC-721 chaincode"
	chaincode.Info.Version = "0.0.1"

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      chaincode,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	if err := server.Start(); err != nil {
		panic("Failed to start erc-721 chaincode. " + err.Error())
	}
}
