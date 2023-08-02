package main

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"math/big"

	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/fabric-chaincode-go/shim"

	"github.com/hyperledger/fabric-protos-go/peer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

var caAccess bool

//IdentityChaincode stuct
type IdentityChaincode struct {
}

//Init method
func (t *IdentityChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	LogMessage(INFO, "---INIT SMART INDUSTRY CONTRACT---", "")

	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *IdentityChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	// Variable

	LogMessage(INFO, "IdentityChaincode function invoked: ", fn)

	if fn == "postIdentity" {
		return t.postIdentity(stub, args) // TODO
	} else if fn == "suspendIdentity" {
		return t.editIdentity(stub, SUSPENDSTATUS, args) // TODO
	} else if fn == "activateIdentity" {
		return t.editIdentity(stub, ACTIVESTATUS, args) // TODO
	} else if fn == "revokeIdentity" {
		return t.editIdentity(stub, REVOKESTATUS, args) // TODO
	} else if fn == "getIdentity" {
		return t.getIdentity(stub, args) // TODO
		// } else if fn == "controlIdentity" {
		//	return t.controlIdentity(stub, args) // TODO
	} else {
		return shim.Error(ValResp(400, "Invalid function name in Invoke"))
	}
}

//*************************************************************
// MANAGE IDENTITY
//*************************************************************

//postIdentity: the input json must contain the object Identity,
//The caller must be a MEMBER/ADMIN!!!

func (t *IdentityChaincode) postIdentity(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	LogMessage(INFO, "IdentityChaincode ON postIdentity", time.Now().Format(time.RFC3339))

	var jsonResp string
	var err error
	var identity Identity = Identity{}
	var epkxy ECDSApublicKeyXY = ECDSApublicKeyXY{}
	var rpken RSApublicKeyEN = RSApublicKeyEN{}
	var byteEpkxy []byte
	var byteRpken []byte
	var ecdsaPublicKey *ecdsa.PublicKey
	var rsaPublicKey *rsa.PublicKey
	var identityController Identity = Identity{}
	var baseEntryController BaseEntry = BaseEntry{}
	var statusEntryController StatusEntry = StatusEntry{}
	var checkControl bool
	var desc string
	var datetime time.Time
	var byteIdentity []byte
	var baseEntryOUT BaseEntry = BaseEntry{}
	var byteBaseEntry []byte
	var statusEntry StatusEntry = StatusEntry{}
	var byteStatusEntry []byte
	var IDKey string
	var found bool
	var role string

	if len(args) != 2 {
		LogMessage(ERROR, "IdentityChaincode postIdentity incorrect numbers of Args: ", strconv.Itoa(len(args)))
		return shim.Error(ValResp(400, "postIdentity error: incorrect numbers of Args"+strconv.Itoa(len(args))))
	}

	// Create Object from args parameters
	err = json.Unmarshal([]byte(args[0]), &identity)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode postIdentity jsonUnmarshal identity err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}

	datetime = time.Now()

	// Control Identity into CA and check the authorization to PostIdentity
	// GetAttributeValue(ROLE) - notfound is KO
	if caAccess {
		found, role, err = isInvokerOperator(stub, ROLE)
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode postIdentity isInvokerOperator(role) err not nil", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}
		if !found {
			LogMessage(ERROR, "IdentityChaincode postIdentity ERROR: ROLE is empty!!!", "")
			return shim.Error(ValResp(404, "IdentityChaincode postIdentity ERROR: ROLE is empty!!!"))
		}
		LogMessage(DEBUG, "IdentityChaincode postIdentity isInvokerOperator by role: ", role)
		/* if role != CREATOR && role != ADMIN {
			LogMessage(ERROR, "IdentityChaincode postIdentity ERROR: ROLE is not authorized to Post Identity!!", role)
			return shim.Error("IdentityChaincode postIdentity ERROR: ROLE is not authorized to Post Identity!!")
		} */

	}

	// baseEntryOut is identity in output
	// identity.BaseEntry is identity in input
	baseEntryOUT = identity.BaseEntry

	// We are going to create a new identity that is'nt the ROOT of the system. ROOT: Address = Controller.
	// NO:
	if identity.BaseEntry.Controller != "" {

		// Control if esist the Controller and if has the authorization to create a new identity
		baseEntryController, err = GetBaseEntry(stub, baseEntryOUT.Controller)
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode PostIdentity GetBaseEntryController err not nil", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}
		if (baseEntryController == BaseEntry{}) {
			LogMessage(ERROR, "postIdentity error: invalid controller (address does not exist): ", baseEntryOUT.Controller)
			return shim.Error(ValResp(404, "postIdentity error: invalid controller (address does not exist)"))
		}

		// BaseEntry of Controller exists. Now we control the StatusEntry of Controller
		statusEntryController, err = GetStatusEntry(stub, baseEntryController.Address, datetime)
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode PostIdentity GetStatusEntryController err not nil", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}
		if (statusEntryController == StatusEntry{}) {
			LogMessage(ERROR, "postIdentity error: invalid controller (invalid status)", baseEntryController.Address)
			return shim.Error(ValResp(404, "postIdentity error: invalid controller (invalid status): "+baseEntryController.Address))
		}

		// Identity Controller exists. Now we control the Chain of Trust
		identityController.BaseEntry = baseEntryController
		identityController.StatusEntry = statusEntryController

		if identityController.BaseEntry.Address != identityController.BaseEntry.Controller {
			checkControl, desc, err = ControlChainOfTrust(stub, identityController.BaseEntry.Address, datetime)

			if err != nil {
				LogMessage(ERROR, "IdentityChaincode PostIdentity ControlChainOfTrust err not nil", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}

			if !checkControl {
				LogMessage(ERROR, "postIdentity error: invalid identity: ", desc)
				return shim.Error(ValResp(400, "postIdentity error: invalid identity: "+desc))
			}
		} else {
			LogMessage(INFO, "IdentityChaincode PostIdentity the SID Controller is a ROOT", identityController.BaseEntry.Address)
		}

		// The chain of trust is ok. The Controller is Active and all controllers connected to him are actives

		// BASE ENTRY ///////////////////////////////////////////////////////////////////////////////////////////
		// Create the Public Key from parameters

		// Switch case for the algorithm of public key (only ECDSA and RSA are admitted at the moment)
		if identity.BaseEntry.KeyType == ECDSA {

			var signature Signature = Signature{}
			err = json.Unmarshal([]byte(args[1]), &signature)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode postIdentity jsonUnmarshal signature err not nil", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}

			LogMessage(DEBUG, "IdentityChaincode postIdentity identity.Controller: ", identity.Controller)
			LogMessage(DEBUG, "IdentityChaincode postIdentity identity.PKeyBlob  : ", identity.PKeyBlob)
			LogMessage(DEBUG, "IdentityChaincode postIdentity signature.R        : ", signature.R)
			LogMessage(DEBUG, "IdentityChaincode postIdentity signature.S        : ", signature.S)

			// create the public key ecdsa of the Controller for the control of the signature
			// Unmarshal in a structure complex with curve details
			err = json.Unmarshal([]byte(identityController.BaseEntry.PKeyBlob), &epkxy)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode postIdentity jsonUnmarshal publicKey ECDSA err not nil", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}

			ecdsaPublicKey = hexToPublicKey(epkxy)
			// validateEcdsaSignature validate a pkeyblob (regenerate) signed with the public key relative at the signature
			valid := validateEcdsaSignature(ecdsaPublicKey, identity.BaseEntry.PKeyBlob, signature)
			if !valid {
				LogMessage(ERROR, "postIdentity error: controller authentication failed", "")
				return shim.Error(ValResp(401, "postIdentity error: controller authentication failed"))
			}
			LogMessage(DEBUG, "IdentityChaincode PostIdentity Validate of Signature of Public Key: TRUE!!!!!!!", "")

			// create the public key ecdsa
			err = json.Unmarshal([]byte(identity.BaseEntry.PKeyBlob), &epkxy)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode postIdentity jsonUnmarshal publicKeyxy err not nil", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}

			byteEpkxy, err = json.Marshal(&epkxy)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode postIdentity jsonMarshal ecdsaPublicKey err not nil", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}
			baseEntryOUT.PKeyBlob = string(byteEpkxy)

		} else if identity.BaseEntry.KeyType == RSA {
			var sign string = ""
			sign = args[1]
			// create the public key rsa of the Controller for the control of the signature
			// Unmarshal in a structure complex with details
			err = json.Unmarshal([]byte(identityController.BaseEntry.PKeyBlob), &rpken)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode postIdentity jsonUnmarshal publicKey RSA err not nil", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}

			rsaPublicKey, err = hexToRSAPublicKey(rpken)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode postIdentity hexToRSAPublicKey publicKey RSA err conv string to int E", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}

			// validateEcdsaSignature validate a pkeyblob (regenerate) signed with the public key relative at the signature
			valid, err := validateRsaSignature(rsaPublicKey, identity.BaseEntry.PKeyBlob, sign)
			if !valid || err != nil {
				LogMessage(ERROR, "postIdentity error: controller authentication failed", err.Error())
				return shim.Error(ValResp(401, "postIdentity error: controller authentication failed"))
			}
			LogMessage(DEBUG, "IdentityChaincode PostIdentity Validate of Signature of RSA Public Key: TRUE!!!!!!!", "")

			// create the public key rsa
			err = json.Unmarshal([]byte(identity.BaseEntry.PKeyBlob), &rpken)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode postIdentity jsonUnmarshal publicKeyen err not nil", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}

			byteRpken, err = json.Marshal(&rpken)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode postIdentity jsonMarshal rsaPublicKey err not nil", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}
			baseEntryOUT.PKeyBlob = string(byteRpken)
		} else {
			LogMessage(ERROR, "IdentityChaincode postIdentity ERROR: algorithm for the Public Key is not admitted!! ", identity.BaseEntry.KeyType)
			return shim.Error(ValResp(400, "postIdentity error: invalid identity (invalid key type)"))
		}

		baseEntryOUT.Address = createAddressFromPKeyBlob(identity.BaseEntry.PKeyBlob)
		LogMessage(DEBUG, "IdentityChaincode postIdentity Address generated: ", baseEntryOUT.Address)

	} else {
		// We are going to create the Root of the system (controller = "") >>>> ONLY ADMIN CAN DO THIS (TO CREATE ROOT)
		if caAccess && role != ADMIN {
			LogMessage(ERROR, "postIdentity error: invalid identity (caller cannot create a chain of trust root)", role)
			return shim.Error(ValResp(400, "postIdentity error: invalid identity (caller cannot create a chain of trust root)!!"))
		}
		if identity.BaseEntry.KeyType == ECDSA {
			err = json.Unmarshal([]byte(identity.BaseEntry.PKeyBlob), &epkxy)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode postIdentity jsonUnmarshal signature err not nil", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}

			byteEpkxy, err = json.Marshal(epkxy)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode postIdentity jsonMarshal epkxy err not nil", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}
			baseEntryOUT.PKeyBlob = string(byteEpkxy)
		} else if identity.BaseEntry.KeyType == RSA {

			err = json.Unmarshal([]byte(identity.BaseEntry.PKeyBlob), &rpken)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode postIdentity jsonUnmarshal PKyBlob in RSA Public Key err not nil", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}

			byteRpken, err = json.Marshal(rpken)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode postIdentity jsonMarshal rpken err not nil", err.Error())
				return shim.Error(ValResp(400, err.Error()))
			}
			baseEntryOUT.PKeyBlob = string(byteRpken)
		} else {
			LogMessage(ERROR, "IdentityChaincode postIdentity ERROR: algorithm for the Public Key is not admitted!! ", identity.BaseEntry.KeyType)
			return shim.Error(ValResp(400, "postIdentity error: invalid identity (invalid key type)"))
		}
		// Create address
		baseEntryOUT.Address = createAddressFromPKeyBlob(identity.BaseEntry.PKeyBlob)
		LogMessage(DEBUG, "IdentityChaincode postIdentity Address generated: ", baseEntryOUT.Address)

		baseEntryOUT.Controller = baseEntryOUT.Address
		LogMessage(DEBUG, "IdentityChaincode postIdentity Address generated for the ROOT: ", baseEntryOUT.Address)

	}
	// Control if the record BaseEntry is there into the Ledger
	// GetBaseEntry with this Address
	// Create the KEY

	IDKey, err = stub.CreateCompositeKey(BASEENTRY, []string{baseEntryOUT.Address})
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode postIdentity CreateCompositeKey err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}

	LogMessage(DEBUG, "IdentityChaincode POSTIdentity for the key:", IDKey)
	byteBaseEntry, err = stub.GetState(IDKey)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode postIdentity GetState err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}
	if byteBaseEntry != nil {
		LogMessage(ERROR, "postIdentity error: invalid identity (address is not unique)", baseEntryOUT.Address)
		return shim.Error(ValResp(400, "postIdentity error: invalid identity (address is not unique)"))
	}

	baseEntryOUT.Created = time.Now().Format(time.RFC3339)
	if caAccess {
		found, baseEntryOUT.CreatedBy, err = isInvokerOperator(stub, UUID)
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode postIdentity isInvokerOperator(role) err not nil", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}
		if !found {
			LogMessage(ERROR, "IdentityChaincode postIdentity ERROR: UUID is empty!!!", "")
			return shim.Error(ValResp(404, "IdentityChaincode postIdentity ERROR: UUID is empty!!!"))
		}
		LogMessage(DEBUG, "IdentityChaincode postIdentity isInvokerOperator by UUID: ", baseEntryOUT.CreatedBy)

	}

	// Convert Object BaseEntry in []Byte
	byteBaseEntry, err = json.Marshal(&baseEntryOUT)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode postIdentity jsonMarshal BaseEntry err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}

	// PUT STATE INTO THE LEDGER
	err = stub.PutState(IDKey, byteBaseEntry)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode postIdentity putState(BaseEntry) err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}
	LogMessage(INFO, "IdentityChaincode postIdentity putState(BaseEntry) is OK with address: ", IDKey)

	// STATUS ENTRY ////////////////////////////////
	statusEntry = identity.StatusEntry
	statusEntry.Address = baseEntryOUT.Address
	statusEntry.ValidFrom = datetime.Format(time.RFC3339)

	if identity.StatusEntry.Status != SUSPENDSTATUS &&
		identity.StatusEntry.Status != ACTIVESTATUS {
		LogMessage(INFO, "postIdentity error: invalid identity (invalid status)", string(identity.StatusEntry.Status))
		return shim.Error(ValResp(400, "postIdentity error: invalid identity (invalid status)"+string(identity.StatusEntry.Status)))
	}

	// Create the KEY
	IDKey, err = stub.CreateCompositeKey(STATUSENTRY, []string{statusEntry.Address, statusEntry.SubAddress})

	// Convert Object StatusEntry in []Byte
	byteStatusEntry, err = json.Marshal(&statusEntry)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode postIdentity jsonMarshal StatusEntry err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}

	// PUT STATE INTO THE LEDGER
	err = stub.PutState(IDKey, byteStatusEntry)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode postIdentity putState(StatusEntry) err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}
	LogMessage(INFO, "IdentityChaincode postIdentity putState(StatusEntry) is OK with address: ", IDKey)

	identity.BaseEntry = baseEntryOUT
	identity.StatusEntry = statusEntry

	byteIdentity, err = json.Marshal(&identity)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode postIdentity jsonMarshal identity err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}

	jsonResp = string(byteIdentity)
	LogMessage(INFO, "IdentityChaincode postIdentity putState is OK with response:", jsonResp)

	return shim.Success([]byte(ValResp(201, string(byteIdentity))))

}

//editIdentity: the input json must contain the object Identity to EDIT,
//The caller must be a MEMBER/ADMIN!!!

func (t *IdentityChaincode) editIdentity(stub shim.ChaincodeStubInterface, funcEdit int, args []string) pb.Response {
	// SUSPENDSTATUS = 2 - ACTIVESTATUS  = 1 - REVOKESTATUS  = 3

	LogMessage(INFO, "IdentityChaincode ON editIdentity with function: ", strconv.Itoa(funcEdit))

	var jsonResp string
	var IDKey string
	var err error
	var identity Identity = Identity{}
	var ecdsaPublicKey *ecdsa.PublicKey
	var rsaPublicKey *rsa.PublicKey
	var byteIdentity []byte
	var baseEntry BaseEntry = BaseEntry{}
	var byteBaseEntry []byte
	var epkxy ECDSApublicKeyXY = ECDSApublicKeyXY{}
	var rpken RSApublicKeyEN = RSApublicKeyEN{}
	var statusEntry StatusEntry = StatusEntry{}
	var byteStatusEntry []byte
	var statusEntries []StatusEntry
	var found bool
	var addrIn string
	var subAddrIn string
	var contrIn string
	var identityController Identity = Identity{}
	var baseEntryController BaseEntry = BaseEntry{}
	var statusEntryController StatusEntry = StatusEntry{}
	var checkControl bool
	var desc string
	var datetime time.Time

	var role string

	// Control args[] Address, subAddress and Controller
	if len(args) != 4 {
		LogMessage(ERROR, "IdentityChaincode editIdentity numbers of Args not correctly!: ", strconv.Itoa(len(args)))
		return shim.Error(ValResp(400, "IdentityChaincode editIdentity numbers of Args not correctly: "+strconv.Itoa(len(args))))
	}
	addrIn = args[0]
	subAddrIn = args[1]
	contrIn = args[2]

	datetime = time.Now()

	if caAccess {
		// Control Identity into CA and check the authorization to editIdentity
		// GetAttributeValue(ROLE) - notfound is KO
		found, role, err = isInvokerOperator(stub, ROLE)
		if err != nil {
			LogMessage(ERROR, "editIdentity error: isInvokerOperator(role) err not nil", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}
		if !found {
			LogMessage(ERROR, "editIdentity error: ROLE is empty!!!", "")
			return shim.Error(ValResp(404, "editIdentity ERROR: ROLE is empty"))
		}
		LogMessage(DEBUG, "editIdentity isInvokerOperator by role: ", role)
		if role != ADMIN {
			LogMessage(ERROR, "editIdentity error: invalid identity (caller cannot create a chain of trust root)", role)
			return shim.Error(ValResp(401, "postIdentity error: invalid identity (caller cannot create a chain of trust root)!"))
		}
	}

	// Create ID KEY
	IDKey, err = stub.CreateCompositeKey(BASEENTRY, []string{addrIn})
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode editIdentity CreateCompositeKey err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}
	LogMessage(DEBUG, "IdentityChaincode editIdentity for the key:", IDKey)
	byteBaseEntry, err = stub.GetState(IDKey)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode editIdentity GetState err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}
	if byteBaseEntry == nil {
		LogMessage(ERROR, "editIdentity error: identity not found", IDKey)
		return shim.Error(ValResp(404, "IdentityChaincode editIdentity baseEntry DATA NOT FOUND func EDIT not possible"))

	}
	err = json.Unmarshal(byteBaseEntry, &baseEntry)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode editIdentity Unmarshal err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}

	// Only your controller (parameter args[2]) can modify the status of a identity
	if baseEntry.Controller != contrIn {
		LogMessage(ERROR, "editIdentity error: controller mismatch", contrIn+" "+baseEntry.Controller)
		return shim.Error(ValResp(401, "editIdentity error: controller mismatch"+contrIn+" "+baseEntry.Controller))
	}

	// GET Identity Controller
	baseEntryController, err = GetBaseEntry(stub, baseEntry.Controller)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode editIdentity GetBaseEntryController err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}
	if (baseEntryController == BaseEntry{}) {
		LogMessage(ERROR, "editIdentity error: invalid controller (address does not exist)", baseEntry.Controller)
		return shim.Error(ValResp(404, "IdentityChaincode editIdentity GetBaseEntryController Address of Controller NOT FOUND - BaseEntry"))
	}

	// BaseEntry of Controller exists. Now we control the StatusEntry of Controller
	statusEntryController, err = GetStatusEntry(stub, baseEntryController.Address, datetime)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode editIdentity GetStatusEntryController err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}
	if (statusEntryController == StatusEntry{}) {
		LogMessage(ERROR, "editIdentity error: invalid controller (invalid status)", baseEntryController.Address)
		return shim.Error(ValResp(404, "IdentityChaincode editIdentity GetStatusEntryController Address of Controller NOT FOUND - StatusEntry"))
	}

	// Identity Controller exists. Now we control the Chain of Trust
	identityController.BaseEntry = baseEntryController
	identityController.StatusEntry = statusEntryController

	if identityController.BaseEntry.Address != identityController.BaseEntry.Controller {
		checkControl, desc, err = ControlChainOfTrust(stub, identityController.BaseEntry.Address, datetime)

		if err != nil {
			LogMessage(ERROR, "IdentityChaincode editIdentity ControlChainOfTrust err not nil", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}

		if !checkControl {
			LogMessage(ERROR, "editIdentity error: invalid identity : ", desc)
			return shim.Error(ValResp(400, "editIdentity error: invalid identity : "+desc))
		}
	} else {
		LogMessage(INFO, "IdentityChaincode editIdentity the SID Controller a ROOT", identity.BaseEntry.Address)
	}

	// The chain of trust is ok. The Controller is Active and all controllers connected to him are actives

	// BASE ENTRY //////////////////////////////////////////////////////////////////////////////////////////

	LogMessage(DEBUG, "IdentityChaincode editIdentity subAddrIn: ", subAddrIn)

	if baseEntry.KeyType == ECDSA {
		var signature Signature = Signature{}
		err = json.Unmarshal([]byte(args[3]), &signature)
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode postIdentity jsonUnmarshal signature err not nil", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}

		// read the public key ecdsa of the Controller for the control of the signature
		err = json.Unmarshal([]byte(identityController.BaseEntry.PKeyBlob), &epkxy)
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode postIdentity jsonUnmarshal publicKey err not nil", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}

		ecdsaPublicKey = hexToPublicKey(epkxy)
		// validateEcdsaSignature validate a pkeyblob signed with the public key relative at the signature
		valid := validateEcdsaSignature(ecdsaPublicKey, addrIn+subAddrIn, signature)
		if !valid {
			LogMessage(ERROR, "editIdentity error: controller authentication failed", "")
			return shim.Error(ValResp(401, "editIdentity error: controller authentication failed"))
		}
		LogMessage(DEBUG, "IdentityChaincode editIdentity Validate of Signature of subAddrIn: TRUE!!!!!!!", "")

	} else if baseEntry.KeyType == RSA {
		var sign string = ""
		sign = args[3]
		// read the public key rsa of the Controller for the control of the signature
		err = json.Unmarshal([]byte(identityController.BaseEntry.PKeyBlob), &rpken)
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode postIdentity jsonUnmarshal publicKey RSA err not nil", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}

		rsaPublicKey, err = hexToRSAPublicKey(rpken)
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode editIdentity hexToRSAPublicKey publicKey RSA err conv string to int E", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}

		// validateEcdsaSignature validate a pkeyblob (regenerate) signed with the public key relative at the signature
		valid, err := validateRsaSignature(rsaPublicKey, addrIn+subAddrIn, sign)
		if !valid || err != nil {
			LogMessage(ERROR, "editIdentity error: controller authentication failed", err.Error())
			return shim.Error(ValResp(401, "editIdentity error: controller authentication failed"))
		}
		LogMessage(DEBUG, "IdentityChaincode EditIdentity Validate of Signature of RSA Public Key: TRUE!!!!!!!", "")

		// validateEcdsaSignature validate a pkeyblob signed with the public key relative at the signature

	} else {
		LogMessage(ERROR, "editIdentity error: invalid identity (invalid key type)", baseEntry.KeyType)
		return shim.Error(ValResp(400, "IdentityChaincode editIdentity KeyType (of pbKey) of this identity is not ECDSA and not RSA, this is a problem."))
	}

	// Search statusEntry for this baseEntry
	statusEntries, err = GetStateByPartialCompositeKeyStatusEntry(stub, []string{baseEntry.Address})
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode editIdentity GetStateByPartialCompositeKeyStatusEntry err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}
	if len(statusEntries) == 0 || statusEntries == nil {
		LogMessage(ERROR, "IdentityChaincode editIdentity GetStateByPartialCompositeKeyStatusEntry DATA NOT FOUND", IDKey)
		return shim.Error(ValResp(404, "editIdentity error: invalid identity (status not found)"))
	}

	statusEntry = statusEntries[len(statusEntries)-1]

	// Control Status into the Ledger
	if statusEntry.Status == REVOKESTATUS {
		if funcEdit != REVOKESTATUS {
			LogMessage(ERROR, "editIdentity error: invalid identity (invalid status)", baseEntry.Address)
			return shim.Error(ValResp(400, "editIdentity error: invalid identity (invalid status)"))
		}
	}

	statusEntry.SubAddress = subAddrIn
	statusEntry.ValidFrom = datetime.Format(time.RFC3339)
	if caAccess {
		found, statusEntry.UpdatedBy, err = isInvokerOperator(stub, UUID)
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode editIdentity isInvokerOperator(role) err not nil", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}
		if !found {
			LogMessage(ERROR, "IdentityChaincode editIdentity ERROR: UUID is empty!!!", "")
			return shim.Error(ValResp(404, "editIdentity error: UUID is empty!!!"))
		}
		LogMessage(DEBUG, "IdentityChaincode editIdentity isInvokerOperator by UUID: ", baseEntry.CreatedBy)
	}

	// Set the value of new status
	if funcEdit == ACTIVESTATUS {
		if statusEntry.Status == ACTIVESTATUS {
			LogMessage(WARNING, "IdentityChaincode editIdentity: identity is already ACTIVE", baseEntry.Address)
			return shim.Error(ValResp(409, "editIdentity error: conflict with the current status (ACTIVE) of the resource."))
		}
		LogMessage(DEBUG, "IdentityChaincode editIdentity: identity NOW is ACTIVE ", baseEntry.Address)
		statusEntry.Status = ACTIVESTATUS
	} else if funcEdit == SUSPENDSTATUS {
		if statusEntry.Status == SUSPENDSTATUS {
			LogMessage(WARNING, "IdentityChaincode editIdentity: identity is already SUSPENDED", baseEntry.Address)
			return shim.Error(ValResp(409, "editIdentity error: conflict with the current status (SUSPENDED) of the resource."))
		}
		statusEntry.Status = SUSPENDSTATUS
		LogMessage(DEBUG, "IdentityChaincode editIdentity: identity NOW is SUSPENDED ", baseEntry.Address)
	} else if funcEdit == REVOKESTATUS {
		if statusEntry.Status == REVOKESTATUS {
			LogMessage(WARNING, "IdentityChaincode editIdentity: identity is already REVOKED", baseEntry.Address)
			return shim.Error(ValResp(409, "editIdentity error: conflict with the current status (REVOKED) of the resource."))
		}
		statusEntry.Status = REVOKESTATUS
		LogMessage(DEBUG, "IdentityChaincode editIdentity: identity NOW is REVOKED ", baseEntry.Address)
	} else {
		LogMessage(ERROR, "IdentityChaincode editIdentity function Edit not compatible: ", strconv.Itoa(funcEdit))
		return shim.Error(ValResp(400, "Invalid invoke function edit name !(ACTIVE, SUSPEND, REVOKE)"))
	}

	// Create ID KEY
	IDKey, err = stub.CreateCompositeKey(STATUSENTRY, []string{statusEntry.Address, statusEntry.SubAddress})
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode editIdentity CreateCompositeKey StatusEntry err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}

	// Convert Object StatusEntry in []Byte
	byteStatusEntry, err = json.Marshal(&statusEntry)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode editIdentity jsonMarshal StatusEntry err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}

	// PUT STATE INTO THE LEDGER
	err = stub.PutState(IDKey, byteStatusEntry)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode editIdentity putState(StatusEntry) err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}
	LogMessage(INFO, "IdentityChaincode editIdentity putState(StatusEntry) is OK with address: ", IDKey)

	identity.BaseEntry = baseEntry
	identity.StatusEntry = statusEntry
	byteIdentity, err = json.Marshal(identity)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode editIdentity Marshal err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}

	jsonResp = string(byteIdentity)
	LogMessage(INFO, "IdentityChaincode editIdentity editStatus is OK with response:", jsonResp)
	return shim.Success([]byte(ValResp(204, string(byteIdentity))))

}

//getIdentity: the input json must contain the KEY of the object Identity to SEARCH, [Identity = BaseEntry + StatusEntry]
//The caller must be a ANY USER!!!

func (t *IdentityChaincode) getIdentity(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	LogMessage(INFO, "IdentityChaincode getIdentity: ", time.Now().Format(time.RFC3339))
	LogMessage(INFO, "IdentityChaincode ON getIdentity with Address: ", args[0])

	var jsonResp string
	var err error
	var baseEntry BaseEntry = BaseEntry{}
	var statusEntry StatusEntry = StatusEntry{}
	var identity Identity = Identity{}
	var byteIdentity []byte
	var datetime time.Time
	var found bool
	var role string
	var checkControl bool
	var desc string

	// Control args[]
	if len(args) != 1 && len(args) != 2 {
		LogMessage(ERROR, "IdentityChaincode getIdentity numbers of Args not correctly!: ", strconv.Itoa(len(args)))
		return shim.Error(ValResp(400, "IdentityChaincode getIdentity ERROR: numbers of Args not correctly: "+strconv.Itoa(len(args))))
	}

	baseEntry.Address = args[0]
	if caAccess {
		// Control Identity into CA and check the authorization to getIdentity
		// GetAttributeValue(ROLE) - notfound is KO
		found, role, err = isInvokerOperator(stub, ROLE)
		if err != nil {
			LogMessage(ERROR, "getIdentity error: isInvokerOperator(role) err not nil", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}
		if !found {
			LogMessage(ERROR, "getIdentity error: ROLE is empty!!!", "")
			return shim.Error(ValResp(404, "MessagingBoardChaincode getIdentity ERROR: ROLE is empty!!!"))
		}
		LogMessage(DEBUG, "getIdentity error: isInvokerOperator by role: ", role)
		if role != ADMIN && role != USER {
			LogMessage(ERROR, "getIdentity error: ROLE is not authorized to get an Identity!!", role)
			return shim.Error(ValResp(401, "IdentityChaincode getIdentity ERROR: ROLE is not authorized to get an Identity!!"))
		}
	}

	baseEntry, err = GetBaseEntry(stub, baseEntry.Address)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode getIdentity GetBaseEntry err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}
	if (baseEntry == BaseEntry{}) {
		LogMessage(ERROR, "ProtectIdentityChaincode getIdentity GetBaseEntry DATA NOT FOUND", args[0])
		return shim.Error(ValResp(404, "ProtectIdentityChaincode getIdentity GetBaseEntry DATA NOT FOUND"))
	}

	identity.BaseEntry = baseEntry

	if len(args) == 1 {
		datetime = time.Now()
	} else { // 2 arguments (Address and Date) -> return the version valid at the Date
		datetime, err = time.Parse(time.RFC3339, args[1])
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode getIdentity time.Parse args[1] Data Format is not correct", args[1])
			return shim.Error(ValResp(400, err.Error()))
		}
	}

	statusEntry, err = GetStatusEntry(stub, baseEntry.Address, datetime)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode getIdentity GetStatusEntry err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}

	if (statusEntry == StatusEntry{}) {
		LogMessage(ERROR, "ProtectIdentityChaincode getIdentity GetStatusEntry DATA NOT FOUND", args[0])
		return shim.Error(ValResp(404, "ProtectIdentityChaincode getIdentity GetStatusEntry DATA NOT FOUND"))
	}
	identity.StatusEntry = statusEntry

	if statusEntry.Status != ACTIVESTATUS {
		LogMessage(ERROR, "IdentityChaincode getIdentity STATUS of IDENTITY is not ACTIVE: ", strconv.Itoa(statusEntry.Status))
		return shim.Error(ValResp(400, "getIdentity error: invalid identity (invalid status)"+strconv.Itoa(statusEntry.Status)))
	}
	if identity.BaseEntry.Address != identity.BaseEntry.Controller {
		checkControl, desc, err = ControlChainOfTrust(stub, identity.BaseEntry.Controller, datetime)

		if err != nil {
			LogMessage(ERROR, "IdentityChaincode getIdentity ControlChainOfTrust err not nil", err.Error())
			return shim.Error(ValResp(400, err.Error()))
		}

		if !checkControl {
			LogMessage(ERROR, "getIdentity error: invalid identity  ", desc)
			return shim.Error(ValResp(400, "editIdentity error: invalid identity : "+desc))
		}
	} else {
		LogMessage(INFO, "IdentityChaincode getIdentity getState is OK and the SID is a ROOT", identity.BaseEntry.Address)
	}

	byteIdentity, err = json.Marshal(identity)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode getIdentity Marshal err not nil", err.Error())
		return shim.Error(ValResp(400, err.Error()))
	}

	jsonResp = string(byteIdentity)
	LogMessage(INFO, "IdentityChaincode getIdentity getState is OK with response:", jsonResp)
	return shim.Success([]byte(ValResp(200, string(byteIdentity))))

	// return shim.Success(byteIdentity)
}

// **********************************************************************
// UTILITY
// **********************************************************************

// LogMessage System Manual Logger
func LogMessage(activity string, message string, parameter string) {

	fmt.Println(time.Now().Format(time.RFC3339)+" "+activity+" MESSAGE: "+message, parameter)

}

// validateRsaSignature validate a pkeyblob signed with the public key relative at the signature
func validateRsaSignature(pbKey *rsa.PublicKey, pKeyBlob string, sign string) (bool, error) {
	hash := sha256.Sum256([]byte(pKeyBlob))
	// validate the public key with the signature
	err := rsa.VerifyPKCS1v15(pbKey, crypto.SHA256, hash[:], []byte(sign))
	if err != nil {
		LogMessage(ERROR, "postIdentity error: controller authentication failed", "")
		LogMessage(ERROR, "IdentityChaincode PostIdentity Validate of Signature - pbKey    : ", strconv.Itoa(pbKey.E)+" "+pbKey.N.String())
		LogMessage(ERROR, "IdentityChaincode PostIdentity Validate of Signature - pKeyBlob : ", pKeyBlob)
		LogMessage(ERROR, "IdentityChaincode PostIdentity Validate of Signature - signature: ", sign)
		return false, err
	}
	return true, nil
}

// validateEcdsaSignature validate a pkeyblob signed with the public key relative at the signature
func validateEcdsaSignature(pbKey *ecdsa.PublicKey, pKeyBlob string, sign Signature) bool {
	hash := sha256.Sum256([]byte(pKeyBlob))
	var esigR = new(big.Int)
	var esigS = new(big.Int)
	var ok bool
	esigR, ok = esigR.SetString(sign.R, 0)
	esigS, ok = esigS.SetString(sign.S, 0)
	if !ok {
		LogMessage(ERROR, "IdentityChaincode PostIdentity Error in signature: ", sign.R+" "+sign.S)
		return ok
	}
	// validate the public key with the signature
	ok = ecdsa.Verify(pbKey, hash[:], esigR, esigS)
	if !ok {
		LogMessage(ERROR, "postIdentity error: controller authentication failed", "")
		LogMessage(ERROR, "IdentityChaincode PostIdentity Validate of Signature - pbKey   : ", pbKey.X.String()+" "+pbKey.Y.String())
		LogMessage(ERROR, "IdentityChaincode PostIdentity Validate of Signature - pKeyBlob: ", pKeyBlob)
		LogMessage(ERROR, "IdentityChaincode PostIdentity Validate of Signature - sign.R  : ", sign.R)
		LogMessage(ERROR, "IdentityChaincode PostIdentity Validate of Signature - sign.S  : ", sign.S)
	}
	return ok
}

// createAddressFromPKeyBlob creates a Addess from pKeyBlob input parameters
func createAddressFromPKeyBlob(pKeyBlob string) string {
	byteAddrCalc32 := sha256.Sum256([]byte(pKeyBlob))
	return base58.Encode([]byte(byteAddrCalc32[:]))
}

// hexToRSAPublicKey creates a public key rsa from e and n parameters
func hexToRSAPublicKey(rsaPK RSApublicKeyEN) (*rsa.PublicKey, error) {
	nBytes, _ := hex.DecodeString(rsaPK.N)
	n := new(big.Int)
	n.SetBytes(nBytes)

	e, err := strconv.Atoi(rsaPK.E)
	if err != nil {
		return nil, err
	}

	pub := new(rsa.PublicKey)
	pub.E = e
	pub.N = n

	return pub, nil

}

// hexToPublicKey creates a public key ecdsa from x and y parameters
func hexToPublicKey(ecdsaSign ECDSApublicKeyXY) *ecdsa.PublicKey {
	xBytes, _ := hex.DecodeString(ecdsaSign.X)
	x := new(big.Int)
	x.SetBytes(xBytes)

	yBytes, _ := hex.DecodeString(ecdsaSign.Y)
	y := new(big.Int)
	y.SetBytes(yBytes)

	pub := new(ecdsa.PublicKey)
	pub.X = x
	pub.Y = y

	pub.Curve = elliptic.P256()

	return pub

}

// GetBaseEntry Gate BaseEntry of a Identity by Address
func GetBaseEntry(stub shim.ChaincodeStubInterface, addr string) (BaseEntry, error) {
	var BE BaseEntry = BaseEntry{}
	var err error
	var IDKey string
	var byteBaseEntry []byte

	// GetBaseEntry with this Address
	// Create the KEY
	IDKey, err = stub.CreateCompositeKey(BASEENTRY, []string{addr})
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode GetBaseEntry CreateCompositeKey err not nil", err.Error())
		return BE, err
	}

	LogMessage(DEBUG, "IdentityChaincode GetBaseEntry for the key:", IDKey)
	byteBaseEntry, err = stub.GetState(IDKey)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode GetBaseEntry GetState err not nil", err.Error())
		return BE, err
	}
	if byteBaseEntry == nil {
		LogMessage(WARNING, "IdentityChaincode GetBaseEntry response is nil", "")
		return BE, nil

	}
	err = json.Unmarshal(byteBaseEntry, &BE)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode GetBaseEntry Unmarshal err not nil", err.Error())
		return BE, err
	}
	LogMessage(INFO, "IdentityChaincode GetBaseEntry response is OK", "")
	return BE, nil
}

// GetStatusEntry Gate StatusEntry of a Identity by Address and Time
func GetStatusEntry(stub shim.ChaincodeStubInterface, addr string, datetime time.Time) (StatusEntry, error) {
	var SE StatusEntry = StatusEntry{}
	var err error
	var statusEntries []StatusEntry
	var validFrom time.Time

	// Search statusEntry
	statusEntries, err = GetStateByPartialCompositeKeyStatusEntry(stub, []string{addr})
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode GetStatusEntry GetStateByPartialCompositeKeyStatusEntry err not nil", err.Error())
		return SE, err
	}
	if len(statusEntries) == 0 || statusEntries == nil {
		LogMessage(ERROR, "IdentityChaincode GetStatusEntry GetStateByPartialCompositeKeyStatusEntry DATA NOT FOUND", addr)
		return SE, nil
	}

	// loop for search the version valid a the date
	validFrom, err = time.Parse(time.RFC3339, statusEntries[0].ValidFrom)

	LogMessage(DEBUG, "IdentityChaincode GetStatusEntry COMPARE ValidFrom[0] Data Format is: ", validFrom.String())
	LogMessage(DEBUG, "IdentityChaincode GetStatusEntry COMPARE datetime par Data Format is: ", datetime.String())

	if err != nil {
		LogMessage(ERROR, "IdentityChaincode GetStatusEntry time.Parse ValidFrom[0] Data Format is not correct", statusEntries[0].ValidFrom)
		return SE, err
	}

	if validFrom.After(datetime) { // datetime is lower all ValidFrom of statusEntry (first date validity)
		SE = StatusEntry{}
		LogMessage(WARNING, "IdentityChaincode GetStatusEntry First statusEntry is greater of datatime parameter: ", datetime.String())
	} else {
		SE = statusEntries[0]

		for i := 1; i < len(statusEntries); i++ {

			validFrom, err = time.Parse(time.RFC3339, statusEntries[i].ValidFrom)
			if err != nil {
				LogMessage(ERROR, "IdentityChaincode GetStatusEntry ParseAny ValidFrom[i] Data Format is not correct", statusEntries[i].ValidFrom)
				return SE, err
			}

			if validFrom.After(datetime) {
				SE = statusEntries[i-1]
				break
			} else {
				SE = statusEntries[i]
			}
		}
	}
	// TEST NOT FOUND
	if SE == (StatusEntry{}) {
		LogMessage(WARNING, "IdentityChaincode GetStatusEntry Loop for data validity - DATA NOT FOUND for date: ", datetime.String())
	}
	return SE, nil
}

// ControlChainOfTrust System of control for the controller of Identity
func ControlChainOfTrust(stub shim.ChaincodeStubInterface, accessAddress string, time time.Time) (bool, string, error) {
	var desc string = ""
	var BE BaseEntry = BaseEntry{}
	var SE StatusEntry = StatusEntry{}
	// var accessAddress string = ""
	var err error
	var i int

	i = 0
	// The first time i'm sure that Address is not Controller (v. Call)
	BE.Address = ""
	BE.Controller = accessAddress
	// The first time i'm sure that Status is Active (v. Call)
	SE.Status = ACTIVESTATUS

	for BE.Address != BE.Controller && SE.Status == ACTIVESTATUS && i < NUMMAXGEN {
		i++
		accessAddress = BE.Controller
		BE, err = GetBaseEntry(stub, accessAddress)
		if err != nil {
			BE, err = GetBaseEntry(stub, accessAddress)
			LogMessage(ERROR, "IdentityChaincode ControlChainOfTrust GetBaseEntry err not nil", err.Error())
			desc = err.Error()
			break
		}
		SE, err = GetStatusEntry(stub, accessAddress, time)
		if err != nil {
			LogMessage(ERROR, "(chain of trust would exceed length limit)", err.Error())
			desc = err.Error()
			break
		}
	}
	if err != nil {
		return false, desc, err
	}
	if i >= NUMMAXGEN {
		LogMessage(ERROR, "(chain of trust would exceed length limit): ", SE.Address)
		desc = "(chain of trust would exceed length limit): " + SE.Address
		return false, desc, nil
	}
	if SE.Status != ACTIVESTATUS {
		LogMessage(ERROR, "(controller is not active or chain of trust is broken): ", SE.Address+" - "+(strconv.Itoa(SE.Status)))
		desc = "(controller is not active or chain of trust is broken): " + SE.Address + " - " + (strconv.Itoa(SE.Status))
		return false, desc, nil
	}
	if BE.Address == BE.Controller {
		LogMessage(INFO, "IdentityChaincode ControlChainOfTrust RETURN IS OK for: ", SE.Address)
		desc = "IdentityChaincode ControlChainOfTrust Found identity Controller Root: " + SE.Address
		return true, desc, nil
	}
	// Never
	desc = "IdentityChaincode ControlChainOfTrust There is some problem with condition exit of loop for " + SE.Address
	return false, desc, nil
}

// GetStateByPartialCompositeKeyStatusEntry return a array of StatusEntry by key
func GetStateByPartialCompositeKeyStatusEntry(stub shim.ChaincodeStubInterface, KEYid []string) ([]StatusEntry, error) {

	var buffer bytes.Buffer
	var err error
	var statusEntry StatusEntry
	var statusEntries []StatusEntry

	resultsIterator, err := stub.GetStateByPartialCompositeKey(STATUSENTRY, KEYid)
	if err != nil {
		LogMessage(ERROR, "IdentityChaincode GetStateByPartialCompositeKeyStatusEntry err not nil", err.Error())
		return nil, err
	}

	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults

	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	var i int

	for i = 0; resultsIterator.HasNext(); i++ {
		responseRange, err := resultsIterator.Next()
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode Into Loop  resultsIterator.Next() for err not nil", strconv.Itoa(i))
			return nil, err
		}

		LogMessage(DEBUG, "IdentityChaincode Into Loop pre stub.SplitCompositeKey", responseRange.Key)
		_, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode Into Loop  stub.SplitCompositeKey for err not nil", responseRange.Key)
			return nil, err
		}

		keyComplete, err := stub.CreateCompositeKey(STATUSENTRY, compositeKeyParts)

		if err != nil {
			LogMessage(ERROR, "IdentityChaincode GetStateByPartialCompositeKeyStatusEntry CreateCompositeKey err not nil", err.Error())
			return nil, err
		}

		LogMessage(DEBUG, "IdentityChaincode Into Loop post stub.SplitCompositeKey", keyComplete)

		// idAsBytes, err := stub.GetState(idReturn)
		idAsBytes, err := stub.GetState(keyComplete)
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode GetStateByPartialCompositeKeyStatusEntry GetState: ", err.Error())
			return nil, err
		}
		if idAsBytes == nil {
			LogMessage(ERROR, "IdentityChaincode GetStateByPartialCompositeKeyStatusentry GetState DATA NOT FOUND for KEY: ", keyComplete)
		}

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(keyComplete)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(idAsBytes))

		err = json.Unmarshal(idAsBytes, &statusEntry)
		if err != nil {
			LogMessage(ERROR, "IdentityChaincode GetStateByPartialCompositeKeyMessage Unmarshal StatusEntry", err.Error())
			return nil, err
		}
		statusEntries = append(statusEntries, statusEntry)

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	sort.Slice(statusEntries, func(i, j int) bool {
		return statusEntries[i].ValidFrom < statusEntries[j].ValidFrom
	})

	return statusEntries, nil

}

// **********************************************************************
// END
// **********************************************************************

// main function starts up the chaincode in the container during instantiate
func main() {
	fmt.Println("---MAIN SECURE IDENTITY CONTRACT---")
	fmt.Println("---Cert. Auth. ACCESS ENABLED---")
	caAccess = false
	if err := shim.Start(new(IdentityChaincode)); err != nil {

		// log.Debug("Main()")
		fmt.Printf("Error starting Contract chaincode: %s", err)
	}
}
