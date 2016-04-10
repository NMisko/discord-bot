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

	// Redis client connection (used for stats)
	rcli *redis.Client

	// Owner
	OWNER string

	// Shard (or -1)
	SHARDS []string = make([]string, 0)
)

func main() {
	var (
		Token = flag.String("t", "", "Discord Authentication Token")
		Redis = flag.String("r", "", "Redis Connection String")
		Shard = flag.String("s", "", "Integers to shard by")
		Owner = flag.String("o", "", "Owner ID")
		err   error
	)
	flag.Parse()

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

	//Supposed to change the name of our bot
	me, err = discord.User("168313836951175168")
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

	log.Info("Setting avatar...")
	me, err = discord.UserUpdate(me.Email, "", "J.A.R.V.I.S", "data:image/png;base64,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", "")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("-- Failed user update")
		return
	}

	// We're running!
	log.Info("Finished!")

	// Wait for a signal to quit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}
