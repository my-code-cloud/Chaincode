package engine

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/s7techlab/cckit/convert"

	"github.com/KompiTech/fabric-cc-core/v2/pkg/kompiguard"
	. "github.com/KompiTech/fabric-cc-core/v2/pkg/konst"
	"github.com/KompiTech/rmap"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pkg/errors"
)

// ContextInterface is extension of contractapi.TransactionContextInterface
type ContextInterface interface {
	contractapi.TransactionContextInterface
	contractapi.SettableTransactionContextInterface
	SetConfFunc(func() Configuration)
	GetConfiguration() Configuration
	GetRegistry() *Registry
	Get(key string) interface{}
	Set(key string, value interface{})
	Stub() shim.ChaincodeStubInterface
	Time() (time.Time, error)
	Params() map[string]interface{}
	Param(name string) (interface{}, error)
	ParamString(name string) (string, error)
	ParamBytes(name string) ([]byte, error)
	ParamInt(name string) (int, error)
	ParamBool(name string) (bool, error)
	Logger() Logger
}

type IDFunc func(cert *x509.Certificate) (string, error)

// Configuration is used to manage configuration of this particular dynamic chaincode (previously called Engine)
type Configuration struct {
	BusinessExecutor          BusinessExecutor // Abstraction for executing business logic
	FunctionExecutor          FunctionExecutor // Abstraction for executing generic functionality
	RecursiveResolveWhitelist rmap.Rmap        // Rmap of asset names that have recursive resolve enabled
	ResolveBlacklist          rmap.Rmap        // Rmap of asset names that are forbidden from being resolved
	ResolveFieldsBlacklist    rmap.Rmap        // Rmap of asset name -> list of fields to not resolve
	CurrentIDFunc             IDFunc           // Function to get identity fingerprint
	PreviousIDFunc            *IDFunc          // Previous function to get identity fingerprint when migration is desired

	// SchemaDefinitionCompatibility is legacy setting, to allow the chaincode to work with older JSONSchemas (draft-07 and older) that are using reusable definitions.
	// Previously, any location for the definitions can be used, but JSONSchema newer than draft-07 allows only "$defs" key to be used.
	// To allow chaincode to work with these older schemas, set the value of SchemaDefinitionCompatibility member to name under which the definitions are stored in schema.
	// Chaincode will then convert its key to standard $defs key to allow use of newer JSONSchema library.
	// These replacements are done only at runtime and not persisted anywhere.
	// If SchemaDefinitionCompatibility is empty string, then no replacements are done.
	// This is the default setting, as the issue only occurs with legacy schemas.
	SchemaDefinitionCompatibility string
}

// HTTPErr is an error with custom HTTP status code
type HTTPErr struct {
	code int
	err  error
}

// BusinessPolicyMember is signature for business logic function
type BusinessPolicyMember = func(ctx ContextInterface, prePatch *rmap.Rmap, postPatch rmap.Rmap) (rmap.Rmap, error)

// FuncKey is the key for policy member - it defines the asset name and its version to execute on
// version -1 has special meaning -> wildcard -> execute on ALL versions (do not use in production or the migrations of business logic will be impossible)
type FuncKey struct {
	Name    string
	Version int
}

// businessPolicy defines which functions should be executed on what asset versions and when
// level 1: {assetName, assetVersion}
// level 2: Stage (execution stage)
// level 3: list of functions to execute in this order
type businessPolicy = map[FuncKey]map[Stage][]BusinessPolicyMember

// BusinessExecutor holds configuration for business logic policy and can execute it on asset instance
type BusinessExecutor struct {
	policy businessPolicy
}

type RegistryInterface interface {
	BulkUpsertItems(items []bulkItem) error
	UpsertItem(registryItemToUpsert Rmap, assetName string) (Rmap, int, error)
	GetThisIdentity() (Rmap, error)
	GetThisIdentityResolved() (Rmap, error)
	GetItem(name string, requestedVersion int) (Rmap, int, error)
	MakeAsset(name, id string, version int) (Rmap, error)
	MarkAssetAsExisting(name, id string, data Rmap) error
	GetAsset(name, id string, resolve bool, failOnNotFound bool) (Rmap, error)
	ListItems() ([]string, error)
	ListSingletons() ([]string, error)
	PutAsset(asset Rmap, isCreate bool) error
	GetQueryIterator(name string, query Rmap, bookmark string, pageSize int) (IteratorInterface, string, error)
	QueryAssets(name string, query Rmap, bookmark string, resolve bool, paginate bool, pageSize int) ([]Rmap, string, error)
	DeleteAsset(asset Rmap) error
	GetAssetHistory(asset Rmap) ([]Rmap, error)
	UpsertSingleton(singletonItemToUpsert Rmap, singletonName string) (int, error)
	BulkUpsertSingletons(items []bulkItem) error
	ExistsSingleton(name string, version int) (bool, error)
	ExistsAsset(name, id string) (bool, error)
	GetSingleton(name string, version int) (Rmap, int, error)
}

// FunctionPolicyMember is signature for all functions in policy
type FunctionPolicyMember = func(ctx ContextInterface, input rmap.Rmap, output rmap.Rmap) (rmap.Rmap, error)

// FunctionPolicy is definition of list of functions to execute for some name
type FunctionPolicy map[string][]FunctionPolicyMember

// FunctionExecutor holds configuration of callable functions
type FunctionExecutor struct {
	policy FunctionPolicy
}

type (
	// FromBytesTransformer is used after getState operation for convert value
	FromBytesTransformer func(bb []byte, config ...interface{}) (interface{}, error)

	// ToBytesTransformer is used before putState operation for convert payload
	ToBytesTransformer func(v interface{}, config ...interface{}) ([]byte, error)

	// KeyTransformer is used before putState operation for convert key
	KeyTransformer func(Key) (Key, error)

	// StringTransformer is used before setEvent operation for convert name
	StringTransformer func(string) (string, error)

	Serializer interface {
		ToBytes(interface{}) ([]byte, error)
		FromBytes(serialized []byte, target interface{}) (interface{}, error)
	}

	ProtoSerializer struct {
	}

	JSONSerializer struct {
	}
)

func ConvertFromBytes(bb []byte, config ...interface{}) (interface{}, error) {
	// conversion is not needed
	if len(config) == 0 {
		return bb, nil
	}
	return convert.FromBytes(bb, config[0])
}

func ConvertToBytes(v interface{}, _ ...interface{}) ([]byte, error) {
	return convert.ToBytes(v)
}

// KeyAsIs returns string parts of composite key
func KeyAsIs(key Key) (Key, error) {
	return key, nil
}

func NameAsIs(name string) (string, error) {
	return name, nil
}

func (ps *ProtoSerializer) ToBytes(entry interface{}) ([]byte, error) {
	return proto.Marshal(entry.(proto.Message))
}

func (ps *ProtoSerializer) FromBytes(serialized []byte, target interface{}) (interface{}, error) {
	return convert.FromBytes(serialized, target)
}

func (js *JSONSerializer) ToBytes(entry interface{}) ([]byte, error) {
	return json.Marshal(entry)
}

func (js *JSONSerializer) FromBytes(serialized []byte, target interface{}) (interface{}, error) {
	return convert.JSONUnmarshalPtr(serialized, target)
}

// Error creates new initialized HttpError
func Error(code int, err error) HTTPErr {
	return HTTPErr{
		code: code,
		err:  err,
	}
}

// ErrorBadRequest returns new HttpError with HTTP status code 400
func ErrorBadRequest(message string) HTTPErr {
	return Error(http.StatusBadRequest, fmt.Errorf(message))
}

// ErrorForbidden returns new HttpError with HTTP status code 403
func ErrorForbidden(message string) HTTPErr {
	return Error(http.StatusForbidden, fmt.Errorf(message))
}

// ErrorNotFound returns new HttpError with HTTP status code 404
func ErrorNotFound(message string) HTTPErr {
	return Error(http.StatusNotFound, fmt.Errorf(message))
}

// ErrorConflict returns new HttpError with HTTP status code 409
func ErrorConflict(message string) HTTPErr {
	return Error(http.StatusConflict, fmt.Errorf(message))
}

// ErrorUnprocessableEntity returns new HttpError with HTTP status code 422
func ErrorUnprocessableEntity(message string) HTTPErr {
	return Error(http.StatusUnprocessableEntity, fmt.Errorf(message))
}

// Error implements error interface
func (e HTTPErr) Error() string {
	return fmt.Sprintf("%s|||%d", e.err.Error(), e.code)
}

// this file contains different pieces of repeated code for use in engine package only
// everything here should be private (lowercase first letter)!
func newRmapFromDestination(ctx ContextInterface, docType, key, destination string, failOnNotFound bool) (rmap.Rmap, error) {
	if destination == StateDestinationValue {
		return newRmapFromState(ctx, key, failOnNotFound)
	}
	return newRmapFromPrivateData(ctx, docType, key, failOnNotFound)
}

func newRmapFromPrivateData(ctx ContextInterface, collectionName, key string, failOnNotFound bool) (rmap.Rmap, error) {
	collectionName = strings.ToUpper(collectionName)
	assetBytes, err := ctx.Stub().GetPrivateData(collectionName, key)
	if err != nil {
		return rmap.Rmap{}, errors.Wrap(err, "r.ctx.Stub().GetPrivateData() failed")
	}

	if len(assetBytes) == 0 {
		if failOnNotFound {
			return rmap.Rmap{}, ErrorNotFound(fmt.Sprintf("private data entry not found, collection: %s, key: %s", collectionName, strings.Replace(key, "\x00", "", -1)))
		}
		return rmap.NewEmpty(), nil
	}

	return rmap.NewFromBytes(assetBytes)
}

// Helper to create Rmap from State
// Do not want to make rmap dependent on cckit
func newRmapFromState(ctx ContextInterface, key string, failOnNotFound bool) (rmap.Rmap, error) {
	assetBytes, err := ctx.Stub().GetState(key)
	if err != nil {
		return rmap.Rmap{}, errors.Wrap(err, "r.ctx.Stub().GetState() failed")
	}

	if len(assetBytes) == 0 {
		if failOnNotFound {
			return rmap.Rmap{}, ErrorNotFound(fmt.Sprintf("state entry not found: %s", strings.Replace(key, "\x00", "", -1)))
		}
		return rmap.NewEmpty(), nil
	}

	return rmap.NewFromBytes(assetBytes)
}

func putRmapToPrivateData(ctx ContextInterface, collectionName, key string, isCreate bool, rm rmap.Rmap) error {
	collectionName = strings.ToUpper(collectionName)
	// get existing data. If key does not exists, length is 0
	existingData, err := ctx.Stub().GetPrivateData(collectionName, key)
	if err != nil {
		return errors.Wrap(err, "ctx.Stub().GetPrivateData() failed")
	}

	if isCreate && len(existingData) != 0 {
		// when creating, it is error if key already exists
		return ErrorConflict(fmt.Sprintf("private data key already exists: %s", strings.Replace(key, ZeroByte, "", -1)))
	} else if !isCreate && len(existingData) == 0 {
		// when updating, it is error if key does not exists
		return ErrorConflict(fmt.Sprintf("attempt to update non-existent private data key: %s", strings.Replace(key, ZeroByte, "", -1)))
	}

	return ctx.Stub().PutPrivateData(collectionName, key, rm.Bytes())
}

func putRmapToState(ctx ContextInterface, key string, isCreate bool, rm rmap.Rmap) error {
	// get existing data. If key does not exists, length is 0
	existingData, err := ctx.Stub().GetState(key)
	if err != nil {
		return errors.Wrap(err, "ctx.Stub().GetState() failed")
	}

	if isCreate && len(existingData) != 0 {
		// when creating, it is error if key already exists
		return ErrorConflict(fmt.Sprintf("state key already exists: %s", strings.Replace(key, ZeroByte, "", -1)))
	} else if !isCreate && len(existingData) == 0 {
		// when updating, it is error if key does not exists
		return ErrorConflict(fmt.Sprintf("attempt to update non-existent state key: %s", strings.Replace(key, ZeroByte, "", -1)))
	}

	return ctx.Stub().PutState(key, rm.Bytes())
}

// enforceCustomAccess loads identity for this and checks, if identity can do action on object
// this is used for methods where there is no related asset for inferring object
func enforceCustomAccess(reg *Registry, object, action string) error {
	thisIdentity, err := reg.GetThisIdentityResolved()
	if err != nil {
		return errors.Wrap(err, "reg.GetThisIdentityResolved() failed")
	}

	subject, err := AssetGetID(thisIdentity)
	if err != nil {
		return errors.Wrap(err, "konst.AssetGetID(thisIdentity) failed")
	}

	kmpg, err := kompiguard.New()
	if err != nil {
		return errors.Wrap(err, "kompiguard.New() failed")
	}

	if err := kmpg.LoadRoles(thisIdentity); err != nil {
		return errors.Wrap(err, "kmpg.LoadRoles() failed")
	}

	granted, reason, err := kmpg.EnforceCustom(object, subject, action, nil)
	if err != nil {
		return errors.Wrap(err, "kompiguard.New().EnforceCustom() failed")
	}

	if !granted {
		return ErrorForbidden(reason)
	}

	return nil
}

// enforceAssetAccess loads identity for this and enforces standard action for some asset
func enforceAssetAccess(reg *Registry, asset rmap.Rmap, action string) error {
	thisIdentity, err := reg.GetThisIdentityResolved()
	if err != nil {
		return errors.Wrap(err, "reg.GetThisIdentityResolved() failed")
	}

	kmpg, err := kompiguard.New()
	if err != nil {
		return errors.Wrap(err, "kompiguard.New() failed")
	}

	granted, reason, err := kmpg.EnforceAsset(asset, thisIdentity, action)
	if err != nil {
		return errors.Wrap(err, "kompiguard.New().EnforceAsset() failed")
	}

	if !granted {
		return ErrorForbidden(reason)
	}

	return nil
}

func GetMyFingerprint(ctx ContextInterface) (string, error) {
	myCert, err := cid.GetX509Certificate(ctx.Stub())
	if err != nil {
		return "", errors.Wrap(err, "cid.GetX509Certificate() failed")
	}

	//call function instead of using hardcoded SHA512 as before
	return ctx.GetConfiguration().CurrentIDFunc(myCert)
}

func GetMyOrgName(ctx ContextInterface) (string, error) {
	myCert, err := cid.GetX509Certificate(ctx.Stub())
	if err != nil {
		return "", err
	}

	//testing certs have issuer CN identical to subject CN ({user_id}.{org_id}.kompitech.com)
	//real certs have issuer CN: {org_id}.kompitech.com, subject CN: {user_id}.{org_id}.kompitech.com
	//this code will always return {org_id}.kompitech.com in both cases
	fields := strings.Split(myCert.Issuer.CommonName, ".")

	if len(fields) == 3 { // {org_id}.kompitech.com , no transformation needed
		return myCert.Issuer.CommonName, nil
	} else if len(fields) == 4 { // {user_id}.{org_id}.kompitech.com , remove {user_id}
		return strings.Join(fields[1:], "."), nil
	} else {
		return "", fmt.Errorf("unexpected cert.Issuer.CommonName: %s", myCert.Issuer.CommonName)
	}
}

func MakeUUID(now time.Time) (string, error) {
	rand.Seed(now.UnixNano()) // seed RNG with this TX time, this will make all peers to generate the same UUID

	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return strings.ToLower(fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])), nil
}

// GetUUIDGeneratingFunc returns func that returns unique UUID and returned func can be called repeatedly in one transaction
func GetUUIDGeneratingFunc() (f func(ctx ContextInterface) (string, error)) {
	var i int

	return func(ctx ContextInterface) (string, error) {
		now, err := ctx.Time()
		if err != nil {
			return "", err
		}

		uuid, err := MakeUUID(now.Add(time.Duration(i) * time.Nanosecond))
		if err != nil {
			return "", err
		}

		i++

		return uuid, nil
	}
}

// GetThisIdentityResolved returns identity asset for current user and resolves all roles it has
func (r Registry) GetThisIdentityResolved() (Rmap, error) {
	thisIdentity, err := r.GetThisIdentity()
	if err != nil {
		return Rmap{}, errors.Wrap(err, "r.GetThisIdentity() failed")
	}

	if thisIdentity.Exists(IdentityRolesKey) {
		// manual resolve of roles, to prevent infinite recursion if some user business logic does the same
		roles, err := thisIdentity.GetIterable(IdentityRolesKey)
		if err != nil {
			return Rmap{}, errors.Wrap(err, "thisIdentity.GetIterable() failed")
		}

		for roleIndex, roleI := range roles {
			roleAsset, err := r.GetAsset(RoleAssetName, roleI.(string), false, true)
			if err != nil {
				return Rmap{}, errors.Wrap(err, "reg.GetAsset() failed")
			}

			// set resolved role asset
			if err := thisIdentity.SetJPtr("/"+IdentityRolesKey+"/"+strconv.Itoa(roleIndex), roleAsset.Mapa); err != nil {
				return Rmap{}, errors.Wrap(err, "thisIdentity.SetJPtr() failed")
			}
		}
	}

	return thisIdentity, nil
}
