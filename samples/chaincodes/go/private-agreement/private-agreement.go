package main

import (
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type PrivateAgreement struct {
	ID        string `json:"id"`
	PartyA    string `json:"partyA"`
	PartyB    string `json:"partyB"`
	SecretKey string `json:"secretKey"`
}

func (s *SmartContract) CreatePrivateAgreement(
	ctx contractapi.TransactionContextInterface,
	id string) error {

	transientData, err := ctx.GetStub().GetTransient()
	if err != nil {
		return err
	}

	dataJSON := transientData["agreement"]
	if dataJSON == nil {
		return fmt.Errorf("agreement not found in transient map")
	}

	return ctx.GetStub().PutPrivateData(
		"collectionAgreements",
		id,
		dataJSON,
	)
}

func (s *SmartContract) ReadPrivateAgreement(
	ctx contractapi.TransactionContextInterface,
	id string) ([]byte, error) {

	return ctx.GetStub().GetPrivateData(
		"collectionAgreements",
		id,
	)
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		log.Panic(err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panic(err)
	}
}
