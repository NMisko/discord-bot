package main

import (
	"flag"
	"fmt"
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
	var (
		sound    *Sound
		stype    int = TYPE_AIRHORN
		khaled   bool
		ourShard = true
	)

	if len(m.Content) <= 0 || (m.Content[0] != '!' && len(m.Mentions) != 1) {
		return
	}

	parts := strings.Split(strings.ToLower(m.Content), " ")

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

	if len(m.Mentions) > 0 {
		if m.Mentions[0].ID == s.State.Ready.User.ID && m.Author.ID == OWNER && len(parts) > 0 {
			if scontains(parts[len(parts)-1], "stats") && ourShard {
				users := 0
				for _, guild := range s.State.Ready.Guilds {
					users += len(guild.Members)
				}

				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
					"I'm in %v servers with %v users.",
					len(s.State.Ready.Guilds),
					users))
			} else if scontains(parts[len(parts)-1], "status") {
				guilds := 0
				for _, guild := range s.State.Ready.Guilds {
					if shardContains(guild.ID) {
						guilds += 1
					}
				}
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
					"Shard %v contains %v servers",
					strings.Join(SHARDS, ","),
					guilds))
			} else if len(parts) >= 3 && scontains(parts[len(parts)-2], "die") {
				shard := parts[len(parts)-1]
				if len(SHARDS) == 0 || scontains(shard, SHARDS...) {
					log.Info("Got DIE request, exiting...")
					s.ChannelMessageSend(m.ChannelID, ":ok_hand: goodbye cruel world")
					os.Exit(0)
				}
			} else if scontains(parts[len(parts)-1], "aps") && ourShard {
				s.ChannelMessageSend(m.ChannelID, ":ok_hand: give me a sec m8")
				go calculateAirhornsPerSecond(m.ChannelID)
			} else if scontains(parts[len(parts)-1], "where") && ourShard {
				s.ChannelMessageSend(m.ChannelID,
					fmt.Sprintf("its a me, shard %v", string(guild.ID[len(guild.ID)-5])))
			}
			return
		}
	}

	if !ourShard {
		return
	}
	if parts[0] == "!elo" {
		region := "euw"
		if (len(parts) > 1) {
			if (len(parts) > 2) {
				switch (strings.ToLower(parts[2])) {
					case "na": region = "na"
					case "br": region = "br"
					case "kr": region = ""
					case "eune": region = "eune"
					case "jp": region = "jp"
					case "tr": region = "tr"
					case "oce": region = "oce"
					case "las": region = "las"
					case "ru" : region = "ru"
				}
			}

			name := strings.ToLower(parts[1])
			if len(name) > 0 {
				summoner := GetSummonerElo(parts[1], region)
				if(summoner.rank == "") {
					s.ChannelMessageSend(m.ChannelID, "Could not find player.")
					return
				}
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("http:%s", summoner.rankImage))
				if(summoner.rank != "Unranked") {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**%s** %sLP", summoner.rank, summoner.lp))
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Wins: **%s** Losses: **%s** Winrate: **%s**", summoner.wins, summoner.losses, strings.Join([]string{summoner.winratio, "%"}, "")))
				} else {
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**%s**", summoner.rank))
				}

				if(name == "uznick") {
					s.ChannelMessageSend(m.ChannelID, "But deserves Challenjour, Kappa.")
				}
				if(name == "MehdiD") {
					s.ChannelMessageSend(m.ChannelID, "also, best Amumu EUW")
				}
			}
		return
		} else {
			return
		}
	}

	if scontains(parts[0], COMMANDS...) {
		// Support !airhorn <sound>
		if len(parts) > 1 {
			for _, s := range AIRHORNS {
				if parts[1] == s.Name {
					sound = s
				}
			}

			if sound == nil {
				return
			}
		}

		// Select mode
		if scontains(parts[0], "!cena", "!johncena", "!nycto") {
			stype = TYPE_CENA
		} else if scontains(parts[0], "!eb", "!ethanbradberry", "!h3h3") {
			stype = TYPE_ETHAN
		} else if scontains(parts[0], "!anotha", "!anothaone") {
			khaled = true
		}

		go enqueuePlay(m.Author, guild, sound, khaled, stype)
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

	// Preload all the sounds
	log.Info("Preloading sounds...")
	for _, sound := range AIRHORNS {
		AIRHORN_SOUND_RANGE += sound.Weight
		sound.Load()
	}

	log.Info("Preloading loyalty...")
	for _, sound := range KHALED {
		KHALED_SOUND_RANGE += sound.Weight
		sound.Load()
	}

	log.Info("PRELOADING THE JOHN CENA")
	for _, sound := range CENA {
		CENA_SOUND_RANGE += sound.Weight
		sound.Load()
	}

	log.Info("I'm ethan bradberry!")
	for _, sound := range ETHAN {
		ETHAN_SOUND_RANGE += sound.Weight
		sound.Load()
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
	log.Info("AIRHORNBOT is ready to horn it up.")

	// Wait for a signal to quit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}
