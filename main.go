package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var Token string

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	if Token == "" {
		fmt.Println("Flag missing! Must specify bot token with -t flag.")
		return
	}

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating discord session: " + err.Error())
		return
	}

	dg.Identify.Intents = discordgo.IntentGuildMessages
	//dg.Identify.Intents += discordgo.IntentDirectMessages

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening discord connection: " + err.Error())
		return
	}

	fmt.Println("Bot running! Use Ctrl-C to exit.")
	sigchannel := make(chan os.Signal, 1)
	signal.Notify(sigchannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sigchannel

	dg.Close()
}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	fmt.Printf("[%s] Message Created: %s\n", time.Now().String(), message.Content)
	//session.ChannelMessageDelete(message.ChannelID, message.ID, )
}
