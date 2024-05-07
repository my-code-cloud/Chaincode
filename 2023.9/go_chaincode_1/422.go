package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type ReportInfo struct {
	ReportName  string `json:"reportName"`
	AttackPoint string `json:"attackPoint"`
	Reporter    string `json:"reporter"`
}

type ReportCertification struct {
	CertificationURL string `json:"certificationurl"`
	Reporter         string `json:"reporter"`
	CreateDate       string `json:"createdate"`
}

type UserRating struct {
	User    string  `json:"user"`
	Average float64 `json:"average"`
	Rates   []Rate  `json:"rates"`
}
type Rate struct {
	ProjectTitle string  `json:"projecttitle"`
	Score        float64 `json:"score"`
}

// func (s *SmartContract) getReport(ctx contractapi.TransactionContextInterface, reportSeq string) (string, error) {
// 	ReportAsBytes, err := ctx.GetStub().GetState(reportSeq)

// 	if err != nil {
// 		return "", fmt.Errorf("Failed to read from world state. %s", err.Error())
// 	}

// 	if ReportAsBytes == nil {
// 		return "", fmt.Errorf("%s does not exist", reportSeq)
// 	}

// 	return string(ReportAsBytes[:]), nil
// }

func (s *SmartContract) SubmitReport(ctx contractapi.TransactionContextInterface, reportName string, attackPoint string, reporter string) error {
	var report = ReportInfo{ReportName: reportName, AttackPoint: attackPoint, Reporter: reporter}
	reportAsBytes, _ := json.Marshal(report)
	byteBuffer := make([]byte, 3)
	v, size := utf8.DecodeRune(byteBuffer)
	reportAsEnc := utf8.EncodeRune(reportAsBytes, size)
	reportAsMD5 := md5.Sum(reportAsEnc)

	fmt.Println(reportAsEnc)
	fmt.Println(reportAsMD5)
	//	r, size := utf8.DecodeRune(byteBuffer)
	// var report = Report{ReportSeq: reportSeq, ReportName: reportName, AttackReport: attackReport, CreatedDate: time.Now()}
	// var hashTemp = reportName + attackReport
	// reportMd5 := md5.Sum([]byte(hashTemp))
	// reportAsBytes, _ := json.Marshal(reportMd5)

	// 	rInt := '한'
	// 	byteBuffer := make([]byte, 3)
	// 한글이 3바이트이기 때문에 바이트 배열로 3칸을 할당해줍니다.
	//	r, size := utf8.DecodeRune(byteBuffer)
	// v, size := utf8.DecodeRuneInString(x)
	// 	n := utf8.EncodeRune(byteBuffer, rInt)
	return ctx.GetStub().PutState(reportName, reportAsBytes)
}

func (s *SmartContract) AddUser(ctx contractapi.TransactionContextInterface, username string) error {

	var user = UserRating{User: username, Average: 0}
	userAsBytes, _ := json.Marshal(user)

	return ctx.GetStub().PutState(username, userAsBytes)
}

func (s *SmartContract) AddRating(ctx contractapi.TransactionContextInterface, username string, prjTitle string, prjscore string) error {

	// getState User
	userAsBytes, err := ctx.GetStub().GetState(username)

	if err != nil {
		return err
	} else if userAsBytes == nil { // no State! error
		return fmt.Errorf("\"Error\":\"User does not exist: " + username + "\"")
	}
	// state ok
	user := UserRating{}
	err = json.Unmarshal(userAsBytes, &user)
	if err != nil {
		return err
	}
	// create rate structure
	newRate, _ := strconv.ParseFloat(prjscore, 64)
	var Rate = Rate{ProjectTitle: prjTitle, Score: newRate}

	rateCount := float64(len(user.Rates))

	user.Rates = append(user.Rates, Rate)

	user.Average = (rateCount*user.Average + newRate) / (rateCount + 1)
	// update to User World state
	userAsBytes, err = json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to Marshaling: %v", err)
	}

	err = ctx.GetStub().PutState(username, userAsBytes)
	if err != nil {
		return fmt.Errorf("failed to AddRating: %v", err)
	}
	return nil
}

func (s *SmartContract) ReadRating(ctx contractapi.TransactionContextInterface, username string) (string, error) {

	UserAsBytes, err := ctx.GetStub().GetState(username)

	if err != nil {
		return "", fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if UserAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", username)
	}

	// user := new(UserRating)
	// _ = json.Unmarshal(UserAsBytes, &user)

	return string(UserAsBytes[:]), nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create teamate chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting teamate chaincode: %s", err.Error())
	}
}
