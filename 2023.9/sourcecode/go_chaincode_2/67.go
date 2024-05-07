package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"

)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}

// Product :  Define the Product structure
type Product struct {
	ProductName  string `json:"productname"`
	ProductClass  string `json:"productclass"`
	Producer string `json:"producer"`
	ProductionDate  string `json:"productiondate"`
	ProducerCheckoutDate  string `json:"producercheckoutdate"`
	Transporter  string `json:"transporter"`
	TransporterEntryDate  string `json:"transporterentrydate"`
	TransporterCheckoutDate  string `json:"transportercheckoutdate"`
	Warehouse string `json:"warehouse"`
	WarehouseEntryDate string `json:"warehouseentrydate"`
	WarehouseCheckoutDate string `json:"warehousecheckoutdate"`
	Status  string `json:"status"`
}

type ProductPrivateDetails struct {
	ProductId   string `json:"productid"`
	Status  string `json:"status"`
}

type Sensor struct {
	IotName  string `json:"iotname"`
	IotSensorId  string `json:"iotsensorid"`
	Temperature  string `json:"temperature"`
	Humidity string `json:"humidity"`
	Pressure  string `json:"pressure"`
	TimeStamp  string `json:"timestamp"`
}

// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("fabcar_cc")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "queryProduct":
		return s.queryProduct(APIstub, args)
	case "createProduct":
		return s.createProduct(APIstub, args)
	case "queryAllProduct":
		return s.queryAllProduct(APIstub)
	case "changeProductStatus":
		return s.changeProductStatus(APIstub, args)
	case "getHistoryForAsset":
		return s.getHistoryForAsset(APIstub, args)
	case "queryProductByStatus":
		return s.queryProductByStatus(APIstub, args)
	case "createSensorData":
		return s.createSensorData(APIstub, args)
	case "changeSensorData":
		return s.changeSensorData(APIstub, args)
	case "querySensorData":
		return s.querySensorData(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}

	// return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryProduct(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	productAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(productAsBytes)
}

func (s *SmartContract) querySensorData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	sensorAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(sensorAsBytes)
}

func (s *SmartContract) createSensorData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
	
	dt:= time.Now()
	
	var sensor = Sensor{IotName: args[0], Temperature: args[1], Humidity: args[2], Pressure: args[3]}
	
	sensor.IotSensorId=args[4]
	sensor.TimeStamp=dt.Format("01-02-2006 15:04:05")
	
	sensorAsBytes, _ := json.Marshal(sensor)
	APIstub.PutState(args[0], sensorAsBytes)
	
	indexName := "status~key"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{sensor.IotSensorId, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)
	
	return shim.Success(sensorAsBytes)
	
}

func (s *SmartContract) changeSensorData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	dt := time.Now()

	sensorAsBytes, _ := APIstub.GetState(args[0])
	sensor := Sensor{}

	json.Unmarshal(sensorAsBytes, &sensor)
	
	sensor.Temperature= args[1]
	sensor.Humidity=args[2]
	sensor.Pressure=args[3]
	sensor.TimeStamp=dt.Format("01-02-2006 15:04:05")

	sensorAsBytes, _ = json.Marshal(sensor)
	APIstub.PutState(args[0], sensorAsBytes)

	return shim.Success(sensorAsBytes)
}

func (s *SmartContract) createProduct(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	dt:= time.Now()
	var product = Product{ProductName: args[1], ProductClass: args[2], Producer: args[3]}
	product.Status= "Uretildi"
	product.ProductionDate=dt.Format("01-02-2006 15:04:05")
	product.ProducerCheckoutDate="n/a"
	product.Transporter="n/a" 
	product.TransporterEntryDate="n/a"
	product.TransporterCheckoutDate="n/a"
	product.Warehouse="n/a"
	product.WarehouseEntryDate="n/a"
	product.WarehouseCheckoutDate= "n/a"

	productAsBytes, _ := json.Marshal(product)
	APIstub.PutState(args[0], productAsBytes)

	indexName := "status~key"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{product.Status, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)

	return shim.Success(productAsBytes)
}

func (S *SmartContract) queryProductByStatus(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	status := args[0]

	statusAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("status~key", []string{status})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer statusAndIdResultIterator.Close()

	var i int
	var id string

	var products []byte
	bArrayMemberAlreadyWritten := false

	products = append([]byte("["))

	for i = 0; statusAndIdResultIterator.HasNext(); i++ {
		responseRange, err := statusAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositeKeyParts[1]
		assetAsBytes, err := APIstub.GetState(id)

		if bArrayMemberAlreadyWritten == true {
			newBytes := append([]byte(","), assetAsBytes...)
			products = append(products, newBytes...)

		} else {
			// newBytes := append([]byte(","), carsAsBytes...)
			products = append(products, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	products = append(products, []byte("]")...)

	return shim.Success(products)
}

func (s *SmartContract) queryAllProduct(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "PRODUCT0"
	endKey := "PRODUCT999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllProduct:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) changeProductStatus(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	dt := time.Now()

	productAsBytes, _ := APIstub.GetState(args[0])
	product := Product{}

	json.Unmarshal(productAsBytes, &product)

	if product.Status == "Uretildi" {
		product.ProducerCheckoutDate=dt.Format("01-02-2006 15:04:05")
		product.Transporter= args[1]
		product.Status= "Nakliyede"
		product.TransporterEntryDate=dt.Format("01-02-2006 15:04:05")

	  } else if product.Status == "Nakliyede" {

		product.Warehouse= args[1]
		product.TransporterCheckoutDate=dt.Format("01-02-2006 15:04:05")
		product.WarehouseEntryDate=dt.Format("01-02-2006 15:04:05")
		product.Status= "Depoda"
	  }

	productAsBytes, _ = json.Marshal(product)
	APIstub.PutState(args[0], productAsBytes)

	return shim.Success(productAsBytes)
}

func (t *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	productName := args[0]

	resultsIterator, err := stub.GetHistoryForKey(productName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")

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

	fmt.Printf("- getHistoryForAsset returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}