/*	This file contains all high level commands of the bot. */

package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	log "github.com/Sirupsen/logrus"
	//yt "github.com/kkdai/youtube"
)

type MessageAck struct {
	MessageID string `json:"message_id"`
	ChannelID string `json:"channel_id"`
}

var (
	DEFAULT_LOL_REGION string = "euw"
	COIN_FACES_PATHS          = []string{
		"images/Head.png",
		"images/Tail.png",
	}
	JARVIS_ANSWERS = []string{
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
	PATH_TO_CLASSIFY_EXEC = "tensorflow/imagenet/classify_image.py"

	youtubeQueues      map[string]*StringQueue = make(map[string]*StringQueue)
	youtubeDownloading map[string]*StringQueue = make(map[string]*StringQueue)

	polls map[string]*Poll = make(map[string]*Poll)
)

/* 	Sends messages with information about the given players LoL rank.
Takes input of form: <Username> <Region>
*/
func elo(input []string, s *discordgo.Session, m *discordgo.MessageCreate, riotkey string) {
	region := DEFAULT_LOL_REGION

	p := NewParser(input)
	if !p.nextToken() {
		return
	}
	p.nextToken()
	name := strings.ToLower(p.Token)
	p.nextToken()
	switch strings.ToLower(p.Token) {
	case "na":
		region = "na"
	case "br":
		region = "br"
	case "kr":
		region = ""
	case "eune":
		region = "eune"
	case "jp":
		region = "jp"
	case "tr":
		region = "tr"
	case "oce":
		region = "oce"
	case "las":
		region = "las"
	case "ru":
		region = "ru"
	}
	summoner := GetSummoner(name, region, riotkey)
	if summoner.ID == 0 {
		s.ChannelMessageSend(m.ChannelID, "Could not find player.")
		return
	}
	league := GetLeague(strconv.Itoa(summoner.ID), region, riotkey)
	if league.Tier != "Unranked" {
		entry := league.Entry[0]
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**%s %s** %sLP", capitalize(strings.ToLower(league.Tier)), entry.Division, strconv.Itoa(entry.LP)))
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Wins: **%s** Losses: **%s**", strconv.Itoa(entry.Wins), strconv.Itoa(entry.Losses)))
	} else {
		s.ChannelMessageSend(m.ChannelID, "**Unranked**")
	}
	switch name {
	case "uznick":
		s.ChannelMessageSend(m.ChannelID, "But deserves Challenjour, Kappa.")
	case "mehdid":
		s.ChannelMessageSend(m.ChannelID, "Also, best Amumu EUW.")
	case "flakelol":
		s.ChannelMessageSend(m.ChannelID, "Also, best Shen EUW.")
	}
}

/* 	Sends messages with information about the given cities weather
Takes input of form <city> <time>|<country> <country>
Time can be either "today" or "tomorrow"
Country is its code (e.g. "de" for germany)
*/
func weather(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	setTime := false
	country := "de"
	forecastStart := 0 //Today
	forecastRange := 3 //0,1,2
	days := []string{"Today", "Tomorrow", "In two days", "In three days", "In four days", "In five days", "In six days", "In a week"}

	p := NewParser(input)
	if !p.nextToken() {
		return
	}
	city := p.Token

	if p.nextToken() {
		switch strings.ToLower(p.Token) {
		case "today":
			forecastRange = 1
			setTime = true
		case "tomorrow":
			forecastRange = 1
			forecastStart = 1
			setTime = true
		default:
			country = p.Token
		}
	}
	if p.nextToken() && setTime {
		country = p.Token
	}

	weather := GetWeather(city, country)
	if (forecastStart + forecastRange) > len(weather.Forecast.Time) {
		s.ChannelMessageSend(m.ChannelID, "Can't get information in that timerange or for that city.")
		return
	}
	if len(weather.Forecast.Time) > forecastRange {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Weather forecast for: **%s**", capitalize(city)))
		for i := forecastStart; i < (forecastStart + forecastRange); i++ {
			d := weather.Forecast.Time[i]
			text := fmt.Sprintf("%s there will be %s with a temperature ranging from %s°C to %s°C and a %s from the %s.", days[i], bold(d.Symbol.Name), bold(d.Temperature.Min), bold(d.Temperature.Max), bold(strings.ToLower(d.WindSpeed.Name)), bold(strings.ToLower(d.WindDirection.Name)))
			s.ChannelMessageSend(m.ChannelID, text)
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "Can't find information about this city.")
		return
	}
}

/* 	Answers the question 8-Ball Style randomly.
 */
func jarvis(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if string(input[len(input)-1][len(input[len(input)-1])-1]) == "?" {
		if m.Author.ID == "119818300308848651" { //Für Miguel
			s.ChannelMessageSend(m.ChannelID, "No.")
		} else {
			s.ChannelMessageSend(m.ChannelID, JARVIS_ANSWERS[rand.Intn(len(JARVIS_ANSWERS))])
		}
	}
}

/*	Uploads a random coin face.
 */
func coin(s *discordgo.Session, m *discordgo.MessageCreate) {
	file, err := os.Open(COIN_FACES_PATHS[rand.Intn(len(COIN_FACES_PATHS))])
	if err != nil {
		log.Warning(err)
	}
	s.ChannelFileSend(m.ChannelID, "Coin.png", file)
}

/*	Sends a message with the result of a six-sided dice roll.
 */
func dice(s *discordgo.Session, m *discordgo.MessageCreate) {
	answers := []string{"1", "2", "3", "4", "5", "6"}
	s.ChannelMessageSend(m.ChannelID, answers[rand.Intn(len(answers))])
}

func help(s *discordgo.Session, m *discordgo.MessageCreate) {
	data := "Commands for J.A.R.V.I.S.: \n ''!help'' - to list all commands \n ''!coin'' - to flip a coin, \n ''!dice'' - to roll a dice, \n ''!elo'' <summonername> - to show the current LoL rank of the Summoner \n ''!remindme'' <seconds> <message> or ''!rm'' <seconds> <message> - to remind you for something \n Emotes: \n ''!kappa'' \n ''!erwinross''"
	s.ChannelMessageSend(m.ChannelID, data)
}

/* Queues up a Youtube video, whose sound is played in the voice channel the command caller is in. Downloads the entire Youtube video locally, which might take a while, based on the internet connection.
 */
func queueYoutube(input []string, s *discordgo.Session, m *discordgo.MessageCreate, g *discordgo.Guild) {
	var (
		titleOut    []byte
		filenameOut []byte
		err         error
	)
	if len(input) < 1 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !play <link>")
		return
	}
	link := input[0]
	log.Info("Downloading ", link)

	if filenameOut, err = exec.Command("youtube-dl", "-f", "140", link, "--get-filename").Output(); err != nil {
		log.Info("Error calling youtube-dl command (only to get id): ", err)
	}
	file := strings.Replace(string(filenameOut), "\n", "", -1) //replace all new lines
	log.Info("--get-filename (with newlines removed): " + file)

	//THIS RETURNS A NEWLINE AT THE END
	if titleOut, err = exec.Command("youtube-dl", "-f", "140", link, "--get-title").Output(); err != nil {
		log.Info("Error calling youtube-dl command (only to get id): ", err)
	}
	title := strings.Replace(string(titleOut), "\n", "", -1) //replace all new lines
	log.Info("--get-title (with newlines removed): " + title)

	if _, ok := youtubeDownloading[g.ID]; ok {
		log.Info("Enqueuing (not a new queue)")
		youtubeDownloading[g.ID].enqueue(title)
	} else {
		log.Info("Enqueuing into a new queue")
		youtubeDownloading[g.ID] = newStringQueue(20)
		youtubeDownloading[g.ID].enqueue(title)
	}

	log.Info("Starting download.")
	if err = exec.Command("youtube-dl", "-f", "140", link).Run(); err != nil {
		log.Info("Error calling youtube-dl command: ", err)
	}
	log.Info("Finished download.")

	youtubeDownloading[g.ID].remove(title)

	if _, ok := youtubeQueues[g.ID]; ok {
		err = youtubeQueues[g.ID].enqueue(title)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Queue full. sry m8")
		}
	} else {
		youtubeQueues[g.ID] = newStringQueue(20)
		youtubeQueues[g.ID].enqueue(title)
	}

	log.Info("File: " + file)

	file = fmt.Sprintf("./%s", file)

	sound := createSound(link)
	sound.Load(file)

	go enqueuePlay(m.Author, g, sound, title)

	err = os.Remove(file)
	if err != nil {
		log.Warning("Unsuccessfull deletion of file.")
		log.Warning(err)
	}
}

/* Skips the current song.
 */
func nextYoutube(s *discordgo.Session, m *discordgo.MessageCreate, g *discordgo.Guild) {
	s.ChannelMessageSend(m.ChannelID, "Skipping song.")
	next(g.ID)
}

func currentsong(s *discordgo.Session, m *discordgo.MessageCreate, g *discordgo.Guild) {
	if queue, ok := youtubeQueues[g.ID]; ok {
		if queue.length() > 0 {
			s.ChannelMessageSend(m.ChannelID, "Current song: "+queue.peek())
			return
		}
	}
	s.ChannelMessageSend(m.ChannelID, "No song playing")
}

func printQueue(s *discordgo.Session, m *discordgo.MessageCreate, g *discordgo.Guild) {
	message := ""

	yq, okq := youtubeQueues[g.ID]
	ydq, okd := youtubeDownloading[g.ID]

	if okq {
		i := 0
		for _, y := range yq.toArray() {
			if y != "" {
				i++
				message = message + strconv.Itoa(i) + ". " + y + " \n"
			}
		}
	}
	if okd {
		i := 0
		for _, y := range ydq.toArray() {
			if y != "" {
				if i == 0 {
					message = message + "Downloading: "
				}
				i++
				message = message + y + " \n"
			}
		}
	}
	s.ChannelMessageSend(m.ChannelID, message)
}

func startPoll(input []string, s *discordgo.Session, m *discordgo.MessageCreate, g *discordgo.Guild) {
	_, exists := polls[g.ID]
	if !exists {
		var voters []string
		polls[g.ID] = &Poll{"", nil, voters}
	}
	if polls[g.ID].description != "" {
		s.ChannelMessageSend(m.ChannelID, "There's already a poll! End it with !endpoll")
		return
	}
	description := strings.Join(input, " ")
	if description == "" {
		s.ChannelMessageSend(m.ChannelID, "Needs a description!")
		return
	}
	var voters []string
	polls[g.ID] = &Poll{description, make(map[string]int), voters}
	s.ChannelMessageSend(m.ChannelID, "Added poll: \""+description+"\" \nEnter !vote <yourvote> to vote!")
}

func vote(input []string, s *discordgo.Session, m *discordgo.MessageCreate, g *discordgo.Guild) {
	_, exists := polls[g.ID]
	if !exists {
		var voters []string
		polls[g.ID] = &Poll{"", nil, voters}
	}
	if polls[g.ID].description == "" {
		s.ChannelMessageSend(m.ChannelID, "There's no poll! Start a poll with !startpoll")
		return
	}
	if contains(m.Author.ID, polls[g.ID].voters) {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> : You already voted once...", m.Author.ID))
		return
	}
	polls[g.ID].vote(strings.Join(input, " "))
	polls[g.ID].voters = append(polls[g.ID].voters, m.Author.ID)
}

func endPoll(s *discordgo.Session, m *discordgo.MessageCreate, g *discordgo.Guild) {
	_, exists := polls[g.ID]
	if !exists {
		var voters []string
		polls[g.ID] = &Poll{"", nil, voters}
	}
	if polls[g.ID].description == "" {
		s.ChannelMessageSend(m.ChannelID, "There's no poll! Start a poll with !startpoll")
		return
	}
	s.ChannelMessageSend(m.ChannelID, "Ending poll: \""+polls[g.ID].description+"\"")
	s.ChannelMessageSend(m.ChannelID, "Result: \n"+bold(polls[g.ID].getResult()))
	polls[g.ID] = &Poll{"", nil, nil}
}
