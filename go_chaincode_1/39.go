package main


import (
	"encoding/json"
	"fmt"
	"time"
	"log"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)
type SmartContract struct {
	contractapi.Contract
}
// key -> Asset -> 수강권(grew)
// WS: ID, 수강권NAME, 학원ID, 수강권정보, OWNER
// status: purchased, enrolled, registered, verified, purchased, expired
type Voucher struct {
	VoucherID string `json:"voucher_id"`
	AcademyName string `json:"academy_name"`
	VaildDate string 	`json:"vaild_date"`
	Owner string `json:"owner"`
	Status string `json:"status"`
}

// history 결과 저장 구조체
type HistoryQueryResult struct {
	Record    *Voucher    `json:"record"`
	TxId     string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

//수강권 생성(등록자, 수강권이름, 학원이름, 수강권정보)

func (s *SmartContract) EnrollLecture(ctx contractapi.TransactionContextInterface, vid string, academy string, valid string, owner string) error {
	lecture := Voucher{
		VoucherID:   vid,
		AcademyName:  academy,
		VaildDate:	valid,
		Owner:	owner,
		Status: "registered",
	}
	
	lectureBytes, _ := json.Marshal(lecture)
	
	return ctx.GetStub().PutState(vid, lectureBytes)
}

//정보 가져오기
func (s *SmartContract) GetVoucher(ctx contractapi.TransactionContextInterface, vid string) (*Voucher, error) {
	lectureBytes, err := ctx.GetStub().GetState(vid)
	
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}
	
	if lectureBytes == nil {
		return nil, fmt.Errorf("%s does not exist", vid)
	}
	
	lecture := new(Voucher)
	_ = json.Unmarshal(lectureBytes, lecture)
	
	return lecture, nil
}



//판매등록
func (s *SmartContract) SellVoucher(ctx contractapi.TransactionContextInterface, vid string, owner string) error {
	lecture, err := s.GetVoucher(ctx, vid)
	
	if err != nil {
		return err
	}
	if lecture.Owner  != owner {
		return fmt.Errorf("not vaild owner")
	}
	lecture.Status = "enrolled"

	lectureBytes, _ := json.Marshal(lecture)
	
	return ctx.GetStub().PutState(vid, lectureBytes)
}





//판매의 학원 검증
func (s *SmartContract) VerifyVoucher(ctx contractapi.TransactionContextInterface, vid string, academy string) error {
	lecture, err := s.GetVoucher(ctx, vid)
	
	if err != nil {
		return err
	}
	if lecture.AcademyName  != academy {
		return fmt.Errorf("not vaild academy")
	}
	lecture.Status = "verified"

	lectureBytes, _ := json.Marshal(lecture)
	
	return ctx.GetStub().PutState(vid, lectureBytes)
}


//구매(강의이름,판매자,구매자)
func (s *SmartContract) PurchaseVoucher(ctx contractapi.TransactionContextInterface, vid string, oldowner string, newowner string) error {
	
	lecture, err := s.GetVoucher(ctx, vid)
	
	if err != nil {
		return err
	}
	
	if lecture.Status != "verified" {
		return fmt.Errorf("voucher is not verified ")
	}
	if lecture.Status == "expired" {
		return fmt.Errorf("voucher is expired ")
	}
	if lecture.Owner != oldowner {
		return fmt.Errorf("voucher is not belong to  "+oldowner)
	}

	lecture.Status = "purchased"
	lecture.Owner = newowner
	
	lectureBytes, _ := json.Marshal(lecture)
	ctx.GetStub().PutState(vid, lectureBytes)
	
	return nil
}

// Interface : 수강권 판매등록, 판매검증, 구매, 사용/무효화

//무효화

func (s *SmartContract) ExpireVoucher(ctx contractapi.TransactionContextInterface, vid string) error {
	lecture, err := s.GetVoucher(ctx, vid)

	if err != nil {
		return err
	}
	
	lecture.Status = "expired"

	lectureBytes, _ := json.Marshal(lecture)
	
	ctx.GetStub().PutState(vid, lectureBytes)
	
	return nil
}

//기록 검색
func (t *SmartContract) GetHistory(ctx contractapi.TransactionContextInterface, key string) ([]HistoryQueryResult, error) {
	log.Printf("GetHistory: ID %v", key)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(key)
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

		var asset Voucher
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, err
			}
		} else {
			asset = Voucher{
				VoucherID: key,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &asset,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}


func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create simpleasset chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting simpleasset chaincode: %s", err.Error())
	}
}

