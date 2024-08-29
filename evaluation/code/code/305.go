package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SimplePaymentApp SmartContract provides functions for managing an Asset
type SimplePaymentApp struct {
	contractapi.Contract
}

//Account struct
type Account struct {
	DocType         string
	AccountName     string
	AccountID       string
	AccountType     string
	AccountNo       int
	CurrentBalance  int
	Kyc             bool
	ListTransaction []string
}

//Auth Structure
type Auth struct {
	DocType     string
	UserName    string
	Password    string
	AccountType string
	Bank        string
	AccountNo   string
}

//Transactions struct
type Transactions struct {
	DocType       string
	TransactionID string
	FromID        string
	ToID          string
	Amount        string
}

//AccountStatement Struct
type AccountStatement struct {
	Balance   int       `json:"balance"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// UserPrivateDetails : Captures Account Users Personal Information
type UserPrivateDetails struct {
	DocType string `json:"doctype"`
	UserID  string `json:"userID"`
	Name    string `json:"Name"`
	Address string `json:"address"`
	Sex     string `json:"sex"`
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

//StringWithCharset will return a random string of given length and charset
func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

//String returns a random strng of defined length
func String(length int) string {
	return StringWithCharset(length, charset)
}

//RegisterAuth will register a new user for login
func (s *SimplePaymentApp) RegisterAuth(ctx contractapi.TransactionContextInterface) (*Auth, error) {
	// Handling the private Data of the user
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return nil, fmt.Errorf("Error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["login"]
	if !ok {
		return nil, fmt.Errorf("User details not found in the transient map input")
	}
	type accountUserTransientInput struct {
		UserName string `json:"username"`
		Password string `json:"password"`
		Bank     string `json:"bank"`
	}

	var accountUserInput accountUserTransientInput
	err = json.Unmarshal(transientAssetJSON, &accountUserInput)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(accountUserInput.UserName) == 0 {
		return nil, fmt.Errorf("Name field must be a non-empty string")
	}
	if len(accountUserInput.Password) == 0 {
		return nil, fmt.Errorf("Password must be a non-empty string")
	}
	if len(accountUserInput.Bank) == 0 {
		return nil, fmt.Errorf("Bank field must be a non-empty string")
	}

	id := accountUserInput.UserName

	collectionName, _ := getCollectionName(ctx)

	// Check if user already exists
	userAsBytes, err := ctx.GetStub().GetPrivateData(collectionName, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get user info: %v", err)
	}
	if userAsBytes != nil {
		fmt.Println("User already exists ")
		return nil, fmt.Errorf("This User already exists %v", id)
	}

	userInfo := Auth{
		DocType:     "Auth",
		UserName:    accountUserInput.UserName,
		Password:    accountUserInput.Password,
		AccountType: "user",
		Bank:        accountUserInput.Bank,
		AccountNo:   "0",
	}

	userInfoBytes, err := json.Marshal(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall user info: %v", err)
	}

	err = ctx.GetStub().PutPrivateData(collectionName, id, userInfoBytes)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

//Login returns after validating the auth
func (s *SimplePaymentApp) Login(ctx contractapi.TransactionContextInterface, userName string, password string) (*Auth, error) {
	collection, err := getCollectionName(ctx)
	userAsBytes, err := ctx.GetStub().GetPrivateData(collection, userName)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %v", err)
	}

	if userAsBytes == nil {
		return nil, fmt.Errorf("failed to get user infor %v", err)
	}

	var userPrivateDetails Auth
	err = json.Unmarshal(userAsBytes, &userPrivateDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return &userPrivateDetails, nil
}

// InitLedger adds a base set of assets to the ledger
func (s *SimplePaymentApp) InitLedger(ctx contractapi.TransactionContextInterface) error {
	listTran1 := []string{}
	listTran2 := []string{}
	listTran3 := []string{}

	accounts := []Account{
		{DocType: "Account", AccountName: "sanath", AccountID: "1", AccountType: "savings", AccountNo: 1111, CurrentBalance: 90000, Kyc: true, ListTransaction: listTran1},
		{DocType: "Account", AccountName: "devyansh", AccountID: "2", AccountType: "salary", AccountNo: 1112, CurrentBalance: 50000, Kyc: true, ListTransaction: listTran2},
		{DocType: "Account", AccountName: "shreyansh", AccountID: "3", AccountType: "salary", AccountNo: 1113, CurrentBalance: 60000, Kyc: true, ListTransaction: listTran3},
	}

	for _, account := range accounts {
		accountJSON, err := json.Marshal(account)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(account.AccountID, accountJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}
	return nil

}

//InitAuth will create dummy accounts
func (s *SimplePaymentApp) InitAuth(ctx contractapi.TransactionContextInterface) error {

	accounts := []Auth{
		{DocType: "Auth", UserName: "admin", Password: "adminpw", AccountType: "operator", Bank: "HDFC", AccountNo: "0"},
		{DocType: "Auth", UserName: "sanath", Password: "pass", AccountType: "user", Bank: "HDFC", AccountNo: "1"},
	}
	collectionName, _ := getCollectionName(ctx)

	for _, account := range accounts {
		accountJSON, err := json.Marshal(account)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutPrivateData(collectionName, account.UserName, accountJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}
	return nil

}

// InitTransactions adds a base set of assets to the ledger
func (s *SimplePaymentApp) InitTransactions(ctx contractapi.TransactionContextInterface) error {

	transactions := []Transactions{
		{DocType: "transaction", TransactionID: "s9f39hdqf", FromID: "1111", ToID: "1112", Amount: "100"},
		{DocType: "transaction", TransactionID: "2974hhd89", FromID: "1112", ToID: "1113", Amount: "6969"},
		{DocType: "transaction", TransactionID: "3jf93j0g8", FromID: "1113", ToID: "1111", Amount: "420"},
	}

	for _, transaction := range transactions {
		transactionJSON, err := json.Marshal(transaction)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(transaction.TransactionID, transactionJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}
	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SimplePaymentApp) CreateAsset(ctx contractapi.TransactionContextInterface, name string, id string, atype string, ano int, currbalance int, kyc bool, listtransaction []string, increment int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	account := Account{
		AccountName:     name,
		AccountID:       id,
		AccountType:     atype,
		AccountNo:       ano,
		CurrentBalance:  currbalance,
		Kyc:             kyc,
		ListTransaction: listtransaction,
	}
	accountJSON, err := json.Marshal(account)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(id, accountJSON)
}

//ReadUserDetails : Read the user private details
func (s *SimplePaymentApp) ReadUserDetails(ctx contractapi.TransactionContextInterface, userID string) (*UserPrivateDetails, error) {

	collectionName, _ := getCollectionName(ctx)

	// Check if user already exists
	userAsBytes, err := ctx.GetStub().GetPrivateData(collectionName, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %v", err)
	}

	if userAsBytes == nil {
		return nil, fmt.Errorf("failed to get user info %v", err)
	}

	var userPrivateDetails UserPrivateDetails
	err = json.Unmarshal(userAsBytes, &userPrivateDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return &userPrivateDetails, nil
}

// RegisterAcc issues a new asset to the world state with given details.
func (s *SimplePaymentApp) RegisterAcc(ctx contractapi.TransactionContextInterface, id string, Name string) (*Account, error) {
	// Handling the private Data of the user
	accountBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to read Account Info from world state: %v", err)
	}

	if accountBytes != nil {
		return nil, fmt.Errorf("The User account already exists for user")
	}

	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("the asset %s already exists", id)
	}
	iid, err := strconv.Atoi(id)
	account := Account{
		DocType:         "Account",
		AccountName:     Name,
		AccountID:       id,
		AccountType:     "Savings",
		AccountNo:       iid,
		CurrentBalance:  1000,
		Kyc:             true,
		ListTransaction: []string{},
	}
	accountJSON, err := json.Marshal(account)
	if err != nil {
		return nil, err
	}

	err = ctx.GetStub().PutState(id, accountJSON)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// Register issues a new asset to the world state with given details.
func (s *SimplePaymentApp) Register(ctx contractapi.TransactionContextInterface) (*UserPrivateDetails, error) {
	// Handling the private Data of the user
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return nil, fmt.Errorf("Error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		return nil, fmt.Errorf("User details not found in the transient map input")
	}
	type accountUserTransientInput struct {
		Name    string `json:"name"`
		ID      string `json:"ID"`
		Address string `json:"address"`
		Sex     string `json:"sex"`
	}

	var accountUserInput accountUserTransientInput
	err = json.Unmarshal(transientAssetJSON, &accountUserInput)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(accountUserInput.Name) == 0 {
		return nil, fmt.Errorf("Name field must be a non-empty string")
	}
	if len(accountUserInput.ID) == 0 {
		return nil, fmt.Errorf("id must be a non-empty string")
	}
	if len(accountUserInput.Address) == 0 {
		return nil, fmt.Errorf("Address field must be a non-empty string")
	}
	if len(accountUserInput.Sex) == 0 {
		return nil, fmt.Errorf("Sex field must be a non-empty string")
	}

	id := accountUserInput.ID

	collectionName, _ := getCollectionName(ctx)

	// Check if user already exists
	userAsBytes, err := ctx.GetStub().GetPrivateData(collectionName, id)
	if err != nil {
		return nil, fmt.Errorf("Failed to get user info: %v", err)
	}
	if userAsBytes != nil {
		fmt.Println("User already exists ")
		return nil, fmt.Errorf("This User already exists %v", id)
	}

	userInfo := UserPrivateDetails{
		DocType: "UserPrivateDetails",
		UserID:  id,
		Name:    accountUserInput.Name,
		Address: accountUserInput.Address,
		Sex:     accountUserInput.Sex,
	}

	userInfoBytes, err := json.Marshal(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall user info: %v", err)
	}

	err = ctx.GetStub().PutPrivateData(collectionName, id, userInfoBytes)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// Summary returns the asset stored in the world state with given id.
func (s *SimplePaymentApp) Summary(ctx contractapi.TransactionContextInterface, id string) (*Account, error) {
	accountJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if accountJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var account Account
	err = json.Unmarshal(accountJSON, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// Balance returns the asset stored in the world state with given id.
func (s *SimplePaymentApp) Balance(ctx contractapi.TransactionContextInterface, id string) (int, error) {
	accountJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}
	if accountJSON == nil {
		return 0, fmt.Errorf("the asset %s does not exist", id)
	}

	var account Account
	err = json.Unmarshal(accountJSON, &account)
	if err != nil {
		return 0, err
	}

	return account.CurrentBalance, nil
}

// AddAmount updates an existing asset in the world state with provided parameters.
func (s *SimplePaymentApp) AddAmount(ctx contractapi.TransactionContextInterface, toid string, amount int) error {
	exists, err := s.AssetExists(ctx, toid)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", toid)
	}

	account, _ := s.Summary(ctx, toid)

	account.CurrentBalance = account.CurrentBalance + amount
	accountJSON, err := json.Marshal(account)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(toid, accountJSON)
}

// Transfer updates an existing asset in the world state with provided parameters.
func (s *SimplePaymentApp) Transfer(ctx contractapi.TransactionContextInterface, fromid string, toid string, amount string) error {
	exists, err := s.AssetExists(ctx, toid)
	exists2, err2 := s.AssetExists(ctx, fromid)
	if err != nil || err2 != nil {
		if err != nil {
			return err
		}
		return err2
	}
	if !exists || !exists2 {
		if !exists {
			return fmt.Errorf("the asset %s does not exist", toid)
		}
		if !exists2 {
			return fmt.Errorf("the asset %s does not exist", fromid)
		}
	}

	//Fetching Data
	accountFrom, _ := s.Summary(ctx, fromid)
	accountTo, _ := s.Summary(ctx, toid)

	if !accountFrom.Kyc {
		return fmt.Errorf("The Sender's acount is not KYC enabled")
	}
	amt, err := strconv.Atoi(amount)
	//Check if ammount is sufficient If not then return the message that insuffient ammount
	if accountFrom.CurrentBalance < amt {
		return fmt.Errorf("Insufficient Balance in account ID: %s, The amout is %d", fromid, accountFrom.CurrentBalance)
	}
	//else transfer the amout
	accountFrom.CurrentBalance = accountFrom.CurrentBalance - amt
	accountTo.CurrentBalance = accountTo.CurrentBalance + amt

	transactionID := fmt.Sprintf("%s%s%d%d", accountFrom.AccountID, accountTo.AccountID, accountFrom.CurrentBalance, accountTo.CurrentBalance)
	accountFrom.ListTransaction = append(accountFrom.ListTransaction, transactionID)
	transaction := Transactions{
		DocType:       "transaction",
		TransactionID: transactionID,
		FromID:        accountFrom.AccountID,
		ToID:          accountTo.AccountID,
		Amount:        amount,
	}
	transactionJSON, err := json.Marshal(transaction)
	if err != nil {
		return err
	}

	fmt.Println(accountFrom, accountTo)

	accountFromJSON, err := json.Marshal(accountFrom)
	accountToJSON, err2 := json.Marshal(accountTo)

	if err != nil || err2 != nil {
		if err != nil {
			return err
		}
		return err2
	}
	puterr := ctx.GetStub().PutState(fromid, accountFromJSON)
	puterr2 := ctx.GetStub().PutState(toid, accountToJSON)
	putter3 := ctx.GetStub().PutState(transactionID, transactionJSON)
	if puterr != nil || puterr2 != nil || putter3 != nil {
		if puterr != nil {
			return fmt.Errorf("error in putting Fromaccount data")
		}
		return fmt.Errorf("error in putting Toaccount data")
	}
	return nil
	//return ctx.GetStub().PutState(fromid, accountFromJSON), ctx.GetStub().PutState(toid, accountToJSON)
	//return fmt.Errorf("amount has been transfered check the accounts")
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SimplePaymentApp) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	accountJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return accountJSON != nil, nil
}

//GetAllAssets returns all the assests from the blockchain
func (s *SimplePaymentApp) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Account, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var accounts []*Account
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var account Account
		err = json.Unmarshal(queryResponse.Value, &account)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, &account)
	}

	return accounts, nil
}

//UpdateKYC will update the kyc
func (s *SimplePaymentApp) UpdateKYC(ctx contractapi.TransactionContextInterface, id string) (string, error) {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("the asset %s does not exist", id)
	}

	account, _ := s.Summary(ctx, id)

	account.Kyc = true
	accountJSON, err := json.Marshal(account)
	if err != nil {
		return "", err
	}
	err = ctx.GetStub().PutState(id, accountJSON)
	if err != nil {
		return "", err
	}
	return "Success", nil
}

// GetAccountStatement returns the chain of custody for an asset since issuance.
func (s *SimplePaymentApp) GetAccountStatement(ctx contractapi.TransactionContextInterface, id string) ([]AccountStatement, error) {

	log.Printf("GetAccountStatement: ID %v", id)
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var records []AccountStatement
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var account Account
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &account)
			if err != nil {
				return nil, err
			}
		} else {
			account = Account{
				AccountID: id,
			}
		}
		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}
		record := AccountStatement{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Balance:   account.CurrentBalance,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}
	return records, nil
}

// GetAllAccountsOwner returns the accounts of owner CouchDB Functions
func (s *SimplePaymentApp) GetAllAccountsOwner(ctx contractapi.TransactionContextInterface, owner string) ([]*Account, error) {
	queryString := fmt.Sprintf(`{"selector":{"DocType":"Account","AccountName":"%s"}}`, owner)
	return getQueryResultForQueryString(ctx, queryString)
}

// GetAllTransactions returns the accounts of owner CouchDB Functions
func (s *SimplePaymentApp) GetAllTransactions(ctx contractapi.TransactionContextInterface) ([]*Account, error) {
	queryString := fmt.Sprintf(`{"selector":{"DocType":"transaction"}}`)
	return getQueryResultForQueryString(ctx, queryString)
}

// GetAllAccounts returns the accounts of owner CouchDB Functions
func (s *SimplePaymentApp) GetAllAccounts(ctx contractapi.TransactionContextInterface) ([]*Account, error) {
	queryString := fmt.Sprintf(`{"selector":{"DocType":"Account"}}`)
	return getQueryResultForQueryString(ctx, queryString)
}

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Account, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// constructQueryResponseFromIterator constructs a slice of assets from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Account, error) {
	var assets []*Account
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset Account
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// getCollectionName is an internal helper function to get collection of submitting client identity.
func getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get the MSP ID of submitting client identity
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get verified MSPID: %v", err)
	}

	orgCollection := clientMSPID + "UserCollection"

	return orgCollection, nil
}

//ReadUserFromCollection : Read the user private details
func (s *SimplePaymentApp) ReadUserFromCollection(ctx contractapi.TransactionContextInterface, collection string, userID string) (*UserPrivateDetails, error) {

	// Check if user already exists
	userAsBytes, err := ctx.GetStub().GetPrivateData(collection, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %v", err)
	}

	if userAsBytes == nil {
		return nil, fmt.Errorf("failed to get user infor %v", err)
	}

	var userPrivateDetails UserPrivateDetails
	err = json.Unmarshal(userAsBytes, &userPrivateDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return &userPrivateDetails, nil
}
