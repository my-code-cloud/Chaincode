package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing Vaccination Records
type SmartContract struct {
	contractapi.Contract
}

type Patient struct {
	ID           string        `json:"id"`
	FirstName    string        `json:"firstName"`
	LastName     string        `json:"lastName"`
	DateOfBirth  string        `json:"dateOfBirth"`
	Vaccinations []Vaccination `json:"vaccinations"`
}

type Vaccine struct {
	ID      string `json:"id"`
	Disease string `json:"disease"`
	Brand   string `json:"brand"`
}

type VaccinationListing struct {
	Vaccines []Vaccine `json:"vaccines"`
}

type Vaccination struct {
	ID                 string  `json:"id"`
	Vaccine            Vaccine `json:"vaccine"`
	Date               string  `json:"date"`
	HealthCareProvider string  `json:"healthCareProvider"`
}

func (s *SmartContract) PatientExists(ctx contractapi.TransactionContextInterface, nationalID string, country string) (bool, error) {
	id := createID(nationalID, country)
	patientJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("Failed to read from world state: %v", err)
	}

	return patientJSON != nil, nil
}

func (s *SmartContract) CreatePatient(ctx contractapi.TransactionContextInterface, nationalID string, country string, firstName string, lastName string, dateOfBirth string) error {
	id := createID(nationalID, country)
	exists, err := s.PatientExists(ctx, nationalID, country)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("The patient with id: %s and country %s exists", nationalID, country)
	}

	patient := Patient{
		ID:           id,
		FirstName:    firstName,
		LastName:     lastName,
		DateOfBirth:  dateOfBirth,
		Vaccinations: []Vaccination{},
	}
	patientJSON, err := json.Marshal(patient)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, patientJSON)
}

func (s *SmartContract) ReadPatient(ctx contractapi.TransactionContextInterface, nationalID string, country string) (*Patient, error) {
	id := createID(nationalID, country)
	patientJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state: %v", err)
	}
	if patientJSON == nil {
		return nil, fmt.Errorf("Patient not found.")
	}

	var patient Patient
	err = json.Unmarshal(patientJSON, &patient)
	if err != nil {
		return nil, err
	}

	return &patient, nil
}

func (s *SmartContract) ReadVaccinationListing(ctx contractapi.TransactionContextInterface) (*VaccinationListing, error) {
	vaccinationListingJSON, err := ctx.GetStub().GetState("vaccinationListing")
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state: %v", err)
	}
	if vaccinationListingJSON == nil {
		return nil, fmt.Errorf("The vaccination listing does not exist")
	}

	var vaccinationListing VaccinationListing
	err = json.Unmarshal(vaccinationListingJSON, &vaccinationListing)
	if err != nil {
		return nil, err
	}

	return &vaccinationListing, nil
}

func (s *SmartContract) AddVaccinationToPatient(ctx contractapi.TransactionContextInterface, nationalID string, country string, disease string, brand string, vaccineDate string, healthCareProvider string) (*Patient, error) {
	patient, err := s.ReadPatient(ctx, nationalID, country)
	if err != nil {
		return nil, err
	}

	validVaccines, err := s.ReadVaccinationListing(ctx)

	if err == nil {
		for _, validVaccine := range validVaccines.Vaccines {
			if validVaccine.Disease == disease && validVaccine.Brand == brand {

				vaccineId := uuid.New().String()

				vaccination := Vaccination{
					ID:                 vaccineId,
					Vaccine:            validVaccine,
					Date:               vaccineDate,
					HealthCareProvider: healthCareProvider,
				}

				patient.Vaccinations = append(patient.Vaccinations, vaccination)

				patientJSON, err := json.Marshal(patient)

				if err != nil {
					return nil, err
				}

				ctx.GetStub().PutState(patient.ID, patientJSON)
				return patient, nil

			}
		}
		return nil, fmt.Errorf("The vaccine is not valid")
	}

	return nil, err
}

// Create ID from nationalID and country
func createID(nationalID string, country string) string {
	hash := sha256.Sum256([]byte(nationalID + country))
	return fmt.Sprintf("%x", hash)
}

func (s *SmartContract) SyncVaccinationListing(ctx contractapi.TransactionContextInterface, vaccinesJSON string) error {
	vaccines := []Vaccine{}

	err := json.Unmarshal([]byte(vaccinesJSON), &vaccines)
	vaccinationListing := VaccinationListing{
		Vaccines: vaccines,
	}
	vaccinationListingJSON, err := json.Marshal(vaccinationListing)
	if err != nil {
		return fmt.Errorf("Failed unmarshalling vaccination listing: %v", err)
	}

	return ctx.GetStub().PutState("vaccinationListing", vaccinationListingJSON)
}

func (s *SmartContract) GetPatientByID(ctx contractapi.TransactionContextInterface, id string) (*Patient, error) {
	patientJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state: %v", err)
	}

	patient := Patient{}

	jsonMarshallingError := json.Unmarshal(patientJSON, &patient)

	if jsonMarshallingError != nil {
		return nil, jsonMarshallingError
	}

	return &patient, nil
}

func main() {
	vaccinationChainCode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		log.Panicf("Error creating chaincode: %v", err)
	}

	if err := vaccinationChainCode.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}
