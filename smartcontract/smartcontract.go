package smartcontract

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	smartcontracterrors "github.com/zacharyfrederick/admin/types/errors"
)

type AdminContract struct {
	contractapi.Contract
}

type SmartContractContext contractapi.TransactionContextInterface

type Modeler interface {
	ToJSON() ([]byte, error)
	FromJSON([]byte) error
	GetID() string
}

func SaveState(ctx contractapi.TransactionContextInterface, m Modeler) error {
	modelJSON, err := m.ToJSON()
	if err != nil {
		return smartcontracterrors.SaveStateError
	}
	return ctx.GetStub().PutState(m.GetID(), modelJSON)
}

func LoadState(data []byte, m Modeler) error {
	err := m.FromJSON(data)
	if err != nil {
		return smartcontracterrors.LoadStateError
	}
	return nil
}
