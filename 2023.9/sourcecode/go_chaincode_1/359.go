/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a appointment
type SmartContract struct {
	contractapi.Contract
}

// Appointment describes basic details of what makes up a appointment
type App struct {
	AppointmentId   string `json:"appointmentid"`
	To              string `json:"to"`
	From            string `json:"from"`
	Date            string `json:"date"`
	Org             string `json:"org"`
	Content         string `json:"content"`
	UpdatedByUserId string `json:"updatedbyuserId"`
	UpdatedByOrg    string `json:"updatedbyorg"`
}

// InitLedger adds a base set of cars to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	_ = ctx.GetStub().PutState("transationCount", []byte(strconv.Itoa(1)))
	return nil
}

// CreateCar adds a new car to the world state with given details
func (s *SmartContract) CreateAppointment(ctx contractapi.TransactionContextInterface, appointmentId string, to string, from string, date string, org string, content string) error {

	var appointment = App{
		AppointmentId: appointmentId,
		To:            to,
		From:          from,
		Date:          date,
		Org:           org,
		Content:       content,
		UpdatedByUserId:``,
		UpdatedByOrg:``,
	}
	appointmentAsBytes, _ := json.Marshal(appointment)
	id := appointmentId

	return ctx.GetStub().PutState(id, appointmentAsBytes)
}

func (s *SmartContract) QueryAppointment(ctx contractapi.TransactionContextInterface, appointmentId string) *App {
	appointmentAsBytes, err := ctx.GetStub().GetState(apointmentId)

	if err != nil {
		return nil
	}
	if appointmentAsBytes == nil {
		return nil
	}

	appointment := new(App)
	_ = json.Unmarshal(appointmentAsBytes, appointment)
	return appointment
}

func (s *SmartContract) QueryAllAppointment(ctx contractapi.TransactionContextInterface, org string) ([]string, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []string{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}
		if queryResponse.Key == `transationCount` {
			continue
		}

		appointment := new(App)
		_ = json.Unmarshal(queryResponse.Value, appointment)
		if appointment.To == org {
			results = append(results, queryResponse.Key)
		}
	}

	return results, nil
}

func (s *SmartContract) GetAppointmentHistory(ctx contractapi.TransactionContextInterface, appointmentId string) (string, error) {

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(appointmentId)
	if err != nil {
		return `nil`, fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return `nil`, fmt.Errorf(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return buffer.String(), nil
}

func (s *SmartContract) UpdateAppointment(ctx contractapi.TransactionContextInterface, appointmentId string, content string, updatedByUserId string, org string) string {
	app := s.QueryAppointment(ctx, appointmentId)

	app.Content = content
	app.UpdatedByUserId = updatedByUserId
	app.UpdatedByOrg=org
	appAsBytes, _ := json.Marshal(tran)
	err1 := ctx.GetStub().PutState(appointmentId, appAsBytes)
	if err1 != nil {
		return `err`
	}
	return `true`

}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create appointment chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting appointment chaincode: %s", err.Error())
	}
}
