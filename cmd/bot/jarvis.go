/*	This file contains all high level commands of the bot. */

package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	clever "github.com/ugjka/cleverbot-go"

	log "github.com/Sirupsen/logrus"
)

type MessageAck struct {
	MessageID string `json:"message_id"`
	ChannelID string `json:"channel_id"`
}

var (
	DEFAULT_LOL_REGION string = "euw1"
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

	polls             map[string]*Poll           = make(map[string]*Poll)
	conversation      map[string]*clever.Session = make(map[string]*clever.Session)
	conversationTimer map[string]int             = make(map[string]int)
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
	name := strings.ToLower(p.Token)

	if p.nextToken() {
		region = strings.ToLower(p.Token)
		if !contains(region, []string{"ru", "kr"}) {
			if contains(region, []string{"euw", "eun", "br", "oc", "jp", "na", "tr", "la"}) {
				region = region + "1"
			} else {
				s.ChannelMessageSend(m.ChannelID, "Could not find region: "+region)
				return
			}
		}
	}

	summoner := GetSummoner(name, region, riotkey)
	if summoner.Status.Code == 404 {
		s.ChannelMessageSend(m.ChannelID, "Could not find player.")
		return
	}

	leagues := GetLeague(strconv.Itoa(summoner.ID), region, riotkey)
	message := ""
	if len(leagues) == 0 {
		s.ChannelMessageSend(m.ChannelID, "**Unranked**")
		return
	}
	for _, league := range leagues {
		if league.Tier != "Unranked" {
			message = message + fmt.Sprintf("%s: **%s %s**\n", league.Type, capitalize(strings.ToLower(league.Tier)), league.Rank)
		}
	}
	s.ChannelMessageSend(m.ChannelID, message)
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
func jarvis(input []string, s *discordgo.Session, m *discordgo.MessageCreate, g *discordgo.Guild, cleverbotKey string) {
	if len(input) < 1 {
		return
	}
	conversationTimer[g.ID] = 100 //In seconds.

	if conversation[g.ID] == nil {
		log.Info("New Cleverbot conversation started.")
		conversation[g.ID] = clever.New(cleverbotKey)
		go tryToCloseConversation(g)
	}
	sentence := strings.Replace(strings.Join(input, " "), "\x03", "", -1) //replacing annoying \x03
	answer, err := conversation[g.ID].Ask(sentence)
	if err != nil {
		log.Info("Cleverbot error")
		log.Info(err)
		s.ChannelMessageSend(m.ChannelID, "...")
	}

	s.ChannelMessageSend(m.ChannelID, answer)
}

func tryToCloseConversation(g *discordgo.Guild) {
	for conversationTimer[g.ID] > 0 {
		time.Sleep(5 * time.Second)
		conversationTimer[g.ID] = conversationTimer[g.ID] - 5
	}
	log.Info("Conversation reset.")
	conversation[g.ID].Reset()
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
	data := "Commands for J.A.R.V.I.S.: \n ''!help'' - list all commands \n ''!dice'' - to roll a dice, \n ''!elo'' <summonername> - show the current LoL rank of the Summoner \n ''!remindme'' <seconds> <message> or ''!rm'' <seconds> <message> - remind you something sometime \n Emotes: \n ''!play(s)'' - play a youtube song \n ''!queue'' - show the current queue \n ''!skip'' - skip a song \n ''!loop'' - loop songs \n ''!auto'' - auto play songs"
	s.ChannelMessageSend(m.ChannelID, data)
}

func queueAndDeleteYoutube(input []string, s *discordgo.Session, m *discordgo.MessageCreate, g *discordgo.Guild, authorID string) {
	if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		log.Warning("Couldn't delete message.")
	}
	if len(input) < 1 {
		return
	}
	link := input[0]
	if strings.Contains(link, "&") || !strings.Contains(link, "www.youtube.com/watch?v=") {
		return
	} else {
		s.ChannelMessageSend(m.ChannelID, "Queuing song.")
	}
	queueYoutube(input, s, m, g.ID, authorID)
}

/* Queues up a Youtube video, whose sound is played in the voice channel the command caller is in. Downloads the entire Youtube video locally, which might take a while, based on the internet connection.
 */
func queueYoutube(input []string, s *discordgo.Session, m *discordgo.MessageCreate, guildID string, authorID string) {
	var (
		titleOut    []byte
		filenameOut []byte
		err         error
	)
	if len(input) < 1 {
		if s != nil {
			s.ChannelMessageSend(m.ChannelID, "Usage: !play <link>")
		}
		return
	}
	link := input[0]

	if strings.Contains(link, "&") || !strings.Contains(link, "www.youtube.com/watch?v=") {
		if s != nil {
			s.ChannelMessageSend(m.ChannelID, "That link doesn't look right...")
		}
		return
	}

	log.Info("Downloading ", link)

	if filenameOut, err = exec.Command("youtube-dl", "-f", "140", link, "--get-filename", "--restrict-filenames").Output(); err != nil {
		log.Info("Error calling youtube-dl command (only to get id): ", err)
		if s != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Download failed :( Sorry <@%s>", m.Author.ID))
		}
		return
	}
	file := strings.Replace(string(filenameOut), "\n", "", -1) //replace all new lines
	log.Info("--get-filename (with newlines removed): " + file)

	//THIS RETURNS A NEWLINE AT THE END
	if titleOut, err = exec.Command("youtube-dl", "-f", "140", link, "--get-title").Output(); err != nil {
		log.Info("Error calling youtube-dl command (only to get id): ", err)
		if s != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Download failed :( Sorry <@%s>", m.Author.ID))
		}
		return
	}
	title := strings.Replace(string(titleOut), "\n", "", -1) //replace all new lines
	log.Info("--get-title (with newlines removed): " + title)

	if _, ok := youtubeDownloading[guildID]; ok {
		log.Info("Enqueuing (not a new queue)")
		youtubeDownloading[guildID].enqueue(title)
	} else {
		log.Info("Enqueuing into a new queue")
		youtubeDownloading[guildID] = newStringQueue(20)
		youtubeDownloading[guildID].enqueue(title)
	}

	log.Info("Starting download.")
	if err = exec.Command("youtube-dl", "-f", "140", link, "--restrict-filenames").Run(); err != nil {
		log.Info("Error calling youtube-dl command: ", err)
		if s != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Download failed :( Sorry <@%s>", m.Author.ID))
		}
		return
	}
	log.Info("Finished download.")

	youtubeDownloading[guildID].remove(title)

	if _, ok := youtubeQueues[guildID]; ok {
		err = youtubeQueues[guildID].enqueue(title)
		if err != nil {
			if s != nil {
				s.ChannelMessageSend(m.ChannelID, "Queue full. sry m8")
			}
		}
	} else {
		youtubeQueues[guildID] = newStringQueue(20)
		youtubeQueues[guildID].enqueue(title)
	}

	log.Info("File: " + file)

	file = fmt.Sprintf("./%s", file)

	sound := createSound(link)
	sound.Load(file)

	go enqueueSound(authorID, guildID, sound, title, link)

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

/* Loops current songs.
 */
func loopYoutube(s *discordgo.Session, m *discordgo.MessageCreate, g *discordgo.Guild) {
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Looping: %t.", !loops[g.ID]))
	loop(g.ID)
}

/* Autoplays songs.
 */
func autoYoutube(s *discordgo.Session, m *discordgo.MessageCreate, g *discordgo.Guild) {
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Auto playing: %t.", !autos[g.ID]))
	auto(g.ID)
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
	if (!okq || yq.length() == 0) && (!okd || ydq.length() == 0) {
		message = message + "Queue is empty\n"
	}
	message = message + "Looping is set to: " + strconv.FormatBool(loops[g.ID]) + "\n"
	message = message + "Autoplaying is set to: " + strconv.FormatBool(autos[g.ID])
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
	if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		log.Warning("Couldn't delete message.")
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
