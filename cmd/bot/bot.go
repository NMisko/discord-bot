package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
)

var (
	// discordgo session
	discord *discordgo.Session
	me      *discordgo.User
	RiotKey = flag.String("k", "", "Riot API Key")

	ADMINS = []string{
		"118830934710681602", //Beni
		"118641605837062144", //Nicola
	}

	RESTRICTED = []string{}

	BANNED = []string{}
)

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Info("Recieved READY payload")
	s.UpdateStatus(0, "League Of Legends")
}

func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if !shardContains(event.Guild.ID) {
		return
	}

	if event.Guild.Unavailable == true {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			return
		}
	}
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Info("message created: ", m.Content)
	log.Info("Author: ", m.Author.ID)
	var (
		ourShard = true
		//adminMode = false
	)

	if len(m.Content) <= 0 || (m.Content[0] != '!' && m.Content[0] != '<') { //@J.A.R.V.I.S = <@168313836951175168>
		return
	}

	// if contains(m.Author.ID, ADMINS) {
	// 	adminMode = true
	// }

	parts := strings.Split(m.Content, " ")

	channel, _ := discord.State.Channel(m.ChannelID)
	if channel == nil {
		log.WithFields(log.Fields{
			"channel": m.ChannelID,
			"message": m.ID,
		}).Warning("Failed to grab channel")
		return
	}

	guild, _ := discord.State.Guild(channel.GuildID)
	if guild == nil {
		log.WithFields(log.Fields{
			"guild":   channel.GuildID,
			"channel": channel,
			"message": m.ID,
		}).Warning("Failed to grab guild")
		return
	}

	// If we're in sharding mode, test whether this message is relevant to us
	if !shardContains(channel.GuildID) {
		ourShard = false
	}

	if !ourShard {
		return
	}

	if contains(m.Author.ID, BANNED) {
		return
	}

	switch strings.ToLower(parts[0]) {
	case "!weather":
		weather(parts[1:], s, m)
	case "<@168313836951175168>":
		jarvis(parts[1:], s, m)
	case "!dice":
		dice(s, m)
	case "!elo":
		elo(parts[1:], s, m, *RiotKey)
	case "!remindme":
		remindme(parts[1:], s, m)
	case "!rm":
		remindme(parts[1:], s, m)
	case "!help":
		help(s, m)
	case "!queue":
		printQueue(s, m, guild)
	case "!song":
		currentsong(s, m, guild)
	case "!startpoll":
		startPoll(parts[1:], s, m, guild)
	case "!vote":
		vote(parts[1:], s, m, guild)
	case "!endpoll":
		endPoll(s, m, guild)
	case "!coin":
		coin(s, m)
	}

	//My ears
	if contains(m.Author.ID, RESTRICTED) {
		return
	}

	switch strings.ToLower(parts[0]) {
	case "!play":
		go queueYoutube(parts[1:], s, m, guild)
	case "!plays":
		go queueAndDeleteYoutube(parts[1:], s, m, guild)
	case "!skip":
		nextYoutube(s, m, guild)
	}
}

func main() {
	var (
		Token = flag.String("t", "", "Discord Authentication Token")
		Owner = flag.String("o", "", "Owner ID")
		err   error
	)
	flag.Parse()

	if *Token == "" || *RiotKey == "" {
		flag.PrintDefaults()
		return
	}

	if *Owner != "" {
		OWNER = *Owner
	}

	if err = os.Mkdir("temp", 0777); err != nil {
		log.Warning("Error creating temp directory: ", err, "\nTemp directory is used to store downloaded data and will be later deleted.")
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

	discord.AddHandler(onReady)
	discord.AddHandler(onGuildCreate)
	discord.AddHandler(onMessageCreate)

	err = discord.Open()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to create discord websocket connection")
		return
	}

	// We're running!
	log.Info("JARVIS READY.")

	// Wait for a signal to quit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-c

	if err := os.Remove("temp"); err != nil {
		log.Warning("Unsuccessful deletion of temp folder: ", err, "\nDelete it manually.")
	}
}
