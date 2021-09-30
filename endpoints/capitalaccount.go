package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/zacharyfrederick/admin/types"
)

func (w *EndpointWrapper) PostCapitalAccountEndpoint(c *gin.Context) {
	var createCapitalAccountRequest types.CreateCapitalAccountRequest

	err := c.BindJSON(&createCapitalAccountRequest)
	if err != nil {
		fmt.Printf("%v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
		return
	}

	validRequest := types.ValidateCreateCapitalAccountRequest(&createCapitalAccountRequest)
	if !validRequest {
		c.JSON(http.StatusBadRequest, gin.H{"error": "improperly formatted request"})
		return
	}

	capitalAccountId := uuid.NewV4().String()
	hasPerformanceFees := fmt.Sprintf("%t", createCapitalAccountRequest.HasPerformanceFees)
	result, err := w.Contract.SubmitTransaction("CreateCapitalAccount", capitalAccountId, createCapitalAccountRequest.Fund, createCapitalAccountRequest.Investor, hasPerformanceFees, createCapitalAccountRequest.PerformanceRate)
	if err != nil {
		errorString := fmt.Sprintf("error submitting request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorString})
		fmt.Println(result)
		return
	}

	c.JSON(http.StatusOK, gin.H{"capitalAccountId": capitalAccountId})
}

func (w *EndpointWrapper) GetCapitalAccountByIdEndpoint(c *gin.Context) {
	capitalAccountId := c.Param("id")
	result, err := w.Contract.EvaluateTransaction("QueryCapitalAccountById", capitalAccountId)

	if err != nil {
		errorString := fmt.Sprintf("error evaluating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorString})
		fmt.Println(result)
	}

	if len(result) == 0 {
		c.JSON(http.StatusOK, "")
		return
	}

	var capitalAccount types.CapitalAccount
	jsonErr := json.Unmarshal(result, &capitalAccount)

	if jsonErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error unmarshaling json"})
		return
	}
	c.JSON(http.StatusOK, capitalAccount)
}

func (w *EndpointWrapper) PostCapitalAccountActionEndpoint(c *gin.Context) {
	var createCapitalAccountActionRequest types.CreateCapitalAccountActionRequest

	err := c.BindJSON(&createCapitalAccountActionRequest)
	if err != nil {
		fmt.Printf("%v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
		return
	}

	validRequest := types.ValidateCreateCapitalAccountActionRequest(&createCapitalAccountActionRequest)
	if !validRequest {
		c.JSON(http.StatusBadRequest, gin.H{"error": "improperly formatted request"})
		return
	}

	transactionId := uuid.NewV4().String()
	full := fmt.Sprintf("%t", createCapitalAccountActionRequest.Full)
	period := fmt.Sprintf("%d", createCapitalAccountActionRequest.Period)

	result, err := w.Contract.SubmitTransaction("CreateCapitalAccountAction", transactionId, createCapitalAccountActionRequest.CapitalAccount, createCapitalAccountActionRequest.Type, createCapitalAccountActionRequest.Amount, full, createCapitalAccountActionRequest.Date, period)
	if err != nil {
		errorString := fmt.Sprintf("error submitting request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorString})
		fmt.Println(result)
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactionId": transactionId})
}

func (w *EndpointWrapper) GetCapitalAccountActionByIdEndpoint(c *gin.Context) {
	transactionId := c.Param("id")
	result, err := w.Contract.EvaluateTransaction("QueryCapitalAccountActionById", transactionId)

	if err != nil {
		errorString := fmt.Sprintf("error evaluating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorString})
		fmt.Println(result)
	}

	if len(result) == 0 {
		c.JSON(http.StatusOK, "")
		return
	}

	var capitalAccountAction types.CapitalAccountAction
	jsonErr := json.Unmarshal(result, &capitalAccountAction)

	if jsonErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error unmarshaling json"})
		return
	}
	c.JSON(http.StatusOK, capitalAccountAction)
}
