package utils

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	obj, err := ctx.GetStub().GetState(id)

	if err != nil {
		return false, err
	}

	return obj != nil, nil
}

func GetAssetType(ctx contractapi.TransactionContextInterface, id string) (string, error) {
	obj, err := ctx.GetStub().GetState(id)

	if err != nil {
		return "", err
	}

	if obj == nil {
		return "", fmt.Errorf("an object with that id does not exist")
	}

	data := make(map[string]string)

	conversionErr := json.Unmarshal(obj, &data)
	if conversionErr != nil {
		return "", err
	}

	docType, ok := data["docType"]

	if !ok {
		return "", fmt.Errorf("the docType could not be found")
	}

	return string(docType), nil
}
