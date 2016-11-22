package main

import (
	"flag"
	"io/ioutil"
	"encoding/base64"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
)

var (
	// discordgo session
	discord *discordgo.Session

	me *discordgo.User
)

// Sets nickname, profile picture and status of given bot.
func main() {
	var (
		Token     = flag.String("t", "", "discord authentication token")
		ID        = flag.String("i", "", "bot id")
		Nickname  = flag.String("n", "", "new bot nickname")
		Imagefile = flag.String("f", "", "new bot profile image file")
		Status    = flag.String("s", "", "new bot status")
		err       error
	)
	flag.Parse()

	if *Token == "" || *ID == "" || *Nickname == "" || *Imagefile == "" || *Status == "" {
		flag.PrintDefaults()
		return
	}

	// Create a discord session
	log.Info("Starting discord session...")
	discord, err = discordgo.New("Bot " + *Token)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to create discord session")
		return
	}

	//Supposed to change the name of our bot
	me, err = discord.User(*ID)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal(" -- Failed to get user.")
		return
	}

	err = discord.Open()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to create discord websocket connection")
		return
	}

	var avatar []byte
	if avatar,err = ioutil.ReadFile(*Imagefile); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("-- Failed reading file")
		return
	}

	transmittedAvatarData := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(avatar))

	log.Info("Setting avatar and nickname.")
	me, err = discord.UserUpdate(me.Email, "", *Nickname, transmittedAvatarData, "")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("-- Failed user update")
		return
	}

	err = discord.UpdateStatus(0, *Status)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("-- Failed status update")
		return
	}

	log.Info("Finished!")
}
