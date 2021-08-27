package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/zacharyfrederick/admin/types"
)

func (a *EndpointWrapper) PostFundEndpoint(c *gin.Context) {
	var createFundRequest types.CreateFundRequest

	err := c.BindJSON(&createFundRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameter"})
		return
	}

	validRequest := types.ValidateCreateFundRequest(&createFundRequest)
	if !validRequest {
		c.JSON(http.StatusBadRequest, gin.H{"error": "improperly formatted inceptionDate"})
	}

	fundId := uuid.NewV4().String()

	result, err := a.Contract.SubmitTransaction("CreateFund", fundId, createFundRequest.Name, createFundRequest.InceptionDate)
	if err != nil {
		errorString := fmt.Sprintf("error submitting request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorString})
		fmt.Println(result)
	}

	c.JSON(http.StatusOK, gin.H{"fundId": fundId})
}

func (a *EndpointWrapper) GetFundByIdEndpoint(c *gin.Context) {
	fundId := c.Param("id")
	result, err := a.Contract.EvaluateTransaction("QueryFundById", fundId)
	if err != nil {
		errorString := fmt.Sprintf("error evaluating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorString})
		fmt.Println(result)
	}

	if len(result) == 0 {
		c.JSON(http.StatusOK, "")
		return
	}

	var fund types.Fund
	jsonErr := json.Unmarshal(result, &fund)

	if jsonErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error unmarshaling json"})
		fmt.Println(jsonErr)
		return
	}
	c.JSON(http.StatusOK, fund)
}

func (a *EndpointWrapper) GetFundAction(c *gin.Context) {
	fundId := c.Param("id")
	action := c.Param("action")

	fmt.Println(fundId)

	switch action {
	case "/investors":
		fmt.Println("investors")
	case "/capitalaccounts":
		fmt.Println("capitalaccounts")
	case "/portfolios":
		fmt.Println("portfolios")
	case "/capitalaccountactions":
		fmt.Println("capitalaccountactions")
	case "/portfolioactions":
		fmt.Println("portfolioactions")
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action supplied"})
		return
	}

	c.JSON(http.StatusOK, "success")
}
