package main

import (
	"flag"
	"os"
	"os/signal"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
	redis "gopkg.in/redis.v3"
)

var (
	// discordgo session
	discord *discordgo.Session
	me *discordgo.User
	RiotKey = flag.String("k", "", "Riot API Key")
)

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Info("Recieved READY payload")
	s.UpdateStatus(0, "League Of Legends")
}

func onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if !shardContains(event.Guild.ID) {
		return
	}

	if event.Guild.Unavailable != nil {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			s.ChannelMessageSend(channel.ID, "BEEP BOOP BOOTING")
			s.ChannelMessageSend(channel.ID, "https://media.giphy.com/media/3o85g3yQa2iG2Rdq1O/giphy.gif")
			return
		}
	}
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Info("message created: ", m.Content)
	var (
		ourShard = true
	)

	if (len(m.Content) <= 0 || (m.Content[0] != '!' && m.Content[0] != '<')) { //@J.A.R.V.I.S = <@168313836951175168>
		return
	}
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

	switch strings.ToLower(parts[0]) {
		case "!weather": weather(parts[1:], s, m)
		case "<@168313836951175168>": jarvis(parts[1:], s, m)
		case "!coin": coin(s, m)
		case "!dice": dice(s, m)
		case "!elo": elo(parts[1:], s, m, *RiotKey)
		//case "!whatis": classifyImage(parts[1:], s, m)
	}
}


func main() {
	var (
		Token = flag.String("t", "", "Discord Authentication Token")
		Redis = flag.String("r", "", "Redis Connection String")
		Shard = flag.String("s", "", "Integers to shard by")
		Owner = flag.String("o", "", "Owner ID")
		err   error
	)
	flag.Parse()

	if(*Token == "" || *RiotKey == "") {
		flag.PrintDefaults()
		return
	}

	if *Owner != "" {
		OWNER = *Owner
	}

	// Make sure shard is either empty, or an integer
	if *Shard != "" {
		SHARDS = strings.Split(*Shard, ",")

		for _, shard := range SHARDS {
			if _, err := strconv.Atoi(shard); err != nil {
				log.WithFields(log.Fields{
					"shard": shard,
					"error": err,
				}).Fatal("Invalid Shard")
				return
			}
		}
	}

	// If we got passed a redis server, try to connect
	if *Redis != "" {
		log.Info("Connecting to redis...")
		rcli = redis.NewClient(&redis.Options{Addr: *Redis, DB: 0})
		_, err = rcli.Ping().Result()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Failed to connect to redis")
			return
		}
	}

	// Create a discord session
	log.Info("Starting discord session...")
	discord, err = discordgo.New(*Token)
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
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}
