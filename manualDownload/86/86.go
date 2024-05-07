package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	logger "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
)

const (
	CREATE_DATA_CALL_FOR_UPDATE_JSON                     = `{"id": "Data_Call_123","version":"1","name":"Data_Call_123","intentToPublish":true,"IsLocked":false,"isLatest":true,"isShowParticipants":true,"description":"","purpose":"","lineOfBusiness":"","deadline":"2018-11-13T18:30:00.000Z","premiumFromDate":"2018-11-13T18:30:00.000Z","premiumToDate":"2018-11-13T18:30:00.000Z","lossFromDate":"2018-11-13T18:30:00.000Z","lossToDate":"2018-11-13T18:30:00.000Z","jurisdiction":"","proposedDeliveryDate":"2018-11-13T18:30:00.000Z","updatedBy":"","updatedTs":"2018-11-13T18:30:00.000Z","detailedCriteria":"","eligibilityRequirement":"","status":"DRAFT","type":"","comments":"","forumURL":""}`
	CREATE_DATA_CALL_FOR_ISSUE_JSON                      = `{"id": "Data_Call_1234","version":"1","name":"Data_Call_1234","intentToPublish":true,"IsLocked":true,"isLatest":true,"isShowParticipants":true,"description":"","purpose":"","lineOfBusiness":"","deadline":"2018-11-13T18:30:00.000Z","premiumFromDate":"2018-11-13T18:30:00.000Z","premiumToDate":"2018-11-13T18:30:00.000Z","lossFromDate":"2018-11-13T18:30:00.000Z","lossToDate":"2018-11-13T18:30:00.000Z","jurisdiction":"","proposedDeliveryDate":"2018-11-13T18:30:00.000Z","updatedBy":"","updatedTs":"2018-11-13T18:30:00.000Z","detailedCriteria":"","eligibilityRequirement":"","status":"DRAFT","type":"","comments":"","forumURL":""}`
	CREATE_DATA_CALL_EMPTY_ID_JSON                       = `{"id": "","version":"1","name":"Data_Call_123","intentToPublish":true,"IsLocked":true,"isLatest":true,"isShowParticipants":true,"description":"","purpose":"","lineOfBusiness":"","deadline":"2018-11-13T18:30:00.000Z","premiumFromDate":"2018-11-13T18:30:00.000Z","premiumToDate":"2018-11-13T18:30:00.000Z","lossFromDate":"2018-11-13T18:30:00.000Z","lossToDate":"2018-11-13T18:30:00.000Z","jurisdiction":"","proposedDeliveryDate":"2018-11-13T18:30:00.000Z","updatedBy":"","updatedTs":"2018-11-13T18:30:00.000Z","detailedCriteria":"","eligibilityRequirement":"","status":"DRAFT","type":"","comments":"","forumURL":""}`
	UPDATE_DATA_CALL_VALID_JSON                          = `{"id": "Data_Call_123","version":"1","name":"Data_Call_123","intentToPublish":true,"IsLocked":true,"isLatest":true,"isShowParticipants":true,"description":"","purpose":"","lineOfBusiness":"","deadline":"2018-11-13T18:30:00.000Z","premiumFromDate":"2018-11-13T18:30:00.000Z","premiumToDate":"2018-11-13T18:30:00.000Z","lossFromDate":"2018-11-13T18:30:00.000Z","lossToDate":"2018-11-13T18:30:00.000Z","jurisdiction":"","proposedDeliveryDate":"2018-11-13T18:30:00.000Z","updatedBy":"","updatedTs":"2018-11-13T18:30:00.000Z","detailedCriteria":"","eligibilityRequirement":"","status":"CANCELLED","type":"","comments":"","forumURL":""}`
	UPDATE_DATA_CALL_EMPTY_ID_JSON                       = `{"id": "","version":"1","name":"Data_Call_123","intentToPublish":true,"IsLocked":true,"isLatest":true,"isShowParticipants":true,"description":"","purpose":"","lineOfBusiness":"","deadline":"2018-11-13T18:30:00.000Z","premiumFromDate":"2018-11-13T18:30:00.000Z","premiumToDate":"2018-11-13T18:30:00.000Z","lossFromDate":"2018-11-13T18:30:00.000Z","lossToDate":"2018-11-13T18:30:00.000Z","jurisdiction":"","proposedDeliveryDate":"2018-11-13T18:30:00.000Z","updatedBy":"","updatedTs":"2018-11-13T18:30:00.000Z","detailedCriteria":"","eligibilityRequirement":"","status":"DRAFT","type":"","comments":"","forumURL":""}`
	ISSUE_DATA_CALL_VALID_JSON                           = `{"id": "Data_Call_1234","version":"1","name":"Data_Call_1234","intentToPublish":true,"IsLocked":true,"isLatest":true,"isShowParticipants":true,"description":"","purpose":"","lineOfBusiness":"","deadline":"2018-11-13T18:30:00.000Z","premiumFromDate":"2018-11-13T18:30:00.000Z","premiumToDate":"2018-11-13T18:30:00.000Z","lossFromDate":"2018-11-13T18:30:00.000Z","lossToDate":"2018-11-13T18:30:00.000Z","jurisdiction":"","proposedDeliveryDate":"2018-11-13T18:30:00.000Z","updatedBy":"","updatedTs":"2018-11-13T18:30:00.000Z","detailedCriteria":"","eligibilityRequirement":"","status":"ISSUED","type":"","comments":"","forumURL":""}`
	ISSUE_DATA_CALL_EMPTY_ID_JSON                        = `{"id": "","version":"1","name":"Data_Call_123","intentToPublish":true,"IsLocked":true,"isLatest":true,"isShowParticipants":true,"description":"","purpose":"","lineOfBusiness":"","deadline":"2018-11-13T18:30:00.000Z","premiumFromDate":"2018-11-13T18:30:00.000Z","premiumToDate":"2018-11-13T18:30:00.000Z","lossFromDate":"2018-11-13T18:30:00.000Z","lossToDate":"2018-11-13T18:30:00.000Z","jurisdiction":"","proposedDeliveryDate":"2018-11-13T18:30:00.000Z","updatedBy":"","updatedTs":"2018-11-13T18:30:00.000Z","detailedCriteria":"","eligibilityRequirement":"","status":"ISSUED","type":"","comments":"","forumURL":""}`
	SAVE_NEW_DRAFT_JSON                                  = `{"id": "Data_Call_123","version":"1","name":"Data_Call_123","intentToPublish":true,"IsLocked":false,"isLatest":true,"isShowParticipants":true,"description":"","purpose":"","lineOfBusiness":"","deadline":"2018-11-13T18:30:00.000Z","premiumFromDate":"2018-11-13T18:30:00.000Z","premiumToDate":"2018-11-13T18:30:00.000Z","lossFromDate":"2018-11-13T18:30:00.000Z","lossToDate":"2018-11-13T18:30:00.000Z","jurisdiction":"","proposedDeliveryDate":"2018-11-13T18:30:00.000Z","updatedBy":"","updatedTs":"2018-11-13T18:30:00.000Z","detailedCriteria":"","eligibilityRequirement":"","status":"DRAFT","type":"","comments":"","forumURL":""}`
	SAVE_NEW_DRAFT_EMPTY_ID_JSON                         = `{"id": "","version":"1","name":"Data_Call_123","intentToPublish":true,"IsLocked":true,"isLatest":true,"isShowParticipants":true,"description":"","purpose":"","lineOfBusiness":"","deadline":"2018-11-13T18:30:00.000Z","premiumFromDate":"2018-11-13T18:30:00.000Z","premiumToDate":"2018-11-13T18:30:00.000Z","lossFromDate":"2018-11-13T18:30:00.000Z","lossToDate":"2018-11-13T18:30:00.000Z","jurisdiction":"","proposedDeliveryDate":"2018-11-13T18:30:00.000Z","updatedBy":"","updatedTs":"2018-11-13T18:30:00.000Z","detailedCriteria":"","eligibilityRequirement":"","status":"DRAFT","type":"","comments":"","forumURL":""}`
	GET_DATA_CALL_BY_ID_AND_VERSION_EMPTY_ID_JSON        = `{"id":"","version":"1"}`
	GET_DATA_CALL_VERSIONS_BY_ID_EMPTY_ID_JSON           = `{"id":"Data_Call_123","startIndex":0,"pageSize":0}`
	GET_DATA_CALL_VERSIONS_BY_ID_VALID_JSON              = `{"id":"Data_Call_123","startIndex":0,"pageSize":0}`
	GET_DATA_CALL_BY_ID_AND_VERSION_VALID_JSON           = `{"id":"Data_Call_123","version":"1"}`
	GET_DATA_CALL_BY_ID_AND_VERSION_FOR_ISSUE_VALID_JSON = `{"id":"Data_Call_1234","version":"1"}`
	GET_DATA_CALL_BY_ID_AND_VERSION_FOR_DRAFT_JSON       = `{"id":"Data_Call_123","version":"2"}`
	LIST_DATA_CALL_BY_CRITERIA_EMPTY_STATUS_JSON         = `{"status":"","version":"","startIndex":0,"pageSize":0}`
	CREATE_REPORT_EMPTY_ID_JSON                          = `{"dataCallID":"","dataCallVersion":"1","hash":"hash_1","status":"CANDIDATE","url":"","createdBy":"test","createdTs":"2018-11-14T10:10:04.535Z"}`
	CREATE_REPORT_VALID_JSON                             = `{"dataCallID":"Data_Call_1","dataCallVersion":"1","reportVersion":"1","hash":"hash_1","status":"CANDIDATE","url":"","createdBy":"test","createdTs":"2018-11-14T10:10:04.535Z","updatedTs":"2018-11-14T10:10:04.535Z"}`
	GET_REPORT_BY_ID_FOR_CREATE_JSON                     = `{"dataCallID":"Data_Call_1","dataCallVersion":"1","hash":"hash_1"}`
	UPDATE_REPORT_VALID_JSON                             = `{"dataCallID":"Data_Call_1","dataCallVersion":"1","reportVersion":"1","hash":"hash_1","isLocked":true,"status":"ACCEPTED","url":"","createdBy":"test","createdTs":"2018-11-14T10:10:04.535Z","updatedTs":"2018-11-14T10:10:04.535Z"}`
	UPDATE_REPORT_EMPTY_ID_JSON                          = `{"dataCallID":"","dataCallVersion":"1","hash":"hash_1","status":"ACCEPTED","url":"","createdBy":"test","createdTs":"2018-11-14T10:10:04.535Z"}`
	GET_REPORT_BY_CRITERIA_JSON                          = `{"dataCallID":"Data_Call_1","dataCallVersion":"1","status":""}`
	CREATE_DATA_CALL_LIKE_TEST                           = `{"id": "Data_Call_123455","version":"1","name":"Data_Call_123455","intentToPublish":true,"IsLocked":true,"isLatest":true,"isShowParticipants":true,"description":"","purpose":"","lineOfBusiness":"","deadline":"2018-11-13T18:30:00.000Z","premiumFromDate":"2018-11-13T18:30:00.000Z","premiumToDate":"2018-11-13T18:30:00.000Z","lossFromDate":"2018-11-13T18:30:00.000Z","lossToDate":"2018-11-13T18:30:00.000Z","jurisdiction":"","proposedDeliveryDate":"2018-11-13T18:30:00.000Z","updatedBy":"","updatedTs":"2018-11-13T18:30:00.000Z","detailedCriteria":"","eligibilityRequirement":"","status":"CANCELLED","type":"","comments":"","forumURL":""}`
	LIKE_JSON_MULTI_CARRIER                              = `{"datacallID":"Data_Call_123","dataCallVersion":"1","organizationType":"Carrier","OrganizationID":"12345","updatedTs":"2018-11-14T10:10:04.535Z","updatedBy":"user@user.com","liked":true}`
	LIKE_JSON_CARRIER1                                   = `{"datacallID":"Data_Call_123","dataCallVersion":"1","organizationType":"Carrier","organizationID":"12345","updatedTs":"2018-11-14T10:10:04.535Z","updatedBy":"user@user.com","liked":true}`
	UNLIKE_JSON_MULTI_CARRIER                            = `{"datacallID":"Data_Call_123","dataCallVersion":"1","organizationType":"Carrier","OrganizationID":"12345","updatedTs":"2018-11-14T10:10:04.535Z","updatedBy":"user@user.com","liked":false}`
	LIKE_ENTRY_JSON                                      = `{"datacallID":"Data_Call_123455","dataCallVersion":"1","delta":1,"updatedTs":"2018-11-14T10:10:04.535Z","liked":true}`
	COUNT_LIKES_JSON                                     = `{"datacallID":"Data_Call_123455","dataCallVersion":"1","updatedTs":"2018-11-14T10:10:04.535Z","liked":true}`
	CONSENT_TEST_DATA_WITHOUT_DELTA                      = `{"datacallID":"Data_Call_123","dataCallVersion":"1","carrierID":"12345","carrierName":"Hypermutual","createdTs":"2018-11-13T18:30:00.000Z","createdBy":"abc"}`
	CONSENT_TEST_DATA_DATA_CALL_DOES_NOT_EXIST           = `{"datacallID":"Data_Call_12345","dataCallVersion":"1","carrierID":"12345","carrierName":"Hypermutual","createdTs":"2018-11-13T18:30:00.000Z","createdBy":"abc"}`
	CONSENT_TEST_DATA_CARRIER1                           = `{"datacallID":"Data_Call_123","dataCallVersion":"1","carrierID":"12345","carrierName":"Hypermutual","createdTs":"2018-11-13T18:30:00.000Z","createdBy":"abc"}`
	CONSENT_TEST_DATA_MULTICARRIERs                      = `{"datacallID":"Data_Call_123","dataCallVersion":"1","carrierID":"12345","carrierName":"Hypermutual","createdTs":"2018-11-13T18:30:00.000Z","createdBy":"abc"}`
	CONSENT_TEST_DATA_WITH_DELTA                         = `{"datacallID":"Data_Call_123","dataCallVersion":"1","carrierID":"12345","carrierName":"Hypermutual","createdTs":"2018-11-13T18:30:00.000Z","createdBy":"abc","delta":1}`
	LIST_CONSENT_CRITERIA_JSON                           = `{"consent": {"datacallID": "Data_Call_123","dataCallVersion": "1", "carrierID":"12345"}, "channelIDs":["aais-carrier1"]}`
	LIST_CONSENT_CRITERIA_NEW_JSON                       = `{"datacallID": "Data_Call_123","dataCallVersion": "1", "channelList":[{"channelName":"aais-carrier1", "chaincodeName": "openidl-cc-default"}, {"channelName":"aais-carrier1", "chaincodeName": "openidl-cc-default"}]}`
	GET_LIKE_CRITERIA_JSON                               = `{"like": {"datacallID": "Data_Call_123","dataCallVersion": "1", "OrganizationID":"12345"}, "channelIDs":["aais-carrier1"]}`
	LIST_LIKE_CRITERIA_JSON                              = `{"datacallID": "Data_Call_123","dataCallVersion": "1", "channelList":[{"channelName":"aais-carrier1", "chaincodeName": "openidl-cc-default"}]}`
	CREATE_DATACALL_LOG_ENTRY                            = `{"dataCallID":"Data_Call_1","dataCallVersion":"1","actionID":"DATA_CALL_ISSUED","action":"Issued","actionTs":"2018-11-01T18:30:00.000Z","updatedBy":"user@aaisonline.com"}`
	SAVE_INSURANCE_HASH_EMPTY_ID_JSON                    = `{"batchId":"","hash":"test"}`
	SAVE_INSURANCE_HASH_VALID_JSON                       = `{"batchId":"Insurance-batch-1", "chunkId":"chunk1", "hash":"test","carrierId":"12345"}`
	SAVE_INSURANCE_DATA_VALID_JSON                       = `{"batchId":"batch_123","dataCallId":"Data_Call_123","dataCallVersion":"1","carrierId":"12345","pageNumber":1,"value":{"key1":"value1","key2":"value2","key3":"value3","key4":["val1","val2","val3"]}}`
	SAVE_INSURANCE_DATA_EMPTY_CARRIER_ID_JSON            = `{"batchId":"batch_123","dataCallId":"Data_Call_123","dataCallVersion":"1","carrierId":"","pageNumber":1,"value":{"key1":"value1","key2":"value2","key3":"value3","key4":["val1","val2","val3"]}}`
)

// openIDLCC is a chaincode component that supports the mainFunction operations for the openIDL network
type openIDLTestCC struct {
	carriers map[string]Carrier
	SmartContract
}

// SmartContract provides functions for managing an Asset
type SmartContract struct {
}

// openIDLCC is a chaincode component that supports the main operations for the openIDL network
type openIDLCC struct {
	carriers map[string]Carrier
}

var crossInvocationChannels Channels

type timestamp struct {
	time.Time
}

func (sd *timestamp) UnmarshalJSON(input []byte) error {
	strInput := string(input)
	strInput = strings.Trim(strInput, `"`)
	//newTime, err := time.Parse("2006/01/02 15:04:05", strInput)
	newTime, err := time.Parse(time.RFC3339, strInput)
	if err != nil {
		return err
	}

	sd.Time = newTime
	return nil
}

//type timestamp time.Time

// Prefixes
const (
	EXTRACTION_PATTERN_PREFIX                   = "Extraction_Pattern_"
	AUDIT_INSURANCE_TRANSACTIONAL_RECORD_PREFIX = "transactional-data-ingested-"
	INSURANCE_TRANSACTIONAL_RECORD_PREFIX       = "transactional-data-"
	INSURANCE_HASH_PREFIX                       = "hash-evidence-"
	DATA_CALL_PREFIX                            = "DataCall_Key_"
	DATA_CALLCOUNT_PREFIX                       = "DataCallCount_Key_"
	CARRIER_PREFIX                              = "Carrier_Key_"
	CONSENT_PREFIX                              = "Consent_Key_"
	CONSENT_DOCUMENT_TYPE                       = "Consent_Document_"
	REPORT_PREFIX                               = "Report_Key_"
	DOCUMENT_TYPE                               = "DataCall_Document"
	DOCUMENTCOUNT_TYPE                          = "DataCall_DocumentCount"
	REPORT_DOCUMENT_TYPE                        = "Report_Document_"
	LATEST_VERSION                              = "latest"
	STATUS_DRAFT                                = "DRAFT"
	STATUS_ISSUED                               = "ISSUED"
	STATUS_ABANDONED                            = "ABANDONED"
	STATUS_CANCELLED                            = "CANCELLED"
	STATUS_CANDIDATE                            = "CANDIDATE"
	STATUS_ACCEPTED                             = "ACCEPTED"
	STATUS_PUBLISHED                            = "PUBLISHED"
	STATUS_WITHHELD                             = "WITHHELD"
	LIKE_PREFIX                                 = "Like_Key_"
	LIKE_DOCUMENT_TYPE                          = "Like_Document_"
	PAGINATION_DEFAULT_START_INDEX              = 0
	DATACALL_LOG_PREFIX                         = "DataCallLog_Key_"
	DATACALL_LOG_DOCUMENT                       = "DataCallLog_Document_"
	ATTRIBUTE_NAME                              = "orgType"
	CARRIER_ORGANISATION_TYPE                   = "carrier"
	ADVISORY_ORGANISATION_TYPE                  = "advisory"
)

// Channels
const (
	DEFAULT_CHANNEL        = "defaultchannel"
	DEFAULT_CHAINCODE_NAME = "openidl-cc-default"
	// DEFAULT_CHAINCODE_NAME = "openidl-chaincode/defaultchannel"
	LOGGING_LEVEL = "LOGGING_LEVEL"
)

//channel and chaincode map for cross-channel query
var ccName = map[string]string{
	"aais-faircover": "openidl-cc-aais-faircover",
	"aais-carrier1":  "openidl-cc-aais-carrier1",
}

// Event Names
const (
	TOGGLE_LIKE_EVENT                        = "ToggleLikeEvent"
	CREATE_CONSENT_EVENT                     = "ConsentedEvent"
	SET_EXTRACTION_PATTERN_EVENT             = "ExtractionPatternSpecified"
	INSURANCE_RECORD_AND_AUDIT_CREATED_EVENT = "TransactionalDataAvailable"
)

type DataCallList struct {
	DataCalls        []DataCallExtended `json:"dataCallsList"`
	TotalNoOfRecords int                `json:"totalNoOfRecords"`
}

//struct to store audit record
type InsuranceRecordAudit struct {
	DataCallId      string `json:"dataCallId"`
	DataCallVersion string `json:"dataCallVersion"`
	CarrierId       string `json:"carrierId"`
}

//todo--add validation logic to match ext_pattern for value field  ValueValue---Records
//struct to strore Insurance Data value
type InsuranceData struct {
	PageNumber      int           `json:"pageNumber"`
	CarrierId       string        `json:"carrierId"`
	DataCallId      string        `json:"dataCallId"`
	DataCallVersion string        `json:"dataCallVersion"`
	Records         []interface{} `json:"records"`
	CreatedTs       timestamp     `json:"createdTs"`
} //map[string]interface{}

// struct to store Insurance data hash
type InsuranceDataHash struct {
	BatchId   string    `json:"batchId"`
	CarrierId string    `json:"carrierId"`
	ChunkId   string    `json:"chunkId"`
	Hash      string    `json:"hash"`
	CreatedTs timestamp `json:"createdTs"`
}

type InsuranceDataResponse struct {
	Records     []InsuranceData `jaon:"records"`
	NoOfRecords int             `json:"noOfRecords"`
}

type GetInsuranceData struct {
	ChannelName     string `json:"channelName"`
	DataCallId      string `json:"dataCallId"`
	DataCallVersion string `json:"dataCallVersion"`
	CarrierId       string `json:"carrierId"`
	StartIndex      int    `json:"startIndex"`
	PageSize        int    `json:"pageSize"`
	PageNumber      int    `json:"pageNumber"`
}

// struct to return as payload, when InsuranceRecord and Audit has been created
type InsuranceRecordEventPayload struct {
	ChannelName     string `json:"channelName"`
	DataCallId      string `json:"dataCallId"`
	DataCallVersion string `json:"dataCallVersion"`
	CarrierId       string `json:"carrierId"`
	PageNumber      int    `json:"pageNumber"`
}

//struct to return as ExtractionPattern event payload
type ExtractionPatternPayload struct {
	DataCallId          string `json:"dataCallId"`
	DataCallVsersion    string `json:"dataCallVersion"`
	ExtractionPatternID string `json:"extractionPatternID"`
	//ExtractionPattern ExtractionPattern `json:"extractionPattern"`
	ExtPatternTs timestamp `json:"extPatternTs"`
}

type ExtractionPatternIsSetPayload struct {
	IsSet             bool              `json:"isSet"`
	ExtractionPattern ExtractionPattern `json:"extractionPattern"`
}

type CouchDBView struct {
	Definition string `json:"definition"`
	Group      bool   `json:"group"`
}

type View struct {
	Map    string `json:"map"`
	Reduce string `json:"reduce"`
}

//struct to record ExtractionPattern
type ExtractionPattern struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CouchDBView `json:"couchDBView"`
}
type ExtPattern struct {
	ExtractionPatternID   string `json:"extractionPatternID"`
	ExtractionPatternName string `json:"extractionPatternName"`
	Description           string `json:"description"`
	//ViewDefinition        View      `json:"viewDefinition"`
	ViewDefinition struct {
		Map    string `json:"map"`
		Reduce string `json:"reduce"`
	} `json:"viewDefinition"`
	PremiumFromDate  string    `json:"premiumFromdate"`
	LossFromDate     string    `json:"lossFromdate"`
	Jurisdiction     string    `json:"jurisdication"`
	Insurance        string    `json:"insurance"`
	DbType           string    `json:"dbType"`
	Version          string    `json:"version"`
	IsActive         bool      `json:"isActive"`
	EffectiveStartTs timestamp `json:"effectiveStartTs,omitempty"`
	EffectiveEndTs   timestamp `json:"effectiveEndTs,omitempty"`
	UpdatedTs        timestamp `json:"updatedTs,omitempty"`
	UpdatedBy        string    `json:"updatedBy"`
}

/*type ExtractionPattern struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Documents   struct {
		Agreement []string `json:"agreement"`
		Claim     []string `json:"claim"`
	} `json:"documents"`
}*/

// An extened version of dataCall which contains count of likes and consents as no need to
// store count of likes and consents with actual data call model
type DataCallExtended struct {
	DataCall     DataCall `json:"dataCalls"`
	Reports      []Report `json:"reportsList"`
	NoOfConsents int      `json:"NoOfConsents"`
	NoOfLikes    int      `json:"NoOfLikes"`
}

// Carrier object
type Carrier struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DataCall object
type DataCall struct {
	ID                     string    `json:"id"`
	Version                string    `json:"version"`
	Name                   string    `json:"name"`
	IntentToPublish        bool      `json:"intentToPublish"`
	IsLocked               bool      `json:"isLocked"`
	IsLatest               bool      `json:"isLatest"`
	IsShowParticipants     bool      `json:"isShowParticipants"`
	Description            string    `json:"description"`
	Purpose                string    `json:"purpose"`
	LineOfBusiness         string    `json:"lineOfBusiness"`
	Deadline               timestamp `json:"deadline,omitempty"`
	PremiumFromDate        timestamp `json:"premiumFromDate,omitempty"`
	PremiumToDate          timestamp `json:"premiumToDate,omitempty"`
	LossFromDate           timestamp `json:"lossFromDate,omitempty"`
	LossToDate             timestamp `json:"lossToDate,omitempty"`
	Jurisdiction           string    `json:"jurisdiction"`
	ProposedDeliveryDate   timestamp `json:"proposedDeliveryDate,omitempty"`
	UpdatedBy              string    `json:"updatedBy"`
	UpdatedTs              timestamp `json:"updatedTs,omitempty"`
	DetailedCriteria       string    `json:"detailedCriteria"`
	EligibilityRequirement string    `json:"eligibilityRequirement"`
	Status                 string    `json:"status"`
	Type                   string    `json:"type"`
	Comments               string    `json:"comments"`
	ForumURL               string    `json:"forumURL"`
	LikeCount              int       `json:"likeCount"`
	ConsentCount           int       `json:"consentCount"`
	ExtractionPatternName  string    `json:"extractionPatternName"`
	ExtractionPatternID    string    `json:"extractionPatternID"`
	ExtractionPatternTs    timestamp `json:"extractionPatternTs"`
}

// DataCallCount object
type DataCallCount struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	ISSUED    int    `json:"issued"`
	DRAFT     int    `json:"draft"`
	CANCELLED int    `json:"cancelled"`
}

// SearchCriteria Struct for ListDataCallsByCriteria
type SearchCriteria struct {
	StartIndex int    `json:"startIndex"`
	PageSize   int    `json:"pageSize"`
	Version    string `json:"version"`
	Status     string `json:"status"`
	SearchKey  string `json:"searchKey"`
}

// SearchCriteria Struct for GetDataCallVersionsById
type GetDataCallVersions struct {
	ID         string `json:"id"`
	StartIndex int    `json:"startIndex"`
	PageSize   int    `json:"pageSize"`
	Status     string `json:"status"`
}

// SearchCriteria Struct for GetDataCallByIdAndVersion
type GetDataCall struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// SearchCriteria Struct for GetDataCallByIdAndVersion
type GetDataCallCount struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type ToggleDataCallCount struct {
	OriginalStatus string `json:"originalStatus"`
	NewStatus      string `json:"newStatus"`
}

//Struct for GetReportById
type GetReportById struct {
	DataCallID      string `json:"dataCallID"`
	DataCallVersion string `json:"dataCallVersion"`
	Hash            string `json:"hash"`
}

//Struct for GetHighestOrderReportByDataCall
type GetHighestOrderReport struct {
	DataCallID      string `json:"dataCallID"`
	DataCallVesrion string `json:"dataCallVersion"`
}

// struct to get extraction pattern by id
type GetExtractionPatternById struct {
	ExtractionPatternID string `json:"extractionPatternID"`
	DbType              string `json:"dbType"`
}
type GetDataCallAndExtractionPattern struct {
	DataCallID      string `json:"dataCallID"`
	DataCallVersion string `json:"dataCallVersion"`
	DbType          string `json:"dbType"`
}
type DataCallAndExtractionPatternResponse struct {
	Jurisdiction      string     `json:"jurisdiction"`
	IsSet             bool       `json:"isSet"`
	ExtractionPattern ExtPattern `json:"extractionPattern"`
}
type ExtractionPatternId struct {
	Id []string `json:"id"`
}

type Like struct {
	DatacallID       string    `json:"datacallID"`
	DataCallVersion  string    `json:"dataCallVersion"`
	OrganizationType string    `json:"organizationType"`
	OrganizationID   string    `json:"organizationID"`
	UpdatedTs        timestamp `json:"updatedTs"`
	UpdatedBy        string    `json:"updatedBy"`
	Liked            bool      `json:"liked"`
}

type LikeCountEntry struct {
	DatacallID      string    `json:"datacallID"`
	DataCallVersion string    `json:"dataCallVersion"`
	UpdatedTs       timestamp `json:"updatedTs"`
	Liked           bool      `json:"liked"`
	Delta           int       `json:"delta"`
}

type UpdateLikeAndConsentCountReq struct {
	DataCallID      string `json:"dataCallID"`
	DataCallVersion string `json:"dataCallVersion"`
}

// Node fabric-network module not allowing a separate argument in chaincodeInvoice, moving as part of Request of Like and Consent
/*type ListLikeRequest struct {
	Like       Like     `json:"like"`
	ChannelIDs []string `json:"channelIDs"`
}*/

type ListLikeRequest struct {
	DataCallID      string `json:"dataCallID"`
	DataCallVersion string `json:"dataCallVersion"`
	ChannelList     []struct {
		ChannelName   string `json:"channelName"`
		ChaincodeName string `json:"chaincodeName"`
	} `json:"channelList"`
}
type ListLikeResponse struct {
	Like             Like   `json:"like"`
	OrganizationName string `json:"organizationName"`
}

type Consent struct {
	DatacallID      string `json:"datacallID"`
	DataCallVersion string `json:"dataCallVersion"`
	CarrierID       string `json:"carrierID"`
	//CarrierName     string `json:"carrierName"`
	CreatedTs timestamp `json:"createdTs"`
	CreatedBy string    `json:"createdBy"`
	Status    string    `json:"status"`
}

type UpdateConsentStatus struct {
	DataCallID      string `json:"dataCallID"`
	DataCallVersion string `json:"dataCallVersion"`
	CarrierID       string `json:"carrierID"`
	Status          string `json:"status"`
}

type ConsentCountEntry struct {
	DatacallID      string    `json:"datacallID"`
	DataCallVersion string    `json:"dataCallVersion"`
	UpdatedTs       timestamp `json:"updatedTs"`
	Delta           int       `json:"delta"`
}

type GetConsentsByDataCallRequest struct {
	DataCallID      string `json:"dataCallID"`
	DataCallVersion string `json:"dataCallVersion"`
}
type GetLikesByDataCallRequest struct {
	DataCallID      string `json:"dataCallID"`
	DataCallVersion string `json:"dataCallVersion"`
}
type ListConsentRequest struct {
	DataCallID      string `json:"dataCallID"`
	DataCallVersion string `json:"dataCallVersion"`
	ChannelList     []struct {
		ChannelName   string `json:"channelName"`
		ChaincodeName string `json:"chaincodeName"`
	} `json:"channelList"`
}

type GetConsentByDataCallAndOrganizationRequest struct {
	Consent    Consent  `json:"consent"`
	ChannelIDs []string `json:"channelIDs"`
}

type GetLikeByDataCallAndOrganizationRequest struct {
	Like       Like     `json:"like"`
	ChannelIDs []string `json:"channelIDs"`
}

type ListConsentResponse struct {
	Consent     Consent `json:"consent"`
	CarrierName string  `json:"carrierName"`
}

type Report struct {
	DataCallID      string    `json:"dataCallID"`
	DataCallVersion string    `json:"dataCallVersion"`
	ReportVersion   string    `json:"reportVersion"`
	Hash            string    `json:"hash"`
	Status          string    `json:"status"`
	IsLocked        bool      `json:"isLocked"`
	Url             string    `json:"url"`
	CreatedBy       string    `json:"createdBy"`
	CreatedTs       timestamp `json:"createdTs,omitempty"`
	UpdatedTs       timestamp `json:"updatedTs,omitempty"`
}

type ListReportsCriteria struct {
	DataCallID      string `json:"dataCallID"`
	DataCallVersion string `json:"dataCallVersion"`
	Status          string `json:"status"`
	StartIndex      int    `json:"startIndex"`
	PageSize        int    `json:"pageSize"`
}

// Node fabric-network module not allowing a separate argument in chaincodeInvoice, moving as part of Request of Like and Consent
// TODO: Comeback to it.
type Channels struct {
	ChannelIDs []string `json:"channelIDs"`
}

type DataCallAction struct {
	ActionID   string
	ActionDesc string
}

var (
	ActionIssued             = DataCallAction{"DATA_CALL_ISSUED", "Data Call Issued."}
	ActionDeliveryDateUpdate = DataCallAction{"DATA_CALL_DELIVERY_DATE_UPDATED", "Report Delivery Date Updated."}
	ActionReportCandidate    = DataCallAction{"DATA_CALL_CANDIDATE_REPORT_DELIVERED", "Candidate Report Delivered."}
	ActionReportAccepted     = DataCallAction{"DATA_CALL_ACCEPTED", "Report Accepted."}
	ActionReportPublished    = DataCallAction{"DATA_CALL_PUBLISHED", "Report Published."}
	ActionReportWithheld     = DataCallAction{"DATA_CALL_WITHHELD", "Report Withheld."}
)

type DataCallLog struct {
	DataCallID      string    `json:"dataCallID"`
	DataCallVersion string    `json:"dataCallVersion"`
	ActionID        string    `json:"actionID"`
	Action          string    `json:"action"`
	ActionTs        timestamp `json:"actionTs,omitempty"`
	UpdatedBy       string    `json:"updatedBy"`
}

// Changed to Transaction id being generated uniquely for even test cases
var TxIdSeed int = 0

type CouchDBMockStub struct {
	*shimtest.MockStub
	ObjectType string
}

func NewCouchDBMockStateRangeQueryIterator(queryresults []queryresult.KV) *CouchDBMockStateRangeQueryIterator {
	iter := new(CouchDBMockStateRangeQueryIterator)
	if !iter.HasNext() {
		return iter
	}
	iter.QueryResults = queryresults
	iter.CurrentIndex = 0
	iter.Closed = false
	return iter
}

type CouchDBMockStateRangeQueryIterator struct {
	QueryResults []queryresult.KV
	CurrentIndex int
	Closed       bool
}

func NewCouchDBMockStub(name string, cc shim.Chaincode) *CouchDBMockStub {
	mock := shimtest.NewMockStub(name, cc)
	cmock := CouchDBMockStub{mock, ""}
	return &cmock
}

type CouchDBQuery struct {
	Selector map[string]string `json:"selector"`
}

func MockInit(stub *CouchDBMockStub, function string, args []byte) pb.Response {
	mockInvokeArgs := [][]byte{[]byte(function), args}
	txId := generateTransactionId()
	res := stub.MockStub.MockInit(txId, mockInvokeArgs)
	return res
}

func checkInvoke(t *testing.T, stub *CouchDBMockStub, function string, args []byte) pb.Response {
	fmt.Println("inside checkInvoke")
	mockInvokeArgs := [][]byte{[]byte(function), args}
	fmt.Println("mockinvokeArgs ", mockInvokeArgs)
	txId := generateTransactionId()
	fmt.Println("txId ", txId)
	res := stub.MockInvoke(txId, mockInvokeArgs)
	fmt.Println("res ", res)
	if res.Status != shim.OK {
		t.FailNow()
	}
	fmt.Println("res ", res)
	return res
}
func checkInvokeForResetLedger(t *testing.T, stub *CouchDBMockStub, function string) pb.Response {
	mockInvokeArgs := [][]byte{[]byte(function)}
	txId := generateTransactionId()
	res := stub.MockInvoke(txId, mockInvokeArgs)
	if res.Status != shim.OK {
		t.FailNow()
	}
	return res
}

// TODO: What is this function supposed to do?
// It is not asserting anything at the moment... anything missing here?
func checkInvoke_forError(t *testing.T, stub *CouchDBMockStub, function string, args []byte) pb.Response {
	mockInvokeArgs := [][]byte{[]byte(function), args}
	txId := generateTransactionId()
	res := stub.MockInvoke(txId, mockInvokeArgs)
	if res.Status != shim.OK {
	}
	return res
}

func (stub *CouchDBMockStub) GetQueryResult(query string) (shim.StateQueryIteratorInterface, error) {
	// Not implemented since the mock engine does not have a query engine.
	// However, a very simple query engine that supports string matching
	// could be implemented to test that the framework supports queries

	fmt.Printf("%+v\n", query)
	cdbquery := CouchDBQuery{}
	json.Unmarshal([]byte(query), &cdbquery)
	fmt.Printf("%+v\n", cdbquery)
	fmt.Printf("object type %s\n", stub.ObjectType)
	iter, err := stub.GetStateByPartialCompositeKey(stub.ObjectType, []string{})

	defer iter.Close()
	if err != nil {
		return nil, err
	}
	//	fmt.Printf("%+v\n", iter)
	//	fmt.Printf("%+v\n", stub.State)
	if iter.HasNext() {
		return iter, nil
	}
	ret := []queryresult.KV{}
	for iter.HasNext() {
		kv, err := iter.Next()
		if err != nil {
			return nil, err
		}
		fmt.Printf("%+v\n", kv)
		// TODO: choose items which matches query condition.
		ret = append(ret, *kv)
	}
	retiter := NewCouchDBMockStateRangeQueryIterator(ret)
	fmt.Printf("%+v\n", retiter)
	return retiter, nil
	//	return nil, errors.New("Not Implemented")
}

// HasNext returns true if the range query iterator contains additional keys
// and values.
func (iter *CouchDBMockStateRangeQueryIterator) HasNext() bool {
	if iter.Closed {
		// previously called Close()
		return false
	}
	if iter.CurrentIndex >= len(iter.QueryResults) {
		return false
	}
	return true
}

// Next returns the next key and value in the range query iterator.
func (iter *CouchDBMockStateRangeQueryIterator) Next() (*queryresult.KV, error) {
	if iter.Closed == true {
		return nil, errors.New("MockStateRangeQueryIterator.Next() called after Close()")
	}

	if iter.HasNext() == false {
		return nil, errors.New("MockStateRangeQueryIterator.Next() called when it does not HaveNext()")
	}
	ret := &iter.QueryResults[iter.CurrentIndex]
	iter.CurrentIndex++
	return ret, nil
	//	return nil, errors.New("MockStateRangeQueryIterator.Next() went past end of range")
}

// Close closes the range query iterator. This should be called when done
// reading from the iterator to free up resources.
func (iter *CouchDBMockStateRangeQueryIterator) Close() error {
	iter.Closed = true
	return nil
}

func generateTransactionId() string {
	TxIdSeed++
	s := strconv.Itoa(TxIdSeed)
	return s
}

//returns true(if orgType has access for that function)
func checkAccessForOrg(stub shim.ChaincodeStubInterface, function string) (bool, error) {

	//getting the certificate attribute
	organisationType, ok, err := cid.GetAttributeValue(stub, ATTRIBUTE_NAME)
	logger.Info(fmt.Sprintf("checkAccessForOrg: Checking access for %v organisation for %v function", organisationType, function))
	if err != nil {
		errStr := fmt.Sprintf("checkAccessForOrg: There was an error trying to retrieve the attribute %v", ATTRIBUTE_NAME)
		logger.Error("checkAccessForOrg: There was an error trying to retrieve the attribute ", err)
		return false, errors.New(errStr)
	}

	if !ok {
		errStr := fmt.Sprintf("checkAccessForOrg: The client identity does not possess the attribute for %v", ATTRIBUTE_NAME)
		logger.Error(errStr)
		return false, errors.New(errStr)
	}
	accessControlMap :=
		map[string][]string{
			"CreateExtractionPattern": {ADVISORY_ORGANISATION_TYPE},
			"UpdateExtractionPattern": {ADVISORY_ORGANISATION_TYPE},
			"CreateDataCall":          {ADVISORY_ORGANISATION_TYPE},
			"SaveNewDraft":            {ADVISORY_ORGANISATION_TYPE},
			"UpdateDataCall":          {ADVISORY_ORGANISATION_TYPE},
			"IssueDataCall":           {ADVISORY_ORGANISATION_TYPE},
			"ToggleLike":              {ADVISORY_ORGANISATION_TYPE, CARRIER_ORGANISATION_TYPE},
			"CreateConsent":           {ADVISORY_ORGANISATION_TYPE, CARRIER_ORGANISATION_TYPE},
			"CreateReport":            {ADVISORY_ORGANISATION_TYPE},
			"UpdateReport":            {ADVISORY_ORGANISATION_TYPE},
		}

	value, ok := accessControlMap[function]
	if ok {
		return contains(value, organisationType), nil

	} else {
		//doent have the function(menas that function doenst require access control check)
		//errStr := fmt.Sprintf("checkAccessForOrg: The organisation %v doesn't have access for function %v", organisationType, function)
		return true, nil
	}

}

// Contains tells whether an array arr contains a value searchElement.
func contains(arr []string, searchElement string) bool {
	for _, val := range arr {
		if searchElement == val {
			return true
		}
	}
	return false
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	logger.Info("Initledger: enter")
	defer logger.Debug("InitLedger: exit")

	var logLevelConfig = os.Getenv(LOGGING_LEVEL)
	if logLevelConfig != "" {
		// logger.SetLevel(logLevelConfig)
		/*temporary*/
		logger.SetLevel(logger.DebugLevel)
	} else {
		logger.SetLevel(logger.DebugLevel)
	}
	return nil
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data, so be careful to avoid a scenario where you
// inadvertently clobber your ledger's data!
func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("Init: enter")
	defer logger.Debug("Init: exit")

	var logLevelConfig = os.Getenv(LOGGING_LEVEL)
	if logLevelConfig != "" {
		// logger.SetLevel(logLevelConfig)
		/*temporary*/
		logger.SetLevel(logger.DebugLevel)
	} else {
		logger.SetLevel(logger.DebugLevel)
	}

	/*init_args := stub.GetStringArgs()
	logger.Debug("Init args > ", init_args)
	if len(init_args) == 2 {
		initChannelsJson := init_args[1]
		logger.Debug("Channels Argument > ", initChannelsJson)
		json.Unmarshal([]byte(initChannelsJson), &crossInvocationChannels)
		logger.Debug("Channels List Marshalled Successfully > ", crossInvocationChannels)
	}*/

	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The 'set'
// method may create a new asset by specifying a new key-value pair.
//func (s *SmartContract) Invoke2(ctx contractapi.TransactionContextInterface, id string) error {
func (this *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("Invoke: enter")
	defer logger.Debug("Invoke: exit")
	//function and parameters
	function, args := stub.GetFunctionAndParameters()

	hasAccess, err := checkAccessForOrg(stub, function)

	if err != nil {
		logger.Error(err)
		return shim.Error(err.Error())
	}
	if !hasAccess {
		errStr := fmt.Sprintf("checkAccessForOrg: The organisation doesn't have access for function %v", function)
		return shim.Error(errors.New(errStr).Error())
	}

	logger.Info("Invoke: function: ", function)

	switch function {
	case "Ping":
		return this.Ping(stub)
	case "ListDataCallsByCriteria":
		return this.ListDataCallsByCriteria(stub, args[0])
	case "ListMatureDataCalls":
		return this.ListMatureDataCalls(stub)
	case "CreateDataCall":
		return this.CreateDataCall(stub, args[0])
	case "SaveNewDraft":
		return this.SaveNewDraft(stub, args[0])
	case "UpdateDataCall":
		return this.UpdateDataCall(stub, args[0])
	case "IssueDataCall":
		return this.IssueDataCall(stub, args[0])
	case "GetDataCallVersionsById":
		return this.GetDataCallVersionsById(stub, args[0])
	case "GetDataCallByIdAndVersion":
		return this.GetDataCallByIdAndVersion(stub, args[0])
	case "ToggleLike":
		return this.ToggleLike(stub, args[0])
	case "CreateLikeCountEntry":
		return this.CreateLikeCountEntry(stub, args[0])
	case "CountLikes":
		return this.CountLikes(stub, args[0])
	case "CreateConsent":
		return this.CreateConsent(stub, args[0])
	case "CreateConsentCountEntry":
		return this.CreateConsentCountEntry(stub, args[0])
	case "CountConsents":
		return this.CountConsents(stub, args[0])
	case "ListConsentsByDataCall":
		return this.ListConsentsByDataCall(stub, args[0])
	case "GetConsentsByDataCall":
		return this.GetConsentsByDataCall(stub, args[0])
	case "GetHashById":
		return this.GetHashById(stub, args[0])
	case "GetConsentByDataCallAndOrganization":
		return this.GetConsentByDataCallAndOrganization(stub, args)
	case "ListLikesByDataCall":
		return this.ListLikesByDataCall(stub, args[0])
	case "GetLikesByDataCall":
		return this.GetLikesByDataCall(stub, args[0])
	case "GetLikeByDataCallAndOrganization":
		return this.GetLikeByDataCallAndOrganization(stub, args)
	case "SaveAndIssueDataCall":
		return this.SaveAndIssueDataCall(stub, args[0])
	case "CreateReport":
		return this.CreateReport(stub, args[0])
	case "UpdateReport":
		return this.UpdateReport(stub, args[0])
	case "ListReportsByCriteria":
		return this.ListReportsByCriteria(stub, args[0])
	case "ResetWorldState":
		return this.ResetWorldState(stub)
	case "LogDataCallTransaction":
		return this.LogDataCallTransaction(stub, args[0])
	case "GetDataCallTransactionHistory":
		return this.GetDataCallTransactionHistory(stub, args[0])
	case "GetReportById":
		return this.GetReportById(stub, args[0])
	case "ListDataCallTransactionHistory":
		return this.ListDataCallTransactionHistory(stub, args[0])
	case "SaveInsuranceDataHash":
		return this.SaveInsuranceDataHash(stub, args[0])
	case "CheckExtractionPatternIsSet":
		return this.CheckExtractionPatternIsSet(stub, args[0])
	case "SaveInsuranceData":
		return this.SaveInsuranceData(stub, args)
	case "CheckInsuranceDataExists":
		return this.CheckInsuranceDataExists(stub, args[0])
	case "GetExtractionPatternByIds":
		return this.GetExtractionPatternByIds(stub, args[0])
	case "GetInsuranceData":
		return this.GetInsuranceData(stub, args[0])
	case "CreateExtractionPattern":
		return this.CreateExtractionPattern(stub, args[0])
	case "UpdateExtractionPattern":
		return this.UpdateExtractionPattern(stub, args[0])
	case "GetDataCallAndExtractionPattern":
		return this.GetDataCallAndExtractionPattern(stub, args[0])
	case "ListExtractionPatterns":
		return this.ListExtractionPatterns(stub)
	case "UpdateLikeCountForDataCall":
		return this.UpdateLikeCountForDataCall(stub, args[0])
	case "UpdateConsentCountForDataCall":
		return this.UpdateConsentCountForDataCall(stub, args[0])
	case "ToggleDataCallCount":
		return this.ToggleDataCallCount(stub, args[0])
	case "GetDataCallCount":
		return this.GetDataCallCount(stub, args[0])
	case "UpdateDataCallCount":
		return this.UpdateDataCallCount(stub, args[0])
	case "SearchDataCalls":
		return this.SearchDataCalls(stub, args[0])
	case "UpdateConsentStatus":
		return this.UpdateConsentStatus(stub, args[0])
	default:
		//error
		return shim.Error("Invalid Function: " + function)
	}

}

// Ping simply returns a string as a way to validate that the chaincode component is up and running.
func (s *SmartContract) Ping(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Ping: enter")
	defer logger.Debug("Ping: exit")
	return shim.Success([]byte("Ping OK"))
}

func main() {
	// err := shim.Start(new(openIDLCC))
	// if err != nil {
	// 	logger.Error("Error starting openIDLCC: %s", err)
	// }
	// assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	// if err != nil {
	// 	logger.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	// }

	// if err := assetChaincode.Start(); err != nil {
	// 	logger.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	// }
	if err := shim.Start(new(SmartContract)); err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}

func (this *openIDLTestCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init: enter")
	defer logger.Debug("Init: exit")

	init_args := stub.GetStringArgs()
	logger.Debug("Init args > ", init_args)
	if len(init_args) == 2 {
		initChannelsJson := init_args[1]
		logger.Debug("Channels Argument > ", initChannelsJson)
		json.Unmarshal([]byte(initChannelsJson), &crossInvocationChannels)
		logger.Debug("Channels List Marshalled Successfully > ", crossInvocationChannels)
	}

	// this.carriers = GetCarriersMap()

	return shim.Success(nil)
}

func (this *openIDLTestCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Invoke: enter")
	defer logger.Debug("Invoke: exit")
	os.Setenv(LOGGING_LEVEL, "DEBUG")
	//function and parameters
	function, args := stub.GetFunctionAndParameters()

	logger.Debug("Invoke: function: ", function)

	if function == "CreateDataCall" {
		return this.CreateDataCall(stub, args[0])
	} else if function == "ListDataCallsByCriteria" {
		return this.ListDataCallsByCriteria(stub, args[0])
	} else if function == "SaveNewDraft" {
		return this.SaveNewDraft(stub, args[0])
	} else if function == "GetDataCallVersionsById" {
		return this.GetDataCallVersionsById(stub, args[0])
	} else if function == "GetDataCallByIdAndVersion" {
		return this.GetDataCallByIdAndVersion(stub, args[0])
	} else if function == "UpdateDataCall" {
		return this.UpdateDataCall(stub, args[0])
	} else if function == "IssueDataCall" {
		return this.IssueDataCall(stub, args[0])
	} else if function == "ToggleLike" {
		return this.ToggleLike(stub, args[0])
	} else if function == "CreateLikeCountEntry" {
		return this.CreateLikeCountEntry(stub, args[0])
	} else if function == "CountLikes" {
		return this.CountLikes(stub, args[0])
	} else if function == "CreateConsent" {
		return this.CreateConsent(stub, args[0])
	} else if function == "CreateConsentCountEntry" {
		return this.CreateConsentCountEntry(stub, args[0])
	} else if function == "CountConsents" {
		return this.CountConsents(stub, args[0])
	} else if function == "ListConsentsByDataCall" {
		return this.ListConsentsByDataCall(stub, args[0])
	} else if function == "GetConsentByDataCallAndOrganization" {
		return this.GetConsentByDataCallAndOrganization(stub, args)
	} else if function == "ListLikesByDataCall" {
		return this.ListLikesByDataCall(stub, args[0])
	} else if function == "GetLikesByDataCall" {
		return this.GetLikesByDataCall(stub, args[0])
	} else if function == "GetConsentsByDataCall" {
		return this.GetConsentsByDataCall(stub, args[0])
	} else if function == "GetLikeByDataCallAndOrganization" {
		return this.GetLikeByDataCallAndOrganization(stub, args)
	} else if function == "SaveAndIssueDataCall" {
		return this.SaveAndIssueDataCall(stub, args[0])
	} else if function == "CreateReport" {
		return this.CreateReport(stub, args[0])
	} else if function == "UpdateReport" {
		return this.UpdateReport(stub, args[0])
	} else if function == "ListReportsByCriteria" {
		return this.ListReportsByCriteria(stub, args[0])
	} else if function == "ResetWorldState" {
		return this.ResetWorldState(stub)
	} else if function == "LogDataCallTransaction" {
		return this.LogDataCallTransaction(stub, args[0])
	} else if function == "GetDataCallTransactionHistory" {
		return this.GetDataCallTransactionHistory(stub, args[0])
	} else if function == "GetReportById" {
		return this.GetReportById(stub, args[0])
	} else if function == "ListDataCallTransactionHistory" {
		return this.ListDataCallTransactionHistory(stub, args[0])
	} else if function == "SaveInsuranceDataHash" {
		return this.SaveInsuranceDataHash(stub, args[0])
	} else if function == "GetHashById" {
		return this.GetHashById(stub, args[0])
	} else if function == "SaveInsuranceData" {
		return this.SaveInsuranceData(stub, args)
	}

	return shim.Error("Invalid Function: " + function)
}

//Test for SaveInsuranceHash
//This function tests whether it returns error for empty BatchId.
//This function tests whether it returns 200 for Success.
//This function tests whether it stores hash.
func Test_SaveInsuranceHash_Should_Save_Insurance_Data_Hash(t *testing.T) {
	fmt.Println("Test_SaveInsuranceHash_Should_Save_Insurance_Data_Hash")
	scc := new(openIDLTestCC)
	stub := NewCouchDBMockStub("OpenIDLMockStub", scc)

	//test for SaveInsuranceHash
	//Step-1: For ERROR- Check whether it returns error for empty ID
	res_err_saveHash := checkInvoke_forError(t, stub, "SaveInsuranceDataHash", []byte(SAVE_INSURANCE_HASH_EMPTY_ID_JSON))
	var err_message_for_saveHash = res_err_saveHash.Message
	if res_err_saveHash.Status != shim.OK {
		assert.Equal(t, "BatchId should not be Empty", err_message_for_saveHash, "Test_SaveInsuranceHash: For Empty Id")
	} else {
		t.FailNow()
	}

	//Step-2: For SUCCESS- Check whether it returns 200
	res_saveHash := checkInvoke(t, stub, "SaveInsuranceDataHash", []byte(SAVE_INSURANCE_HASH_VALID_JSON))
	if res_saveHash.Status != shim.OK {
		logger.Error("SaveInsuranceHash failed with message res.Message: ", string(res_saveHash.Message))
		fmt.Println("SaveInsuranceHash failed with message res.Message: ", string(res_saveHash.Message))
		t.FailNow()
	}
	var saveHash_returnCode = int(res_saveHash.Status)
	assert.Equal(t, 200, saveHash_returnCode, "Test_SaveInsuranceHash: Function's success, status code 200.")

	//Step-3: For SUCCESS- Check whether input object matches output object
	res_getHashById := checkInvoke(t, stub, "GetHashById", []byte(SAVE_INSURANCE_HASH_VALID_JSON))
	var input_saveHash InsuranceDataHash
	json.Unmarshal([]byte(SAVE_INSURANCE_HASH_VALID_JSON), &input_saveHash)
	var output_saveHash InsuranceDataHash
	err_saveHash := json.Unmarshal(res_getHashById.Payload, &output_saveHash)
	if err_saveHash != nil {
		logger.Error("Test_SaveInsuranceHash: Error during json.Unmarshal for GetReportById: ", err_saveHash)
		t.FailNow()
	}
	assert.True(t, reflect.DeepEqual(input_saveHash, output_saveHash))

}

//Test for SaveInsuranceData
//This function tests whether it returns error for empty BatchId.
//This function tests whether it returns 200 for Success.
//This function tests whether it stores insurance data.
func Test_SaveInsuranceData_Should_Save_Insurance_Data(t *testing.T) {
	fmt.Println("Test_SaveInsuranceData_Should_Save_Insurance_Data")
	scc := new(openIDLTestCC)
	stub := NewCouchDBMockStub("OpenIDLMockStub", scc)

	//test for SaveInsuranceData
	//Step-1: For ERROR- Check whether it returns error for empty ID
	res_err_saveData := checkInvoke_forError(t, stub, "SaveInsuranceData", []byte(SAVE_INSURANCE_DATA_EMPTY_CARRIER_ID_JSON))
	var err_message_for_saveData = res_err_saveData.Message
	if res_err_saveData.Status != shim.OK {
		assert.Equal(t, "CarrierId should not be Empty", err_message_for_saveData, "Test_SaveInsuranceData: For Empty Id")
	} else {
		t.FailNow()
	}

	//Step-2: For SUCCESS- Check whether it returns 200
	res_saveData := checkInvoke(t, stub, "SaveInsuranceData", []byte(SAVE_INSURANCE_DATA_VALID_JSON))
	if res_saveData.Status != shim.OK {
		logger.Error("SaveInsuranceData failed with message res.Message: ", string(res_saveData.Message))
		fmt.Println("SaveInsuranceData failed with message res.Message: ", string(res_saveData.Message))
		t.FailNow()
	}
	var saveData_returnCode = int(res_saveData.Status)
	assert.Equal(t, 200, saveData_returnCode, "Test_SaveInsuranceData: Function's success, status code 200.")
}

// generateVersion is a helper function for generating a version number.
func generateVersion(number int) string {
	return strconv.Itoa(number + 1)
}

// CreateDataCall creates a new DataCall object. This method receives as a parameter a DataCall object in JSON format.
// Success: nil
// Error: {"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (s *SmartContract) CreateDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CreateDataCall: enter")
	defer logger.Debug("CreateDataCall: exit")
	logger.Debug("CreaCreateDataCall json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}

	var dataCall DataCall
	err := json.Unmarshal([]byte(args), &dataCall)
	if dataCall.ID == "" {
		return shim.Error("Id cant not be empty!!")
	}
	if err != nil {
		logger.Error("CreateDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CreateDataCall: Error during json.Unmarshal").Error())
	}
	if dataCall.Status == STATUS_DRAFT {
		dataCall.IsLocked = false
	}
	if dataCall.Status == STATUS_ISSUED || dataCall.Status == STATUS_CANCELLED {
		dataCall.IsLocked = true
	}
	logger.Debug("Unmarshalled object ", dataCall)
	dataCall.IsLatest = true
	//dataCall.IsLocked = "false"
	dataCall.Version = generateVersion(0)

	var pks []string = []string{DATA_CALL_PREFIX, dataCall.ID, dataCall.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)
	//dataCallKey := DATA_CALL_PREFIX + dataCall.ID + dataCall.Version
	logger.Info("In data call create ", dataCallKey)
	// Checking the ledger to confirm that the dataCall doesn't exist
	prevDataCall, _ := stub.GetState(dataCallKey)

	if prevDataCall != nil {
		logger.Error("CreateDataCall: Data Call already exist for the data call with ID: " + dataCallKey)
		return shim.Error("Data Call already exist for the data call with ID: " + dataCallKey)
	}

	dataCallAsBytes, _ := json.Marshal(dataCall)
	err = stub.PutState(dataCallKey, dataCallAsBytes)

	// update datacalllog
	//var dataCallCreateLog DataCallLog
	if dataCall.Status == STATUS_ISSUED {
		dataCallIssuedLog := DataCallLog{dataCall.ID, dataCall.Version, ActionIssued.ActionID, ActionIssued.ActionDesc, dataCall.UpdatedTs, dataCall.UpdatedBy}
		dataCallIssuedLogAsBytes, _ := json.Marshal(dataCallIssuedLog)
		logger.Info((dataCallIssuedLogAsBytes))
		// this.LogDataCallTransaction(stub, string(dataCallIssuedLogAsBytes))
	}

	if err != nil {
		logger.Error("Error commiting data for key: ", dataCallKey)
		return shim.Error("Error committing data for key: " + dataCallKey)
	}

	//change the count of the statuses
	logger.Info("Toggling ", dataCall.Status)
	toggleDataCallCount := ToggleDataCallCount{"", dataCall.Status}
	datacallIssueLogAsBytes, _ := json.Marshal(toggleDataCallCount)
	logger.Info((datacallIssueLogAsBytes))
	// dataCallCountAsBytes := this.ToggleDataCallCount(stub, string(datacallIssueLogAsBytes))
	// logger.Info("The reply from toggle is ", dataCallCountAsBytes)

	return shim.Success(nil)

}

// ListDataCallsByCriteria retrives all data calls that match given criteria. If startindex and pageSize are not provided,
// this method returns the complete list of data calls. If version = latest, the it returns only latest version of a data call
// using the specified criteria. If version = all, it returns all data calls with their versions as individual items in list.
// params {json}: {
//  "startIndex":"optional",
//  "pageSize":"optional",
//  "version": "latest or all"
//  "status" :"DRAFT OR ISSUED OR CANCELLED"}
// Success {byte[]}: byte[]
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (this *SmartContract) ListDataCallsByCriteria(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("ListDataCallsByCriteria: enter")
	defer logger.Debug("ListDataCallsByCriteria: exit")
	logger.Debug("ListDataCallsByCriteria json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}
	var searchCriteria SearchCriteria
	err := json.Unmarshal([]byte(args), &searchCriteria)
	if err != nil {
		logger.Error("ListDataCallsByCriteria: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("ListDataCallsByCriteria: Error during json.Unmarshal").Error())
	}
	logger.Debug("ListDataCallsByCriteria: Unmarshalled object ", searchCriteria)
	startIndex := searchCriteria.StartIndex
	pageSize := searchCriteria.PageSize
	version := searchCriteria.Version
	var isLatest string
	if version == LATEST_VERSION {
		isLatest = "true"
	} else {
		isLatest = "false"
	}
	status := searchCriteria.Status
	if status == "" {
		logger.Error("ListDataCallsByCriteria: Status not present, You must pass Status in agument")
		return shim.Error("ListDataCallsByCriteria: Status not present, You must pass Status in agument")
	}
	var queryStr string

	if status == STATUS_ISSUED {
		queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\"},\"use_index\":[\"deadline\", \"deadlineIndex\"],\"sort\":[{\"deadline\": \"desc\"}],\"limit\":%d,\"skip\":%d}", DATA_CALL_PREFIX, status, pageSize, startIndex)
	} else if status == STATUS_DRAFT {
		queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\", \"isLatest\":%s, \"isLocked\":false},\"use_index\":[\"deadline\", \"deadlineIndex\"],\"sort\":[{\"updatedTs\": \"desc\"}],\"limit\":%d,\"skip\":%d}", DATA_CALL_PREFIX, status, isLatest, pageSize, startIndex)
	} else if status == STATUS_CANCELLED {
		queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\"},\"use_index\":[\"deadline\", \"deadlineIndex\"],\"sort\":[{\"deadline\": \"desc\"}],\"limit\":%d,\"skip\":%d}", DATA_CALL_PREFIX, status, pageSize, startIndex)
	}

	logger.Info("ListDataCallsByCriteria: Selector ", queryStr)
	logger.Info("startIndex", "pageSize", "version", "status", startIndex, pageSize, version, status)

	var dataCalls []DataCall
	startTime := time.Now()
	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryStr, int32(pageSize), "")
	elapsedTime := time.Since(startTime)
	logger.Info("RESPONSE META DATA ", responseMetadata.FetchedRecordsCount)
	logger.Info("Time consumed to get Data Calls", elapsedTime)
	defer resultsIterator.Close()
	if err != nil {
		logger.Error("Failed to get state for all the data calls")
		return shim.Error("Failed to get state for all the data calls")
	}

	if !resultsIterator.HasNext() {
		dataCallsAsByte, _ := json.Marshal(dataCalls)
		logger.Debug("ListDataCallsByCriteria: dataCallsAsByte", dataCallsAsByte)
		//return shim.Error(errors.New("ListDataCallsByCriteria :DataCall not found ").Error())
		return shim.Success(dataCallsAsByte)
	}

	for resultsIterator.HasNext() {
		dataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate data call")
			return shim.Error("Failed to iterate data call")
		}

		var dataCall DataCall
		err = json.Unmarshal([]byte(dataCallAsBytes.GetValue()), &dataCall)
		if err != nil {
			return shim.Error("Failed to unmarshal data call: " + err.Error())
		}

		dataCalls = append(dataCalls, dataCall)
	}

	//var paginatedDataCall []DataCall
	//var paginatedDataCalls []DataCall
	//paginatedDataCalls = paginate(dataCalls, startIndex, pageSize)
	var dataCallList DataCallList

	//getting the count
	getDataCallCount := GetDataCallCount{"123456", "1"}
	datacallIssueLogAsBytes, _ := json.Marshal(getDataCallCount)
	dataCallCountAsBytes := this.GetDataCallCount(stub, string(datacallIssueLogAsBytes))
	var dataCallCount DataCallCount
	err = json.Unmarshal(dataCallCountAsBytes.Payload, &dataCallCount)
	logger.Info("The retrieved data is ", dataCallCount)

	if status == STATUS_ISSUED {
		dataCallList.TotalNoOfRecords = dataCallCount.ISSUED
	} else if status == STATUS_DRAFT {
		dataCallList.TotalNoOfRecords = dataCallCount.DRAFT
	} else if status == STATUS_CANCELLED {
		dataCallList.TotalNoOfRecords = dataCallCount.CANCELLED
	}

	var IdAndVersionMap map[string]string
	IdAndVersionMap = make(map[string]string)
	var dataCallIDs []string
	var dataCallVersions []string
	for dataCallIndex := 0; dataCallIndex < len(dataCalls); dataCallIndex++ {
		if dataCallIndex == 0 {
			IdAndVersionMap[dataCalls[dataCallIndex].ID] = dataCalls[dataCallIndex].Version
			dataCallIDs = append(dataCallIDs, `"`+dataCalls[dataCallIndex].ID+`"`)
			dataCallVersions = append(dataCallVersions, `"`+dataCalls[dataCallIndex].Version+`"`)
		} else {
			IdAndVersionMap[dataCalls[dataCallIndex].ID] = dataCalls[dataCallIndex].Version
			dataCallIDs = append(dataCallIDs, `,`+`"`+dataCalls[dataCallIndex].ID+`"`)
			dataCallVersions = append(dataCallVersions, `,`+`"`+dataCalls[dataCallIndex].Version+`"`)
		}
	}

	startTimeForAll := time.Now()
	/*
		//get ConsentCount map
		startTimeForConsent := time.Now()
		consentCounts := GetConsentsCount(stub, dataCallIDs)
		elapsedTimeForConsent := time.Since(startTimeForConsent)
		logger.Info("Time consumed for consent", elapsedTimeForConsent)
		startTimeForLike := time.Now()
		likeCounts := GetLikesCount(stub, dataCallIDs, IdAndVersionMap)
		elapsedTimeForLike := time.Since(startTimeForLike)
		logger.Info("Time consumed for Like", elapsedTimeForLike)
	*/

	startTimeForReport := time.Now()
	latestReports := GetLatestaReport(stub, dataCallIDs, dataCallVersions)
	elapsedTimeForReport := time.Since(startTimeForReport)
	logger.Info("Time consumed for Report", elapsedTimeForReport)

	elapsedTimeForAll := time.Since(startTimeForAll)
	logger.Info("Time consumed for all", elapsedTimeForAll)
	logger.Info("Final ===========", elapsedTime)
	for dataCallIndex := 0; dataCallIndex < len(dataCalls); dataCallIndex++ {
		var dataCallExtended DataCallExtended
		dataCallExtended.DataCall = dataCalls[dataCallIndex]
		dataCallExtended.Reports = append(dataCallExtended.Reports, latestReports[dataCalls[dataCallIndex].ID])
		//dataCallExtended.NoOfConsents = consentCounts[paginatedDataCalls[dataCallIndex].ID]
		//dataCallExtended.NoOfLikes = likeCounts[paginatedDataCalls[dataCallIndex].ID]
		dataCallList.DataCalls = append(dataCallList.DataCalls, dataCallExtended)

	}

	dataCallsAsByte, _ := json.Marshal(dataCallList)
	return shim.Success(dataCallsAsByte)

}

// ListMatureDataCalls retrives all data calls that has matured in last 24 hours.
// using the specified criteria. If version = all, it returns all data calls with their versions as individual items in list.
// Success {byte[]}: byte[]
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (this *SmartContract) ListMatureDataCalls(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("ListMatureDataCalls: enter")
	defer logger.Debug("ListMatureDataCalls: exit")
	status := STATUS_ISSUED
	var queryStr string
	queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\"},\"use_index\":[\"deadline\", \"deadlineIndex\"],\"sort\":[{\"deadline\": \"desc\"}]}", DATA_CALL_PREFIX, status)

	var dataCalls []DataCall
	startTime := time.Now()
	resultsIterator, err := stub.GetQueryResult(queryStr)
	elapsedTime := time.Since(startTime)
	logger.Info("Time consumed to get Data Calls", elapsedTime)
	defer resultsIterator.Close()
	if err != nil {
		logger.Error("Failed to get state for all the data calls")
		return shim.Error("Failed to get state for all the data calls")
	}

	if !resultsIterator.HasNext() {
		dataCallsAsByte, _ := json.Marshal(dataCalls)
		logger.Debug("ListDataCallsByCriteria: dataCallsAsByte", dataCallsAsByte)
		//return shim.Error(errors.New("ListDataCallsByCriteria :DataCall not found ").Error())
		return shim.Success(dataCallsAsByte)
	}

	for resultsIterator.HasNext() {
		dataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate data call")
			return shim.Error("Failed to iterate data call")
		}

		var dataCall DataCall
		err = json.Unmarshal([]byte(dataCallAsBytes.GetValue()), &dataCall)
		if err != nil {
			return shim.Error("Failed to unmarshal data call: " + err.Error())
		}
		// If mature date in last 24 hours, add them to datacalls

		startDate := startTime.Truncate(24*time.Hour).AddDate(0, 0, -1)
		endDate := startTime.Truncate(24 * time.Hour)
		if (dataCall.Deadline.After(startDate) && dataCall.Deadline.Before(endDate)) || dataCall.Deadline.Equal(startTime) {
			dataCalls = append(dataCalls, dataCall)
		}
	}

	var dataCallList DataCallList

	//getting the count
	getDataCallCount := GetDataCallCount{"123456", "1"}
	datacallIssueLogAsBytes, _ := json.Marshal(getDataCallCount)
	dataCallCountAsBytes := this.GetDataCallCount(stub, string(datacallIssueLogAsBytes))
	var dataCallCount DataCallCount
	err = json.Unmarshal(dataCallCountAsBytes.Payload, &dataCallCount)
	logger.Info("The retrieved data is ", dataCallCount)

	if status == STATUS_ISSUED {
		dataCallList.TotalNoOfRecords = dataCallCount.ISSUED
	} else if status == STATUS_DRAFT {
		dataCallList.TotalNoOfRecords = dataCallCount.DRAFT
	} else if status == STATUS_CANCELLED {
		dataCallList.TotalNoOfRecords = dataCallCount.CANCELLED
	}

	var IdAndVersionMap map[string]string
	IdAndVersionMap = make(map[string]string)
	var dataCallIDs []string
	var dataCallVersions []string
	for dataCallIndex := 0; dataCallIndex < len(dataCalls); dataCallIndex++ {
		if dataCallIndex == 0 {
			IdAndVersionMap[dataCalls[dataCallIndex].ID] = dataCalls[dataCallIndex].Version
			dataCallIDs = append(dataCallIDs, `"`+dataCalls[dataCallIndex].ID+`"`)
			dataCallVersions = append(dataCallVersions, `"`+dataCalls[dataCallIndex].Version+`"`)
		} else {
			IdAndVersionMap[dataCalls[dataCallIndex].ID] = dataCalls[dataCallIndex].Version
			dataCallIDs = append(dataCallIDs, `,`+`"`+dataCalls[dataCallIndex].ID+`"`)
			dataCallVersions = append(dataCallVersions, `,`+`"`+dataCalls[dataCallIndex].Version+`"`)
		}
	}

	startTimeForAll := time.Now()
	/*
		//get ConsentCount map
		startTimeForConsent := time.Now()
		consentCounts := GetConsentsCount(stub, dataCallIDs)
		elapsedTimeForConsent := time.Since(startTimeForConsent)
		logger.Info("Time consumed for consent", elapsedTimeForConsent)
		startTimeForLike := time.Now()
		likeCounts := GetLikesCount(stub, dataCallIDs, IdAndVersionMap)
		elapsedTimeForLike := time.Since(startTimeForLike)
		logger.Info("Time consumed for Like", elapsedTimeForLike)
	*/

	startTimeForReport := time.Now()
	latestReports := GetLatestaReport(stub, dataCallIDs, dataCallVersions)
	elapsedTimeForReport := time.Since(startTimeForReport)
	logger.Info("Time consumed for Report", elapsedTimeForReport)

	elapsedTimeForAll := time.Since(startTimeForAll)
	logger.Info("Time consumed for all", elapsedTimeForAll)
	logger.Info("Final ===========", elapsedTime)
	for dataCallIndex := 0; dataCallIndex < len(dataCalls); dataCallIndex++ {
		var dataCallExtended DataCallExtended
		dataCallExtended.DataCall = dataCalls[dataCallIndex]
		dataCallExtended.Reports = append(dataCallExtended.Reports, latestReports[dataCalls[dataCallIndex].ID])
		//dataCallExtended.NoOfConsents = consentCounts[paginatedDataCalls[dataCallIndex].ID]
		//dataCallExtended.NoOfLikes = likeCounts[paginatedDataCalls[dataCallIndex].ID]
		dataCallList.DataCalls = append(dataCallList.DataCalls, dataCallExtended)

	}

	dataCallsAsByte, _ := json.Marshal(dataCallList)
	return shim.Success(dataCallsAsByte)

}

//helper function for pagination
func paginate(dataCall []DataCall, startIndex int, pageSize int) []DataCall {
	if startIndex == 0 {
		startIndex = PAGINATION_DEFAULT_START_INDEX
	}
	// no pageSize specified then return all results
	if pageSize == 0 {
		pageSize = len(dataCall)
		return dataCall
	}
	limit := func() int {
		if startIndex+pageSize > len(dataCall) {
			return len(dataCall)
		} else {
			return startIndex + pageSize - 1
		}
	}

	start := func() int {
		if startIndex > len(dataCall) {
			return len(dataCall) - 1
		} else {
			return startIndex - 1
		}
	}
	logger.Debug("ListDataCallsByCriteria: Getting Records from index", start(), " to index ", limit())
	return dataCall[start():limit()]
}

// function-name: SaveNewDraft (invoke)
// params {DataCall json}
// Success :nil
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : Creates/Update new version of the data call.

func (this *SmartContract) SaveNewDraft(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("SaveNewDraft: enter")
	defer logger.Debug("SaveNewDraft: exit")
	logger.Debug("SaveNewDraft json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}
	var dataCalls []DataCall
	var currentDataCall DataCall
	//var prevDataCall DataCall
	err := json.Unmarshal([]byte(args), &currentDataCall)
	if err != nil {
		logger.Error("SaveNewDraft: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("SaveNewDraft: Error during json.Unmarshal").Error())
	}

	logger.Debug("Unmarshalled object ", currentDataCall)

	if currentDataCall.ID == "" {
		return shim.Error("Id can not be empty")
	}
	var pks []string = []string{DATA_CALL_PREFIX, currentDataCall.ID}

	//Step-1: getting the previous data for the data call
	resultsIterator, errmsg := stub.GetStateByPartialCompositeKey(DOCUMENT_TYPE, pks)
	if errmsg != nil {
		logger.Error("SaveNewDraft: Failed to get state for previous data calls")
		return shim.Error("SaveNewDraft: Failed to get state for previous data calls")
	}
	defer resultsIterator.Close()
	logger.Debug("SaveNewDraft: ", resultsIterator)
	for resultsIterator.HasNext() {
		prevDataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate data call")
			return shim.Error("Failed to iterate data call")
		}

		var prevDataCall DataCall
		err = json.Unmarshal([]byte(prevDataCallAsBytes.GetValue()), &prevDataCall)
		if err != nil {
			logger.Error("Failed to unmarshal data call ")
			return shim.Error("Failed to unmarshal data call: " + err.Error())
		}
		//fetching the latest record
		if prevDataCall.IsLatest == true {
			dataCalls = append(dataCalls, prevDataCall)
		}
	}

	//Step-2: setting the previous DataCall with isLatest to false and
	dataCalls[0].IsLatest = false
	//dataCalls[0].IsLocked = "false"
	prevDataCallVersion, _ := strconv.Atoi(dataCalls[0].Version)

	//creating composite key to save the previous DataCall
	var pkForPrevDataCall []string = []string{DATA_CALL_PREFIX, dataCalls[0].ID, dataCalls[0].Version}
	prevDataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pkForPrevDataCall)
	fmt.Println(prevDataCallKey)
	logger.Debug("SaveNewDraft: previousDatacallKey > ", prevDataCallKey)

	prevDataCallAsBytes, _ := json.Marshal(dataCalls[0])
	//saving the previous DataCall
	err = stub.PutState(prevDataCallKey, prevDataCallAsBytes)
	logger.Debug("SaveNewDraft: Previous Datacall saved!")
	if err != nil {
		logger.Error("Error commiting the previous DataCall")
		return shim.Error("Error commiting the previous DataCall")
	}

	//Step-4: saving the draft with updating new version and setting isLatest to true
	currentDataCall.IsLatest = true
	currentDataCall.IsLocked = false
	//currentDataCallVersion, _ := strconv.Atoi(currentDataCall.Version)
	currentDataCall.Version = generateVersion(prevDataCallVersion)

	var pkForCurrentDataCall []string = []string{DATA_CALL_PREFIX, currentDataCall.ID, currentDataCall.Version}
	currentDataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pkForCurrentDataCall)

	currentDataCallAsBytes, _ := json.Marshal(currentDataCall)
	err = stub.PutState(currentDataCallKey, currentDataCallAsBytes)
	if err != nil {
		return shim.Error("Error committing Current DataCall for key: " + currentDataCallKey)
	}
	logger.Debug("SaveNewDraft: Latest Datacall saved!")

	if currentDataCall.Status != dataCalls[0].Status {
		//change the count of the statuses
		logger.Info("Toggling ", currentDataCall.Status, " and ", dataCalls[0].Status)
		toggleDataCallCount := ToggleDataCallCount{dataCalls[0].Status, currentDataCall.Status}
		datacallIssueLogAsBytes, _ := json.Marshal(toggleDataCallCount)
		dataCallCountAsBytes := this.ToggleDataCallCount(stub, string(datacallIssueLogAsBytes))
		logger.Info("The reply from toggle is ", dataCallCountAsBytes)
	}

	return shim.Success(nil)

}

// function-name: GetDataCallVersionsById(invoke)
// params {json}: {
// "id":"mandatory",
// "startIndex":"optional",
// "pageSize":"optional"}
// Success {byte[]}: byte[]  - List of data calls
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : returns a datacall of specifc version and id.
// if pageSize is blank returns all versions

func (this *SmartContract) GetDataCallVersionsById(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetDataCallVersionsById: enter")
	defer logger.Debug("GetDataCallVersionsById: exit")

	var getDataCallVersions GetDataCallVersions
	err := json.Unmarshal([]byte(args), &getDataCallVersions)
	if err != nil {
		logger.Error("GetDataCallVersionsById: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetDataCallVersionsById: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshalled object ", getDataCallVersions)

	var dataCalls []DataCall
	if getDataCallVersions.ID == "" {
		return shim.Error("GetDataCallVersionsById: ID is Empty")
	}
	logger.Debug("GetDataCallVersionsById: Status", getDataCallVersions.Status)
	//fmt.Println(startIndex, pageSize)
	//queryStr := fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"id\":\"%s\"}}", DATA_CALL_PREFIX, getDataCallVersions.ID)
	//queryStr := fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"id\":\"%s\"},\"sort\":[{\"version\":\"desc\"}]}", DATA_CALL_PREFIX, getDataCallVersions.ID)
	queryStr := fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"id\":\"%s\"},\"use_index\":\"_design/versionDoc\",\"sort\":[{\"version\": \"desc\"}]}", DATA_CALL_PREFIX, getDataCallVersions.ID)
	resultsIterator, err := stub.GetQueryResult(queryStr)

	if err != nil {
		logger.Error("GetDataCallVersionsById: Failed to get Data Calls")
		return shim.Error("Failed to get Data Calls : " + err.Error())
	}
	defer resultsIterator.Close()
	logger.Debug("GetDataCallVersionsById: Iterating over datacalls versions")
	for resultsIterator.HasNext() {
		dataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate data call")
			return shim.Error("Failed to iterate data call")
		}

		var dataCall DataCall
		err = json.Unmarshal([]byte(dataCallAsBytes.GetValue()), &dataCall)

		logger.Debug("GetDataCallVersionsById: DataCall > ", dataCall.ID)
		if err != nil {
			return shim.Error("Failed to unmarshal data call: " + err.Error())
		}

		if len(getDataCallVersions.Status) > 0 {
			// if status is present in input add as per status
			if getDataCallVersions.Status == dataCall.Status {
				dataCalls = append(dataCalls, dataCall)
			}
		} else {
			// else add all data calls in response
			dataCalls = append(dataCalls, dataCall)
		}

	}
	dataCallsAsByte, _ := json.Marshal(dataCalls)
	return shim.Success(dataCallsAsByte)

}

// function-name: GetDataCallByIdAndVersion (invoke)
// params {json}: {
// "id":"mandatory",
// "version": "mandatory"}
// Success {byte[]}: byte[]  - List of data calls
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : returns a datacall of specifc version and id.

func (s *SmartContract) GetDataCallByIdAndVersion(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetDataCallByIdAndVersion: enter")
	defer logger.Debug("GetDataCallByIdAndVersion: exit")

	var getDataCall GetDataCall
	err := json.Unmarshal([]byte(args), &getDataCall)
	if err != nil {
		logger.Error("GetDataCallVersionsById: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetDataCallVersionsById: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshalled object ", getDataCall)

	if getDataCall.ID == "" || getDataCall.Version == "" {
		return shim.Error("ID and Version can not be Empty")
	}

	var pks []string = []string{DATA_CALL_PREFIX, getDataCall.ID, getDataCall.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)
	dataCallAsBytes, err := stub.GetState(dataCallKey)
	if err != nil {
		logger.Error("Error retreiving data for key ", dataCallKey)
		return shim.Error("Error retreiving data for key" + dataCallKey)
	}
	return shim.Success(dataCallAsBytes)

}

// function-name: UpdateDataCall (invoke)
// params {DataCall json}
// Success :nil
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : Updates the data call, without creating new version.
//   This needs to be invoked when we need to change delivery date

func (this *SmartContract) UpdateDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("UpdateDataCall: enter")
	defer logger.Debug("UpdateDataCall: exit")
	logger.Debug("UpdateDataCall json received : ", args)

	var dataCall DataCall
	var isCancelled bool
	err := json.Unmarshal([]byte(args), &dataCall)
	if err != nil {
		logger.Error("UpdateDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("UpdateDataCall: Error during json.Unmarshal").Error())
	}
	if dataCall.ID == "" {
		return shim.Error("ID should not be Empty")

	}
	if dataCall.Version == "" {
		return shim.Error("Version should not be Empty")

	}
	var pks []string = []string{DATA_CALL_PREFIX, dataCall.ID, dataCall.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)
	dataCallAsBytes, err := stub.GetState(dataCallKey)
	if err != nil {
		return shim.Error("Error retreiving data for key" + dataCallKey)
	}

	var prevDataCall DataCall
	err = json.Unmarshal(dataCallAsBytes, &prevDataCall)
	if err != nil {
		return shim.Error("Failed to unmarshal data call: " + err.Error())
	}
	if prevDataCall.Status == STATUS_DRAFT {
		if dataCall.Status == STATUS_CANCELLED {
			isCancelled = true
			prevDataCall.Status = dataCall.Status
			prevDataCall.IsLocked = true
		}
		prevDataCall.ForumURL = dataCall.ForumURL

	} else if prevDataCall.Status == STATUS_ISSUED {
		if prevDataCall.ProposedDeliveryDate != dataCall.ProposedDeliveryDate {
			// update datacalllog
			datacallIssueLog := DataCallLog{dataCall.ID, dataCall.Version, ActionDeliveryDateUpdate.ActionID,
				ActionDeliveryDateUpdate.ActionDesc, dataCall.UpdatedTs, dataCall.UpdatedBy}
			datacallIssueLogAsBytes, _ := json.Marshal(datacallIssueLog)
			this.LogDataCallTransaction(stub, string(datacallIssueLogAsBytes))
		}
		prevDataCall.ProposedDeliveryDate = dataCall.ProposedDeliveryDate
		prevDataCall.ForumURL = dataCall.ForumURL
		prevDataCall.ExtractionPatternID = dataCall.ExtractionPatternID
		prevDataCall.ExtractionPatternTs = dataCall.ExtractionPatternTs
		prevDataCall.IsLocked = true
	}

	//updating all the dataCalls with isLocked = true, when isCancelled
	if isCancelled {
		var pks []string = []string{DATA_CALL_PREFIX, dataCall.ID}

		//Step-1: getting the previous data for the data call
		resultsIterator, errmsg := stub.GetStateByPartialCompositeKey(DOCUMENT_TYPE, pks)
		if errmsg != nil {
			logger.Error("Failed to get state for previous data calls")
			return shim.Error("Failed to get state for previous data calls")
		}
		defer resultsIterator.Close()
		for resultsIterator.HasNext() {
			prevDataCallAsBytes, err := resultsIterator.Next()
			if err != nil {
				return shim.Error("Failed to iterate data call")
			}

			var prevDataCall DataCall
			err = json.Unmarshal([]byte(prevDataCallAsBytes.GetValue()), &prevDataCall)
			if err != nil {
				logger.Error("Failed to unmarshal data call: ")
				return shim.Error("Failed to unmarshal data call: " + err.Error())
			}

			//dont update the current update as it is already being updated
			if prevDataCall.Version != dataCall.Version {
				prevDataCall.IsLocked = true
				//creating composite key to save the previous DataCall
				var pkForPrevDataCall []string = []string{DATA_CALL_PREFIX, prevDataCall.ID, prevDataCall.Version}
				prevDataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pkForPrevDataCall)
				prevDataAsBytes, _ := json.Marshal(prevDataCall)
				err = stub.PutState(prevDataCallKey, prevDataAsBytes)
				if err != nil {
					logger.Error("Error commiting the previous DataCall")
					return shim.Error("Error commiting the previous DataCall")
				}
			}
		}

	}

	//prevDataCall.ForumURL = dataCall.ForumURL
	//creating composite key to save the previous DataCall
	var pkForPrevDataCall []string = []string{DATA_CALL_PREFIX, prevDataCall.ID, prevDataCall.Version}
	prevDataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pkForPrevDataCall)
	fmt.Println(prevDataCallKey)
	prevDataAsBytes, _ := json.Marshal(prevDataCall)
	err = stub.PutState(prevDataCallKey, prevDataAsBytes)
	if err != nil {
		logger.Error("Error commiting the previous DataCall")
		return shim.Error("Error commiting the previous DataCall")
	}

	//Prepare response to emit
	//patterns := GetExtractionPatternsMap() //GetExtractionPatternById(dataCall.ExtractionPatternID)
	var extPatternResponse ExtractionPatternPayload
	extPatternResponse.DataCallId = dataCall.ID
	extPatternResponse.DataCallVsersion = dataCall.Version
	//extPatternResponse.ExtractionPattern = patterns[dataCall.ExtractionPatternID]
	extPatternResponse.ExtractionPatternID = dataCall.ExtractionPatternID
	extPatternResponse.ExtPatternTs = dataCall.ExtractionPatternTs

	extPatternResponseAsBytes, _ := json.Marshal(extPatternResponse)

	_ = stub.SetEvent(SET_EXTRACTION_PATTERN_EVENT, extPatternResponseAsBytes)

	if dataCall.Status != prevDataCall.Status {
		//change the count of the statuses
		logger.Info("Toggling ", dataCall.Status, " and ", prevDataCall.Status)
		toggleDataCallCount := ToggleDataCallCount{prevDataCall.Status, dataCall.Status}
		datacallIssueLogAsBytes, _ := json.Marshal(toggleDataCallCount)
		dataCallCountAsBytes := this.ToggleDataCallCount(stub, string(datacallIssueLogAsBytes))
		logger.Info("The reply from toggle is ", dataCallCountAsBytes)
	}

	return shim.Success(nil)

}

// function-name: IssueDataCall (invoke)
// params {DataCall json}
// Success :nil
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : Updates the status of DataCall to STATUS_ISSUED
// set all the versions of DataCall to isLocked="true"

func (this *SmartContract) IssueDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("IssueDataCall: enter")
	defer logger.Debug("IssueDataCall: exit")
	logger.Debug("IssueDataCall json received : ", args)

	var dataCall DataCall
	//var key_for_response string
	err := json.Unmarshal([]byte(args), &dataCall)
	if err != nil {
		logger.Error("IssueDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("IssueDataCall: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshelled object ", dataCall)
	if dataCall.ID == "" {
		return shim.Error("ID should not be Empty")

	}
	if dataCall.Version == "" {
		return shim.Error("Version should not be Empty")

	}
	var pks []string = []string{DATA_CALL_PREFIX, dataCall.ID}

	//Step-1: getting the previous data for the data call
	resultsIterator, errmsg := stub.GetStateByPartialCompositeKey(DOCUMENT_TYPE, pks)
	if errmsg != nil {
		logger.Error("Failed to get state for previous data calls")
		return shim.Error("Failed to get state for previous data calls")
	}
	defer resultsIterator.Close()
	for resultsIterator.HasNext() {
		prevDataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Failed to iterate data call")
		}

		var prevDataCall DataCall
		err = json.Unmarshal([]byte(prevDataCallAsBytes.GetValue()), &prevDataCall)
		if err != nil {
			logger.Error("Failed to unmarshal data call: ")
			return shim.Error("Failed to unmarshal data call: " + err.Error())
		}
		//updating the particular version which is coming in current DataCall
		if prevDataCall.IsLocked == true {
			return shim.Error("DataCall is Locked, as it has already been Issued or Cancelled")

		} else if prevDataCall.Version == dataCall.Version {
			if prevDataCall.Status == STATUS_DRAFT {
				if dataCall.Status != STATUS_ISSUED {
					return shim.Error("Invalid Status ")
				}
				prevDataCall.Status = dataCall.Status
				prevDataCall.IsLocked = true

			}
		}
		prevDataCall.IsLocked = true
		//creating composite key to save the previous DataCall
		var pkForPrevDataCall []string = []string{DATA_CALL_PREFIX, prevDataCall.ID, prevDataCall.Version}
		prevDataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pkForPrevDataCall)
		//key_for_response = prevDataCallKey
		prevDataAsBytes, _ := json.Marshal(prevDataCall)
		err = stub.PutState(prevDataCallKey, prevDataAsBytes)

		// update datacalllog
		datacallIssueLog := DataCallLog{dataCall.ID, dataCall.Version, ActionIssued.ActionID,
			ActionIssued.ActionDesc, dataCall.UpdatedTs, dataCall.UpdatedBy}
		datacallIssueLogAsBytes, _ := json.Marshal(datacallIssueLog)
		this.LogDataCallTransaction(stub, string(datacallIssueLogAsBytes))

		if err != nil {
			logger.Error("Error commiting the previous DataCall")
			return shim.Error("Error commiting the previous DataCall")
		}

		if dataCall.Status != prevDataCall.Status {
			//change the count of the statuses
			logger.Info("Toggling ", dataCall.Status, " and ", prevDataCall.Status)
			toggleDataCallCount := ToggleDataCallCount{prevDataCall.Status, dataCall.Status}
			datacallIssueLogAsBytes, _ := json.Marshal(toggleDataCallCount)
			dataCallCountAsBytes := this.ToggleDataCallCount(stub, string(datacallIssueLogAsBytes))
			logger.Info("The reply from toggle is ", dataCallCountAsBytes)
		}

	}

	return shim.Success(nil)

}

// TODO Remove this function as it must be done from API End through orchestration
// function-name: SaveAndIssueDataCall (invoke)
// params {DataCall json}
// Success :nil
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : Updates the status of DataCall to STATUS_ISSUED
// set all the versions of DataCall to isLocked="true"

func (this *SmartContract) SaveAndIssueDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("SaveAndIssueDataCall: enter")
	defer logger.Debug("SaveAndIssueDataCall: exit")
	logger.Debug("SaveAndIssueDataCall json received : ", args)

	var dataCall DataCall
	err := json.Unmarshal([]byte(args), &dataCall)

	if err != nil {
		logger.Error("SaveAndIssueDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("SaveAndIssueDataCall: Error during json.Unmarshal").Error())
	}

	dataCall.Status = STATUS_DRAFT

	saveDraftJson, _ := json.Marshal(dataCall)
	saveDraftResponse := this.SaveNewDraft(stub, string(saveDraftJson))
	if saveDraftResponse.Status != 200 {
		logger.Error("SaveAndIssueDataCall: Unable to Save a new Draft: ", saveDraftResponse.Message)
		return shim.Error(errors.New("SaveAndIssueDataCall: Unable to Save a new Draft").Error())
	}

	logger.Debug("SaveAndIssueDataCall: Successfully Saved a new datacall, proceeding to issue data call")
	issueDataCall := dataCall
	issueVersion, err := strconv.Atoi(dataCall.Version)
	issueDataCall.Version = generateVersion(issueVersion)
	issueDataCall.Status = STATUS_ISSUED

	issueDataCallJson, _ := json.Marshal(issueDataCall)
	logger.Debug("SaveAndIssueDataCall: issueDataCallJson > ", issueDataCallJson)
	issueDataCallResponse := this.IssueDataCall(stub, string(issueDataCallJson))
	if issueDataCallResponse.Status != 200 {
		logger.Error("SaveAndIssueDataCall: Unable to Save a new Draft: ", issueDataCallResponse.Message)
		return shim.Error(errors.New("SaveAndIssueDataCall: Unable to Save a new Draft").Error())
	}
	return shim.Success(issueDataCallResponse.Payload)
}

func (this *SmartContract) toggleDataCallCount(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Info("REACHED INSIDE************** ")

	version := generateVersion(0)

	var pks []string = []string{DATA_CALL_PREFIX, "12345678", version}
	countPatternKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)

	fmt.Println("KEY IS [")
	fmt.Println(countPatternKey)
	logger.Info(countPatternKey)
	fmt.Println("] KEY IS")
	dataCallAsBytes, err := stub.GetState(countPatternKey)
	if err != nil {
		return shim.Error("Error retreiving data for key" + countPatternKey)
	}

	var prevDataCall DataCallCount
	err = json.Unmarshal(dataCallAsBytes, &prevDataCall)
	if err != nil {
		return shim.Error("Failed to unmarshal data call: " + err.Error())
	}
	logger.Info("GOT THE RESPONSE ", prevDataCall)

	return shim.Success(nil)

}

// func (this *openIDLCC) putDataCallCount(stub shim.ChaincodeStubInterface, args string) pb.Response {
// 	logger.Debug("putDataCallCount: enter")
// 	defer logger.Debug("putDataCallCount: exit")
// 	logger.Info("putDataCallCount json received : ", args)
// 	if len(args) < 1 {
// 		return shim.Error("Incorrect number of arguments!!")
// 	}

// 	var pks []string = []string{DATA_CALL_PREFIX, "COUNT", "10"}
// 	countPatternKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)

// 	args3 := map[string]int{"ISSUED": 1,"DRAFT": 0, "CANCELLED": 0}
// 	var dataCallCount = DataCallCount{counts: args3}
// 	fmt.Println("REACHED INSIDE 3 " + countPatternKey)

// 	logger.Info("REACHED INSIDE 4 ", dataCallCount)

// 	prevDataCallAsBytes, _ := json.Marshal(dataCallCount)
// 	logger.Info("REACHED INSIDE 5 ", dataCallCount)
// 	fmt.Println(prevDataCallAsBytes)
// 	//saving the previous DataCall
// 	err := stub.PutState(countPatternKey, prevDataCallAsBytes)
// 	logger.Info("SaveNewDraft: Previous Datacall saved!")
// 	if err != nil {
// 		logger.Info("SaveNewDraft: Previous Datacall saved!111111")
// 		logger.Error("Error commiting the previous DataCall")
// 		return shim.Error("Error commiting the previous DataCall")
// 	}
// 	logger.Info("SaveNewDraft: Previous Datacall saved!2222222")
// 	dataCallAsBytes, err := stub.GetState(countPatternKey)
// 	fmt.Println("SaveNewDraft: Previous Datacall saved!33333333 ")
// 	fmt.Println(dataCallAsBytes)
// 	if err != nil {
// 		logger.Info("SaveNewDraft: Previous Datacall saved!44444444444")
// 		return shim.Error("Error retreiving data for key")
// 	}

// 	var prevDataCall DataCallCount
// 	logger.Info("SaveNewDraft: Previous Datacall saved!55555555")
// 	err = json.Unmarshal(dataCallAsBytes, &prevDataCall)
// 	logger.Info("SaveNewDraft: Previous Datacall saved!66666666")
// 	if err != nil {
// 		logger.Info("SaveNewDraft: Previous Datacall saved!777777")
// 		return shim.Error("Failed to unmarshal data call: " + err.Error())
// 	}
// 	logger.Info("GOT THE RESPONSE ", prevDataCall)

// 	return shim.Success(nil)

// }

func (this *SmartContract) ToggleDataCallCount(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("ToggleDataCallCount: enter")
	defer logger.Debug("ToggleDataCallCount: exit")
	logger.Debug("ToggleDataCallCount json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}

	var toggleDataCallCount ToggleDataCallCount
	err := json.Unmarshal([]byte(args), &toggleDataCallCount)
	if err != nil {
		logger.Error("ToggleDataCallCount: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("ToggleDataCallCount: Error during json.Unmarshal").Error())
	}
	logger.Info("Unmarshalled object ", toggleDataCallCount)

	getDataCallCount := GetDataCallCount{"123456", "1"}
	datacallIssueLogAsBytes, _ := json.Marshal(getDataCallCount)
	dataCallCountAsBytes := this.GetDataCallCount(stub, string(datacallIssueLogAsBytes))
	var dataCallCount DataCallCount
	err = json.Unmarshal(dataCallCountAsBytes.Payload, &dataCallCount)

	logger.Info("The retrieved data is ", dataCallCount)

	if toggleDataCallCount.OriginalStatus == "ISSUED" {
		dataCallCount.ISSUED = dataCallCount.ISSUED - 1
	} else if toggleDataCallCount.OriginalStatus == "DRAFT" {
		dataCallCount.DRAFT = dataCallCount.DRAFT - 1
	} else if toggleDataCallCount.OriginalStatus == "CANCELLED" {
		dataCallCount.CANCELLED = dataCallCount.CANCELLED - 1
	}

	if toggleDataCallCount.NewStatus == "ISSUED" {
		dataCallCount.ISSUED = dataCallCount.ISSUED + 1
	} else if toggleDataCallCount.NewStatus == "DRAFT" {
		dataCallCount.DRAFT = dataCallCount.DRAFT + 1
	} else if toggleDataCallCount.NewStatus == "CANCELLED" {
		dataCallCount.CANCELLED = dataCallCount.CANCELLED + 1
	}

	dataCallCount.Version = generateVersion(0)

	var pks []string = []string{DATA_CALLCOUNT_PREFIX, dataCallCount.ID, dataCallCount.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENTCOUNT_TYPE, pks)
	//dataCallKey := DATA_CALL_PREFIX + dataCall.ID + dataCall.Version
	logger.Info("In data call create ", dataCallKey)
	// Checking the ledger to confirm that the dataCall doesn't exist

	dataCallAsBytes, _ := json.Marshal(dataCallCount)
	err = stub.PutState(dataCallKey, dataCallAsBytes)

	if err != nil {
		logger.Error("Error commiting data for key: ", dataCallKey)
		return shim.Error("Error committing data for key: " + dataCallKey)
	}

	return shim.Success(nil)

}

func (this *SmartContract) GetDataCallCount(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetDataCallCount: enter")
	defer logger.Debug("GetDataCallCount: exit")

	var getDataCallCount GetDataCallCount
	err := json.Unmarshal([]byte(args), &getDataCallCount)
	if err != nil {
		logger.Error("GetDataCallCount: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetDataCallCount: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshalled object ", getDataCallCount)

	var pks []string = []string{DATA_CALLCOUNT_PREFIX, getDataCallCount.ID, getDataCallCount.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENTCOUNT_TYPE, pks)
	dataCallAsBytes, err := stub.GetState(dataCallKey)
	if err != nil {
		logger.Error("Error retreiving data for key ", dataCallKey)
		return shim.Error("Error retreiving data for key" + dataCallKey)
	}
	return shim.Success(dataCallAsBytes)

}

func (this *SmartContract) UpdateDataCallCount(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("UpdateDataCallCount: enter")
	defer logger.Debug("UpdateDataCallCount: exit")
	logger.Debug("UpdateDataCallCount json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}

	var dataCallCount DataCallCount
	err := json.Unmarshal([]byte(args), &dataCallCount)
	if err != nil {
		logger.Error("UpdateDataCallCount: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("UpdateDataCallCount: Error during json.Unmarshal").Error())
	}
	logger.Info("Unmarshalled object ", dataCallCount)

	dataCallCount.ID = "123456"
	dataCallCount.Version = generateVersion(0)

	var pks []string = []string{DATA_CALLCOUNT_PREFIX, dataCallCount.ID, dataCallCount.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENTCOUNT_TYPE, pks)
	//dataCallKey := DATA_CALL_PREFIX + dataCall.ID + dataCall.Version
	logger.Info("In data call create ", dataCallKey)
	// Checking the ledger to confirm that the dataCall doesn't exist

	dataCallAsBytes, _ := json.Marshal(dataCallCount)
	err = stub.PutState(dataCallKey, dataCallAsBytes)

	if err != nil {
		logger.Error("Error commiting data for key: ", dataCallKey)
		return shim.Error("Error committing data for key: " + dataCallKey)
	}

	return shim.Success(nil)

}

// SearchDataCalls retrives all data calls that match given criteria. If startindex and pageSize are not provided,
// this method returns the complete list of data calls. If version = latest, the it returns only latest version of a data call
// using the specified criteria. If version = all, it returns all data calls with their versions as individual items in list.
// params {json}: {
//  "startIndex":"optional",
//  "pageSize":"optional",
//  "version": "latest or all"
//  "status" :"DRAFT OR ISSUED OR CANCELLED"}
// Success {byte[]}: byte[]
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (this *SmartContract) SearchDataCalls(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("SearchDataCalls: enter")
	defer logger.Debug("SearchDataCalls: exit")
	logger.Debug("SearchDataCalls json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}
	var searchCriteria SearchCriteria
	err := json.Unmarshal([]byte(args), &searchCriteria)
	if err != nil {
		logger.Error("SearchDataCalls: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("SearchDataCalls: Error during json.Unmarshal").Error())
	}
	logger.Debug("SearchDataCalls: Unmarshalled object ", searchCriteria)
	startIndex := searchCriteria.StartIndex
	pageSize := searchCriteria.PageSize
	version := searchCriteria.Version
	searchKey := searchCriteria.SearchKey
	var isLatest string
	if version == LATEST_VERSION {
		isLatest = "true"
	} else {
		isLatest = "false"
	}
	status := searchCriteria.Status
	if status == "" {
		logger.Error("SearchDataCalls: Status not present, You must pass Status in agument")
		return shim.Error("SearchDataCalls: Status not present, You must pass Status in agument")
	}
	var queryStr string

	if status == STATUS_ISSUED {
		queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\",\"$or\":[{\"name\":{\"$regex\":\"%s\"}},{\"description\":{\"$regex\":\"%s\"}},{\"lineOfBusiness\":{\"$regex\":\"%s\"}},{\"jurisdiction\":{\"$regex\":\"%s\"}}]},\"use_index\":[\"_design/deadline\"],\"sort\":[{\"deadline\": \"desc\"}]}", DATA_CALL_PREFIX, status, searchKey, searchKey, searchKey, searchKey)
	} else if status == STATUS_DRAFT {
		queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\", \"isLatest\":%s, \"isLocked\":false, \"$or\":[{\"name\":{\"$regex\":\"%s\"}},{\"description\":{\"$regex\":\"%s\"}},{\"lineOfBusiness\":{\"$regex\":\"%s\"}},{\"jurisdiction\":{\"$regex\":\"%s\"}}]},\"use_index\":[\"_design/updatedTs\"],\"sort\":[{\"updatedTs\": \"desc\"}]}", DATA_CALL_PREFIX, status, isLatest, searchKey, searchKey, searchKey, searchKey)
	} else if status == STATUS_CANCELLED {
		queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\", \"$or\":[{\"name\":{\"$regex\":\"%s\"}},{\"description\":{\"$regex\":\"%s\"}},{\"lineOfBusiness\":{\"$regex\":\"%s\"}},{\"jurisdiction\":{\"$regex\":\"%s\"}}]},\"use_index\":[\"_design/deadline\"],\"sort\":[{\"deadline\": \"desc\"}]}", DATA_CALL_PREFIX, status, searchKey, searchKey, searchKey, searchKey)
	}

	logger.Debug("SearchDataCalls: Selector ", queryStr)
	logger.Debug("startIndex", "pageSize", "version", "status", startIndex, pageSize, version, status)

	var dataCalls []DataCall
	startTime := time.Now()
	resultsIterator, err := stub.GetQueryResult(queryStr)
	elapsedTime := time.Since(startTime)
	logger.Info("Time consumed to get Data Calls", elapsedTime)
	defer resultsIterator.Close()
	if err != nil {
		logger.Error("Failed to get state for all the data calls")
		return shim.Error("Failed to get state for all the data calls")
	}

	if !resultsIterator.HasNext() {
		dataCallsAsByte, _ := json.Marshal(dataCalls)
		logger.Debug("SearchDataCalls: dataCallsAsByte", dataCallsAsByte)
		//return shim.Error(errors.New("SearchDataCalls :DataCall not found ").Error())
		return shim.Success(dataCallsAsByte)
	}

	for resultsIterator.HasNext() {
		dataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate data call")
			return shim.Error("Failed to iterate data call")
		}

		var dataCall DataCall
		err = json.Unmarshal([]byte(dataCallAsBytes.GetValue()), &dataCall)
		if err != nil {
			return shim.Error("Failed to unmarshal data call: " + err.Error())
		}

		dataCalls = append(dataCalls, dataCall)
	}

	//var paginatedDataCall []DataCall
	var paginatedDataCalls []DataCall
	paginatedDataCalls = paginate(dataCalls, startIndex, pageSize)
	var dataCallList DataCallList
	dataCallList.TotalNoOfRecords = len(dataCalls)

	var IdAndVersionMap map[string]string
	IdAndVersionMap = make(map[string]string)
	var dataCallIDs []string
	var dataCallVersions []string
	for dataCallIndex := 0; dataCallIndex < len(paginatedDataCalls); dataCallIndex++ {
		if dataCallIndex == 0 {
			IdAndVersionMap[paginatedDataCalls[dataCallIndex].ID] = paginatedDataCalls[dataCallIndex].Version
			dataCallIDs = append(dataCallIDs, `"`+paginatedDataCalls[dataCallIndex].ID+`"`)
			dataCallVersions = append(dataCallVersions, `"`+paginatedDataCalls[dataCallIndex].Version+`"`)
		} else {
			IdAndVersionMap[paginatedDataCalls[dataCallIndex].ID] = paginatedDataCalls[dataCallIndex].Version
			dataCallIDs = append(dataCallIDs, `,`+`"`+paginatedDataCalls[dataCallIndex].ID+`"`)
			dataCallVersions = append(dataCallVersions, `,`+`"`+paginatedDataCalls[dataCallIndex].Version+`"`)
		}
	}

	startTimeForAll := time.Now()
	/*
		//get ConsentCount map
		startTimeForConsent := time.Now()
		consentCounts := GetConsentsCount(stub, dataCallIDs)
		elapsedTimeForConsent := time.Since(startTimeForConsent)
		logger.Info("Time consumed for consent", elapsedTimeForConsent)
		startTimeForLike := time.Now()
		likeCounts := GetLikesCount(stub, dataCallIDs, IdAndVersionMap)
		elapsedTimeForLike := time.Since(startTimeForLike)
		logger.Info("Time consumed for Like", elapsedTimeForLike)
	*/

	startTimeForReport := time.Now()
	latestReports := GetLatestaReport(stub, dataCallIDs, dataCallVersions)
	elapsedTimeForReport := time.Since(startTimeForReport)
	logger.Info("Time consumed for Report", elapsedTimeForReport)

	elapsedTimeForAll := time.Since(startTimeForAll)
	logger.Info("Time consumed for all", elapsedTimeForAll)

	for dataCallIndex := 0; dataCallIndex < len(paginatedDataCalls); dataCallIndex++ {
		var dataCallExtended DataCallExtended
		dataCallExtended.DataCall = paginatedDataCalls[dataCallIndex]
		dataCallExtended.Reports = append(dataCallExtended.Reports, latestReports[paginatedDataCalls[dataCallIndex].ID])
		//dataCallExtended.NoOfConsents = consentCounts[paginatedDataCalls[dataCallIndex].ID]
		//dataCallExtended.NoOfLikes = likeCounts[paginatedDataCalls[dataCallIndex].ID]
		dataCallList.DataCalls = append(dataCallList.DataCalls, dataCallExtended)

	}

	dataCallsAsByte, _ := json.Marshal(dataCallList)
	return shim.Success(dataCallsAsByte)

}

// ToggleLike Creates and then toggles likes as a boolean value
// for a datacall
func (this *SmartContract) ToggleLike(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("ToggleLike: enter")
	defer logger.Debug("ToggleLike: exit")
	logger.Debug("ToggleLike > stub.GetChannelID > ", stub.GetChannelID())
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!")
	}
	logger.Debug("ToggleLike: Input Like Json > " + args)
	var like Like
	err := json.Unmarshal([]byte(args), &like)

	// TODO: Investigate why it is returning nill despite the fact data call exists in other channel
	// Check if the data call corresponding to this like exists on Global channel

	var getDataCall GetDataCall
	getDataCall.ID = like.DatacallID
	getDataCall.Version = like.DataCallVersion
	getDataCallAsBytes, _ := json.Marshal(getDataCall)
	getDataCallReqJson := string(getDataCallAsBytes)
	logger.Debug("ToggleLike: getDataCallReqJson > ", getDataCallReqJson)
	var GetDataCallByIdAndVersionFunc = "GetDataCallByIdAndVersion"
	getDataCallRequest := ToChaincodeArgs(GetDataCallByIdAndVersionFunc, getDataCallReqJson)
	logger.Debug("ToggleLike: getDataCallRequest", getDataCallRequest)
	getDataCallResponse := stub.InvokeChaincode(DEFAULT_CHAINCODE_NAME, getDataCallRequest, DEFAULT_CHANNEL)
	logger.Debug("ToggleLike: getDataCallResponse > ", getDataCallResponse)
	logger.Debug("ToggleLike: getDataCallResponse.Status ", getDataCallResponse.Status)
	logger.Debug("ToggleLike: getDataCallResponse.Payload", string(getDataCallResponse.Payload))
	if getDataCallResponse.Status != 200 {
		logger.Error("ToggleLike: Unable to Fetch DataCallId and Version due to Error: ", err)
		return shim.Error(errors.New("ToggleLike: Unable to Fetch DataCallId and Version due to Error").Error())
	}

	if len(getDataCallResponse.Payload) <= 0 {
		logger.Error("ToggleLike: No Matching datacallId and datacallVersion specified in Like message")
		return shim.Error(errors.New("No Matching datacallId and datacallVersion specified in Like message").Error())

	}

	if err != nil {
		logger.Error("ToggleLike: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("ToggleLike: Error during json.Unmarshal").Error())
	}

	var pks []string = []string{LIKE_PREFIX, like.DatacallID, like.DataCallVersion, like.OrganizationID}
	likeKey, _ := stub.CreateCompositeKey(LIKE_DOCUMENT_TYPE, pks)

	logger.Info("ToggleLike: Get Like from World State")
	logger.Info("ToggleLike: likeKey > ", likeKey)
	prevLikeAsBytes, _ := stub.GetState(likeKey)
	var likeAsBytes []byte

	logger.Debug("ToggleLike: PreviousLikeAsBytes > ", prevLikeAsBytes)

	if prevLikeAsBytes == nil {

		if like.Liked == false {
			return shim.Error("Can't Unlike DataCall, as it is not Liked")
		}

		// like doesn't exist creating new like
		logger.Info("ToggleLike: No Previous Like Found, Create new like entry")
		likeAsBytes, _ = json.Marshal(like)
		err = stub.PutState(likeKey, likeAsBytes)
		if err != nil {
			return shim.Error("ToggleLike: Error committing data for key: " + likeKey)
		}
		logger.Debug("ToggleLike: Like Committed to World State, Raising a ToggleLikeEvent")
		_ = stub.SetEvent(TOGGLE_LIKE_EVENT, likeAsBytes)

	} else {
		// compare if like has already been performed for a given organization id
		var prevLike Like
		err := json.Unmarshal(prevLikeAsBytes, &prevLike)
		if err != nil {
			return shim.Error("Unable to umarshall previous like for key : " + likeKey)
		}
		logger.Info("ToggleLike: Comparing Previous and new like status, Previous Like and new Likes as follows ", prevLike.Liked, like.Liked)
		if prevLike.Liked != like.Liked {
			//compare if there is a change in state of like, update like for change in state
			logger.Debug("Toggle like status")
			likeAsBytes, _ := json.Marshal(like)
			err = stub.PutState(likeKey, likeAsBytes)
			if err != nil {
				return shim.Error("ToggleLike: Error committing data for key: " + likeKey)
			}
			logger.Debug("ToggleLike: Like Committed to World State, Raising a ToggleLikeEvent")
			_ = stub.SetEvent(TOGGLE_LIKE_EVENT, likeAsBytes)
		}
	}

	return shim.Success(nil)
}

// Returns List of carriers Liked for a specific data call, based on dataCallID and dataCallVersion
// Request param- {"dataCallID":" ", "dataCallVersion":" "}
func (this *SmartContract) GetLikesByDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {

	logger.Debug("GetLikesByDataCall: enter")
	defer logger.Debug("GetLikesByDataCall: exit")
	if len(args) < 1 {
		return shim.Error("GetLikesByDataCall: Incorrect number of arguments!")
	}
	var likeList []ListLikeResponse
	var getLikesByDataCallRequest GetLikesByDataCallRequest
	err := json.Unmarshal([]byte(args), &getLikesByDataCallRequest)

	if err != nil {
		logger.Error("GetLikesByDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetLikesByDataCall: Error during json.Unmarshal").Error())
	}

	if getLikesByDataCallRequest.DataCallID == "" || getLikesByDataCallRequest.DataCallVersion == "" {
		return shim.Error("DataCallID or DataCallVersion can't be empty")
	}

	// Get current channel
	logger.Info("GetLikesByDataCall: currentChannelID > ", stub.GetChannelID())

	var pks []string
	pks = []string{LIKE_PREFIX, getLikesByDataCallRequest.DataCallID, getLikesByDataCallRequest.DataCallVersion}

	logger.Info("GetLikesByDataCall: Key ", pks)
	resultsIterator, err := stub.GetStateByPartialCompositeKey(LIKE_DOCUMENT_TYPE, pks)
	defer resultsIterator.Close()

	if err != nil {
		// proceed without any further action to check consents on other channels
		logger.Error("GetLikesByDataCall: Error fetching like on this channel: ", stub.GetChannelID()+"error ", err)
	}
	// Check the variable existed
	if !resultsIterator.HasNext() {
		// proceed without any further action to check consents on other channels
		logger.Info("GetLikesByDataCall: No like found on current channel, proceed to next channel ")

	} else {
		var i int
		logger.Debug("GetLikesByDataCall: Iterating over list of like")
		for i = 0; resultsIterator.HasNext(); i++ {
			currentLikeAsBytes, nextErr := resultsIterator.Next()
			if nextErr != nil {
				return shim.Error("GetLikesByDataCall: Failed to iterate over like")
			}
			var currentLike Like
			err = json.Unmarshal([]byte(currentLikeAsBytes.GetValue()), &currentLike)
			if err != nil {
				return shim.Error("GetLikesByDataCall: Failed to unmarshal like: " + err.Error())
			}
			var listLikeResponse ListLikeResponse
			listLikeResponse.Like = currentLike
			//listConsentResponse.CarrierName = ""
			likeList = append(likeList, listLikeResponse)

		}
		logger.Info("GetLikesByDataCall: Likes fetched for current channel")
	}
	likeListAsByte, _ := json.Marshal(likeList)
	logger.Debug("GetLikesByDataCall: LikeListAsByte", likeListAsByte)
	return shim.Success(likeListAsByte)
}

// Returns List of carriers Consented for a specific data call, based on dataCallID and dataCallVersion on requested channels
// Request param- {"dataCallID":" ", "dataCallVersion":" ", "channelList":[{"channelName": "channel1","chaincodeName": "openidl-cc-channel1"}]}
func (this *SmartContract) ListLikesByDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {

	logger.Debug("ListLikesByDataCall: enter")
	defer logger.Debug("ListLikesByDataCall: exit")
	if len(args) < 1 {
		return shim.Error("ListLikesByDataCall: Incorrect number of arguments!")
	}

	//listDataCallRequestJson := args[0]
	//logger.Info("ListLikesByDataCall: Request > " + listDataCallRequestJson)
	var likeList []ListLikeResponse
	var listLikeRequest ListLikeRequest
	err := json.Unmarshal([]byte(args), &listLikeRequest)

	if err != nil {
		logger.Error("ListLikesByDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("ListLikesByDataCall: Error during json.Unmarshal").Error())
	}
	logger.Debug("ListLikesByDataCall: Unmarshalled object ", listLikeRequest)

	if listLikeRequest.DataCallID == "" || listLikeRequest.DataCallVersion == "" {
		return shim.Error("DataCallID or DataCallVersion can't be empty")
	}

	var pks []string
	pks = []string{LIKE_PREFIX, listLikeRequest.DataCallID, listLikeRequest.DataCallVersion}

	logger.Info("ListLikesByDataCall: Key ", pks)
	resultsIterator, err := stub.GetStateByPartialCompositeKey(LIKE_DOCUMENT_TYPE, pks)
	defer resultsIterator.Close()

	if err != nil {
		// proceed without any further action to check consents on other channels
		logger.Error("ListLikesByDataCall: Error fetching likes on this channel: ", stub.GetChannelID()+"error ", err)
	}
	// Check the variable existed
	if !resultsIterator.HasNext() {
		// proceed without any further action to check consents on other channels
		logger.Info("ListLikesByDataCall: No like found on current channel, proceed to next channel ")
	} else {
		var i int
		logger.Debug("ListLikesByDataCall: Iterating over list of likes")
		for i = 0; resultsIterator.HasNext(); i++ {
			currentLikeAsBytes, nextErr := resultsIterator.Next()
			if nextErr != nil {
				return shim.Error("ListLikesByDataCall: Failed to iterate over Consent")
			}
			var currentLike Like
			err = json.Unmarshal([]byte(currentLikeAsBytes.GetValue()), &currentLike)
			if err != nil {
				return shim.Error("Failed to unmarshal data call: " + err.Error())
			}
			var listLikeResponse ListLikeResponse
			listLikeResponse.Like = currentLike
			//listConsentResponse.CarrierName = ""
			likeList = append(likeList, listLikeResponse)

		}
		logger.Info("ListLikesByDataCall: likes fetched for current channel, moving on to other channels")
	}

	// Get data from other channels mentioned in the client request and recieved any other channels mentioned in the client request
	//var channels Channels
	//channels.ChannelIDs = make([]string, len(crossInvocationChannels.ChannelIDs))
	//copy(channels.ChannelIDs[:], crossInvocationChannels.ChannelIDs)
	//logger.Debug("ListLikesByDataCall: Requested additional channels data from client >> ", channels.ChannelIDs)
	totalChannels := len(listLikeRequest.ChannelList)
	logger.Debug("ListLikesByDataCall: Requested additional channels data from client >> ", listLikeRequest.ChannelList)
	logger.Debug("ListLikesByDataCall: Total Number of channels > ", totalChannels)

	// Get current channel
	//currentChannelID := stub.GetChannelID()
	//logger.Info("InvokeChaincodeOnChannel: currentChannelID > ", currentChannelID)

	var channenlIndex int
	for channenlIndex = 0; channenlIndex < totalChannels; channenlIndex++ {
		var getLikesReq GetLikesByDataCallRequest
		getLikesReq.DataCallID = listLikeRequest.DataCallID
		getLikesReq.DataCallVersion = listLikeRequest.DataCallVersion
		getLikesReqAsBytes, _ := json.Marshal(getLikesReq)
		getLikesReqJson := string(getLikesReqAsBytes)
		var GetLikesByDataCallFunc = "GetLikesByDataCall"
		getLikesByDataCallRequest := ToChaincodeArgs(GetLikesByDataCallFunc, getLikesReqJson)
		logger.Debug("ListLikesByDataCall: getLikesByDataCallRequest", getLikesByDataCallRequest)
		logger.Info("ListLikesByDataCall: GetLikesByDataCall request json " + getLikesReqJson)
		//var invokeResponse pb.Response

		// fetch only if the requested channel is different from current channel
		//if channels.ChannelIDs[channenlIndex] != currentChannelID {
		logger.Debug("ListLikesByDataCall: Fetching likes from channel ", listLikeRequest.ChannelList[channenlIndex].ChannelName)

		// Modify channel in the request before sending with current channelID to prevent any loops
		//listLikeRequest.ChannelIDs = make([]string, 1)
		//listLikeRequest.ChannelIDs[0] = currentChannelID
		chaincodeName := listLikeRequest.ChannelList[channenlIndex].ChaincodeName
		channelName := listLikeRequest.ChannelList[channenlIndex].ChannelName
		invokeResponse := stub.InvokeChaincode(chaincodeName, getLikesByDataCallRequest, channelName)
		//invokeResponse := InvokeChaincode(stub, chaincodeName, "GetLikesByDataCall", getLikesReqJson, channelName)
		if invokeResponse.Status != 200 {
			logger.Error("ListLikesByDataCall: Unable to Invoke cross channel query GetLikesByDataCall: ", invokeResponse)
			// Do not block functionality and proceed to return the original channels like list
			//}

		} else if len(invokeResponse.Payload) <= 0 {
			logger.Debug("ListLikesByDataCall: ErrorInvokeResponse from another channel ", string(invokeResponse.Payload))
		} else {
			var invokeListLikeResponse []ListLikeResponse
			logger.Debug("ListLikesByDataCall: InvokeResponse from another channel ", string(invokeResponse.Payload))
			json.Unmarshal(invokeResponse.Payload, &invokeListLikeResponse)
			likeList = append(likeList, invokeListLikeResponse...)

		}
		//}
	}

	likeListAsByte, _ := json.Marshal(likeList)
	logger.Debug("ListLikesByDataCall: likesListAsByte", likeListAsByte)
	return shim.Success(likeListAsByte)

}

// GetLikeByDataCallAndOrganization Returns list of likes based on input criteria
func (this *SmartContract) GetLikeByDataCallAndOrganization(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Debug("GetLikeByDataCallAndOrganization: enter")
	defer logger.Debug("GetLikeByDataCallAndOrganization: exit")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!")
	}
	listDataCallRequestJson := args[0]
	logger.Debug("GetLikeByDataCallAndOrganization: Request > " + listDataCallRequestJson)
	var listLikeRequest GetLikeByDataCallAndOrganizationRequest
	var likesList []ListLikeResponse

	err := json.Unmarshal([]byte(listDataCallRequestJson), &listLikeRequest)
	like := listLikeRequest.Like
	if err != nil {
		logger.Error("GetLikeByDataCallAndOrganization: Error during json.Unmarshal: ", err)
		likesListAsByte, _ := json.Marshal(likesList)
		logger.Debug("GetLikeByDataCallAndOrganization: likesListAsByte", likesListAsByte)
	}

	// Create partial composite key to fetch liks based on DataCall Id and DataCall Version
	var pks []string
	pks = []string{LIKE_PREFIX, like.DatacallID, like.DataCallVersion, like.OrganizationID}
	logger.Debug("GetLikeByDataCallAndOrganization: Key ", pks)
	resultsIterator, err := stub.GetStateByPartialCompositeKey(LIKE_DOCUMENT_TYPE, pks)
	defer resultsIterator.Close()

	if err != nil {
		logger.Error("GetLikeByDataCallAndOrganization: No Likes found on current channel due to error, proceed to next channel ", err)
		// proceed without any further action to check consents on other channels
	}
	// Check the variable existed
	if !resultsIterator.HasNext() {
		logger.Error("GetLikeByDataCallAndOrganization: No Likes found on current channel, proceed to next channel ")
		// proceed without any further action to check consents on other channels
	} else {
		var i int
		logger.Debug("ListLikesByDataCall: Iterating over list of likes")
		for i = 0; resultsIterator.HasNext(); i++ {
			currentLikeAsBytes, nextErr := resultsIterator.Next()
			if nextErr != nil {
				return shim.Error("GetLikeByDataCallAndOrganization: Failed to iterate Like Entry call")
			}
			var currentLike Like
			err = json.Unmarshal([]byte(currentLikeAsBytes.GetValue()), &currentLike)
			if err != nil {
				return shim.Error("GetLikeByDataCallAndOrganization: Failed to unmarshal data call: " + err.Error())
			}
			var listLikeResponse ListLikeResponse
			listLikeResponse.Like = currentLike

			// update organization/Carrier Name in the like's response paylod
			listLikeResponse.OrganizationName = ""

			likesList = append(likesList, listLikeResponse)
		}
		logger.Debug("GetLikeByDataCallAndOrganization: Likes fetched for current channel, moving on to other channels")
	}

	likesListAsByte, _ := json.Marshal(likesList)
	logger.Debug("GetLikeByDataCallAndOrganization: likesListAsByte", string(likesListAsByte))
	return shim.Success(likesListAsByte)
}

// Create a new entry of like count
func (this *SmartContract) CreateLikeCountEntry(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CreateLikeCountEntry: enter")
	defer logger.Debug("CreateLikeCountEntry: exit")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!")
	}
	logger.Debug("CreateLikeCountEntry: Input Like Json > " + args)
	var likeCountEntry LikeCountEntry
	err := json.Unmarshal([]byte(args), &likeCountEntry)

	if err != nil {
		logger.Error("CreateLikeCountEntry: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CreateLikeCountEntry: Error during json.Unmarshal").Error())
	}
	transactionId := stub.GetTxID()
	var pks []string = []string{LIKE_PREFIX, likeCountEntry.DatacallID, likeCountEntry.DataCallVersion, transactionId}
	likeKey, _ := stub.CreateCompositeKey(LIKE_DOCUMENT_TYPE, pks)

	// like doesn't exist creating new like
	logger.Debug("CreateLikeCountEntry: Create new like entry")
	likeAsBytes, _ := json.Marshal(likeCountEntry)
	err = stub.PutState(likeKey, likeAsBytes)
	if err != nil {
		return shim.Error("CreateLikeCountEntry: Error committing Like for key: " + likeKey)
	}

	return shim.Success(nil)
}

// Perform count of like count entries
func (this *SmartContract) CountLikes(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CountLikes: enter")
	defer logger.Debug("CountLikes: exit")
	if len(args) < 1 {
		return shim.Error("CountLikes: Incorrect number of arguments!")
	}

	var likeCountEntry LikeCountEntry
	err := json.Unmarshal([]byte(args), &likeCountEntry)

	if err != nil {
		logger.Error("CountLikes: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CountLikes: Error during json.Unmarshal").Error())
	}

	var pks []string = []string{LIKE_PREFIX, likeCountEntry.DatacallID, likeCountEntry.DataCallVersion}
	logger.Info("CountLikes: Key ", pks)
	deltaResultsIterator, deltaErr := stub.GetStateByPartialCompositeKey(LIKE_DOCUMENT_TYPE, pks)

	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("CountLikes: Could not retrieve value for %s: %s", likeCountEntry.DatacallID, deltaErr.Error()))
	}
	defer deltaResultsIterator.Close()

	// Check the variable existed
	if !deltaResultsIterator.HasNext() {
		logger.Info("CountLikes: No Likes found for criteria, returing 0 delta")
		likeCountEntry.Delta = 0
		likeCountEntryAsBytes, _ := json.Marshal(likeCountEntry)
		return shim.Success(likeCountEntryAsBytes)
	}

	var delta = 0
	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {
		currentLikeEntryAsBytes, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error("CountLikes: Failed to iterate Like Entry call")
		}
		var currentLikeCountEntry LikeCountEntry
		err = json.Unmarshal([]byte(currentLikeEntryAsBytes.GetValue()), &currentLikeCountEntry)
		logger.Debug("CountLikes: Count Data > ", currentLikeCountEntry.DatacallID, currentLikeCountEntry.DataCallVersion, currentLikeCountEntry.Delta)
		delta = delta + currentLikeCountEntry.Delta
	}

	likeCountEntry.Delta = delta
	logger.Debug(likeCountEntry.Delta)
	newLikeCountEntry, _ := json.Marshal(likeCountEntry)
	return shim.Success(newLikeCountEntry)
}

//updates like count for a data call based on dataCallID and dataCallVersion
func (this *SmartContract) UpdateLikeCountForDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("UpdateLikeCountForDataCall: enter")
	defer logger.Debug("UpdateLikeCountForDataCall: exit")
	if len(args) < 1 {
		return shim.Error("UpdateLikeCountForDataCall: Incorrect number of arguments!")
	}

	var updateLikeReqest UpdateLikeAndConsentCountReq
	err := json.Unmarshal([]byte(args), &updateLikeReqest)

	if err != nil {
		logger.Error("UpdateLikeCountForDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("UpdateLikeCountForDataCall: Error during json.Unmarshal").Error())
	}
	if updateLikeReqest.DataCallID == "" {
		return shim.Error("DataCallID should not be Empty")

	} else if updateLikeReqest.DataCallVersion == "" {
		return shim.Error("DataCallVersion should not be Empty")

	}

	var pks []string = []string{DATA_CALL_PREFIX, updateLikeReqest.DataCallID, updateLikeReqest.DataCallVersion}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)
	dataCallAsBytes, err := stub.GetState(dataCallKey)
	if err != nil {
		return shim.Error("UpdateLikeCountForDataCall: Error retreiving data for key" + dataCallKey)
	}

	var prevDataCall DataCall
	err = json.Unmarshal(dataCallAsBytes, &prevDataCall)
	if err != nil {
		return shim.Error("UpdateLikeCountForDataCall: Failed to unmarshal data call: " + err.Error())
	}

	//Invoke CountLikes to get the total likes count for the data call
	var getLikesCount LikeCountEntry
	getLikesCount.DatacallID = updateLikeReqest.DataCallID
	getLikesCount.DataCallVersion = updateLikeReqest.DataCallVersion
	getLikesCountJson, _ := json.Marshal(getLikesCount)
	countLikesResponse := this.CountLikes(stub, string(getLikesCountJson))
	if countLikesResponse.Status != 200 || len(countLikesResponse.Payload) <= 0 {
		logger.Error("UpdateLikeCountForDataCall: Unable to CountLikes: ", countLikesResponse)
		return shim.Error(errors.New("UpdateLikeCountForDataCall: Unable to CountLikes").Error())
	}

	var likeCountEntry LikeCountEntry
	json.Unmarshal(countLikesResponse.Payload, &likeCountEntry)
	logger.Debug("UpdateLikeCountForDataCall: InvokeResponse from CountLikes: delta ", (likeCountEntry.Delta))

	//update the dataCall with delta
	prevDataCall.LikeCount = likeCountEntry.Delta
	prevDataAsBytes, _ := json.Marshal(prevDataCall)
	err = stub.PutState(dataCallKey, prevDataAsBytes)
	if err != nil {
		logger.Error("UpdateLikeCountForDataCall: Error updating DataCall for LikeCount")
		return shim.Error("UpdateLikeCountForDataCall: Error updating DataCall for LikeCount: " + err.Error())
	}

	return shim.Success(nil)

}

func GetLikesCount(stub shim.ChaincodeStubInterface, args []string, idAndVersionMap map[string]string) map[string]int {
	logger.Info("Inside GetLikesCount and args are ", args)
	var likeCounts map[string]int
	likeCounts = make(map[string]int)

	var queryStr string
	queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"datacallID\":{\"$in\":[%s]}},\"use_index\":[\"_design/dataCallId\"]}", LIKE_PREFIX, strings.Trim(fmt.Sprint(args), "[]"))

	deltaResultsIterator, _ := stub.GetQueryResult(queryStr)

	defer deltaResultsIterator.Close()

	// Check the variable existed
	if !deltaResultsIterator.HasNext() {
		logger.Info("GetLikesCount: No Likes found for criteria, returning 0 delta")
		return likeCounts
	}

	for deltaResultsIterator.HasNext() {
		likeCountsAsBytes, _ := deltaResultsIterator.Next()
		var tempLikeCounts LikeCountEntry
		_ = json.Unmarshal([]byte(likeCountsAsBytes.GetValue()), &tempLikeCounts)

		if idAndVersionMap[tempLikeCounts.DatacallID] == tempLikeCounts.DataCallVersion {
			likeCounts[tempLikeCounts.DatacallID] = likeCounts[tempLikeCounts.DatacallID] + tempLikeCounts.Delta
		}

		//likeCounts[tempLikeCounts.DatacallID] = likeCounts[tempLikeCounts.DatacallID] + tempLikeCounts.Delta

	}

	return likeCounts
}

func (this *SmartContract) CreateConsent(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CreateConsent: enter")
	defer logger.Debug("CreateConsent: exit")

	//Check if array length is greater than 0
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
	consent := Consent{}
	err := json.Unmarshal([]byte(args), &consent)
	if err != nil {
		return shim.Error("CreateConsent: Error Unmarshalling Controller Call JSON: " + err.Error())
	}

	// TODO: Investigate why it is returning nill despite the fact data call exists in other channel
	// Check if the data call corresponding to this like exists on Global channel

	var dataCall GetDataCall
	dataCall.ID = consent.DatacallID
	dataCall.Version = consent.DataCallVersion
	dataCallAsBytes, _ := json.Marshal(dataCall)
	getDataCallReqJson := string(dataCallAsBytes)
	logger.Debug("CreateConsent: getDataCallReqJson > ", getDataCallReqJson)
	var GetDataCallByIdAndVersionFunc = "GetDataCallByIdAndVersion"
	getDataCallRequest := ToChaincodeArgs(GetDataCallByIdAndVersionFunc, getDataCallReqJson)
	logger.Debug("CreateConsent: getDataCallRequest", getDataCallRequest)
	getDataCallResponse := stub.InvokeChaincode(DEFAULT_CHAINCODE_NAME, getDataCallRequest, DEFAULT_CHANNEL)
	logger.Debug("CreateConsent: getDataCallResponse > ", getDataCallResponse)
	logger.Debug("CreateConsent: getDataCallResponse.Status ", getDataCallResponse.Status)
	logger.Debug("CreateConsent: getDataCallResponse.Payload", string(getDataCallResponse.Payload))
	if getDataCallResponse.Status != 200 {
		logger.Error("CreateConsent: Invalid Data Call ID and Version Specified: ", err)
		return shim.Error(errors.New("CreateConsent: Invalid Data Call ID and Version Specified").Error())
	}

	logger.Debug("Recieved Data Call >> " + string(getDataCallResponse.Payload))
	if len(getDataCallResponse.Payload) <= 0 {
		logger.Error("CreateConsent: No Matching datacallId and datacallVersion specified in Consent message")
		return shim.Error(errors.New("CreateConsent: No Matching datacallId and datacallVersion specified in Consent message").Error())

	}

	pks := []string{CONSENT_PREFIX, consent.DatacallID, consent.DataCallVersion, consent.CarrierID}
	consentKey, _ := stub.CreateCompositeKey(CONSENT_DOCUMENT_TYPE, pks) //CONSENT_PREFIX + consent.DatacallID

	// Checking the ledger to confirm that the controller key doesn't exist
	logger.Debug("CreateConsent: Get Consent from World State")
	previousConsentData, _ := stub.GetState(consentKey)
	logger.Debug("CreateConsent: PreviousConsentAsBytes > ", previousConsentData)

	if previousConsentData != nil {
		logger.Warning("CreateConsent: Consent Already Exist for data call Id: ", consent.DatacallID)
		return shim.Success(nil)
	} else {
		logger.Debug("CreateConsent: Create consent entry")
		consentInBytes, _ := json.Marshal(consent)

		// === Save Controller to state ===
		err = stub.PutState(consentKey, consentInBytes)
		if err != nil {
			return shim.Error("CreateConsent: Error committing data for key: " + consentKey)
		}

		logger.Debug("CreateConsent: Consent Committed to World State, Raising a CreateConsentEvent")

		// Create chaincode event
		_ = stub.SetEvent(CREATE_CONSENT_EVENT, consentInBytes)

	}

	return shim.Success(nil)
}

// Request param- {"dataCallID":"", "dataCallVersion":"", "carrierid":"", "status": }
func (this *SmartContract) UpdateConsentStatus(stub shim.ChaincodeStubInterface, args string) pb.Response {

	logger.Debug("UpdateConsentStatus: Enter")
	if len(args) < 1 {
		return shim.Error("UpdateConsentStatus: Incorrect number of arguments!")
	}
	var consent UpdateConsentStatus
	err := json.Unmarshal([]byte(args), &consent)

	if err != nil {
		logger.Error("UpdateConsentStatus: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("UpdateConsentStatus: Error during json.Unmarshal").Error())
	}

	if consent.DataCallID == "" || consent.DataCallVersion == "" || consent.CarrierID == "" {
		return shim.Error("DataCallID or DataCallVersion or CarrierID can't be empty")
	}

	pks := []string{CONSENT_PREFIX, consent.DataCallID, consent.DataCallVersion, consent.CarrierID}
	consentKey, _ := stub.CreateCompositeKey(CONSENT_DOCUMENT_TYPE, pks)

	logger.Debug("Get Consent from World State")
	consentData, _ := stub.GetState(consentKey)
	logger.Debug("GetConsent: PreviousConsentAsBytes > ", consentData)

	var cc Consent
	err = json.Unmarshal(consentData, &cc)

	// var prevDataCall DataCall
	// err = json.Unmarshal(dataCallAsBytes, &prevDataCall)

	if consentData == nil {
		logger.Error("Error retreiving data for key ", consentKey)
		return shim.Error("Error retreiving data for key" + consentKey)
	} else {
		logger.Debug("Getcosent details for datacall id ", consent.DataCallID)

		cc.Status = consent.Status

		consentDataAsBytes, _ := json.Marshal(cc)
		err = stub.PutState(consentKey, consentDataAsBytes)
		if err != nil {
			logger.Error("Error commiting the cosent status")
			return shim.Error("Error commiting the consent status")
		}

		return shim.Success(nil)
	}

}

func (this *SmartContract) CreateConsentCountEntry(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CreateConsentCountEntry: enter")
	defer logger.Debug("CreateConsentCountEntry: exit")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!")
	}
	logger.Debug("Input Consent Json > " + args)
	var consentCountEntry ConsentCountEntry
	err := json.Unmarshal([]byte(args), &consentCountEntry)

	if err != nil {
		logger.Error("CreateConsentCountEntry: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CreateConsentCountEntry: Error during json.Unmarshal").Error())
	}

	transactionId := stub.GetTxID()
	var pks []string = []string{CONSENT_PREFIX, consentCountEntry.DatacallID, consentCountEntry.DataCallVersion, transactionId}
	consentKey, _ := stub.CreateCompositeKey(CONSENT_DOCUMENT_TYPE, pks)

	// Consent doesn't exist creating new consent
	logger.Debug("CreateConsentCountEntry: Create new consent entry")
	consentAsBytes, _ := json.Marshal(consentCountEntry)
	err = stub.PutState(consentKey, consentAsBytes)
	if err != nil {
		return shim.Error("CreateConsentCountEntry: Error committing Consent for key: " + consentKey)
	}

	return shim.Success(nil)
}

/**
* Function-name : CountConsents (invoke)
* for a datacall
* @params :
{
	"datacallID":"Mandatory",
	"dataCallVersion":"Mandatory",
	"carrierID":""Mandatory,
	"createdTs":"Mandatory",
	"createdBy":"Mandatory"
}
*@property {string} 0 - stringified JSON object.
* * @Success: nil
* @Failure:{"message":"", "errorCode":"sys_err or bus_error"}
* @Description : Counting the number of consents carrier is giving or already given
*/
func (this *SmartContract) CountConsents(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CountConsents: enter")
	defer logger.Debug("CountConsents: exit")
	if len(args) < 1 {
		return shim.Error("CountConsents: Incorrect number of arguments!")
	}

	var consentCountEntry ConsentCountEntry
	err := json.Unmarshal([]byte(args), &consentCountEntry)

	if err != nil {
		logger.Error("CreateConsentCountEntry: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CountConsents: Error during json.Unmarshal").Error())
	}

	var pks []string = []string{CONSENT_PREFIX, consentCountEntry.DatacallID, consentCountEntry.DataCallVersion}
	logger.Info("Key ", pks)
	deltaResultsIterator, deltaErr := stub.GetStateByPartialCompositeKey(CONSENT_DOCUMENT_TYPE, pks)

	if deltaErr != nil {
		return shim.Error(fmt.Sprintf("CountConsents: Could not retrieve value for %s: %s", consentCountEntry.DatacallID, deltaErr.Error()))

	}
	defer deltaResultsIterator.Close()

	// Check the variable existed
	if !deltaResultsIterator.HasNext() {
		logger.Info("CountConsents: No Consents found for criteria, returning 0 delta")
		consentCountEntry.Delta = 0
		consentCountEntryAsBytes, _ := json.Marshal(consentCountEntry)
		return shim.Success(consentCountEntryAsBytes)
	}

	var delta = 0
	var i int
	for i = 0; deltaResultsIterator.HasNext(); i++ {
		currentConsentEntryAsBytes, nextErr := deltaResultsIterator.Next()
		if nextErr != nil {
			return shim.Error("CountConsents: Failed to iterate Consent Entry call")
		}
		var currentConsentCountEntry ConsentCountEntry
		err = json.Unmarshal([]byte(currentConsentEntryAsBytes.GetValue()), &currentConsentCountEntry)
		delta = delta + currentConsentCountEntry.Delta
	}

	consentCountEntry.Delta = delta
	newConsentCountEntry, _ := json.Marshal(consentCountEntry)
	return shim.Success(newConsentCountEntry)
}

//updates consent count for a data call based on dataCallID and dataCallVersion
func (this *SmartContract) UpdateConsentCountForDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("UpdateConsentCountForDataCall: enter")
	defer logger.Debug("UpdateConsentCountForDataCall: exit")
	if len(args) < 1 {
		return shim.Error("UpdateConsentCountForDataCall: Incorrect number of arguments!")
	}

	var updateConsentReqest UpdateLikeAndConsentCountReq
	err := json.Unmarshal([]byte(args), &updateConsentReqest)

	if err != nil {
		logger.Error("UpdateConsentCountForDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("UpdateConsentCountForDataCall: Error during json.Unmarshal").Error())
	}
	if updateConsentReqest.DataCallID == "" {
		return shim.Error("DataCallID should not be Empty")

	} else if updateConsentReqest.DataCallVersion == "" {
		return shim.Error("DataCallVersion should not be Empty")

	}

	var pks []string = []string{DATA_CALL_PREFIX, updateConsentReqest.DataCallID, updateConsentReqest.DataCallVersion}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)
	dataCallAsBytes, err := stub.GetState(dataCallKey)
	if err != nil {
		return shim.Error("UpdateConsentCountForDataCall: Error retreiving data for key" + dataCallKey)
	}

	var prevDataCall DataCall
	err = json.Unmarshal(dataCallAsBytes, &prevDataCall)
	if err != nil {
		return shim.Error("UpdateConsentCountForDataCall: Failed to unmarshal data call: " + err.Error())
	}

	//Invoke CountLikes to get the total likes count for the data call
	var getConsentCount ConsentCountEntry
	getConsentCount.DatacallID = updateConsentReqest.DataCallID
	getConsentCount.DataCallVersion = updateConsentReqest.DataCallVersion
	getConsentCountJson, _ := json.Marshal(getConsentCount)
	countConsentsResponse := this.CountConsents(stub, string(getConsentCountJson))
	if countConsentsResponse.Status != 200 || len(countConsentsResponse.Payload) <= 0 {
		logger.Error("UpdateConsentCountForDataCall: Unable to CountConsents: ", countConsentsResponse)
		return shim.Error(errors.New("UpdateConsentCountForDataCall: Unable to CountConsents").Error())
	}

	var consentCountEntry ConsentCountEntry
	json.Unmarshal(countConsentsResponse.Payload, &consentCountEntry)
	logger.Debug("UpdateConsentCountForDataCall: InvokeResponse from CountConsents: delta ", (consentCountEntry.Delta))

	//update the dataCall with delta
	prevDataCall.ConsentCount = consentCountEntry.Delta
	prevDataAsBytes, _ := json.Marshal(prevDataCall)
	err = stub.PutState(dataCallKey, prevDataAsBytes)
	if err != nil {
		logger.Error("UpdateConsentCountForDataCall: Error updating DataCall for CountConsents")
		return shim.Error("UpdateConsentCountForDataCall: Error updating DataCall for CountConsents: " + err.Error())
	}

	return shim.Success(nil)

}

// Returns List of carriers Consented for a specific data call, based on dataCallID and dataCallVersion
// Request param- {"dataCallID":" ", "dataCallVersion":" "}
func (this *SmartContract) GetConsentsByDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {

	logger.Debug("GetConsentsByDataCall: enter")
	defer logger.Debug("GetConsentsByDataCall: exit")
	if len(args) < 1 {
		return shim.Error("GetConsentsByDataCall: Incorrect number of arguments!")
	}
	var consentList []ListConsentResponse
	var getConsentsByDataCallRequest GetConsentsByDataCallRequest
	err := json.Unmarshal([]byte(args), &getConsentsByDataCallRequest)

	if err != nil {
		logger.Error("GetConsentsByDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetConsentsByDataCall: Error during json.Unmarshal").Error())
	}

	if getConsentsByDataCallRequest.DataCallID == "" || getConsentsByDataCallRequest.DataCallVersion == "" {
		return shim.Error("DataCallID or DataCallVersion can't be empty")
	}

	// Get current channel
	logger.Info("GetConsentsByDataCall: currentChannelID > ", stub.GetChannelID())

	var pks []string
	pks = []string{CONSENT_PREFIX, getConsentsByDataCallRequest.DataCallID, getConsentsByDataCallRequest.DataCallVersion}

	logger.Info("GetConsentsByDataCall: Key ", pks)
	resultsIterator, err := stub.GetStateByPartialCompositeKey(CONSENT_DOCUMENT_TYPE, pks)
	defer resultsIterator.Close()

	if err != nil {
		// proceed without any further action to check consents on other channels
		logger.Error("GetConsentsByDataCall: Error fetching consent on this channel: ", stub.GetChannelID()+"error ", err)
	}
	// Check the variable existed
	if !resultsIterator.HasNext() {
		// proceed without any further action to check consents on other channels
		logger.Info("GetConsentsByDataCall: No Consent found on current channel, proceed to next channel ")

	} else {
		var i int
		logger.Debug("GetConsentsByDataCall: Iterating over list of Consents")
		for i = 0; resultsIterator.HasNext(); i++ {
			currentConsentAsBytes, nextErr := resultsIterator.Next()
			if nextErr != nil {
				return shim.Error("GetConsentsByDataCall: Failed to iterate over Consent")
			}
			var currentConsent Consent
			err = json.Unmarshal([]byte(currentConsentAsBytes.GetValue()), &currentConsent)
			if err != nil {
				return shim.Error("GetConsentsByDataCall: Failed to unmarshal consent: " + err.Error())
			}
			var listConsentResponse ListConsentResponse
			listConsentResponse.Consent = currentConsent
			//listConsentResponse.CarrierName = ""
			consentList = append(consentList, listConsentResponse)

		}
		logger.Info("GetConsentsByDataCall: consents fetched for current channel")
	}
	consentListAsByte, _ := json.Marshal(consentList)
	logger.Debug("GetConsentsByDataCall: consentsListAsByte", consentListAsByte)
	return shim.Success(consentListAsByte)
}

// Returns List of carriers Consented for a specific data call, based on dataCallID and dataCallVersion on requested channels
// Request param- {"dataCallID":" ", "dataCallVersion":" ", "channelList":[{"channelName": "channel1","chaincodeName": "openidl-cc-channel1"}]}
func (this *SmartContract) ListConsentsByDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {

	logger.Debug("ListConsentsByDataCall: enter")
	defer logger.Debug("ListConsentsByDataCall: exit")
	if len(args) < 1 {
		return shim.Error("ListConsentsByDataCall: Incorrect number of arguments!")
	}

	//listDataCallRequestJson := args[0]
	//logger.Info("ListLikesByDataCall: Request > " + listDataCallRequestJson)
	var consentList []ListConsentResponse
	var listConsentRequest ListConsentRequest
	err := json.Unmarshal([]byte(args), &listConsentRequest)

	if err != nil {
		logger.Error("ListConsentsByDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("ListConsentsByDataCall: Error during json.Unmarshal").Error())
	}
	logger.Debug("ListConsentsByDataCall: Unmarshalled object ", listConsentRequest)

	if listConsentRequest.DataCallID == "" || listConsentRequest.DataCallVersion == "" {
		return shim.Error("DataCallID or DataCallVersion can't be empty")
	}

	var pks []string
	pks = []string{CONSENT_PREFIX, listConsentRequest.DataCallID, listConsentRequest.DataCallVersion}

	logger.Info("ListConsentsByDataCall: Key ", pks)
	resultsIterator, err := stub.GetStateByPartialCompositeKey(CONSENT_DOCUMENT_TYPE, pks)
	defer resultsIterator.Close()

	if err != nil {
		// proceed without any further action to check consents on other channels
		logger.Error("ListConsentsByDataCall: Error fetching consent on this channel: ", stub.GetChannelID()+"error ", err)
	}
	// Check the variable existed
	if !resultsIterator.HasNext() {
		// proceed without any further action to check consents on other channels
		logger.Info("ListConsentsByDataCall: No Consent found on current channel, proceed to next channel ")
	} else {
		var i int
		logger.Debug("ListConsentsByDataCall: Iterating over list of Consents")
		for i = 0; resultsIterator.HasNext(); i++ {
			currentConsentAsBytes, nextErr := resultsIterator.Next()
			if nextErr != nil {
				return shim.Error("ListConsentsByDataCall: Failed to iterate over Consent")
			}
			var currentConsent Consent
			err = json.Unmarshal([]byte(currentConsentAsBytes.GetValue()), &currentConsent)
			if err != nil {
				return shim.Error("Failed to unmarshal data call: " + err.Error())
			}
			var listConsentResponse ListConsentResponse
			listConsentResponse.Consent = currentConsent
			//listConsentResponse.CarrierName = ""
			consentList = append(consentList, listConsentResponse)

		}
		logger.Info("ListConsentsByDataCall: Consents fetched for current channel, moving on to other channels")
	}

	// Get data from other channels mentioned in the client request and recieved any other channels mentioned in the client request
	//var channels Channels
	//channels.ChannelIDs = make([]string, len(crossInvocationChannels.ChannelIDs))
	//copy(channels.ChannelIDs[:], crossInvocationChannels.ChannelIDs)
	//logger.Debug("ListConsentsByDataCall: Requested additional channels data from client >> ", channels.ChannelIDs)
	totalChannels := len(listConsentRequest.ChannelList)
	logger.Debug("ListConsentsByDataCall: Requested additional channels data from client >> ", listConsentRequest.ChannelList)
	logger.Debug("ListConsentsByDataCall: Total Number of channels > ", totalChannels)

	// Get current channel
	//currentChannelID := stub.GetChannelID()
	//logger.Info("InvokeChaincodeOnChannel: currentChannelID > ", currentChannelID)

	var channenlIndex int
	for channenlIndex = 0; channenlIndex < totalChannels; channenlIndex++ {
		var getConsentsReq GetConsentsByDataCallRequest
		getConsentsReq.DataCallID = listConsentRequest.DataCallID
		getConsentsReq.DataCallVersion = listConsentRequest.DataCallVersion
		getConsentsReqAsBytes, _ := json.Marshal(getConsentsReq)
		getConsentsReqJson := string(getConsentsReqAsBytes)
		var GetConsentsByDataCallFunc = "GetConsentsByDataCall"
		getConsentsByDataCallRequest := ToChaincodeArgs(GetConsentsByDataCallFunc, getConsentsReqJson)
		logger.Debug("ListLikesByDataCall: getConsentsByDataCallRequest", getConsentsByDataCallRequest)
		logger.Info("ListLikesByDataCall: GetConsentsByDataCall request json " + getConsentsReqJson)
		//var invokeResponse pb.Response

		// fetch only if the requested channel is different from current channel
		//if channels.ChannelIDs[channenlIndex] != currentChannelID {
		logger.Debug("ListConsentsByDataCall: Fetching consent from channel ", listConsentRequest.ChannelList[channenlIndex].ChannelName)

		// Modify channel in the request before sending with current channelID to prevent any loops
		//listConsentRequest.ChannelIDs = make([]string, 1)
		//listConsentRequest.ChannelIDs[0] = currentChannelID
		chaincodeName := listConsentRequest.ChannelList[channenlIndex].ChaincodeName
		channelName := listConsentRequest.ChannelList[channenlIndex].ChannelName
		invokeResponse := stub.InvokeChaincode(chaincodeName, getConsentsByDataCallRequest, channelName)
		//invokeResponse := InvokeChaincode(stub, chaincodeName, "GetConsentsByDataCall", getConsentsReqJson, channelName)
		if invokeResponse.Status != 200 {
			logger.Error("ListConsentsByDataCall: Unable to Invoke cross channel query GetConsentsByDataCall: ", (invokeResponse))
			// Do not block functionality and proceed to return the original channels like list
			//}
		} else if len(invokeResponse.Payload) <= 0 {
			logger.Debug("ListConsentsByDataCall: ErrorInvokeResponse from another channel ", string(invokeResponse.Payload))

		} else {
			var invokeListConsentResponse []ListConsentResponse
			logger.Debug("ListConsentsByDataCall: InvokeResponse from another channel ", string(invokeResponse.Payload))
			json.Unmarshal(invokeResponse.Payload, &invokeListConsentResponse)
			consentList = append(consentList, invokeListConsentResponse...)

		}
		//}
	}

	consentListAsByte, _ := json.Marshal(consentList)
	logger.Debug("ListConsentsByDataCall: consentsListAsByte", consentListAsByte)
	return shim.Success(consentListAsByte)

}

func (this *SmartContract) GetConsentByDataCallAndOrganization(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	logger.Debug("GetConsentByDataCallAndOrganization: enter")
	defer logger.Debug("GetConsentByDataCallAndOrganization: exit")
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!")
	}
	listConsentRequestJson := args[0]
	logger.Debug("GetConsentByDataCallAndOrganization: Request > " + listConsentRequestJson)
	var listConsentRequest GetConsentByDataCallAndOrganizationRequest
	var consentList []ListConsentResponse

	err := json.Unmarshal([]byte(listConsentRequestJson), &listConsentRequest)
	consent := listConsentRequest.Consent
	if err != nil {
		logger.Error("GetConsentByDataCallAndOrganization: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetConsentByDataCallAndOrganization: Error during json.Unmarshal").Error())
	}

	// Create partial composite key to fetch liks based on DataCall Id and DataCall Version
	var pks []string
	pks = []string{CONSENT_PREFIX, consent.DatacallID, consent.DataCallVersion, consent.CarrierID}
	logger.Info("GetConsentByDataCallAndOrganization: Key ", pks)
	resultsIterator, err := stub.GetStateByPartialCompositeKey(CONSENT_DOCUMENT_TYPE, pks)
	defer resultsIterator.Close()

	if err != nil {
		logger.Error("GetConsentByDataCallAndOrganization: No Consent found on current channel due to error, proceed to next channel ", err)
		// proceed without any further action to check consents on other channels
	}
	// Check the variable existed
	if !resultsIterator.HasNext() {
		logger.Info("GetConsentByDataCallAndOrganization: No Consent found on current channel, proceed to next channel ")
		// proceed without any further action to check consents on other channels
	} else {
		var i int
		logger.Debug("GetConsentByDataCallAndOrganization: Iterating over list of Consents")
		for i = 0; resultsIterator.HasNext(); i++ {
			currentConsentAsBytes, nextErr := resultsIterator.Next()
			if nextErr != nil {
				return shim.Error("GetConsentByDataCallAndOrganization: Failed to iterate Consent Entry call")
			}
			var currentConsent Consent
			err = json.Unmarshal([]byte(currentConsentAsBytes.GetValue()), &currentConsent)
			if err != nil {
				return shim.Error("Failed to unmarshal data call: " + err.Error())
			}
			var listConsentResponse ListConsentResponse
			listConsentResponse.Consent = currentConsent
			listConsentResponse.CarrierName = ""
			consentList = append(consentList, listConsentResponse)

		}
		logger.Info("GetConsentByDataCallAndOrganization: Consent Fetched  Returning Response")
	}

	consentListAsByte, _ := json.Marshal(consentList)
	logger.Debug("GetConsentByDataCallAndOrganization: consentsListAsByte", consentListAsByte)
	return shim.Success(consentListAsByte)

}

func GetConsentsCount(stub shim.ChaincodeStubInterface, args []string) map[string]int {
	fmt.Println("Inside GetConsentsCount and args are ", args)
	var consentCounts map[string]int
	consentCounts = make(map[string]int)

	//fetch all the repports based on Id and Version sorted by updatedTs
	var queryStr string
	queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"datacallID\":{\"$in\":[%s]}},\"use_index\":[\"_design/dataCallId\"]}", CONSENT_PREFIX, strings.Trim(fmt.Sprint(args), "[]"))
	deltaResultsIterator, _ := stub.GetQueryResult(queryStr)

	defer deltaResultsIterator.Close()

	// Check the variable existed
	if !deltaResultsIterator.HasNext() {
		logger.Info("CountConsents: No Consents found for criteria, returning 0 delta")
		return consentCounts
	}

	for deltaResultsIterator.HasNext() {
		consentCountAsBytes, _ := deltaResultsIterator.Next()

		var tempConsentCount ConsentCountEntry
		_ = json.Unmarshal([]byte(consentCountAsBytes.GetValue()), &tempConsentCount)

		consentCounts[tempConsentCount.DatacallID] = consentCounts[tempConsentCount.DatacallID] + tempConsentCount.Delta

	}
	return consentCounts
}

// logDataCallTransaction creates a new Transactional Log Entry for a datacall
// Success: nil
// Error: {"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (this *SmartContract) LogDataCallTransaction(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("LogDataCallTransaction: enter")
	defer logger.Debug("LogDataCallTransaction: exit")
	logger.Debug("LogDataCallTransaction json received : ", args)
	if len(args) < 1 {
		return shim.Error("LogDataCallTransaction: Incorrect number of arguments!!")
	}

	var dataCallLog DataCallLog
	dataCallLogAsBytes := []byte(args)
	err := json.Unmarshal(dataCallLogAsBytes, &dataCallLog)
	if dataCallLog.DataCallID == "" || dataCallLog.DataCallVersion == "" {
		return shim.Error("LogDataCallTransaction: DataCallID and DataCallVersion cant not be empty!!")
	}
	if err != nil {
		logger.Error("LogDataCallTransaction: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("LogDataCallTransaction: Error during json.Unmarshal").Error())
	}

	// Create a new entry in log each time
	var pks []string = []string{DATACALL_LOG_PREFIX, dataCallLog.DataCallID, dataCallLog.DataCallVersion, stub.GetTxID()}
	dataCallLogKey, _ := stub.CreateCompositeKey(DATACALL_LOG_DOCUMENT, pks)

	stub.PutState(dataCallLogKey, dataCallLogAsBytes)
	return shim.Success(nil)

}

// GetDataCallTransactionHistory retrives all data calls that match given criteria. If startindex and pageSize are not provided,
// this method returns the complete list of data calls. If version = latest, the it returns only latest version of a data call
// using the specified criteria. If version = all, it returns all data calls with their versions as individual items in list.
// params {json}: {
//  "startIndex":"optional",
//  "pageSize":"optional",
//  "version": "latest or all"
//  "status" :"DRAFT OR ISSUED OR CANCELLED"}
// Success {byte[]}: byte[]
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (this *SmartContract) GetDataCallTransactionHistory(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetDataCallTransactionHistory: enter")
	defer logger.Debug("GetDataCallTransactionHistory: exit")
	logger.Debug("GetDataCallTransactionHistory json received : ", args)
	if len(args) < 1 {
		return shim.Error("GetDataCallTransactionHistory: Incorrect number of arguments!!")
	}
	var getTxHistoryReq DataCallLog
	err := json.Unmarshal([]byte(args), &getTxHistoryReq)
	if err != nil {
		logger.Error("GetDataCallTransactionHistory: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetDataCallTransactionHistory: Error during json.Unmarshal").Error())
	}
	logger.Debug("GetDataCallTransactionHistory: Unmarshalled object ", getTxHistoryReq)
	var pks []string = []string{DATACALL_LOG_PREFIX, getTxHistoryReq.DataCallID, getTxHistoryReq.DataCallVersion}
	resultsIterator, errMsg := stub.GetStateByPartialCompositeKey(DATACALL_LOG_DOCUMENT, pks)
	defer resultsIterator.Close()
	if errMsg != nil {
		logger.Warning("GetDataCallTransactionHistory: Failed to get state for all the data calls")
	}

	var getTxHistoryRes []DataCallLog
	if !resultsIterator.HasNext() {
		dataCallLogsAsByte, _ := json.Marshal(getTxHistoryRes)
		logger.Debug("GetDataCallTransactionHistory: dataCallsAsByte", getTxHistoryRes)
		return shim.Success(dataCallLogsAsByte)
	}

	for resultsIterator.HasNext() {
		dataCallLogsAsByte, _ := resultsIterator.Next()
		var currenDatacallLog DataCallLog
		json.Unmarshal([]byte(dataCallLogsAsByte.GetValue()), &currenDatacallLog)
		getTxHistoryRes = append(getTxHistoryRes, currenDatacallLog)
	}
	getTxHistoryResAsBytes, _ := json.Marshal(getTxHistoryRes)

	return shim.Success(getTxHistoryResAsBytes)

}

func (this *SmartContract) ListDataCallTransactionHistory(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetDataCallTransactionHistory: enter")
	defer logger.Debug("GetDataCallTransactionHistory: exit")
	logger.Debug("GetDataCallTransactionHistory json received : ", args)
	if len(args) < 1 {
		return shim.Error("GetDataCallTransactionHistory: Incorrect number of arguments!!")
	}
	var getTxHistoryReq DataCallLog
	err := json.Unmarshal([]byte(args), &getTxHistoryReq)
	if err != nil {
		logger.Error("GetDataCallTransactionHistory: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetDataCallTransactionHistory: Error during json.Unmarshal").Error())
	}
	logger.Debug("GetDataCallTransactionHistory: Unmarshalled object ", getTxHistoryReq)
	//var pks []string = []string{DATACALL_LOG_PREFIX, getTxHistoryReq.DataCallID, getTxHistoryReq.DataCallVersion}
	queryStr := fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"dataCallID\":\"%s\",\"dataCallVersion\":\"%s\"},\"use_index\":[\"_design/actionTs\"],\"sort\":[{\"actionTs\":\"desc\"}]}", DATACALL_LOG_PREFIX, getTxHistoryReq.DataCallID, getTxHistoryReq.DataCallVersion)
	resultsIterator, errMsg := stub.GetQueryResult(queryStr)
	//resultsIterator, errMsg := stub.GetStateByPartialCompositeKey(DATACALL_LOG_DOCUMENT, pks)
	defer resultsIterator.Close()
	if errMsg != nil {
		logger.Warning("GetDataCallTransactionHistory: Failed to get state for all the data calls")
	}

	var getTxHistoryRes []DataCallLog
	if !resultsIterator.HasNext() {
		dataCallLogsAsByte, _ := json.Marshal(getTxHistoryRes)
		logger.Debug("GetDataCallTransactionHistory: dataCallsAsByte", getTxHistoryRes)
		return shim.Success(dataCallLogsAsByte)
	}

	for resultsIterator.HasNext() {
		dataCallLogsAsByte, _ := resultsIterator.Next()
		var currenDatacallLog DataCallLog
		json.Unmarshal([]byte(dataCallLogsAsByte.GetValue()), &currenDatacallLog)
		getTxHistoryRes = append(getTxHistoryRes, currenDatacallLog)
	}
	getTxHistoryResAsBytes, _ := json.Marshal(getTxHistoryRes)

	return shim.Success(getTxHistoryResAsBytes)

}

// creates extraction patten definition
func (this *SmartContract) CreateExtractionPattern(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CreateExtractionPattern: enter")
	defer logger.Debug("CreateExtractionPattern: exit")
	logger.Debug("CreateExtractionPattern json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}

	var extractionPatten ExtPattern
	err := json.Unmarshal([]byte(args), &extractionPatten)
	if extractionPatten.ExtractionPatternID == "" {
		return shim.Error("ExtractionPatternID cant not be empty!!")
	} else if extractionPatten.DbType == "" {
		return shim.Error("DbType cant not be empty!!")
	} else if extractionPatten.ViewDefinition.Map == "" || extractionPatten.ViewDefinition.Reduce == "" {
		return shim.Error("ViewDefinition cant be empty!!")
	} else if extractionPatten.PremiumFromDate == "" {
		return shim.Error("PremiumFromDate cannot not be Empty")
	} else if extractionPatten.LossFromDate == "" {
		return shim.Error("LossFromDate cannot not be Empty")
	} else if extractionPatten.Jurisdiction == "" {
		return shim.Error("Jurisdiction cannot not be Empty")
	} else if extractionPatten.Insurance == "" {
		return shim.Error("Insurance cannot not be Empty")
	}
	if err != nil {
		logger.Error("CreateExtractionPattern: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CreateExtractionPattern: Error during json.Unmarshal").Error())
	}

	logger.Debug("Unmarshalled object ", extractionPatten)

	extractionPatten.Version = generateVersion(0)
	namespace := EXTRACTION_PATTERN_PREFIX
	extPatternKey, _ := stub.CreateCompositeKey(namespace, []string{extractionPatten.ExtractionPatternID, extractionPatten.DbType})

	// Checking the ledger to confirm that the ExtractionPattern doesn't exist
	prevExtractionPattern, _ := stub.GetState(extPatternKey)

	if prevExtractionPattern != nil {
		logger.Error("CreateExtractionPattern: Extarction Pattern already exist with ID: " + extPatternKey)
		return shim.Error("CreateExtractionPattern:Extarction Pattern already exist with ID: " + extPatternKey)
	}

	extractionPatternAsBytes, _ := json.Marshal(extractionPatten)
	err = stub.PutState(extPatternKey, extractionPatternAsBytes)
	if err != nil {
		return shim.Error("CreateExtractionPattern: Failed to Put Extarction Pattern: " + err.Error())
	}

	return shim.Success(nil)

}

// updates an existing extraction pattern definition
func (this *SmartContract) UpdateExtractionPattern(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("UpdateExtractionPattern: enter")
	defer logger.Debug("UpdateExtractionPattern: exit")
	logger.Debug("UpdateExtractionPattern json received : ", args)

	var extractionPatten ExtPattern
	err := json.Unmarshal([]byte(args), &extractionPatten)
	if err != nil {
		logger.Error("UpdateExtractionPattern: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("UpdateExtractionPattern: Error during json.Unmarshal").Error())
	}
	if extractionPatten.ExtractionPatternID == "" {
		return shim.Error("ExtractionPatternID should not be Empty")

	} else if extractionPatten.DbType == "" {
		return shim.Error("DbType should not be Empty")

	} else if extractionPatten.PremiumFromDate == "" {
		return shim.Error("PremiumFromDate cannot not be Empty")
	} else if extractionPatten.LossFromDate == "" {
		return shim.Error("LossFromDate cannot not be Empty")
	} else if extractionPatten.Jurisdiction == "" {
		return shim.Error("Jurisdiction cannot not be Empty")
	} else if extractionPatten.Insurance == "" {
		return shim.Error("Insurance cannot not be Empty")
	}
	namespace := EXTRACTION_PATTERN_PREFIX
	extPatternKey, _ := stub.CreateCompositeKey(namespace, []string{extractionPatten.ExtractionPatternID, extractionPatten.DbType})
	extractionPatternAsBytes, err := stub.GetState(extPatternKey)
	if err != nil {
		logger.Error("UpdateExtractionPattern:Error retreiving extraction pattern for key: " + extPatternKey)
		return shim.Error("UpdateExtractionPattern: Error retreiving extraction pattern for key" + extPatternKey)
	}

	var prevExtractionPattern ExtPattern
	err = json.Unmarshal(extractionPatternAsBytes, &prevExtractionPattern)
	if err != nil {
		return shim.Error("UpdateExtractionPattern: Failed to unmarshal pattern: " + err.Error())
	}
	prevExtractionPattern.PremiumFromDate = extractionPatten.PremiumFromDate
	prevExtractionPattern.LossFromDate = extractionPatten.LossFromDate
	prevExtractionPattern.Jurisdiction = extractionPatten.Jurisdiction
	prevExtractionPattern.Insurance = extractionPatten.Insurance
	prevExtractionPattern.ViewDefinition = extractionPatten.ViewDefinition
	prevExtractionPattern.UpdatedTs = extractionPatten.UpdatedTs
	prevExtractionPattern.UpdatedBy = extractionPatten.UpdatedBy
	prevExtractionPattern.IsActive = extractionPatten.IsActive
	prevExtractionPattern.ExtractionPatternName = extractionPatten.ExtractionPatternName
	prevExtractionPattern.EffectiveStartTs = extractionPatten.EffectiveStartTs
	prevExtractionPattern.EffectiveEndTs = extractionPatten.EffectiveEndTs
	prevExtractionPattern.Description = extractionPatten.Description
	prevExtractionPatternVersion, _ := strconv.Atoi(prevExtractionPattern.Version)
	prevExtractionPattern.Version = generateVersion(prevExtractionPatternVersion)

	prevExtractionPatternAsBytes, _ := json.Marshal(prevExtractionPattern)
	err = stub.PutState(extPatternKey, prevExtractionPatternAsBytes)
	if err != nil {
		return shim.Error("UpdateExtractionPattern: Failed to Update Extraction Pattern: " + err.Error())
	}

	return shim.Success(nil)

}

// function returns extraction pattern based on id and dbtype
func (this *SmartContract) GetExtractionPatternById(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetExtractionPatternById: enter")
	defer logger.Debug("GetExtractionPatternById: exit")

	var getExtractionPatternById GetExtractionPatternById
	err := json.Unmarshal([]byte(args), &getExtractionPatternById)
	if err != nil {
		logger.Error("GetExtractionPatternById: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetExtractionPatternById: Error during json.Unmarshal").Error())
	}
	logger.Debug("GetExtractionPatternById: Unmarshalled object ", getExtractionPatternById)

	if getExtractionPatternById.ExtractionPatternID == "" || getExtractionPatternById.DbType == "" {
		return shim.Error("GetExtractionPatternById: ExtractionPatternID and DbType can not be Empty")
	}

	namespace := EXTRACTION_PATTERN_PREFIX
	extPatternKey, _ := stub.CreateCompositeKey(namespace, []string{getExtractionPatternById.ExtractionPatternID, getExtractionPatternById.DbType})
	extractionPatternAsBytes, err := stub.GetState(extPatternKey)
	if err != nil {
		logger.Error("GetExtractionPatternById: Error retreiving data for key ", extPatternKey)
		return shim.Error("GetExtractionPatternById: Error retreiving data for key" + extPatternKey)
	}
	return shim.Success(extractionPatternAsBytes)

}

// returns required data call details and extraction pattern
func (this *SmartContract) GetDataCallAndExtractionPattern(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetDataCallAndExtractionPattern: enter")
	defer logger.Debug("GetDataCallAndExtractionPattern: exit")

	var getDataCallAndExtractionPattern GetDataCallAndExtractionPattern
	err := json.Unmarshal([]byte(args), &getDataCallAndExtractionPattern)
	if err != nil {
		logger.Error("GetDataCallAndExtractionPattern: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetDataCallAndExtractionPattern: Error during json.Unmarshal").Error())
	}
	logger.Debug("GetDataCallAndExtractionPattern: Unmarshalled object ", getDataCallAndExtractionPattern)

	if getDataCallAndExtractionPattern.DataCallID == "" {
		return shim.Error("GetDataCallAndExtractionPattern: DataCallID can not be Empty")
	} else if getDataCallAndExtractionPattern.DataCallVersion == "" {
		return shim.Error("GetDataCallAndExtractionPattern: DataCallVersion can not be Empty")
	} else if getDataCallAndExtractionPattern.DbType == "" {
		return shim.Error("GetDataCallAndExtractionPattern: DbType can not be Empty")
	}

	// invoke GetDataCallByIdAndVersion to get dataCall details
	var getDataCall GetDataCall
	getDataCall.ID = getDataCallAndExtractionPattern.DataCallID
	getDataCall.Version = getDataCallAndExtractionPattern.DataCallVersion
	getDataCallJson, _ := json.Marshal(getDataCall)
	getDataCallResponse := this.GetDataCallByIdAndVersion(stub, string(getDataCallJson))
	if getDataCallResponse.Status != 200 || len(getDataCallResponse.Payload) <= 0 {
		logger.Error("GetDataCallAndExtractionPattern: Unable to GetDataCallByIdAndVersion: ", getDataCallResponse.Message)
		return shim.Error(errors.New("GetDataCallAndExtractionPattern: Unable to GetDataCallByIdAndVersion").Error())
	}

	var dataCall DataCall
	json.Unmarshal(getDataCallResponse.Payload, &dataCall)
	logger.Debug("GetDataCallAndExtractionPattern: InvokeResponse from GetDataCallByIdAndVersion: Name ", string(dataCall.Name))

	// setting jurisdiction in dataCallAndExtractionPatternResponse
	var dataCallAndExtractionPatternResponse DataCallAndExtractionPatternResponse
	dataCallAndExtractionPatternResponse.Jurisdiction = dataCall.Jurisdiction

	//check whether extraction pattern is set or not
	if dataCall.ExtractionPatternID == "" {
		dataCallAndExtractionPatternResponse.IsSet = false
	} else {
		dataCallAndExtractionPatternResponse.IsSet = true

		//invoke GetExtractionPatternById to get extraction pattern details
		var getExtractionPatternById GetExtractionPatternById
		getExtractionPatternById.ExtractionPatternID = dataCall.ExtractionPatternID
		getExtractionPatternById.DbType = getDataCallAndExtractionPattern.DbType
		getExtractionPatternByIdJson, _ := json.Marshal(getExtractionPatternById)
		getExtractionPatternByIdResponse := this.GetExtractionPatternById(stub, string(getExtractionPatternByIdJson))
		if getExtractionPatternByIdResponse.Status != 200 || len(getExtractionPatternByIdResponse.Payload) <= 0 {
			logger.Error("GetDataCallAndExtractionPattern: Unable to GetExtractionPatternById: ", getExtractionPatternByIdResponse.Message)
			return shim.Error(errors.New("GetDataCallAndExtractionPattern: Unable to GetExtractionPatternById").Error())
		}
		var extractionPattern ExtPattern
		json.Unmarshal(getExtractionPatternByIdResponse.Payload, &extractionPattern)
		logger.Debug("GetDataCallAndExtractionPattern: InvokeResponse from GetExtractionPatternById: Id ", string(extractionPattern.ExtractionPatternID))
		dataCallAndExtractionPatternResponse.ExtractionPattern = extractionPattern
	}

	dataCallAndExtractionPatternResponseAsBytes, _ := json.Marshal(dataCallAndExtractionPatternResponse)
	return shim.Success(dataCallAndExtractionPatternResponseAsBytes)

}

// GetExtractionPatternsMap simply returns a dictionary/map that contains the pre-defined extraction-patterns
// for the chaincode.
func GetExtractionPatternsMap() map[string]ExtractionPattern {
	var patterns map[string]ExtractionPattern
	patterns = make(map[string]ExtractionPattern)

	//patterns["Pattern_01"] = ExtractionPattern{ID: "Pattern_01", Name: "Standard Annual Homeowners Report", Description: "Provides summations of written premiums and exposures within groupings of ZIP code, policy form code, and property amount of insurance. This extraction pattern is for the reporting year 2017 and returns aggregate data for policy forms 01, 02, 03, 05 and 08.", CouchDBView: CouchDBView{Definition: EXT_PATTERN_VIEW_01, Group: true}}
	//patterns["Pattern_02"] = ExtractionPattern{ID: "Pattern_02", Name: "ILCC Data Extraction Pattern", Description: "Provides summations of written premiums and exposures within groupings of ZIP code, policy form code, and property amount of insurance. This extraction pattern is for the reporting year 2017 and returns aggregate data for policy forms 01, 02, 03, 05 and 08.", CouchDBView: CouchDBView{Definition: EXT_PATTERN_VIEW_02, Group: true}}
	//patterns["Pattern_03"] = ExtractionPattern{ID: "Pattern_03", Name: "New ILCC Data Extraction Pattern", Description: "Provides summations of written premiums and exposures within groupings of ZIP code, policy form code, and property amount of insurance. This extraction pattern is for the reporting year 2017 and returns aggregate data for policy forms 01, 02, 03, 05 and 08.", CouchDBView: CouchDBView{Definition: EXT_PATTERN_VIEW_03, Group: true}}
	//patterns["Pattern_04"] = ExtractionPattern{ID: "Pattern_04", Name: "Industry Test Drive Extraction Pattern", Description: "Provides summarizations of written premiums, exposures and paid loss within the corresponding state, grouped by zip code, liability limit, and cause of loss. This extraction pattern is for the reporting years 2016 – 2018 and returns aggregate data to interrogate causes of loss over liability limits.", CouchDBView: CouchDBView{Definition: EXT_PATTERN_VIEW_04, Group: true}}
	return patterns
}

func (this *SmartContract) ListExtractionPatterns(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("ListExtractionPatterns: enter")
	defer logger.Debug("ListExtractionPatterns: exit")
	var patterns []ExtPattern
	queryStr := fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"isActive\":true}}", EXTRACTION_PATTERN_PREFIX)
	resultsIterator, err := stub.GetQueryResult(queryStr)

	if err != nil {
		logger.Error("ListExtractionPatterns: Failed to get extraction patterns")
		return shim.Error("ListExtractionPatterns: Failed to get extraction patterns : " + err.Error())
	}
	defer resultsIterator.Close()
	logger.Debug("ListExtractionPatterns: Iterating over extraction patterns")
	for resultsIterator.HasNext() {
		dataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate over extraction patterns")
			return shim.Error("Failed to iterate over extraction patterns")
		}

		var pattern ExtPattern
		err = json.Unmarshal([]byte(dataCallAsBytes.GetValue()), &pattern)
		logger.Debug("ListExtractionPatterns: DataCall > ", pattern.ExtractionPatternID)
		if err != nil {
			return shim.Error("ListExtractionPatterns: Failed to unmarshal extraction patterns: " + err.Error())
		}
		patterns = append(patterns, pattern)

	}

	patternsAsBytes, _ := json.Marshal(patterns)
	logger.Info("ListExtractionPatterns: ExtractionPatterns", patterns)

	return shim.Success(patternsAsBytes)

}

func (this *SmartContract) GetExtractionPatternByIds(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetExtractionPatternByIds: enter")
	defer logger.Debug("GetExtractionPatternByIds: exit")
	var patternIds ExtractionPatternId
	var patterns []ExtractionPattern

	err := json.Unmarshal([]byte(args), &patternIds)
	logger.Debug("GetExtractionPatternByIds: Incoming args", args)
	if err != nil {
		logger.Error("GetExtractionPatternByIds: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetExtractionPatternByIds: Error during json.Unmarshal").Error())
	}
	logger.Debug("GetExtractionPatternByIds: Unmarshalled object ", patternIds)

	patternsMap := GetExtractionPatternsMap()

	for _, value := range patternIds.Id {
		if patternValue, found := patternsMap[value]; found {
			logger.Info("Extraction Pattern value found for key", value)
			patterns = append(patterns, patternValue)
		}
	}
	patternsAsBytes, _ := json.Marshal(patterns)
	logger.Info("GetExtractionPatternByIds: ExtractionPatterns", patterns)

	return shim.Success(patternsAsBytes)

}

func (this *SmartContract) CheckExtractionPatternIsSet(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CheckExtractionPatternIsSet: enter")
	defer logger.Debug("CheckExtractionPatternIsSet: exit")

	var isSet bool
	var getDataCall GetDataCall
	err := json.Unmarshal([]byte(args), &getDataCall)
	logger.Debug("CheckExtractionPatternIsSet: Incoming args", args)
	if err != nil {
		logger.Error("CheckExtractionPatternIsSet: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CheckExtractionPatternIsSet: Error during json.Unmarshal").Error())
	}
	logger.Debug("CheckExtractionPatternIsSet: Unmarshalled object ", getDataCall)

	if getDataCall.ID == "" || getDataCall.Version == "" {
		return shim.Error("ID and Version can not be Empty")
	}

	getDataCallAsBytes, _ := json.Marshal(getDataCall)
	getDataCallReqJson := string(getDataCallAsBytes)

	var GetDataCallByIdAndVersionFunc = "GetDataCallByIdAndVersion"
	getDataCallRequest := ToChaincodeArgs(GetDataCallByIdAndVersionFunc, getDataCallReqJson)
	logger.Debug("CheckExtractionPatternIsSet: getDataCallRequest", getDataCallRequest)
	getDataCallResponse := stub.InvokeChaincode(DEFAULT_CHAINCODE_NAME, getDataCallRequest, DEFAULT_CHANNEL)
	logger.Debug("CheckExtractionPatternIsSet: getDataCallResponse > ", getDataCallResponse)
	logger.Debug("CheckExtractionPatternIsSet: getDataCallResponse.Status ", getDataCallResponse.Status)
	logger.Debug("CheckExtractionPatternIsSet: getDataCallResponse.Payload", string(getDataCallResponse.Payload))
	if getDataCallResponse.Status != 200 {
		logger.Error("CheckExtractionPatternIsSet: Unable to Fetch DataCall due to Error: ", err)
		return shim.Error(errors.New("CheckExtractionPatternIsSet: Unable to Fetch DataCall due to Error").Error())
	}

	if len(getDataCallResponse.Payload) <= 0 {
		logger.Error("CheckExtractionPatternIsSet: DataCall Doesnt exist")
		return shim.Error(errors.New("CheckExtractionPatternIsSet:  DataCall Doesnt exist").Error())

	}

	/*var pks []string = []string{DATA_CALL_PREFIX, getDataCall.ID, getDataCall.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)
	dataCallAsBytes, err := stub.GetState(dataCallKey)
	logger.Debug("CheckExtractionPatternIsSet: key ", dataCallKey, "value", dataCallAsBytes)
	if err != nil {
		logger.Error("CheckExtractionPatternIsSet: Error retreiving data for key ", dataCallKey)
		return shim.Error("CheckExtractionPatternIsSet: Error retreiving data for key" + dataCallKey)
	} else if len(dataCallAsBytes) == 0 {
		logger.Error("CheckExtractionPatternIsSet: DataCall Doesnt exist")
		return shim.Error("CheckExtractionPatternIsSet: DataCall Doesnt exist")
	}*/
	var dataCall DataCall
	errMsg := json.Unmarshal(getDataCallResponse.Payload, &dataCall)
	if errMsg != nil {
		logger.Error("CheckExtractionPatternIsSet: Error during json.Unmarshal for response: ", errMsg)
		return shim.Error(errors.New("CheckExtractionPatternIsSet: Error during json.Unmarshal for response: ").Error())
	}
	if dataCall.ExtractionPatternID != "" {
		isSet = true
	}

	//get extraction pattern by ExtractionPatternID
	var extractionPattern ExtractionPattern
	if isSet {
		patternsMap := GetExtractionPatternsMap()
		extractionPattern = patternsMap[dataCall.ExtractionPatternID]
	}
	var extractionPatternIsSetPayload ExtractionPatternIsSetPayload
	extractionPatternIsSetPayload.IsSet = isSet
	extractionPatternIsSetPayload.ExtractionPattern = extractionPattern
	responseAsBytes, _ := json.Marshal(extractionPatternIsSetPayload)
	logger.Debug("CheckExtractionPatternIsSet: responseAsBytes ", responseAsBytes)
	return shim.Success(responseAsBytes)

}

func (this *SmartContract) SaveInsuranceDataHash(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Info("SaveInsuranceDataHash: enter")
	defer logger.Debug("SaveInsuranceDataHash: exit")
	var insurance InsuranceDataHash

	err := json.Unmarshal([]byte(args), &insurance)
	if err != nil {
		logger.Error("SaveInsuranceDataHash: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("SaveInsuranceDataHash: Error during json.Unmarshal").Error())
	}
	if insurance.BatchId == "" {
		return shim.Error("BatchId should not be Empty")
	} else if insurance.Hash == "" {
		return shim.Error("Hash should not be Empty")
	} else if insurance.CarrierId == "" {
		return shim.Error("CarrierId should not be Empty")
	} else if insurance.ChunkId == "" {
		return shim.Error("ChunkId should not be Empty")
	}

	namespacePrefix := INSURANCE_HASH_PREFIX
	//var pks []string = []string{INSURANCE_HASH_PREFIX, insurance.CarrierId, insurance.BatchId}
	key, _ := stub.CreateCompositeKey(namespacePrefix, []string{insurance.CarrierId, insurance.BatchId, insurance.ChunkId})

	//key := insurance.BatchId
	insuranceDataAsBytes, _ := json.Marshal(insurance)
	err = stub.PutState(key, insuranceDataAsBytes)

	if err != nil {
		return shim.Error("SaveInsuranceDataHash: Error committing data for key: " + key)
	}

	return shim.Success(nil)
}

// function-name: GetHashById (invoke)
// params {json}: {
// "id":"mandatory"}
// Success {byte[]}: byte[]  - Report
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : returns a InsuranceHashRecord of specifc batchid

func (this *SmartContract) GetHashById(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetHashById: enter")
	defer logger.Debug("GetHashById: exit")

	var insurance InsuranceDataHash
	err := json.Unmarshal([]byte(args), &insurance)
	if err != nil {
		logger.Error("GetHashById: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetHashById: Error during json.Unmarshal").Error())
	}
	namespacePrefix := INSURANCE_HASH_PREFIX
	key, _ := stub.CreateCompositeKey(namespacePrefix, []string{insurance.CarrierId, insurance.BatchId, insurance.ChunkId})
	if key == "" {
		return shim.Error("GetHashById:BatchId should not be empty")
	}

	insuranceHashAsBytes, err := stub.GetState(key)
	if err != nil {
		logger.Error("Error retreiving data for key ", key)
		return shim.Error("Error retreiving data for key" + key)
	}
	return shim.Success(insuranceHashAsBytes)

}

//this function puts the Insurance data extracted based on extraction pattern into private data collection
func (this *SmartContract) SaveInsuranceData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("SaveInsuranceData: enter")
	defer logger.Info("SaveInsuranceData: exit")
	var insurance InsuranceData
	transientMapKey := INSURANCE_TRANSACTIONAL_RECORD_PREFIX

	//getting trnasientMap data (using INSURANCE_TRANSACTIONAL_RECORD_PREFIX as key)
	InsuranceDataTransMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("SaveInsuranceData: Error getting InsuranceDataTransMap: " + err.Error())
	}
	if _, ok := InsuranceDataTransMap[transientMapKey]; !ok {
		return shim.Error("SaveInsuranceData: Invalid key in the transient map")
	}
	err = json.Unmarshal([]byte(InsuranceDataTransMap[transientMapKey]), &insurance)
	logger.Info("SaveInsuranceData: got transient map")

	if err != nil {
		logger.Error("SaveInsuranceData: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("SaveInsuranceData: Error during json.Unmarshal").Error())
	}
	pageNumber := insurance.PageNumber
	pageNumberAsString := strconv.Itoa(pageNumber)

	if insurance.CarrierId == "" {
		return shim.Error("CarrierId should not be Empty")
	} else if insurance.DataCallId == "" {
		return shim.Error("DataCallId should not be Empty")
	} else if insurance.DataCallVersion == "" {
		return shim.Error("DataCallVersion should not be Empty")
	} else if pageNumber == 0 {
		return shim.Error("PageNumber should not be Empty")
	}

	logger.Info("SaveInsuranceData: all necessary params found")
	//Identify the pdc name based on channelName
	channelName := stub.GetChannelID()
	private_data_collection := getPDCNameByChannelName(channelName)

	namespacePrefix := INSURANCE_TRANSACTIONAL_RECORD_PREFIX
	key, _ := stub.CreateCompositeKey(namespacePrefix, []string{insurance.DataCallId, insurance.DataCallVersion, insurance.CarrierId, pageNumberAsString})

	insuranceDataAsBytes, _ := json.Marshal(insurance)
	err = stub.PutPrivateData(private_data_collection, key, insuranceDataAsBytes)

	logger.Info("SaveInsuranceData: put private data done")
	if err != nil {
		logger.Error("Error commiting pdc data:", err)
		return shim.Error("SaveInsuranceData: Error committing data for key: " + key)
	}

	//insurance data has been ingested, now creating audit record
	var auditRecord InsuranceRecordAudit
	auditRecord.DataCallId = insurance.DataCallId
	auditRecord.DataCallVersion = insurance.DataCallVersion
	auditRecord.CarrierId = insurance.CarrierId

	namespacePrefixForAudit := AUDIT_INSURANCE_TRANSACTIONAL_RECORD_PREFIX
	auditRecordKey, _ := stub.CreateCompositeKey(namespacePrefixForAudit, []string{auditRecord.DataCallId, auditRecord.DataCallVersion, auditRecord.CarrierId})

	auditRecordAsBytes, _ := json.Marshal(auditRecord)
	err = stub.PutState(auditRecordKey, auditRecordAsBytes)
	logger.Info("SaveInsuranceData: put audit key done")

	if err != nil {
		return shim.Error("SaveInsuranceData: Creating Audit Record: Error committing data for key: " + auditRecordKey)
	}

	//audit record has been created now firing chaincode event -->TransactionalDataAvailable
	var eventPayload InsuranceRecordEventPayload
	eventPayload.ChannelName = channelName
	eventPayload.DataCallId = insurance.DataCallId
	eventPayload.DataCallVersion = insurance.DataCallVersion
	eventPayload.CarrierId = insurance.CarrierId
	eventPayload.PageNumber = insurance.PageNumber

	eventPayloadAsBytes, _ := json.Marshal(eventPayload)
	err = stub.SetEvent(INSURANCE_RECORD_AND_AUDIT_CREATED_EVENT, eventPayloadAsBytes)
	if err != nil {
		return shim.Error("SaveInsuranceData: error during creating event")
	}

	logger.Info("SaveInsuranceData: set event done")
	return shim.Success(nil)
}

// this function returns true if InsuranceData exists in pdc for a data call else returns false
func (this *SmartContract) CheckInsuranceDataExists(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CheckInsuranceDataExists: enter")
	defer logger.Debug("CheckInsuranceDataExists: exit")

	var isExist bool
	var getInsuranceData InsuranceRecordAudit
	err := json.Unmarshal([]byte(args), &getInsuranceData)
	if err != nil {
		logger.Error("CheckInsuranceDataExists: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CheckInsuranceDataExists: Error during json.Unmarshal").Error())
	}
	logger.Debug("CheckInsuranceDataExists: Unmarshalled object ", getInsuranceData)

	if getInsuranceData.CarrierId == "" {
		return shim.Error("CarrierId can not be Empty")
	} else if getInsuranceData.DataCallId == "" {
		return shim.Error("DataCallId can not be Empty")
	} else if getInsuranceData.DataCallVersion == "" {
		return shim.Error("DataCallVersion can not be Empty")
	}
	namespace := AUDIT_INSURANCE_TRANSACTIONAL_RECORD_PREFIX
	key, _ := stub.CreateCompositeKey(namespace, []string{getInsuranceData.DataCallId, getInsuranceData.DataCallVersion, getInsuranceData.CarrierId})
	insuranceDataAsBytes, err := stub.GetState(key)
	if err != nil {
		logger.Error("CheckInsuranceDataExists: Error retreiving Insurance data for key ", key)
		return shim.Error("CheckInsuranceDataExists: Error retreiving Insurance data for key" + key)
	}
	if len(insuranceDataAsBytes) > 0 {
		isExist = true
	}
	return shim.Success([]byte(strconv.FormatBool(isExist)))
}

// this function returns Insurance Data fetching from pdc
func (this *SmartContract) GetInsuranceData(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetInsuranceData: enter")
	defer logger.Debug("GetInsuranceData: exit")

	var getInsuranceData GetInsuranceData
	//var insuranceData []InsuranceData
	logger.Debug("args", args)
	err := json.Unmarshal([]byte(args), &getInsuranceData)
	if err != nil {
		logger.Error("GetInsuranceData: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetInsuranceData: Error during json.Unmarshal").Error())
	}
	logger.Debug("GetInsuranceData: Unmarshalled object ", getInsuranceData)

	if getInsuranceData.CarrierId == "" {
		return shim.Error("CarrierId can not be Empty")
	} else if getInsuranceData.DataCallId == "" {
		return shim.Error("DataCallId can not be Empty")
	} else if getInsuranceData.DataCallVersion == "" {
		return shim.Error("DataCallVersion can not be Empty")
	} else if getInsuranceData.ChannelName == "" {
		return shim.Error("ChannelName can not be Empty")
	}
	//startIndex := getInsuranceData.StartIndex
	//pageSize := getInsuranceData.PageSize
	pageNumber := getInsuranceData.PageNumber
	pageNumberAsString := strconv.Itoa(pageNumber)
	namespacePrefix := INSURANCE_TRANSACTIONAL_RECORD_PREFIX
	key, _ := stub.CreateCompositeKey(namespacePrefix, []string{getInsuranceData.DataCallId, getInsuranceData.DataCallVersion, getInsuranceData.CarrierId, pageNumberAsString})
	//Identify the pdc name based on channelID
	channelName := stub.GetChannelID()
	private_data_collection := getPDCNameByChannelName(channelName)

	insuranceDataResponseAsBytes, err := stub.GetPrivateData(private_data_collection, key)
	if err != nil {
		logger.Error("GetInsuranceData: Failed to get Insurance Data due to error", err)
		return shim.Error("GetInsuranceData: Failed to get Insurance Data")
	}
	return shim.Success(insuranceDataResponseAsBytes)
}

// getPDCNameByChannelName is a helper function for getting PDC name
func getPDCNameByChannelName(channelName string) string {
	pdcName := strings.Replace(channelName, "-", "_", -1) + "_pdc"
	return pdcName
}

// CreateReport creates a new report for a particular DataCall Id and version.
// Success: byte[]
// Error: {"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (this *SmartContract) CreateReport(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CreateReport: enter")
	defer logger.Debug("CreateReport: exit")
	var report Report
	var reportExist bool
	var isLocked bool
	var totalReports []Report
	var noOfReports int
	err := json.Unmarshal([]byte(args), &report)
	if err != nil {
		logger.Error("CreateReport: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CreateReport: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshalled object for CreateReport", report)
	if report.DataCallID == "" {
		logger.Error("DataCallID is empty!!")
		return shim.Error("DataCallID is empty!!")
	}
	if report.DataCallVersion == "" {
		logger.Error("DataCallVersion is empty!!")
		return shim.Error("DataCallVersion is empty!!")
	}
	if report.Hash == "" {
		logger.Error("Hash is empty!!")
		return shim.Error("Hash is empty!!")
	}
	report.Status = STATUS_CANDIDATE

	// check whether report exist for the particular DataCallId AND Version
	//getting all the reports for the particular DataCallId and Version
	var pks []string = []string{REPORT_PREFIX, report.DataCallID, report.DataCallVersion}

	//Step-1: getting the previous data for the data call
	resultsIterator, errmsg := stub.GetStateByPartialCompositeKey(REPORT_DOCUMENT_TYPE, pks)
	if errmsg != nil {
		logger.Error("CreateReport: Failed to get state for previous reports")
		return shim.Error("CreateReport: Failed to get state for previous reports")
	}
	defer resultsIterator.Close()
	logger.Debug("CreateReport: ", resultsIterator)

	if !resultsIterator.HasNext() {

	} else {

		//set reportExist to true as report already exists
		reportExist = true
		for resultsIterator.HasNext() {
			prevReportAsBytes, err := resultsIterator.Next()
			if err != nil {
				logger.Error("Failed to iterate reports")
				return shim.Error("Failed to iterate reports")
			}
			var prevReport Report
			err = json.Unmarshal([]byte(prevReportAsBytes.GetValue()), &prevReport)
			if err != nil {
				logger.Error("Failed to unmarshal previous Report")
				return shim.Error("Failed to unmarshal previous Report: " + err.Error())
			}
			isLocked = prevReport.IsLocked
			totalReports = append(totalReports, prevReport)

		}

		//get the noOfReports that will be equal to no of Report version already exists
		noOfReports = len(totalReports)
	}

	if isLocked {
		return shim.Success([]byte("Report can not be created as the report is locked."))
	}

	//check if Report exist then update the ReportVersion by lenngth of reports, else set to 1
	if reportExist {
		report.ReportVersion = generateVersion(noOfReports)
	} else {
		report.ReportVersion = generateVersion(0)
	}

	//create composite key to store Report
	var pk []string = []string{REPORT_PREFIX, report.DataCallID, report.DataCallVersion, report.Hash}
	reportKey, _ := stub.CreateCompositeKey(REPORT_DOCUMENT_TYPE, pk)
	logger.Debug("Composite Key for CreateReport", reportKey)

	//Check leadger if report already exists for this key
	prevReport, _ := stub.GetState(reportKey)
	if prevReport != nil {
		logger.Error("Report already exist for : ", reportKey)
		return shim.Error("Report already exist for : " + reportKey)
	}
	//saving the Report
	reportAsBytes, _ := json.Marshal(report)
	err = stub.PutState(reportKey, reportAsBytes)
	if err != nil {
		logger.Error("Error committing data for key: " + reportKey)
		return shim.Error("Error committing data for key: " + reportKey)
	}

	//Creating DataCallLog when the Report is created for IssuedDataCall
	// dataCallCreateReportLog := DataCallLog{report.DataCallID, report.DataCallVersion, ActionReportCandidate.ActionID, ActionReportCandidate.ActionDesc, report.UpdatedTs, report.CreatedBy}
	// dataCallCreateReportLogAsBytes, _ := json.Marshal(dataCallCreateReportLog)
	// this.LogDataCallTransaction(stub, string(dataCallCreateReportLogAsBytes))

	return shim.Success(nil)
}

// UpdateReport updates the existing report for a particular DataCall Id and version.
// Success: byte[]
// Error: {"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (this *SmartContract) UpdateReport(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("UpdateReport: enter")
	defer logger.Debug("UpdateReport: exit")
	var report Report
	err := json.Unmarshal([]byte(args), &report)
	if err != nil {
		logger.Error("UpdateReport: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("UpdateReport: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshalled object for UpdateReport", report)
	if report.DataCallID == "" {
		logger.Error("DataCallID is empty!!")
		return shim.Error("DataCallID is empty!!")
	}
	if report.DataCallVersion == "" {
		logger.Error("DataCallVersion is empty!!")
		return shim.Error("DataCallVersion is empty!!")
	}
	if report.Hash == "" {
		logger.Error("Hash is empty!!")
		return shim.Error("Hash is empty!!")
	}
	if report.ReportVersion == "" {
		logger.Error("ReportVersion is empty!!")
		return shim.Error("ReportVersion is empty!!")
	}

	//create composite key to store Report
	var pk []string = []string{REPORT_PREFIX, report.DataCallID, report.DataCallVersion, report.Hash}
	reportKey, _ := stub.CreateCompositeKey(REPORT_DOCUMENT_TYPE, pk)
	logger.Debug("Composite Key for CreateReport", reportKey)

	//Check leadger if report already exists for this key
	var prevReport Report
	prevReportAsBytes, _ := stub.GetState(reportKey)
	if prevReportAsBytes == nil {
		logger.Error("Report doesn't exist for : ", reportKey)
		return shim.Error("Report doesn't exist for : " + reportKey)
	}

	err = json.Unmarshal(prevReportAsBytes, &prevReport)
	if err != nil {
		return shim.Error("Failed to unmarshal prev report: " + err.Error())
	}
	// incoming report is candidate while current report is locked no update
	// in coming report is candidate while pre report is not locked update
	// incoming report is accepted or published or withheld and previous is locked or unlocked doesn't matter
	if report.Status == STATUS_CANDIDATE && prevReport.IsLocked == true {
		return shim.Success([]byte("This report can not be updated as the report is locked for updated."))
	}
	report.IsLocked = true

	//fetching all the records for particular Id and Version
	var pks []string = []string{REPORT_PREFIX, report.DataCallID, report.DataCallVersion}

	//Step-1: getting the reports to set isLocked to true
	validateReportIterator, _ := stub.GetStateByPartialCompositeKey(REPORT_DOCUMENT_TYPE, pks)

	defer validateReportIterator.Close()

	if report.Status != STATUS_CANDIDATE {
		for validateReportIterator.HasNext() {
			prevReportAsBytes, err := validateReportIterator.Next()
			if err != nil {
				logger.Error("Failed to iterate reports")
				return shim.Error("Failed to iterate reports")
			}
			var prevReport Report
			err = json.Unmarshal([]byte(prevReportAsBytes.GetValue()), &prevReport)
			if err != nil {
				logger.Error("Failed to unmarshal report ")
				return shim.Error("Failed to unmarshal report: " + err.Error())
			}
			if report.Status == prevReport.Status {
				return shim.Success([]byte("This report can not be updated as the report is locked for updated."))
			}
		}
	}

	resultsIterator, errmsg := stub.GetStateByPartialCompositeKey(REPORT_DOCUMENT_TYPE, pks)
	if errmsg != nil {
		logger.Error("UpdateReport: Failed to get state for previous reports")
		return shim.Error("UpdateReport: Failed to get state for reports")
	}
	defer resultsIterator.Close()
	logger.Debug("UpdateReport: ", resultsIterator)
	for resultsIterator.HasNext() {
		prevReportAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate reports")
			return shim.Error("Failed to iterate reports")
		}

		var prevReport Report
		err = json.Unmarshal([]byte(prevReportAsBytes.GetValue()), &prevReport)
		if err != nil {
			logger.Error("Failed to unmarshal report ")
			return shim.Error("Failed to unmarshal report: " + err.Error())
		}

		//if prevReport hash matches current hash then save the current report
		if prevReport.Hash == report.Hash {
			saveReportAsBytes, _ := json.Marshal(report)
			fmt.Println("this is input report ", report)
			err = stub.PutState(reportKey, saveReportAsBytes)
			if err != nil {
				logger.Error("Error committing data for key: " + reportKey)
				return shim.Error("Error committing data for key: " + reportKey)
			}
		} else {
			//else save the prevReport, updating isLocked to true
			prevReport.IsLocked = true
			//save the prevReport with isLocked to true
			var pkForPrevReport []string = []string{REPORT_PREFIX, prevReport.DataCallID, prevReport.DataCallVersion, prevReport.Hash}
			prevReportKey, _ := stub.CreateCompositeKey(REPORT_DOCUMENT_TYPE, pkForPrevReport)
			savePrevReportAsBytes, _ := json.Marshal(prevReport)
			err = stub.PutState(prevReportKey, savePrevReportAsBytes)
			if err != nil {
				logger.Error("Error committing data for key: " + prevReportKey)
				return shim.Error("Error committing data for key: " + prevReportKey)
			}
		}
	}

	//Creating DataCallLog when the Report is Updated
	var dataCallUpdateReportLog DataCallLog
	if report.Status == STATUS_ACCEPTED {
		dataCallUpdateReportLog = DataCallLog{report.DataCallID, report.DataCallVersion, ActionReportAccepted.ActionID, ActionReportAccepted.ActionDesc, report.UpdatedTs, report.CreatedBy}
	} else if report.Status == STATUS_PUBLISHED {
		dataCallUpdateReportLog = DataCallLog{report.DataCallID, report.DataCallVersion, ActionReportPublished.ActionID, ActionReportPublished.ActionDesc, report.UpdatedTs, report.CreatedBy}
	} else if report.Status == STATUS_WITHHELD {
		dataCallUpdateReportLog = DataCallLog{report.DataCallID, report.DataCallVersion, ActionReportWithheld.ActionID, ActionReportWithheld.ActionDesc, report.UpdatedTs, report.CreatedBy}
	}

	dataCallUpdateReportLogAsBytes, _ := json.Marshal(dataCallUpdateReportLog)
	this.LogDataCallTransaction(stub, string(dataCallUpdateReportLogAsBytes))

	return shim.Success(nil)
}

// ListReportsByCriteria retrives all reports that match given criteria.
// params {json}: {
//  "dataCallId": "mandatory",
//  "dataCallVersion": "mandatory"
//  "startIndex":"optional",
//  "pageSize":"optional",
//  "status" :""}
// Success {byte[]}: byte[]
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}

func (this *SmartContract) ListReportsByCriteria(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("ListReportsByCriteria: enter")
	defer logger.Debug("ListReportsByCriteria: exit")
	var listReportsCriteria ListReportsCriteria
	err := json.Unmarshal([]byte(args), &listReportsCriteria)
	if err != nil {
		logger.Error("ListReportsByCriteria: Error during json.Unmarshal", err)
		return shim.Error(errors.New("ListReportsByCriteria: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshalled object ", listReportsCriteria)

	if listReportsCriteria.DataCallID == "" {
		logger.Error("DataCallID is empty!!")
		return shim.Error("DataCallID is empty!!")
	}
	if listReportsCriteria.DataCallVersion == "" {
		logger.Error("DataCallVersion is empty!!")
		return shim.Error("DataCallVersion is empty!!")
	}
	//fetching all the different hash for a particular ID and Version
	var reports []Report
	var queryStr string
	queryStr = fmt.Sprintf("{\"selector\":{\"dataCallID\":\"%s\",\"dataCallVersion\":\"%s\"},\"use_index\": [\"_design/updatedTs\"],\"sort\":[{\"updatedTs\":\"desc\"}]}", listReportsCriteria.DataCallID, listReportsCriteria.DataCallVersion)

	resultsIterator, err := stub.GetQueryResult(queryStr)
	if err != nil {
		return shim.Error("ListReportsByCriteria: failed to get list of Reports: " + err.Error())
	}

	defer resultsIterator.Close()
	if !resultsIterator.HasNext() {
		reportAsByte, _ := json.Marshal(reports)
		logger.Debug("ListReportsByCriteria: reportAsByte", reportAsByte)
		return shim.Success(reportAsByte)
	}
	for resultsIterator.HasNext() {
		reportAsBytes, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("ListReportsByCriteria: Failed to iterate for Reports ")
		}

		var report Report
		err = json.Unmarshal([]byte(reportAsBytes.GetValue()), &report)
		if err != nil {
			return shim.Error("ListReportsByCriteria: Failed to unmarshal Report: " + err.Error())
		}

		reports = append(reports, report)
	}

	//paginate the reports
	var paginatedReports []Report
	paginatedReports = paginateReport(reports, listReportsCriteria.StartIndex, listReportsCriteria.PageSize)
	reportsAsByte, _ := json.Marshal(paginatedReports)
	return shim.Success(reportsAsByte)
}

//helper function for pagination
func paginateReport(report []Report, startIndex int, pageSize int) []Report {
	if startIndex == 0 {
		startIndex = PAGINATION_DEFAULT_START_INDEX
	}
	// no pageSize specified then return all results
	if pageSize == 0 {
		pageSize = len(report)
		return report
	}
	limit := func() int {
		if startIndex+pageSize > len(report) {
			return len(report)
		} else {
			return startIndex + pageSize
		}
	}

	start := func() int {
		if startIndex > len(report) {
			return len(report) - 1
		} else {
			return startIndex - 1
		}
	}
	return report[start():limit()]
}

//returns the last updated report
func (this *SmartContract) GetHighestOrderReportByDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("ListReportsByCriteria: enter")
	defer logger.Debug("ListReportsByCriteria: exit")
	var getHighestOrderReport GetHighestOrderReport
	//var reports []Report
	err := json.Unmarshal([]byte(args), &getHighestOrderReport)
	if err != nil {
		logger.Error("GetHighestOrderReportByDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetHighestOrderReportByDataCall: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshalled object ", getHighestOrderReport)
	if getHighestOrderReport.DataCallID == "" {
		return shim.Error("GetHighestOrderReportByDataCall: ID is Empty")
	}

	//fetch all the repports based on Id and Version sorted by updatedTs
	var queryStr string
	queryStr = fmt.Sprintf("{\"selector\":{\"dataCallID\":\"%s\",\"dataCallVersion\":\"%s\"},\"use_index\": [\"_design/updatedTs\"],\"sort\":[{\"updatedTs\":\"desc\"}]}", getHighestOrderReport.DataCallID, getHighestOrderReport.DataCallVesrion)
	resultsIterator, err := stub.GetQueryResult(queryStr)

	if err != nil {
		logger.Error("GetDataCallVersionsById: Failed to get Data Calls")
		return shim.Error(errors.New("GetDataCallVersionsById: Failed to get Data Calls").Error())

	}
	defer resultsIterator.Close()
	logger.Debug("GetHighestOrderReportByDataCall: Iterating over Reports versions")

	if !resultsIterator.HasNext() {
		fmt.Println("This is hasNext()", resultsIterator.HasNext())
		logger.Debug("GetHighestOrderReportByDataCall: No Reports found for data call returning empty array")
		//reportAsByte, _ := json.Marshal(reports)
		return shim.Error(errors.New("No Reports Found").Error())
	}
	//for resultsIterator.HasNext() {
	dataCallAsBytes, err := resultsIterator.Next()
	fmt.Println("Next inside n error", dataCallAsBytes, err)
	if err != nil {
		logger.Error("Failed to iterate data call")
		return shim.Error("Failed to iterate data call")
	}
	var tempReport Report

	//tempReport has report sorted based on updatedTs
	err = json.Unmarshal([]byte(dataCallAsBytes.GetValue()), &tempReport)

	logger.Debug("GetHighestOrderReportByDataCall: DataCall > ", getHighestOrderReport.DataCallID)
	if err != nil {
		return shim.Error("GetHighestOrderReportByDataCall: Failed to unmarshal Reports: " + err.Error())
	}
	//reports = append(reports, tempReport)

	//}
	//return only one record(latest, as it is sorted already based on updatedTs)
	reportAsByte, _ := json.Marshal(tempReport)
	return shim.Success(reportAsByte)

}

// function-name: GetReportById (invoke)
// params {json}: {
// "id":"mandatory",
// "version": "mandatory"}
// Success {byte[]}: byte[]  - Report
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : returns a report of specifc version,id and hash.

func (this *SmartContract) GetReportById(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetReportById: enter")
	defer logger.Debug("GetReportById: exit")

	var getReport GetReportById
	err := json.Unmarshal([]byte(args), &getReport)
	if err != nil {
		logger.Error("GetReportById: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetReportById: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshalled object ", getReport)

	if getReport.DataCallID == "" || getReport.DataCallVersion == "" || getReport.Hash == "" {
		return shim.Error("DataCall ID, DataCall Version and hash can not be Empty")
	}

	var pks []string = []string{REPORT_PREFIX, getReport.DataCallID, getReport.DataCallVersion, getReport.Hash}
	reportKey, _ := stub.CreateCompositeKey(REPORT_DOCUMENT_TYPE, pks)
	reportAsBytes, err := stub.GetState(reportKey)
	if err != nil {
		logger.Error("Error retreiving data for key ", reportKey)
		return shim.Error("Error retreiving data for key" + reportKey)
	}
	return shim.Success(reportAsBytes)

}

func GetLatestaReport(stub shim.ChaincodeStubInterface, IDs []string, versions []string) map[string]Report {
	logger.Info("Inside GetLatestaReport and args are ", IDs, versions)
	var latestReport map[string]Report
	latestReport = make(map[string]Report)

	//fetch all the repports based on Id and Version sorted by updatedTs
	var queryStr string

	queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"dataCallID\":{\"$in\":[%s]},\"dataCallVersion\":{\"$in\":[%s]}},\"use_index\":[\"_design/updatedTs\"],\"sort\":[{\"updatedTs\":\"desc\"}]}", REPORT_PREFIX, strings.Trim(fmt.Sprint(IDs), "[]"), strings.Trim(fmt.Sprint(versions), "[]"))

	deltaResultsIterator, _ := stub.GetQueryResult(queryStr)
	defer deltaResultsIterator.Close()

	if !deltaResultsIterator.HasNext() {
		logger.Info("GetLatestReport: No report found for criteria, returning 0 delta")
		return latestReport
	}
	var prevDataCallId string
	for deltaResultsIterator.HasNext() {
		latestReportAsBytes, _ := deltaResultsIterator.Next()

		var tempReport Report
		_ = json.Unmarshal([]byte(latestReportAsBytes.GetValue()), &tempReport)

		if prevDataCallId != tempReport.DataCallID {
			latestReport[tempReport.DataCallID] = tempReport
			prevDataCallId = tempReport.DataCallID
		}

	}

	return latestReport
}

func (this *SmartContract) ResetWorldState(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("ResetWorldState: enter")
	defer logger.Debug("ResetWorldState: exit")

	dataCallsDeleteCount, _ := DeleteStateByKey(stub, DATA_CALL_PREFIX)
	reportsDeleteCount, _ := DeleteStateByKey(stub, REPORT_PREFIX)
	likesDeleteCount, _ := DeleteStateByKey(stub, LIKE_PREFIX)
	consentsDeleteCount, _ := DeleteStateByKey(stub, CONSENT_PREFIX)
	datacallLogDeleteCount, _ := DeleteStateByKey(stub, DATACALL_LOG_PREFIX)
	// logger.Debugf("%s DataCalls, %s Reports, %s Likes, %s Consents Deleted, %s DataCallsLog", dataCallsDeleteCount, reportsDeleteCount, likesDeleteCount, consentsDeleteCount, datacallLogDeleteCount)
	totalRecordsDeleted := dataCallsDeleteCount + reportsDeleteCount + likesDeleteCount + consentsDeleteCount + datacallLogDeleteCount
	logger.Debug("Total Records Deleted: ", totalRecordsDeleted)
	return shim.Success([]byte(strconv.Itoa(totalRecordsDeleted)))
}

//helper function to reset world state
func DeleteStateByKey(stub shim.ChaincodeStubInterface, key string) (int, error) {
	logger.Debug("DeleteStateByKey: enter")
	defer logger.Debug("DeleteStateByKey: exit")
	var document_prefix string
	var document_type string

	switch key {
	case DATA_CALL_PREFIX:
		document_prefix = DATA_CALL_PREFIX
		document_type = DOCUMENT_TYPE
	case REPORT_PREFIX:
		document_prefix = REPORT_PREFIX
		document_type = REPORT_DOCUMENT_TYPE
	case LIKE_PREFIX:
		document_prefix = LIKE_PREFIX
		document_type = LIKE_DOCUMENT_TYPE
	case CONSENT_PREFIX:
		document_prefix = CONSENT_PREFIX
		document_type = CONSENT_DOCUMENT_TYPE
	case DATACALL_LOG_PREFIX:
		document_prefix = DATACALL_LOG_PREFIX
		document_type = DATACALL_LOG_DOCUMENT

	}

	var recordsDeletedCount = 0
	var pks []string = []string{document_prefix}
	iterator, err := stub.GetStateByPartialCompositeKey(document_type, pks)
	//iterator, err := stub.GetQueryResult(queryStr)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get iterator for partial composite key: Error: %s", err.Error())
		logger.Error(errorMsg)
		return 0, err
	}
	// Once we are done with the iterator, we must close it
	defer iterator.Close()
	logger.Debugf("Starting to delete all records with prefix %s", document_prefix)
	for iterator.HasNext() {
		responseRange, err := iterator.Next()
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to get next record from iterator: %s", err.Error())
			logger.Error(errorMsg)
		}
		recordKey := responseRange.GetKey()
		logger.Debugf("About to delete record with key %s", recordKey)
		err = stub.DelState(recordKey)

		if err != nil {
			errorMsg := fmt.Sprintf("Failed to delete record '%d' with key %s: %s", recordsDeletedCount, recordKey, err.Error())
			logger.Error(errorMsg)
			return 0, err
		}

		recordsDeletedCount++
		// logger.Debugf("%s - Successfully deleted record '%d' ", recordsDeletedCount)
	}
	logger.Debug("Finished deleting all records for prefix: ", document_prefix)
	return recordsDeletedCount, nil

}

// ToChaincodeArgs Prepare chaincode arguments for invoke chaincode

func ToChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

// InvokeChaincodeOnChannel invokes a specified chaincode function on a different channel except the current channel
func InvokeChaincodeOnChannel(stub shim.ChaincodeStubInterface, chanicodeName string, funcName string, payload string, channelID string) pb.Response {
	var invokeResponse pb.Response
	currentChannelID := stub.GetChannelID()
	logger.Debug("InvokeChaincodeOnChannel: currentChannelID > ", currentChannelID)
	logger.Debug("InvokeChaincodeOnChannel: Called Channel Id > ", channelID)
	// Do not execute on the same channel
	var channels Channels
	channels.ChannelIDs = make([]string, 1)
	channels.ChannelIDs[0] = channelID
	channelsReqJson, _ := json.Marshal(channels)
	invokeRequest := ToChaincodeArgs(funcName, payload, string(channelsReqJson))
	invokeResponse = stub.InvokeChaincode(chanicodeName, invokeRequest, channelID)
	return invokeResponse
}
