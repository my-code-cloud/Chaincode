package main

import (
  "encoding/json"
  "fmt"
  "log"
	
  "github.com/hyperledger/fabric-contract-api-go/contractapi"
  // "github.com/hyperledger/fabric-chaincode-go/shim"
  // "github.com/hyperledger/fabric-protos-go/peer"
)

const samplesetCollection = "sampleCollection"

// Contract Packate 가져오기
type SmartContract struct {
  contractapi.Contract
}

// 샘플로 사용 할 구조체 정의
type Sample struct {
  ID    string `json:"ID"`
  Name	string `json:"name"`
  Money	int    `json:"money"`
}

// PDC 샘플 구조체
type SamplePrivateData struct {
  ID          string `json:"ID"`
  SampleValue int `json:"sampleValue"`
}

// func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

//     fn, args := stub.GetFunctionAndParameters()

//     if fn == "test" {
//       fmt.Println("test")
//     }
//     fmt.Println(fn)
//     fmt.Println(args)

//     return shim.Success(nil)
// }

// 초기화(선택 : 사용 시 -cci 옵션 사용 ex, ./network.sh deployCC -cci InitLedger)
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
  samples := []Sample{
    {ID: "id0001", Name: "김오토", Money: 10},
    {ID: "id0002", Name: "박크리", Money: 10},
    {ID: "id0003", Name: "황에버", Money: 10},
  }

  for _, sample := range samples {
    sampleJSON, err := json.Marshal(sample)
    if err != nil {
      return err
    }

    err = ctx.GetStub().PutState(sample.ID, sampleJSON)
    if err != nil {
      return fmt.Errorf("failed to put to world state. %v", err)
    }
  }

  return nil
}

// 샘플 생성
func (s *SmartContract) CreateSample(ctx contractapi.TransactionContextInterface, id string, name string, money int) error {

  exists, err := s.SampleExists(ctx, id)
  if err != nil {
    return err
  }

  if exists {
    return fmt.Errorf("the sample %s already exists", id)
  }

  sample := Sample{
    ID:     id,
    Name:   name,
    Money:  money,
  }

  sampleJSON, err := json.Marshal(sample)
  if err != nil {
    return err
  }

  return ctx.GetStub().PutState(id, sampleJSON)
}

//   // // PDC 데이터 생성
//   // func (s *SmartContract) CreatePrivateSample(ctx contractapi.TransactionContextInterface) error {

//   //   // transient map 가져오기 : pdc 데이터를 담는 field
//   //   transientMap, err := ctx.GetStub().GetTransient()
//   //   if err != nil {
//   //     return fmt.Errorf("error getting transient: %v", err)
//   //   }

//   //   // pdc 데이터가 존재하는지 확인
//   //   transientSampleJSON, ok := transientMap["sample_properties"]
//   //   if !ok {
//   //     return fmt.Errorf("sample not fount in the transient map input")
//   //   }

//   //   type sampleTrasientInput struct {
//   //     ID          string  `json:"id"`
//   //     Name        string  `json:"name"`
//   //     Money       int     `json:money`
//   //     SampleValue int     `json:sampleValue`
//   //   }

//   //   var sampleInput sampleTrasientInput
//   //   err = json.Unmarshal(transientSampleJSON, &sampleInput)
//   //   if err != nil {
//   //     return fmt.Errorf("failed to unmarshal JSONl: "%v"", err)
//   //   }
//   //   if len(sampleInput.ID) == 0 {
//   //     return fmt.Errorf("ID field must be a non-emptyj string")
//   //   }
//   //   if len(sampleInput.Name) == 0 {
//   //     return fmt.Errorf("ID field must be a non-emptyj string")
//   //   }
//   //   if len(sampleInput.Money) <= 0 {
//   //     return fmt.Errorf("ID field must be a non-emptyj string")
//   //   }
//   //   if len(sampleInput.SampleValue) <= 0 {
//   //     return fmt.Errorf("ID field must be a non-emptyj string")
//   //   }

// //   // Check if sample already exists
// //   sampleAsBytes, err := ctx.GetStub().GetPrivateData(sampleCollection, sampleInput.ID)
// //   if err != nil {
// //     return fmt.Errorf("failed to get sample: %v", err)
// //   } else if sampleAsBytes != nil {
// //     fmt.Println("Sample already existsL " + sampleInput.ID)
// //     return fmt.Errorf("this sample already exists " + sampleInput.ID)
// //   }
// // }

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadSample(ctx contractapi.TransactionContextInterface, id string) (*Sample, error) {
  sampleJSON, err := ctx.GetStub().GetState(id)
  if err != nil {
    return nil, fmt.Errorf("failed to read from world state: %v", err)
  }
  if sampleJSON == nil {
    return nil, fmt.Errorf("the sample %s does not exist", id)
  }

  var sample Sample
  err = json.Unmarshal(sampleJSON, &sample)
  if err != nil {
    return nil, err
  }

  return &sample, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateSample(ctx contractapi.TransactionContextInterface, id string, name string, money int) error {
  exists, err := s.SampleExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the sample %s does not exist", id)
  }

  // overwriting original asset with new asset
  sample := Sample{
    ID:       id,
    Name:     name,
    Money:	money,
  }
  sampleJSON, err := json.Marshal(sample)
  if err != nil {
    return err
  }

  return ctx.GetStub().PutState(id, sampleJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteSample(ctx contractapi.TransactionContextInterface, id string) error {
  exists, err := s.SampleExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the sample %s does not exist", id)
  }

  return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) SampleExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
  sampleJSON, err := ctx.GetStub().GetState(id)
  if err != nil {
    return false, fmt.Errorf("failed to read from world state: %v", err)
  }

  return sampleJSON != nil, nil
}

// 전체 키 조회
func (s *SmartContract) GetAllSamples(ctx contractapi.TransactionContextInterface) ([]*Sample, error) {
// var buffer bytes.Buffer

  resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
  if err != nil {
    return nil, err
  }
  defer resultsIterator.Close()

  var samples []*Sample
  for resultsIterator.HasNext() {
    queryResponse, err := resultsIterator.Next()
    if err != nil {
      return nil, err
    }

    var sample Sample
    err = json.Unmarshal(queryResponse.Value, &sample)
    if err != nil {
      return nil, err
    }

    samples = append(samples, &sample)
  }

  fmt.Println(samples)
  return samples, nil
}

func main() {
  sampleChaincode, err := contractapi.NewChaincode(&SmartContract{})
  if err != nil {
    log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
  }

  if err := sampleChaincode.Start(); err != nil {
    log.Panicf("Error starting sample-pdc chaincode: %v", err)
  }
}