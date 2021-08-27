package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/zacharyfrederick/admin/types"
)

func (w *EndpointWrapper) PostPortfoliosEndpoint(c *gin.Context) {
	var createPortfolioRequest types.CreatePortfolioRequest

	err := c.BindJSON(&createPortfolioRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
		return
	}

	validRequest := types.ValidateCreatePortfolioRequest(&createPortfolioRequest)
	if !validRequest {
		c.JSON(http.StatusBadRequest, gin.H{"error": "improperly formatted request"})
		return
	}

	porfolioId := uuid.NewV4().String()

	result, err := w.Contract.SubmitTransaction("CreateCapitalAccount", porfolioId, createPortfolioRequest.Fund, createPortfolioRequest.Name)
	if err != nil {
		errorString := fmt.Sprintf("error submitting request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorString})
		fmt.Println(result)
		return
	}

	c.JSON(http.StatusOK, gin.H{"porfolioId": porfolioId})
}

func (w *EndpointWrapper) GetPortfolioByIdEndpoint(c *gin.Context) {
	porfolioId := c.Param("id")
	result, err := w.Contract.EvaluateTransaction("QueryInvestorById", porfolioId)

	if err != nil {
		errorString := fmt.Sprintf("error evaluating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorString})
		fmt.Println(result)
	}

	if len(result) == 0 {
		c.JSON(http.StatusOK, "")
		return
	}

	var portfolio types.Portfolio
	jsonErr := json.Unmarshal(result, &portfolio)

	if jsonErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error unmarshaling json"})
		return
	}
	c.JSON(http.StatusOK, portfolio)
}

func (w *EndpointWrapper) PostPortfolioActionEndpoint(c *gin.Context) {
	var createPortfolioActionRequest types.CreatePortfolioActionRequest

	err := c.BindJSON(&createPortfolioActionRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
		return
	}

	validRequest := types.ValidateCreatePortfolioActionRequest(&createPortfolioActionRequest)
	if !validRequest {
		c.JSON(http.StatusBadRequest, gin.H{"error": "improperly formatted request"})
		return
	}

	transactionId := uuid.NewV4().String()

	period := string(createPortfolioActionRequest.Period)
	result, err := w.Contract.SubmitTransaction("CreatePortfolioAction", transactionId, createPortfolioActionRequest.Fund, createPortfolioActionRequest.Type, createPortfolioActionRequest.Date, period, createPortfolioActionRequest.Name, createPortfolioActionRequest.CUSIP, createPortfolioActionRequest.Amount, createPortfolioActionRequest.Currency)
	if err != nil {
		errorString := fmt.Sprintf("error submitting request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorString})
		fmt.Println(result)
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactionId": transactionId})
}

func (w *EndpointWrapper) GetPortfolioActionByIdEndpoint(c *gin.Context) {
	transactionId := c.Param("id")
	result, err := w.Contract.EvaluateTransaction("QueryPortfolioActionById", transactionId)

	if err != nil {
		errorString := fmt.Sprintf("error evaluating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorString})
		fmt.Println(result)
	}

	if len(result) == 0 {
		c.JSON(http.StatusOK, "")
		return
	}

	var portfolioAction types.PortfolioAction
	jsonErr := json.Unmarshal(result, &portfolioAction)

	if jsonErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error unmarshaling json"})
		return
	}
	c.JSON(http.StatusOK, portfolioAction)
}
