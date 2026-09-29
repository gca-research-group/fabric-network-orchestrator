package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func (s *SmartContract) CreateProduct(ctx contractapi.TransactionContextInterface,
	id string,
	name string,
	description string,
	price float64,
) error {

	exists, err := s.ProductExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("product %s already exists", id)
	}

	product := Product{
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
	}

	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, productJSON)
}

func (s *SmartContract) QueryProductByID(ctx contractapi.TransactionContextInterface,
	id string,
) (*Product, error) {

	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if productJSON == nil {
		return nil, fmt.Errorf("product %s does not exist", id)
	}

	var product Product
	err = json.Unmarshal(productJSON, &product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *SmartContract) ListAllProducts(ctx contractapi.TransactionContextInterface) ([]*Product, error) {

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var products []*Product

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var product Product
		err = json.Unmarshal(queryResponse.Value, &product)
		if err != nil {
			return nil, err
		}

		products = append(products, &product)
	}

	return products, nil
}

func (s *SmartContract) UpdateProduct(ctx contractapi.TransactionContextInterface,
	id string,
	name string,
	description string,
	price float64,
) error {

	exists, err := s.ProductExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("product %s does not exist", id)
	}

	product := Product{
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
	}

	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, productJSON)
}

func (s *SmartContract) DeleteProduct(ctx contractapi.TransactionContextInterface,
	id string,
) error {

	exists, err := s.ProductExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("product %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

func (s *SmartContract) ProductExists(ctx contractapi.TransactionContextInterface,
	id string,
) (bool, error) {

	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}

	return productJSON != nil, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		panic(fmt.Sprintf("Error creating chaincode: %v", err))
	}

	if err := chaincode.Start(); err != nil {
		panic(fmt.Sprintf("Error starting chaincode: %v", err))
	}
}
