package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/open-quantum-safe/liboqs-go/oqs"
)

type Escrow struct{}

func GetCreator(stub shim.ChaincodeStubInterface) string {
	creatorByte, _ := stub.GetCreator()
	fmt.Println(string(creatorByte))
	certStart := bytes.IndexAny(creatorByte, "-----BEGIN")
	if certStart == -1 {
		fmt.Println("No certificate found")
		return ""
	}
	certText := creatorByte[certStart:]
	bl, _ := pem.Decode(certText)
	if bl == nil {
		fmt.Println("Could not decode the PEM structure")
		return ""
	}

	cert, err := x509.ParseCertificate(bl.Bytes)
	if err != nil {
		fmt.Println("ParseCertificate failed")
		return ""
	}
	uname := cert.Subject.CommonName
	return uname
}

func AES_Decrypt(encryptedDate []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte(""), err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	privDate := make([]byte, len(encryptedDate))
	blockMode.CryptBlocks(privDate, encryptedDate)
	privDate = PKCS7UnPadding(privDate)
	return privDate, nil
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func (t *Escrow) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success([]byte("Success invoke and not opter!!"))
}

func (t *Escrow) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fn, args := stub.GetFunctionAndParameters()
	if fn == "Gen_EA_KeyPair" {
		return t.Gen_EA_KeyPair(stub, args)
	} else if fn == "Get_EA_PubKey" {
		return t.Get_EA_PubKey(stub, args)
	} else if fn == "Decap_Shared_Sec" {
		return t.Decap_Shared_Sec(stub, args)
	} else if fn == "Dec_Sec_Data" {
		return t.Dec_Sec_Data(stub, args)
	}

	return shim.Error("Recevied unkown function invocation")
}

func (t *Escrow) Gen_EA_KeyPair(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	t0 := time.Now()
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	// verify identity of the escrower
	creator := GetCreator(stub)
	if creator == "" {
		return shim.Error("The operator of Getcreator is failed!!!")
	}
	if (creator != "Admin@org0.example.com") && (creator != "Admin@org2.example.com") {
		return shim.Error("The escrower doesn't have authority, so it can not modify the information!---creator:" + creator)
	}

	// We use generateKeyPair function to generate publicKey and privateKey about escrower.
	EsPKKeyid := args[0]
	kemName := args[1]
	t1 := time.Now()
	escrow := oqs.KeyEncapsulation{}
	defer escrow.Clean()
	if err := escrow.Init(kemName, nil); err != nil {
		return shim.Error(err.Error())
	}
	publicKey, err := escrow.GenerateKeyPair()
	if err != nil {
		return shim.Error(err.Error())
	}
	elapsed1 := time.Since(t1)
	genKeyTime := strconv.FormatFloat(elapsed1.Seconds(), 'E', -1, 64)

	privateKey := escrow.ExportSecretKey()
	strPriKey := base64.StdEncoding.EncodeToString(privateKey)

	err = stub.PutState(EsPKKeyid, publicKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	elapsed := time.Since(t0)
	allTime := strconv.FormatFloat(elapsed.Seconds(), 'E', -1, 64)
	return shim.Success([]byte("Put public key successfully!!!-------AllTime:" + allTime + "------GeneKeyTime:" + genKeyTime + "------privateKey:" + strPriKey))
}

func (t *Escrow) Get_EA_PubKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	EsPKKeyid := args[0]
	publicKeyBytes, err := stub.GetState(EsPKKeyid)
	if err != nil {
		fmt.Println("Get state error")
		return shim.Error(err.Error())
	}
	return shim.Success(publicKeyBytes)
}

func (t *Escrow) Decap_Shared_Sec(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	t0 := time.Now()
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
	// verify identity of the escrower
	creator := GetCreator(stub)
	if creator == "" {
		return shim.Error("The operator of Getcreator is failed!!!")
	}
	if (creator != "Admin@org0.example.com") && (creator != "Admin@org2.example.com") {
		return shim.Error("The escrower doesn't have authority, so it can not modify the information!---creator:" + creator)
	}

	senderKeyid := args[0]
	EsSSKeyid := args[1]
	id, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error(err.Error())
	}
	kemName := args[3]
	collection := args[4]
	tMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not retrieve transient, err %s", err))
	}
	privateKeyBytes, ok := tMap["PRIVATEKEY"]
	if !ok {
		return shim.Error(fmt.Sprintf("Expected transient KS"))
	}

	queryArgs := [][]byte{[]byte("Get_Sender_Data"), []byte(senderKeyid)}
	response := stub.InvokeChaincode("Sender", queryArgs, "mychannel")
	if response.Status != shim.OK {
		return shim.Error(fmt.Sprintf("failed to query chaincode.got error :%s", response.Payload))
	}
	enConBytes := bytes.Split(response.Payload, []byte("-----"))

	// We use DecapSecret function to decrypt sharedSecret.
	t1 := time.Now()
	escrow := oqs.KeyEncapsulation{}
	defer escrow.Clean()
	if err := escrow.Init(kemName, privateKeyBytes); err != nil {
		return shim.Error(err.Error())
	}
	CT := enConBytes[id]
	sharedSecret, err := escrow.DecapSecret(CT)
	if err != nil {
		return shim.Error(err.Error())
	}
	elapsed1 := time.Since(t1)
	decapSecretTime := strconv.FormatFloat(elapsed1.Seconds(), 'E', -1, 64)

	err = stub.PutPrivateData(collection, EsSSKeyid, sharedSecret)
	if err != nil {
		return shim.Error(err.Error())
	}
	elapsed := time.Since(t0)
	alltime := strconv.FormatFloat(elapsed.Seconds(), 'E', -1, 64)
	return shim.Success([]byte("The escrower has successfully decaped sharedsecret!!!------ALLtime" + alltime + "-------DecapSecretTime:" + decapSecretTime))
}

func (t *Escrow) Dec_Sec_Data(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	t0 := time.Now()
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
	// verify identity of the listener
	creator := GetCreator(stub)
	if creator == "" {
		return shim.Error("The operator of Getcreator is failed!")
	}
	if creator != "Admin@org3.example.com" {
		return shim.Error("The listener doesn't have authority, so it can not modify the information!---creator:" + creator)
	}

	senderKeyid := args[0]
	EsSSKeyid1 := args[1]
	EsSSKeyid2 := args[2]
	collection1 := args[3]
	collection2 := args[4]
	sharedSecret1, err := stub.GetPrivateData(collection1, EsSSKeyid1)
	if err != nil {
		return shim.Error(err.Error())
	}
	sharedSecret2, err := stub.GetPrivateData(collection2, EsSSKeyid2)
	if err != nil {
		return shim.Error(err.Error())
	}
	sharedSecret := make([]byte, len(sharedSecret1))
	for i := 0; i < len(sharedSecret); i++ {
		sharedSecret[i] = sharedSecret1[i] ^ sharedSecret2[i]
	}
	if len(sharedSecret) >= 64 {
		hash := sha256.New()
		hash.Write(sharedSecret)
		sharedSecret = hash.Sum(nil)
	}

	queryArgs := [][]byte{[]byte("Get_Sender_Data"), []byte(senderKeyid)}
	response := stub.InvokeChaincode("Sender", queryArgs, "mychannel")
	if response.Status != shim.OK {
		return shim.Error(fmt.Sprintf("failed to query chaincode.got error :%s", response.Payload))
	}
	enConBytes := bytes.Split(response.Payload, []byte("-----"))
	messageBytes, err := AES_Decrypt(enConBytes[3], sharedSecret)
	if err != nil {
		return shim.Error(err.Error())
	}
	//fmt.Println(string(messageBytes))
	elapsed := time.Since(t0)
	runtime := strconv.FormatFloat(elapsed.Seconds(), 'E', -1, 64)
	return shim.Success([]byte("The listener successfully receive the message!!!---MESSAGE:" + string(messageBytes) + "----runtime:" + runtime))
}

func main() {
	err1 := shim.Start(new(Escrow))
	if err1 != nil {
		fmt.Printf("error starting simple chaincode:%s \n", err1)
	}
}
