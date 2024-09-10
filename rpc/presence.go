package rpc

import (
	"github.com/hugolgst/rich-go/client"
	"github.com/unickorn/discordtidal/log"
	"os"
	"os/signal"
	"syscall"
)

var (
	loggedIn = false
	applicationId = "1282957708487561268"
)

// Init ...
func Init() {
	Login()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		Logout()
		os.Exit(1)
	}()
}

// Login logs into the Discord socket.
func Login() {
	if !loggedIn {
		err := client.Login(applicationId)
		if err != nil {
			log.Log().Fatal("Login Error", err)
		}
		loggedIn = true

		log.Log().Infoln("Logged into discord")
	}
}

// Logout logs out from the Discord socket.
func Logout() {
	if loggedIn {
		client.Logout()
		loggedIn = false

		log.Log().Infoln("Logged out of discord")
	}
}

// Relog logs out and back in.
func Relog() {
	log.Log().Infoln("Reloading presence")

	client.Logout()
	err := client.Login(applicationId)
	if err != nil {
		panic(err)
	}
}
