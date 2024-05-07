package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

func getTxCreatorInfo(stub shim.ChaincodeStubInterface) (string, string, error) {

	//var mspid string
	var err error
	var uuid, role string
	var found bool

	role, found, err = cid.GetAttributeValue(stub, ROLE)
	if err != nil {
		fmt.Printf("Error getting Attribute Value: %s\n", err.Error())
		return "", "", err
	}
	if found == false {
		fmt.Printf("Error getting ROLE --> NOT FOUND!!!\n")
		return "", "", nil
	}
	fmt.Printf("ROLE FOUND: %s\n", role)

	uuid, found, err = cid.GetAttributeValue(stub, UUID)
	if err != nil {
		fmt.Printf("Error getting Attribute Value hf.EnrollmentID: %s\n", err.Error())
		return "", "", err
	}
	if found == false {
		fmt.Printf("Error getting UUID --> NOT FOUND!!!\n")
		return "", "", nil
	}
	fmt.Printf("UUID FOUND: %s\n", uuid)

	//return attrValue1, attrValue2, nil
	return role, uuid, nil
}

func isInvokerOperator(stub shim.ChaincodeStubInterface, attrName string) (bool, string, error) {
	var found bool
	var attrValue string
	var err error

	attrValue, found, err = cid.GetAttributeValue(stub, attrName)
	if err != nil {
		fmt.Printf("Error getting Attribute Value: %s\n", err.Error())
		return false, "", err
	}
	return found, attrValue, nil
}

// ThisIsTheRole return a boolean for the role searched
func ThisIsTheRole(stub shim.ChaincodeStubInterface, roleControl string) (bool, bool, error) {
	var ok bool = false
	var found bool
	var role string
	var err error

	// Who is the owner? Is it authorized to create a new ObjectDistribution (CREATOR)?
	found, role, err = isInvokerOperator(stub, ROLE)
	if err != nil {
		LogMessage("ERROR", "SmartIndustryChaincode ThisIsTheRole isInvokerOperator err not nil", err.Error())
		return false, false, err
	}

	if role == roleControl {
		ok = true
	} else {
		LogMessage("WARNING", "SmartIndustryChaincode postObjectDistribution ERROR: role is not creator (super-user)!!!", role)
	}
	return ok, found, nil

}
