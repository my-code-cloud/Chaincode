package main

import (
"encoding/json"
"fmt"
"time"

"github.com/golang/protobuf/ptypes"
"github.com/hyperledger/fabric-chaincode-go/shim"
"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
contractapi.Contract
}

type Prescription struct {
ID             string `json:"id"`
MedicationName string `json:"medicationName"`
Quantity       string `json:"quantity"`
}

type Illness struct {
ID string `json:"id"`
}

type PatientPrivateDetails struct {
Bill         float32 `json:"bill"`
PatientID    string  `json:"patientID"`
ContractType string  `json:"contractType"`
}

// PaginatedQueryResult structure used for returning paginated query results and metadata
type PaginatedQueryResult struct {
Records             []*Patient `json:"records,omitempty" metadata:"records,optional" `
FetchedRecordsCount int32      `json:"fetchedRecordsCount"`
Bookmark            string     `json:"bookmark"`
}

type Patient struct {
ID               string      `json:"id"`
FirstName        string      `json:"firstName"`
LastName         string      `json:"last_name"`
Email            string      `json:"email"`
Description      string      `json:"description"`
GroupType        string      `json:"groupType"`
Allergies        []string    `json:"allergies,omitempty" metadata:"allergies,optional"`
EmergencyContact string      `json:"emergencyContact"`
Diagnosis        []Diagnosis `json:"diagnosis,omitempty" metadata:"diagnosis,optional" `
DoctorsID        []string    `json:"doctorsId,omitempty" metadata:"doctorsId,optional"`
Inpatient        string      `json:"inpatient,omitempty" metadata:"inpatient,optional"`
}

type Diagnosis struct {
ID           string         `json:"id"`
DoctorsID    string         `json:"doctorsId,omitempty" metadata:"doctorsID,optional"`
Description  string         `json:"description"`
Illness      string         `json:"illness"`
Prescription []Prescription `json:"prescriptions,omitempty" metadata:"prescriptions,optional"`
}

func (s *SmartContract) GetAllPatients(ctx contractapi.TransactionContextInterface) ([]*Patient, error) {
// range query with empty string for startKey and endKey does an
// open-ended query of all assets in the chaincode namespace.
resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
if err != nil {
return nil, err
}
defer resultsIterator.Close()

var assets []*Patient
for resultsIterator.HasNext() {
queryResponse, err := resultsIterator.Next()
if err != nil {
return nil, err
}

var asset Patient
err = json.Unmarshal(queryResponse.Value, &asset)

if err != nil {
return nil, err
}

error := ValidateDoctorsID(ctx, asset.DoctorsID)

if error != nil {
	return nil, err
}
assets = append(assets, &asset)
}

return assets, nil
}

// //----------------------------------------------ADMIN FUNCTIONS ----------------------------------------------//
func (s *SmartContract) CreatePatient(ctx contractapi.TransactionContextInterface, patient Patient) error {

err := ValidateRole(ctx, "doctor")

if err != nil {
return err
}

// transientMap, err := ctx.GetStub().GetTransient()
// if err != nil {
// return fmt.Errorf("error getting transient: %v", err)
// }

// type PatientAllDetails struct {
// Bill         float32 `json:"bill"`
// ContractType string  `json:"contractType"`
// }

// // Asset properties are private, therefore they get passed in transient field, instead of func args
// transientAssetJSON, ok := transientMap["asset_properties"]
// if !ok {
// //log error to stdout
// return fmt.Errorf("asset not found in the transient map input")
// }

// var patientBill PatientAllDetails
// err = json.Unmarshal(transientAssetJSON, &patientBill)
// if err != nil {
// return fmt.Errorf("failed to unmarshal JSON: %v", err)
// }

// if patient.ID == "" {
// return fmt.Errorf("the patient ID cannot be empty")
// }
exists, err := s.PatientExists(ctx, patient.ID)
if err != nil {
return err
}
if exists {
return fmt.Errorf("the patient %s already exists", patient.ID)
}
if patient.FirstName == "" {
return fmt.Errorf("the patient FirstName cannot be empty")
}
if patient.LastName == "" {
return fmt.Errorf("the patient LastName cannot be empty")
}
if patient.Email == "" {
return fmt.Errorf("the patient Email cannot be empty")
}
if patient.Description == "" {
return fmt.Errorf("the patient Description cannot be empty")
}
if patient.GroupType == "" {
return fmt.Errorf("the patient GroupType cannot be empty")
}
if patient.EmergencyContact == "" {
return fmt.Errorf("the patient EmergencyContact cannot be empty")
}

// Marshal patient public info to JSON
patientJSON, err := json.Marshal(patient)
if err != nil {
return err
}

// Persist patient to world state.
err = ctx.GetStub().PutState(patient.ID, patientJSON)
if err != nil {
return fmt.Errorf("failed to put to world state. %v", err)
}

// patientPrivateDetails := PatientPrivateDetails{
// PatientID:    patient.ID,
// Bill:         patientBill.Bill,
// ContractType: patientBill.ContractType,
// }

// // Marshal patient private info to JSON
// patientJSON, err = json.Marshal(patientPrivateDetails)
// if err != nil {
// return err
// }

// orgCollection, err := s.getCollectionName(ctx)
// if err != nil {
// return fmt.Errorf("failed to infer private collection name for the org: %v", err)
// }

// // Persist patient private info to world state.
// err = ctx.GetStub().PutPrivateData(orgCollection, patient.ID, patientJSON)
return err
}

func (s *SmartContract) DeletePatientPrivateData(ctx contractapi.TransactionContextInterface, id string) error {
exists, err := s.PatientExists(ctx, id)
if err != nil {
return err
}
if !exists {
return fmt.Errorf("the patient %s does not exist", id)
}

err = ValidateRole(ctx, "admin")
if err != nil {
return err
}

orgCollection, err := s.getCollectionName(ctx)
if err != nil {
return fmt.Errorf("failed to infer private collection name for the org: %v", err)
}

patientBill, err := ctx.GetStub().GetPrivateData(orgCollection, id)
if err != nil {
return fmt.Errorf("%v", err)
}

if patientBill == nil {
return fmt.Errorf("the patient %s does not exist", id)
}

// Delete patient from private collection.
err = ctx.GetStub().DelPrivateData(orgCollection, id)
if err != nil {
return fmt.Errorf("failed to delete from private state. %v", err)
}
return nil
}

func (*SmartContract) GetPatientsByRangeWithPagination(ctx contractapi.TransactionContextInterface, pageSize int, bookmark string) (*PaginatedQueryResult, error) {

resultsIterator, responseMetadata, err := ctx.GetStub().GetStateByRangeWithPagination("", " ", int32(pageSize), bookmark)
if err != nil {
return nil, err
}
defer resultsIterator.Close()

assets, err := constructQueryResponseFromIterator(ctx, resultsIterator)
if err != nil {
return nil, err
}

return &PaginatedQueryResult{
Records:             assets,
FetchedRecordsCount: responseMetadata.FetchedRecordsCount,
Bookmark:            responseMetadata.Bookmark,
}, nil
}

// //----------------------------------------------Doctor FUNCTIONS ----------------------------------------------//
func (s *SmartContract) ReadPatientById(ctx contractapi.TransactionContextInterface, id string) (*Patient, error) {
return s.ReadPatientByIdWithBool(ctx, id, true)
}

func (s *SmartContract) ReadPatientByIdWithBool(ctx contractapi.TransactionContextInterface, id string, flag bool) (*Patient, error) {
patientJSON, err := ctx.GetStub().GetState(id)
if err != nil {
return nil, fmt.Errorf("failed to read from world state: %v", err)
}
if patientJSON == nil {
return nil, fmt.Errorf("the patient %s does not exist", id)
}

var patient Patient
err = json.Unmarshal(patientJSON, &patient)
if err != nil {
return nil, fmt.Errorf("failed to deserialize patient" + id)
}

if flag == true {
error := ValidateDoctorsID(ctx, patient.DoctorsID)

if error != nil {
return nil, fmt.Errorf("%v", error)
}
}
return &patient, nil
}

func (s *SmartContract) CreateDiagnosis(ctx contractapi.TransactionContextInterface,
diagnosisJson string, patientId string) (*Patient, error) {

var diagnosis Diagnosis
err := json.Unmarshal([]byte(diagnosisJson), &diagnosis)
if err != nil {
return nil, fmt.Errorf("failed to unmarshal diagnosis JSON: %v", err)
}

patient, err := s.ReadPatientByIdWithBool(ctx, patientId, true)
if err != nil {
return nil, err
}

diagnosis.DoctorsID, err = ctx.GetClientIdentity().GetID()
if err != nil {
return nil, fmt.Errorf("failed to get client identity: %v", err)
}

patient.Diagnosis = append(patient.Diagnosis, diagnosis)

patientJSON, err := json.Marshal(patient)
if err != nil {
return nil, fmt.Errorf("Failed to serialize patient")
}

ctx.GetStub().PutState(patientId, patientJSON)

return patient, nil
}

func (s *SmartContract) UpdateDiagnosis(ctx contractapi.TransactionContextInterface,
patientId string,
diagnosisJson string) error {

var diagnosis Diagnosis
err := json.Unmarshal([]byte(diagnosisJson), &diagnosis)
if err != nil {
return fmt.Errorf("failed to unmarshal diagnosis JSON: %v", err)
}

patient, err := s.ReadPatientByIdWithBool(ctx, patientId, false)
if err != nil {
return fmt.Errorf("failed to get patient: %v", err)
}

err = ValidateDoctorsID(ctx, []string{diagnosis.DoctorsID})

if err != nil {
return fmt.Errorf("%v", err)
}

for i, d := range patient.Diagnosis {
if d.ID == diagnosis.ID {
patient.Diagnosis[i].Description = diagnosis.Description
patient.Diagnosis[i].Illness = diagnosis.Illness
patient.Diagnosis[i].Prescription = diagnosis.Prescription
break
}
}

patientJSON, err := json.Marshal(patient)
if err != nil {
return fmt.Errorf("failed to serialize patient")
}

ctx.GetStub().PutState(patientId, patientJSON)

return nil
}

// //----------------------------------------------COMMON DOCTORS&ADMINS FUNCTIONs ----------------------------------------------//
func (s *SmartContract) UpdatePatient(ctx contractapi.TransactionContextInterface,
patientJson string) error {

var patient Patient
err := json.Unmarshal([]byte(patientJson), &patient)
if err != nil {
return fmt.Errorf("failed to unmarshal patient JSON: %v", err)
}

exists, err := s.PatientExists(ctx, patient.ID)
if err != nil {
return err
}
if !exists {
return fmt.Errorf("the patient %s does not exist", patient.ID)
}

err1 := ValidateDoctorsID(ctx, patient.DoctorsID)
err2 := ValidateRole(ctx, "admin")

if err1 != nil && err2 != nil {
return fmt.Errorf("%v %v", err1, err2)
}

return ctx.GetStub().PutState(patient.ID, []byte(patientJson))
}

func (s *SmartContract) ReadPatientsByDoctorsID(ctx contractapi.TransactionContextInterface) ([]*Patient, error) {
doctorId, err := ctx.GetClientIdentity().GetID()

if err != nil {
return nil, fmt.Errorf("failed to get client identity: %v", err)
}

queryString := fmt.Sprintf(`{"selector": {"doctorsID": {"$elemMatch": {"$eq": "%s"}}}}`, doctorId)
return getQueryResultForQueryString(ctx, queryString)
}

// //----------------------------------------------NURSE&&SECRETARY FUNCTIONS ----------------------------------------------//
func (s *SmartContract) ReadPatientPerscriptions(ctx contractapi.TransactionContextInterface, patientId string) ([]Prescription, error) {
patient, err := s.ReadPatientByIdWithBool(ctx, patientId, false)

if err != nil {
return nil, fmt.Errorf("failed to get patient: %v", err)
}

err1 := ValidateRole(ctx, "nurse")
err2 := ValidateDoctorsID(ctx, patient.DoctorsID)

if err1 != nil && err2 != nil {
return nil, fmt.Errorf("%v %v", err1, err2)
}

prescription := make([]Prescription, 0)

for _, d := range patient.Diagnosis {
prescription = append(prescription, d.Prescription...)
}
return prescription, nil
}

func (s *SmartContract) ReadPatientBill(ctx contractapi.TransactionContextInterface, patientId string) (PatientPrivateDetails, error) {
collectionName, err := s.getCollectionName(ctx)
if err != nil {
return PatientPrivateDetails{}, err
}
flag, err := s.PatientExists(ctx, patientId)

if err != nil {
return PatientPrivateDetails{}, err
}

if !flag {
return PatientPrivateDetails{}, fmt.Errorf("the patient %s does not exist", patientId)
}

err1 := ValidateRole(ctx, "secretary")
err2 := ValidateRole(ctx, "admin")

if err1 != nil && err2 != nil {
return PatientPrivateDetails{}, fmt.Errorf("%v %v", err1, err2)
}

billAsBytes, error := ctx.GetStub().GetPrivateData(collectionName, patientId)
if error != nil {
return PatientPrivateDetails{}, fmt.Errorf("failed to get patient: %v", error)
}
if billAsBytes == nil {
return PatientPrivateDetails{}, fmt.Errorf("the patient %s does not exist", patientId)
}

// unmarshal the bill
var bill PatientPrivateDetails
err = json.Unmarshal(billAsBytes, &bill)
if err != nil {
return PatientPrivateDetails{}, fmt.Errorf("failed to deserialize patient: %v", err)
}

return bill, nil
}

func (s *SmartContract) UpdatePatientBill(ctx contractapi.TransactionContextInterface) error {

transientMap, err := ctx.GetStub().GetTransient()
if err != nil {
return fmt.Errorf("error getting transient: %v", err)
}

// Asset properties are private, therefore they get passed in transient field, instead of func args
transientAssetJSON, ok := transientMap["asset_properties"]
if !ok {
//log error to stdout
return fmt.Errorf("asset not found in the transient map input")
}

var bill PatientPrivateDetails
err = json.Unmarshal(transientAssetJSON, &bill)
if err != nil {
return fmt.Errorf("failed to unmarshal asset: %v", err)
}

flag, err := s.PatientExists(ctx, bill.PatientID)
if err != nil {
return fmt.Errorf("failed to get patient: %v", err)
}

if !flag {
return fmt.Errorf("the patient %s does not exist", bill.PatientID)
}

err = ValidateRole(ctx, "secretary")

if err != nil {
return fmt.Errorf("%v", err)
}

collectionName, err := s.getCollectionName(ctx)
if err != nil {
return err
}

billAsBytes, err := json.Marshal(bill)
if err != nil {
return fmt.Errorf("failed to serialize bill: %v", err)
}

return ctx.GetStub().PutPrivateData(collectionName, bill.PatientID, billAsBytes)
}

// //----------------------------------------------(*************__PATIENT__*************)----------------------------------------------\\
func (s *SmartContract) TransferPatient(ctx contractapi.TransactionContextInterface, id string, oldDoctor string, newDoctor string) error {
err := ValidatePatientID(ctx, id)
if err != nil {
return err
}

patient, err := s.ReadPatientByIdWithBool(ctx, id, false)
if !contains(patient.DoctorsID, oldDoctor) {
return fmt.Errorf("Doctor with id:" + oldDoctor + " does belong to this patient doctors or doesn't exist")
}

var j int = -1

for i, x := range patient.DoctorsID {
if x == oldDoctor {
j = i
}
}

patient.DoctorsID[j] = newDoctor

patientJSON, err := json.Marshal(patient)
if err != nil {
return err
}

return ctx.GetStub().PutState(id, patientJSON)
}

// //----------------------------------------------(*************_u_t_i_l_*************)----------------------------------------------\\
func (s *SmartContract) GetOrganization(ctx contractapi.TransactionContextInterface) (string, error) {
return ctx.GetClientIdentity().GetMSPID()
}

func (s *SmartContract) getDoctorID(ctx contractapi.TransactionContextInterface) (value string, err error) {
value, found, err := ctx.GetClientIdentity().GetAttributeValue("doctorId")

if err != nil {
return "", fmt.Errorf("failed to get doctorId: %v", err)
}

if !found || value == "" {
return "", fmt.Errorf("doctorId not found")
}

return value, nil
}

func ValidateDoctorsID(ctx contractapi.TransactionContextInterface, doctorsID []string) (err error) {
err = ValidateRole(ctx, "doctor")
if err != nil {
return err
}
//value, err := ctx.GetClientIdentity().GetID()
value , found , err := ctx.GetClientIdentity().GetAttributeValue("userId")
if err != nil {
	return err
	}
if !found || value == "" {
	return  fmt.Errorf("userID not found")
	}	
if !contains(doctorsID, value) {
return fmt.Errorf("the doctor %s does not belong to the group", value)
}
return nil
}

func ValidatePatientID(ctx contractapi.TransactionContextInterface, patientID string) (err error) {
err = ValidateRole(ctx, "patient")
if err != nil {
return err
}
//value, err := ctx.GetClientIdentity().GetID()
value , found , err := ctx.GetClientIdentity().GetAttributeValue("userId")
if err != nil {
	return err
	}
if !found || value == "" {
	return  fmt.Errorf("userID not found")
	}	
if patientID != value {
return fmt.Errorf("the patient %s does not belong to the group", value)
}
return nil
}

func ValidateRole(ctx contractapi.TransactionContextInterface, role string) (err error) {
error := ctx.GetClientIdentity().AssertAttributeValue("role", role)

if error != nil {
return fmt.Errorf("failed to get role: %v", error)
}

return nil
}

func (s *SmartContract) PatientExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
PatientJSON, err := ctx.GetStub().GetState(id)
if err != nil {
return false, fmt.Errorf("failed to read from world state: %v", err)
}

return PatientJSON != nil, nil
}

func (s *SmartContract) getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {

// Get the MSP ID of submitting client identity
clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
if err != nil {
return "", fmt.Errorf("failed to get verified MSPID: %v", err)
}

// Create the collection name
orgCollection := clientMSPID + "PrivateCollection"

return orgCollection, nil
}

func contains(list []string, elem string) bool {
for _, e := range list {
if e == elem {
return true
}
}
return false
}

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Patient, error) {
resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
if err != nil {
return nil, err
}
defer resultsIterator.Close()

return constructQueryResponseFromIterator(ctx, resultsIterator)
}

func constructQueryResponseFromIterator(ctx contractapi.TransactionContextInterface, resultsIterator shim.StateQueryIteratorInterface) ([]*Patient, error) {
var patients []*Patient
for resultsIterator.HasNext() {
queryResult, err := resultsIterator.Next()
if err != nil {
return nil, err
}
var patient Patient
err = json.Unmarshal(queryResult.Value, &patient)
if err != nil {
return nil, err
}
error := ValidateDoctorsID(ctx, patient.DoctorsID)

if error != nil {
continue
}
patients = append(patients, &patient)
}

return patients, nil
}

// //----------------------------------------------(*************_QUERY_HISTORY_*************)----------------------------------------------\\
// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResult struct {
Record    *Patient  `json:"record"`
TxId      string    `json:"txId"`
Timestamp time.Time `json:"timestamp"`
IsDelete  bool      `json:"isDelete"`
}

func (s *SmartContract) GetPatientHistory(ctx contractapi.TransactionContextInterface, patientID string) ([]HistoryQueryResult, error) {
//TODO check if it flag should be true or false
s.ReadPatientByIdWithBool(ctx, patientID, true)

resultsIterator, err := ctx.GetStub().GetHistoryForKey(patientID)
if err != nil {
return nil, err
}
defer resultsIterator.Close()

var records []HistoryQueryResult
for resultsIterator.HasNext() {
response, err := resultsIterator.Next()
if err != nil {
return nil, err
}

var patient Patient
if len(response.Value) > 0 {
err = json.Unmarshal(response.Value, &patient)
if err != nil {
return nil, err
}
} else {
patient = Patient{
ID: patientID,
}
}

timestamp, err := ptypes.Timestamp(response.Timestamp)
if err != nil {
return nil, err
}

record := HistoryQueryResult{
TxId:      response.TxId,
Timestamp: timestamp,
Record:    &patient,
IsDelete:  response.IsDelete,
}
records = append(records, record)
}

return records, nil
}

// //----------------------------------------------(*************IN_PATIENT*************)----------------------------------------------\\
func (s *SmartContract) InPatientCheckIn(ctx contractapi.TransactionContextInterface, patientID string) error {
patient, err := s.ReadPatientByIdWithBool(ctx, patientID, false)

if err != nil {
return err
}

err = ValidateRole(ctx, "secretary")
if err != nil {
return err
}

patient.Inpatient, err = ctx.GetClientIdentity().GetMSPID()

patientJSON, err := json.Marshal(patient)
if err != nil {
return err
}

return ctx.GetStub().PutState(patientID, patientJSON)
}

func (s *SmartContract) InPatientCheckOut(ctx contractapi.TransactionContextInterface, patientID string) error {

patient, err := s.ReadPatientByIdWithBool(ctx, patientID, false)
if err != nil {
return err
}

err = ValidateRole(ctx, "secretary")
if err != nil {
return err
}

patient.Inpatient = ""

patientJSON, err := json.Marshal(patient)
if err != nil {
return err
}

return ctx.GetStub().PutState(patientID, patientJSON)
}
