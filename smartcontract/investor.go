package smartcontract

import (
	"github.com/zacharyfrederick/admin/types"
	smartcontracterrors "github.com/zacharyfrederick/admin/types/errors"
	"github.com/zacharyfrederick/admin/utils"
)

func (s *AdminContract) CreateInvestor(
	ctx SmartContractContext,
	investorId string,
	name string,
) error {
	objExists, err := utils.AssetExists(ctx, investorId)
	if err != nil {
		return smartcontracterrors.ReadingWorldStateError
	}
	if objExists {
		return smartcontracterrors.IdAlreadyInUseError
	}
	investor := types.CreateDefaultInvestor(investorId, name)
	return SaveState(ctx, &investor)
}

func (s *AdminContract) QueryInvestorById(
	ctx SmartContractContext,
	investorId string,
) (*types.Investor, error) {
	investorJson, err := ctx.GetStub().GetState(investorId)
	if err != nil {
		return nil, smartcontracterrors.ReadingWorldStateError
	}
	if investorJson == nil {
		return nil, nil
	}
	var investor types.Investor
	err = LoadState(investorJson, &investor)
	if err != nil {
		return nil, err
	}
	return &investor, nil
}
