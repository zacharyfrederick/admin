package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/zacharyfrederick/admin/types"
)

func (w *EndpointWrapper) PostInvestorEndpoint(c *gin.Context) {
	var createInvestorRequest types.CreateInvestorRequest

	err := c.BindJSON(&createInvestorRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
		return
	}

	validRequest := types.ValidateCreateInvestorRequest(&createInvestorRequest)
	if !validRequest {
		c.JSON(http.StatusBadRequest, gin.H{"error": "improperly formatted request"})
		return
	}

	investorId := uuid.NewV4().String()

	result, err := w.Contract.SubmitTransaction("CreateInvestor", investorId, createInvestorRequest.Name)
	if err != nil {
		errorString := fmt.Sprintf("error submitting request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorString})
		fmt.Println(result)
		return
	}

	c.JSON(http.StatusOK, gin.H{"investorId": investorId})

}

func (w *EndpointWrapper) GetInvestorByIdEndpoint(c *gin.Context) {
	investorId := c.Param("id")
	result, err := w.Contract.EvaluateTransaction("QueryInvestorById", investorId)

	if err != nil {
		errorString := fmt.Sprintf("error evaluating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorString})
		fmt.Println(result)
	}

	if len(result) == 0 {
		c.JSON(http.StatusOK, "")
		return
	}

	var investor types.Investor
	jsonErr := json.Unmarshal(result, &investor)

	if jsonErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error unmarshaling json"})
		return
	}
	c.JSON(http.StatusOK, investor)
}
