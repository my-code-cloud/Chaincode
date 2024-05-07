package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type KittyContract struct {
	contractapi.Contract
}

type Kitty struct {
	Genes         uint64    `json:"genes"`
	BirthTime     time.Time `json:"birth_time"`
	CooldownEnd   time.Time `json:"cooldown_end"`
	MatronID      uint64    `json:"matron_id"`
	SireID        uint64    `json:"sire_id"`
	SiringWithID  uint64    `json:"siring_with_id"`
	CooldownIndex uint8     `json:"cooldown_index"`
	Generation    uint64    `json:"generation"`
}

type KittyList []Kitty

var cooldowns = []time.Duration{
	1 * time.Second,
	2 * time.Second,
	5 * time.Second,
	10 * time.Second,
	30 * time.Second,
	1 * time.Minute,
	2 * time.Minute,
	4 * time.Minute,
	8 * time.Minute,
	16 * time.Minute,
	1 * time.Hour,
	2 * time.Hour,
	4 * time.Hour,
	7 * time.Hour,
}

var kitties = KittyList{Kitty{}}

const kittiesNAME = "kitties"

var kittyIndexToOwner = []string{""}

const kittyIndexToOwnerNAME = "kittyIndexToOwner"

var kittyIndexToApproved = []string{""}

const kittyIndexToApprovedNAME = "kittyIndexToApproved"

var sireAllowedToAddress = []string{""}

const sireAllowedToAddressNAME = "kittyIndexToAddress"

var pregnantKitties uint64 = 0

const pregnantKittiesNAME = "pregnantKitties"

var g_event = map[string]interface{}{}

func transfer(ctx contractapi.TransactionContextInterface, from, to string, kittyID uint64) error {
	kittyIndexToOwner[kittyID] = to

	kittyIndexToApproved[kittyID] = ""
	sireAllowedToAddress[kittyID] = ""

	payload := map[string]interface{}{"from": from, "to": to, "kittyID": kittyID}
	g_event["Transfer"] = payload

	return nil
}

func createKitty(ctx contractapi.TransactionContextInterface, matronID, sireID, generation, genes uint64, owner string) (uint64, error) {
	cooldownIndex := uint8(generation / 2)
	if cooldownIndex > 13 {
		cooldownIndex = 13
	}

	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return 0, err
	}
	now, err := ptypes.Timestamp(txTimestamp)
	if err != nil {
		return 0, err
	}

	kitty := Kitty{
		Genes:         genes,
		BirthTime:     now,
		CooldownEnd:   now.Add(cooldowns[cooldownIndex]),
		MatronID:      matronID,
		SireID:        sireID,
		SiringWithID:  0,
		CooldownIndex: cooldownIndex,
		Generation:    generation,
	}
	kitties = append(kitties, kitty)
	newKittenID := uint64(len(kitties) - 1)

	kittyIndexToOwner = append(kittyIndexToOwner, "")
	kittyIndexToApproved = append(kittyIndexToApproved, "")
	sireAllowedToAddress = append(sireAllowedToAddress, "")

	payload := map[string]interface{}{"owner": owner, "newKittenID": newKittenID, "matronID": matronID, "sireID": sireID, "genes": genes}
	g_event["Birth"] = payload

	if err := transfer(ctx, "", owner, newKittenID); err != nil {
		return 0, err
	}

	return newKittenID, nil
}

func owns(kittyID uint64, owner string) (bool, error) {
	return kittyIndexToOwner[kittyID] == owner, nil
}

func approvedFor(kittyID uint64, account string) bool {
	return kittyIndexToApproved[kittyID] == account
}

func approve(kittyID uint64, account string) error {
	kittyIndexToApproved[kittyID] = account
	return nil
}

func (c *KittyContract) Transfer(ctx contractapi.TransactionContextInterface, to string, kittyID uint64) error {
	from, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}
	from = getClientID(from)
	isOwner, err := owns(kittyID, from)
	if err != nil {
		return err
	}
	if isOwner {
		return transfer(ctx, from, to, kittyID)
	}
	return fmt.Errorf("Transfer initiated not by owner. Caller: %s, Owner: %s", from, kittyIndexToOwner[kittyID])
}

func (c *KittyContract) Approve(ctx contractapi.TransactionContextInterface, kittyID uint64, account string) error {
	userID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}
	userID = getClientID(userID)
	isOwner, err := owns(kittyID, userID)
	if err != nil {
		return err
	}
	if isOwner {
		return approve(kittyID, account)
	}
	return fmt.Errorf("Approve initiated not by owner. Caller: %s, Owner: %s", userID, kittyIndexToOwner[kittyID])
}

func (c *KittyContract) TotalSupply(ctx contractapi.TransactionContextInterface) (uint64, error) {
	return uint64(len(kittyIndexToOwner)) - 1, nil
}

func (c *KittyContract) OwnerOf(ctx contractapi.TransactionContextInterface, kittyID uint64) (string, error) {
	if int(kittyID) >= len(kittyIndexToOwner) {
		return "", fmt.Errorf("Index too large. KittID %d not known.", kittyID)
	}
	owner := kittyIndexToOwner[kittyID]
	return owner, nil
}

func (c *KittyContract) TokensOfOwner(ctx contractapi.TransactionContextInterface, owner string) ([]uint64, error) {
	var supply []uint64 = []uint64{}
	for k, v := range kittyIndexToOwner {
		if v == owner {
			supply = append(supply, uint64(k))
		}
	}
	return supply, nil
}

func (c *KittyContract) PregnantKitties(ctx contractapi.TransactionContextInterface) (uint64, error) {
	return pregnantKitties, nil
}

func isReadyToGiveBirth(ctx contractapi.TransactionContextInterface, matron *Kitty) (bool, error) {
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return false, err
	}
	now, err := ptypes.Timestamp(txTimestamp)
	if err != nil {
		return false, err
	}

	return (matron.SiringWithID != 0) && (matron.CooldownEnd.Before(now)), nil
}

func isReadyToBreed(ctx contractapi.TransactionContextInterface, kitty Kitty) (bool, error) {
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return false, err
	}
	now, err := ptypes.Timestamp(txTimestamp)
	if err != nil {
		return false, err
	}

	return kitty.SiringWithID == 0 && kitty.CooldownEnd.Before(now), nil
}

func (c *KittyContract) IsReadyToBreed(ctx contractapi.TransactionContextInterface, kittyID uint64) (bool, error) {
	if !(kittyID > 0) || int(kittyID) >= len(kitties) {
		return false, fmt.Errorf("No kitty with this ID %d available.", kittyID)
	}

	kitty := kitties[kittyID]

	return isReadyToBreed(ctx, kitty)
}

func isSiringPermitted(matronID, sireID uint64) (bool, error) {
	matronOwner := kittyIndexToOwner[matronID]
	sireOwner := kittyIndexToOwner[sireID]

	return matronOwner == sireOwner || sireAllowedToAddress[sireID] == matronOwner, nil
}

func triggerCooldown(ctx contractapi.TransactionContextInterface, kittyID uint64) error {
	if int(kittyID) >= len(kitties) {
		return fmt.Errorf("No kitty with this ID %d available.", kittyID)
	}

	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}
	now, err := ptypes.Timestamp(txTimestamp)
	if err != nil {
		return err
	}

	cooldownIndex := kitties[kittyID].CooldownIndex

	kitties[kittyID].CooldownEnd = now.Add(cooldowns[cooldownIndex])

	if cooldownIndex < 13 {
		kitties[kittyID].CooldownIndex = cooldownIndex + 1
	}

	return nil
}

func (c *KittyContract) ApproveSiring(ctx contractapi.TransactionContextInterface, kittyID uint64, siringPartner string) error {
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	clientID = getClientID(clientID)

	ok, err := owns(kittyID, clientID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Caller must be the owner of the matron kitty.")
	}

	sireAllowedToAddress[kittyID] = siringPartner

	return nil
}

func (c *KittyContract) IsPregnant(ctx contractapi.TransactionContextInterface, kittyID uint64) (bool, error) {
	if !(kittyID > 0) || int(kittyID) >= len(kitties) {
		return false, fmt.Errorf("No kitty with this ID %d available.", kittyID)
	}

	kitty := kitties[kittyID]

	return kitty.SiringWithID != 0, nil
}

func isValidMatingPair(sireID, matronID uint64) (bool, error) {
	if int(matronID) >= len(kitties) {
		return false, fmt.Errorf("No kitty with this ID %d available.", matronID)
	}

	if int(sireID) >= len(kitties) {
		return false, fmt.Errorf("No kitty with this ID %d available.", sireID)
	}

	if sireID == matronID {
		return false, nil
	}

	matron := kitties[matronID]
	sire := kitties[sireID]

	if matron.MatronID == sireID || matron.SireID == sireID {
		return false, nil
	}
	if sire.MatronID == matronID || sire.SireID == matronID {
		return false, nil
	}

	if sire.Generation == 0 || matron.Generation == 0 {
		return true, nil
	}

	if sire.MatronID == matron.MatronID || sire.MatronID == matron.SireID {
		return false, nil
	}
	if sire.SireID == matron.MatronID || sire.SireID == matron.SireID {
		return false, nil
	}

	return true, nil
}

func (c *KittyContract) CanBreedWith(ctx contractapi.TransactionContextInterface, sireID, matronID uint64) (bool, error) {
	if int(matronID) >= len(kitties) || matronID == 0 {
		return false, fmt.Errorf("No matron with this ID %d available.", matronID)
	}

	if int(sireID) >= len(kitties) || sireID == 0 {
		return false, fmt.Errorf("No sire with this ID %d available.", sireID)
	}

	ok, err := isValidMatingPair(matronID, sireID)
	if err != nil || !ok {
		return false, err
	}

	ok, err = isSiringPermitted(matronID, sireID)
	if err != nil || !ok {
		return false, err
	}

	return true, nil
}

func breedWith(ctx contractapi.TransactionContextInterface, sireID, matronID uint64) error {
	if int(matronID) >= len(kitties) {
		return fmt.Errorf("No kitty with this ID %d available.", matronID)
	}

	if int(sireID) >= len(kitties) {
		return fmt.Errorf("No kitty with this ID %d available.", sireID)
	}

	matron := &kitties[matronID]

	matron.SiringWithID = sireID
	triggerCooldown(ctx, matronID)
	triggerCooldown(ctx, sireID)

	sireAllowedToAddress[sireID] = ""
	sireAllowedToAddress[matronID] = ""

	pregnantKitties++

	owner := kittyIndexToOwner[matronID]

	g_event["Pregnant"] = map[string]interface{}{"owner": owner, "matronID": matronID, "sireID": sireID, "matronCooldown": matron.CooldownEnd}

	return nil
}

func (c *KittyContract) BreedWithAuto(ctx contractapi.TransactionContextInterface, sireID, matronID uint64) error {
	if int(matronID) >= len(kitties) {
		return fmt.Errorf("No kitty with this ID %d available.", matronID)
	}

	if int(sireID) >= len(kitties) {
		return fmt.Errorf("No kitty with this ID %d available.", sireID)
	}

	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}

	clientID = getClientID(clientID)

	matronOwner := kittyIndexToOwner[matronID]
	matron := kitties[matronID]
	sire := kitties[sireID]

	if matronOwner != clientID {
		return fmt.Errorf("Caller must be the owner of the matron kitty. Caller: %s, Owner: %s", clientID, matronOwner)
	}

	if ok, err := isSiringPermitted(matronID, sireID); err != nil || !ok {
		return fmt.Errorf("Siring is not permitted for marton %d and sire %d.", matronID, sireID)
	}

	if ok, err := isReadyToBreed(ctx, matron); err != nil || !ok {
		return fmt.Errorf("Provided marton with id %d is not ready to breed.", matronID)
	}

	if ok, err := isReadyToBreed(ctx, sire); err != nil || !ok {
		return fmt.Errorf("Provided sire with id %d is not ready to breed.", sireID)
	}

	if ok, err := isValidMatingPair(sireID, matronID); err != nil || !ok {
		return fmt.Errorf("Matron with id %d and sire with id %d are no valid mating pair. Shame on you! Don't try to breed these cats.", matronID, sireID)
	}

	return breedWith(ctx, sireID, matronID)
}

func mixGenes(matronGenes, sireGenes uint64) uint64 {
	return matronGenes ^ sireGenes // ^ will implement xor
}

func (c *KittyContract) GiveBirth(ctx contractapi.TransactionContextInterface, matronID uint64) (uint64, error) {
	if int(matronID) >= len(kitties) {
		return 0, fmt.Errorf("No kitty with this ID available.")
	}

	matron := &kitties[matronID]

	if ok, err := isReadyToGiveBirth(ctx, matron); err != nil || !ok {
		return 0, fmt.Errorf("Matron is not yet ready to give birth.")
	}

	sireID := matron.SiringWithID
	sire := kitties[sireID]

	parentGeneration := matron.Generation
	if parentGeneration < sire.Generation {
		parentGeneration = sire.Generation
	}

	childGenes := mixGenes(matron.Genes, sire.Genes)
	owner := kittyIndexToOwner[matronID]
	kittyID, err := createKitty(ctx, matronID, sireID, parentGeneration+1, childGenes, owner)
	if err != nil {
		return 0, err
	}

	matron.SiringWithID = 0

	pregnantKitties--

	return kittyID, nil
}

func (c *KittyContract) TransferFrom(ctx contractapi.TransactionContextInterface, from, to string, kittyID uint64) error {
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return err
	}
	clientID = getClientID(clientID)

	if approvedFor(kittyID, clientID) {
		return transfer(ctx, from, to, kittyID)
	}

	return fmt.Errorf("Caller is not approved. Caller: %s", clientID)
}

func (c *KittyContract) CreateKitty(ctx contractapi.TransactionContextInterface, matronID, sireID, generation, genes uint64, owner string) error {
	_, err := createKitty(ctx, matronID, sireID, generation, genes, owner)
	return err
}

func (c *KittyContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	log.Println("Entering InitLedger: Doing essentially nothing...")
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	ktct := KittyContract{}
	ktct.BeforeTransaction = BeforeTransaction
	ktct.AfterTransaction = AfterTransaction
	assetChaincode, err := contractapi.NewChaincode(&ktct)
	if err != nil {
		log.Panicf("Error creating HyperKitty chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting HyperKitty chaincode: %v", err)
	}
}
