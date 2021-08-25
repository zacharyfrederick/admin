package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"

	"github.com/zacharyfrederick/admin/smartcontract"
)

func main() {
	adminContractContract := new(smartcontract.AdminContract)
	adminContractContract.Info.Version = "0.0.1"
	adminContractContract.Info.Description = "My Smart Contract"
	adminContractContract.Info.License = new(metadata.LicenseMetadata)
	adminContractContract.Info.License.Name = "Apache-2.0"
	adminContractContract.Info.Contact = new(metadata.ContactMetadata)
	adminContractContract.Info.Contact.Name = "Zachary Frederick"

	chaincode, err := contractapi.NewChaincode(adminContractContract)
	chaincode.Info.Title = "admin_contract chaincode"
	chaincode.Info.Version = "0.0.1"

	if err != nil {
		panic("Could not create chaincode from AdminContract" + err.Error())
	}

	err = chaincode.Start()

	if err != nil {
		panic("Failed to start chaincode. " + err.Error())
	}
}
