package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/exec"
	"vpn-agent/controllers"
	"vpn-agent/vpn"
)

func install(protocol string) error {
	p := vpn.AvailableInstallers[protocol]
	if p == nil {
		return fmt.Errorf("%s: unknown protocol", protocol)
	}

	commands := p.Commands()

	for _, command := range commands {
		cmd := exec.Command("/bin/bash", "-c", command)
		cmd.Env = os.Environ()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()

		if err != nil {
			return err
		}
	}

	return nil
}

//middlewares

func Authorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken, tokenExists := os.LookupEnv("AUTH_TOKEN")
		if !tokenExists {
			log.Println("Auth token not exists")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var token string
		if len(c.Request.Header["Authorization"]) > 0 {
			token = c.Request.Header["Authorization"][0]
		} else {
			token = ""
		}

		if token != authToken {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}

func main() {
	action := os.Args[len(os.Args)-1]

	switch action {
	case "install":
		protocol := flag.String("protocol", "", "vpn protocol")
		flag.Parse()

		if *protocol == "" {
			fmt.Println("Protocol is empty")
			return
		}

		if err := install(*protocol); err != nil {
			fmt.Println("Install error: " + err.Error())
		}
	case "run":
		r := gin.Default()

		r.GET("/healthcheck", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{})
		})

		r.Use(Authorization())

		account := r.Group("/account")
		{
			account.POST("/create", controllers.CreateAccount)
			account.POST("/delete", controllers.DeleteAccount)
		}

		r.Run()
	default:
		fmt.Println(action + ": unknown command")
	}
}
