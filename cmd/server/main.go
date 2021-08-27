package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/zacharyfrederick/admin/endpoints"
	"github.com/zacharyfrederick/admin/web"
)

func main() {
	path, err := os.Getwd()
	if err == nil {
		fmt.Println(path)
	}
	adminServer, err := web.ConnectToNetwork()
	if err != nil {
		log.Fatalf("Could not connect to the network: %v", err)
	}

	endpointWrapper := &endpoints.EndpointWrapper{AdminServer: adminServer}

	router := gin.Default()
	router.POST("/funds", endpointWrapper.PostFundEndpoint)
	router.GET("/funds/:id", endpointWrapper.GetFundByIdEndpoint)

	router.POST("/investors", endpointWrapper.PostInvestorEndpoint)
	router.GET("/investors/:id", endpointWrapper.GetInvestorByIdEndpoint)

	router.POST("/capitalaccounts", endpointWrapper.PostCapitalAccountEndpoint)
	router.GET("/capitalaccounts/:id", endpointWrapper.GetCapitalAccountByIdEndpoint)

	router.POST("/portfolios", endpointWrapper.PostPortfoliosEndpoint)
	router.GET("/portfolios/:id", endpointWrapper.GetPortfolioByIdEndpoint)

	router.POST("/capitalaccountactions", endpointWrapper.PostCapitalAccountActionEndpoint)
	router.GET("/capitalaccountactions/:id", endpointWrapper.GetCapitalAccountActionByIdEndpoint)

	router.POST("/portfolioactions", endpointWrapper.PostPortfolioActionEndpoint)
	router.GET("/portfolioactions/:id", endpointWrapper.GetPortfolioActionByIdEndpoint)

	router.Run()
}
