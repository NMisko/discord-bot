package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"math/rand"
	"io"

	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
	redis "gopkg.in/redis.v3"
)

var (
	// discordgo session
	discord *discordgo.Session
	me *discordgo.User

	DEFAULT_LOL_REGION string = "euw"
)

type Message interface {
    Send(s *discordgo.Session, m *discordgo.MessageCreate)
}

type MixedMessage struct {
     Content []Message
}
func (message MixedMessage) Send(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, message := range message.Content {
		message.Send(s, m)
	}
}
//Put Message structs in jarvis.go file, keep functions here.
type FileMessage struct {
    file io.Reader
    name string
}
func (f FileMessage) Send(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelFileSend(m.ChannelID, f.name, f.file)
}

type TextMessage struct {
    text string
}
func (t TextMessage) Send(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, t.text)
}

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

func elo(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	region := DEFAULT_LOL_REGION
	if (len(input) > 0) {
		if (len(input) > 2) {
			switch (strings.ToLower(input[1])) {
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

		name := strings.ToLower(input[0])

		if len(name) > 0 {
			summoner := GetSummonerElo(name, region)
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
			switch name {
				case "uznick": s.ChannelMessageSend(m.ChannelID, "But deserves Challenjour, Kappa.")
				case "mehdid": s.ChannelMessageSend(m.ChannelID, "also, best Amumu EUW")
				case "flakelol": s.ChannelMessageSend(m.ChannelID, "also, best Shen EUW")
			}
		}
	}
	return
}

func weather(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	forecastStart := 0; //Today
	forecastRange := 3; //0,1,2
	country := "de"
	days := []string{"Today", "Tomorrow", "In two days", "In three days", "In four days", "In five days", "In six days", "In a week"}

	if (len(input) > 1) { //has 2 arguments
		if (strings.ToLower(input[1]) == "today") {
			forecastRange = 1
		} else if (strings.ToLower(input[1]) == "tomorrow") {
			forecastStart = 1
			forecastRange = 1
		} else { country = input[1] }
	}
	if (len(input) > 0) {
		city := input[0]
		weather := GetWeather(city, country)
		if ((forecastStart+forecastRange) > len(weather.Forecast.Time)) {
			s.ChannelMessageSend(m.ChannelID, "Can't get information in that timerange or for that city.")
			return
		}

		//print out next 3 days
		if (len(weather.Forecast.Time) > forecastRange) {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Weather forecast for: **%s**", capitalize(city)))
			for i := forecastStart; i < (forecastStart+forecastRange); i++ {
				d := weather.Forecast.Time[i]
				text := fmt.Sprintf("%s: **%s**        min: **%s**°C max: **%s**°C        **%s** from **%s**", days[i], capitalize(d.Symbol.Name), d.Temperature.Min, d.Temperature.Max, strings.ToLower(d.WindSpeed.Name), strings.ToLower(d.WindDirection.Name))
				s.ChannelMessageSend(m.ChannelID, text)
			}
		} else {
			s.ChannelMessageSend(m.ChannelID, "Can't find information about this city.")
			return
	 	}
	}
}

func jarvis(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if (string(input[len(input)-1][len(input[len(input)-1])-1]) == "?") {
		//rand.Seed(50)
		answers := []string {
			"It is certain.",
			"It is decidedly so.",
			"Without a doubt.",
			"Yes definitely.",
			"You may rely on it.",
			"As I see it yes.",
			"Most likely.",
			"Outlook good.",
			"Yes, of course.",
			"Yes.",
			"Signs point to yes.",
			"Ask again later.",
			"Better not tell you now.",
			"Cannot predict now.",
			"Concentrate and ask again.",
			"Don't count on it.",
			"No.",
			"Never.",
			"Not in this universe.",
			"Dont ask me..",
			"What are you? Fucking gay?",
			"My reply is no.",
			"My sources say no.",
			"Outlook not so good.",
			"Very doubtful.",
		}
	s.ChannelMessageSend(m.ChannelID, answers[rand.Intn(len(answers))])
	}
}

func coin(s *discordgo.Session, m *discordgo.MessageCreate) {
	files := []string{
		"images/Head.png",
		"images/Tail.png",
	}
	file, err := os.Open(files[rand.Intn(len(files))])
	if err != nil { log.Warning(err) }
	s.ChannelFileSend(m.ChannelID, "Coin.png", file)
}

func dice(s *discordgo.Session, m *discordgo.MessageCreate) {
	answers := []string {
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
	}
	s.ChannelMessageSend(m.ChannelID, answers[rand.Intn(len(answers))])
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Info("message created: ", m.Content)
	var (
		ourShard = true
	)

	if (len(m.Content) <= 0 || (m.Content[0] != '!' && m.Content[0] != '<')) { //@J.A.R.V.I.S = <@168313836951175168>
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

	if !ourShard {
		return
	}

	switch parts[0] {
		case "!elo": elo(parts[1:], s, m)
		case "!weather": weather(parts[1:], s, m)
		case "<@168313836951175168>": jarvis(parts[1:], s, m)
		case "!coin": coin(s, m)
		case "!dice": dice(s, m)
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
