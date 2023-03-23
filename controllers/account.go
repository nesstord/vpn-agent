package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"vpn-agent/vpn"
)

func CreateAccount(c *gin.Context) {
	var abstractAccount vpn.Account
	account, err := abstractAccount.NewAccount()
	if err != nil {
		fmt.Println("Cannot get new account: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong!"})
		return
	}

	data, err := account.Create()
	if err != nil {
		fmt.Println("Creating account error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func DeleteAccount(c *gin.Context) {
	var request struct {
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var abstractAccount vpn.Account
	account, err := abstractAccount.NewAccount()
	if err != nil {
		fmt.Println("Cannot get new account: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong!"})
		return
	}

	data, err := account.Delete(request.Password)
	if err != nil {
		fmt.Println("Deleting account error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Something went wrong!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}
