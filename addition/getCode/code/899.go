package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
)

// SmartContract :
type SmartContract struct {
}

// Product :
type Product struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	Stock       int     `json:"stock"`
	Owner       string  `json:"owner"`
	BatchNumber int     `json:"batchnumber"`
	Qrcode      string  `json:"Qrcode"`
	Trace       string  `json:"trace"`
}

type productPrivateDetails struct {
	Owner string `json:"owner"`
	Trace string `json:"trace"`
}

// Transaction :
type Transaction struct {
	CreatedAt    string  `json:"created_at"`
	From         string  `json:"from"`
	To           string  `json:"to"`
	Product      string  `json:"product"`
	Stock        int     `json:"stock"`
	Payment      float64 `json:"payment"`
	Organisation string  `json:"organisation"`
	Picked       int     `json:"picked"`
}

type transactionPrivateDetails struct {
	From    string  `json:"from"`
	Payment float64 `json:"payment"`
}

// Init : function
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("pickngo_cc")

// Invoke : fucntion
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "queryProduct":
		return s.queryProduct(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	case "createProduct":
		return s.createProduct(APIstub, args)
	case "queryAllProducts":
		return s.queryAllProducts(APIstub)
	case "changeProductOwner":
		return s.changeProductOwner(APIstub, args)
	case "getHistoryForAsset":
		return s.getHistoryForAsset(APIstub, args)
	case "queryProductsByOwner":
		return s.queryProductsByOwner(APIstub, args)
	case "restictedMethod":
		return s.restictedMethod(APIstub, args)
	case "test":
		return s.test(APIstub, args)
	case "createPrivateProduct":
		return s.createPrivateProduct(APIstub, args)
	case "readPrivateProduct":
		return s.readPrivateProduct(APIstub, args)
	case "updatePrivateData":
		return s.updatePrivateData(APIstub, args)
	case "readProductPrivateDetails":
		return s.readProductPrivateDetails(APIstub, args)
	case "createPrivateProductImplicitForOrg1":
		return s.createPrivateProductImplicitForOrg1(APIstub, args)
	case "createPrivateProductImplicitForOrg2":
		return s.createPrivateProductImplicitForOrg2(APIstub, args)
	case "queryPrivateDataHash":
		return s.queryPrivateDataHash(APIstub, args)
	case "createTransactionAction":
		return s.createTransactionAction(APIstub, args)
	case "queryTransaction":
		return s.queryTransaction(APIstub, args)
	case "queryAllTransactions":
		return s.queryAllTransactions(APIstub)
	case "updateTransactionPickedData":
		return s.updateTransactionPickedData(APIstub, args)
	case "queryPrivateTransactionDataHash":
		return s.queryPrivateTransactionDataHash(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}

	// return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryProduct(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	ProductAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(ProductAsBytes)
}

func (s *SmartContract) readPrivateProduct(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	// collectionProducts, collectionProductPrivateDetails, _implicit_org_Org1MSP, _implicit_org_Org2MSP
	ProductAsBytes, err := APIstub.GetPrivateData(args[0], args[1])
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[1] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if ProductAsBytes == nil {
		jsonResp := "{\"Error\":\"Product private details does not exist: " + args[1] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(ProductAsBytes)
}

func (s *SmartContract) readPrivateProductIMpleciteForOrg1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	ProductAsBytes, _ := APIstub.GetPrivateData("_implicit_org_Org1MSP", args[0])
	return shim.Success(ProductAsBytes)
}
func (s *SmartContract) readProductPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	productAsBytes, err := APIstub.GetPrivateData("collectionproductPrivateDetails", args[0])

	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[0] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if productAsBytes == nil {
		jsonResp := "{\"Error\":\"Marble private details does not exist: " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(productAsBytes)
}

func (s *SmartContract) test(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	ProductAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(ProductAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	Products := []Product{
		Product{Name: "Coconuts oil", Category: "Prius", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 401, Qrcode: "None", Owner: "Tomoko"},
		Product{Name: "Unga", Category: "Mustang", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 402, Qrcode: "None", Owner: "Brad"},
		Product{Name: "Omo", Category: "Tucson", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 403, Qrcode: "None", Owner: "Jin Soo"},
		Product{Name: "Harpic", Category: "Passat", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 404, Qrcode: "None", Owner: "Max"},
		Product{Name: "Yoghurt", Category: "S", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 405, Qrcode: "None", Owner: "Adriana"},
		Product{Name: "Milk", Category: "205", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 406, Qrcode: "None", Owner: "Michel"},
		Product{Name: "Kiwi", Category: "S22L", Price: 13.56, Image: "https://picsum.photos/id/237/200/300", Stock: 13, Trace: "inStore", BatchNumber: 407, Qrcode: "None", Owner: "Aarav"},
	}

	i := 0
	for i < len(Products) {
		ProductAsBytes, _ := json.Marshal(Products[i])
		APIstub.PutState("Product"+strconv.Itoa(i), ProductAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createPrivateProduct(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type productTransientInput struct {
		Name        string `json:"name"`
		Category    string `json:"category"`
		Price       string `json:"price"`
		Image       string `json:"image"`
		Stock       string `json:"stock"`
		Owner       string `json:"owner"`
		Trace       string `json:"trace"`
		BatchNumber string `json:"batchnumber"`
		Qrcode      string `json:"Qrcode"`
		Key         string `json:"key"`
	}
	if len(args) != 0 {
		return shim.Error("1111111----Incorrect number of arguments. Private marble data must be passed in transient map.")
	}

	logger.Infof("11111111111111111111111111")

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}

	productDataAsBytes, ok := transMap["product"]
	if !ok {
		return shim.Error("product must be a key in the transient map")
	}
	logger.Infof("********************8   " + string(productDataAsBytes))

	if len(productDataAsBytes) == 0 {
		return shim.Error("333333 -marble value in the transient map must be a non-empty JSON string")
	}

	logger.Infof("2222222")

	var productInput productTransientInput
	err = json.Unmarshal(productDataAsBytes, &productInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(productDataAsBytes) + "Error is : " + err.Error())
	}

	logger.Infof("3333")

	if len(productInput.Key) == 0 {
		return shim.Error("key field must be a non-empty string")
	}
	if len(productInput.Name) == 0 {
		return shim.Error("Name field must be a non-empty string")
	}
	if len(productInput.Category) == 0 {
		return shim.Error("category field must be a non-empty string")
	}

	pricecheck, _ := strconv.ParseFloat(productInput.Price, 64)
	if pricecheck <= 0 {
		return shim.Error("price field must be a non-empty float")
	}
	if len(productInput.Owner) == 0 {
		return shim.Error("owner field must be a non-empty string")
	}
	if len(productInput.Image) == 0 {
		return shim.Error("Image field must be a non-empty string")
	}
	stockcheck, _ := strconv.Atoi(productInput.Stock)
	if stockcheck <= 0 {
		return shim.Error("stock field must be a non-empty int")
	}
	if len(productInput.Trace) == 0 {
		return shim.Error("Trace field must be a non-empty string")
	}

	logger.Infof("444444")

	// ==== Check if product already exists ====
	productAsBytes, err := APIstub.GetPrivateData("collectionproducts", productInput.Key)
	if err != nil {
		return shim.Error("Failed to get marble: " + err.Error())
	} else if productAsBytes != nil {
		fmt.Println("This product already exists: " + productInput.Key)
		return shim.Error("This product already exists: " + productInput.Key)
	}

	logger.Infof("55555")

	price, _ := strconv.ParseFloat(productInput.Price, 64)
	stock, _ := strconv.Atoi(productInput.Stock)
	batchnumber, _ := strconv.Atoi(productInput.BatchNumber)

	var product = Product{Name: productInput.Name, Category: productInput.Category, Price: price, Image: productInput.Image, Stock: stock, Trace: productInput.Trace, BatchNumber: batchnumber, Qrcode: productInput.Qrcode, Owner: productInput.Owner}

	productAsBytes, err = json.Marshal(product)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.PutPrivateData("collectionproducts", productInput.Key, productAsBytes)
	if err != nil {
		logger.Infof("6666666")
		return shim.Error(err.Error())
	}

	productPrivateDetails := &productPrivateDetails{Owner: productInput.Owner, Trace: productInput.Trace}

	productPrivateDetailsAsBytes, err := json.Marshal(productPrivateDetails)
	if err != nil {
		logger.Infof("77777")
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("collectionproductPrivateDetails", productInput.Key, productPrivateDetailsAsBytes)
	if err != nil {
		logger.Infof("888888")
		return shim.Error(err.Error())
	}

	return shim.Success(productAsBytes)
}

func (s *SmartContract) updatePrivateData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	type productTransientInput struct {
		Owner string `json:"owner"`
		Trace string `json:"trace"`
		Key   string `json:"key"`
	}
	if len(args) != 0 {
		return shim.Error("1111111----Incorrect number of arguments. Private marble data must be passed in transient map.")
	}

	logger.Infof("11111111111111111111111111")

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}

	productDataAsBytes, ok := transMap["product"]
	if !ok {
		return shim.Error("product must be a key in the transient map")
	}
	logger.Infof("********************8   " + string(productDataAsBytes))

	if len(productDataAsBytes) == 0 {
		return shim.Error("333333 -marble value in the transient map must be a non-empty JSON string")
	}

	logger.Infof("2222222")

	var productInput productTransientInput
	err = json.Unmarshal(productDataAsBytes, &productInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(productDataAsBytes) + "Error is : " + err.Error())
	}

	productPrivateDetails := &productPrivateDetails{Owner: productInput.Owner, Trace: productInput.Trace}

	productPrivateDetailsAsBytes, err := json.Marshal(productPrivateDetails)
	if err != nil {
		logger.Infof("77777")
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("collectionproductPrivateDetails", productInput.Key, productPrivateDetailsAsBytes)
	if err != nil {
		logger.Infof("888888")
		return shim.Error(err.Error())
	}

	return shim.Success(productPrivateDetailsAsBytes)

}

//create function for products
func (s *SmartContract) createProduct(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
	price, _ := strconv.ParseFloat(args[3], 64)
	stock, _ := strconv.Atoi(args[5])
	batchnumber, _ := strconv.Atoi(args[7])
	var product = Product{Name: args[1], Category: args[2], Price: price, Image: args[4], Stock: stock, Owner: args[6], BatchNumber: batchnumber, Qrcode: args[8], Trace: args[9]}

	productAsBytes, _ := json.Marshal(product)
	APIstub.PutState(args[0], productAsBytes)

	indexName := "owner~key"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{product.Owner, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)

	return shim.Success(productAsBytes)
}

func (s *SmartContract) queryProductsByOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	owner := args[0]

	ownerAndIDResultIterator, err := APIstub.GetStateByPartialCompositeKey("owner~key", []string{owner})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer ownerAndIDResultIterator.Close()

	var i int
	var id string

	var products []byte
	bArrayMemberAlreadyWritten := false

	products = append([]byte("["))

	for i = 0; ownerAndIDResultIterator.HasNext(); i++ {
		responseRange, err := ownerAndIDResultIterator.Next()
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
			// newBytes := append([]byte(","), productsAsBytes...)
			products = append(products, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : %s ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	products = append(products, []byte("]")...)

	return shim.Success(products)
}

func (s *SmartContract) queryAllProducts(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "Product0"
	endKey := "Product999"

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

	fmt.Printf("- queryAllProducts:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) restictedMethod(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "approver" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as APPROVER have access this method!")
	} else {
		if len(args) != 1 {
			return shim.Error("Incorrect number of arguments. Expecting 1")
		}

		productAsBytes, _ := APIstub.GetState(args[0])
		return shim.Success(productAsBytes)
	}

}

func (s *SmartContract) changeProductOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	productAsBytes, _ := APIstub.GetState(args[0])
	product := Product{}

	json.Unmarshal(productAsBytes, &product)
	product.Owner = args[1]

	productAsBytes, _ = json.Marshal(product)
	APIstub.PutState(args[0], productAsBytes)

	return shim.Success(productAsBytes)
}

func (s *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	productName := args[0]

	resultsIterator, err := stub.GetHistoryForKey(productName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
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

	fmt.Printf("- getHistoryForAsset returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) createPrivateProductImplicitForOrg1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 10 {
		return shim.Error("Incorrect arguments. Expecting 10 arguments")
	}

	price, _ := strconv.ParseFloat(args[3], 64)
	stock, _ := strconv.Atoi(args[5])
	batchnumber, _ := strconv.Atoi(args[7])
	var product = Product{Name: args[1], Category: args[2], Price: price, Image: args[4], Stock: stock, Owner: args[6], BatchNumber: batchnumber, Qrcode: args[8], Trace: args[9]}

	productAsBytes, _ := json.Marshal(product)
	// APIstub.PutState(args[0], productAsBytes)

	err := APIstub.PutPrivateData("_implicit_org_Org1MSP", args[0], productAsBytes)
	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	return shim.Success(productAsBytes)
}

func (s *SmartContract) createPrivateProductImplicitForOrg2(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 10 {
		return shim.Error("Incorrect arguments. Expecting 10 arguments")
	}

	price, _ := strconv.ParseFloat(args[3], 64)
	stock, _ := strconv.Atoi(args[5])
	batchnumber, _ := strconv.Atoi(args[7])
	var product = Product{Name: args[1], Category: args[2], Price: price, Image: args[4], Stock: stock, Owner: args[6], BatchNumber: batchnumber, Qrcode: args[8], Trace: args[9]}

	productAsBytes, _ := json.Marshal(product)
	APIstub.PutState(args[0], productAsBytes)

	err := APIstub.PutPrivateData("_implicit_org_Org2MSP", args[0], productAsBytes)
	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	return shim.Success(productAsBytes)
}

func (s *SmartContract) createTransactionAction(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 9 {
		return shim.Error("Incorrect arguments. Expecting 9 arguments")
	}
	payment, _ := strconv.ParseFloat(args[6], 64)
	stock, _ := strconv.Atoi(args[5])
	var transaction = Transaction{CreatedAt: args[1], From: args[2], To: args[3], Product: args[4], Stock: stock, Payment: payment, Organisation: args[7], Picked: 0}
	transactionAsBytes, _ := json.Marshal(transaction)
	APIstub.PutState(args[0], transactionAsBytes)

	indexName := "key~owner~Organisation"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{args[0], transaction.To, args[7]})
	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)

	return shim.Success(transactionAsBytes)
}

func (s *SmartContract) updateTransactionPickedData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	transactionAsBytes, _ := APIstub.GetState(args[0])
	transaction := Transaction{}
	json.Unmarshal(transactionAsBytes, &transaction)
	productAsBytes, _ := APIstub.GetState(args[1])
	product := Product{}

	json.Unmarshal(productAsBytes, &product)
	topick, _ := strconv.Atoi(args[2])
	var amount = transaction.Picked + topick
	if amount <= transaction.Stock {
		if product.Stock > topick {
			product.Stock = product.Stock - topick
			productAsBytes, _ = json.Marshal(product)
			fmt.Println(productAsBytes)
			APIstub.PutState(args[1], productAsBytes)
			transaction.Picked = amount
			transactionAsBytes, _ = json.Marshal(transaction)
			fmt.Println(transactionAsBytes)
			APIstub.PutState(args[0], transactionAsBytes)
			return shim.Success(transactionAsBytes)
		}
	}
	return shim.Error("Failed to Update Transaction")
}

func (s *SmartContract) queryAllTransactions(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "Transaction0"
	endKey := "Transaction999"

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

	fmt.Printf("- queryAllTransactions:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryTransaction(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	transactionAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(transactionAsBytes)
}

func (s *SmartContract) queryPrivateDataHash(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	productAsBytes, _ := APIstub.GetPrivateDataHash(args[0], args[1])
	return shim.Success(productAsBytes)
}

func (s *SmartContract) queryPrivateTransactionDataHash(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	transactionAsBytes, _ := APIstub.GetPrivateDataHash(args[0], args[1])
	return shim.Success(transactionAsBytes)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
