package main
import (
    "fmt"
    "math/rand"
    "github.com/bwmarrin/discordgo"
    "strings"
    "os"
    log "github.com/Sirupsen/logrus"
)
var (
    DEFAULT_LOL_REGION string = "euw"
    COIN_FACES_PATHS = []string{
        "images/Head.png",
        "images/Tail.png",
    }
    JARVIS_ANSWERS = []string {
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
)

// name [region]
func elo(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	region := DEFAULT_LOL_REGION

    p := NewParser(input)
    if(!p.nextToken()) { return }
    p.nextToken()
    name := strings.ToLower(p.Token)
    p.nextToken()
    switch (strings.ToLower(p.Token)) {
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
        case "mehdid": s.ChannelMessageSend(m.ChannelID, "Also, best Amumu EUW.")
        case "flakelol": s.ChannelMessageSend(m.ChannelID, "Also, best Shen EUW.")
    }
}

// city [time|country] [country]
func weather(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
    setTime := false
    country := "de"
    forecastStart := 0; //Today
    forecastRange := 3; //0,1,2
    days := []string{"Today", "Tomorrow", "In two days", "In three days", "In four days", "In five days", "In six days", "In a week"}

    p := NewParser(input)
    if (!p.nextToken()) {return}
    city := p.Token

    if(p.nextToken()) {
        switch strings.ToLower(p.Token){
            case "today": forecastRange = 1; setTime = true
            case "tomorrow": forecastRange = 1; forecastStart = 1; setTime = true
            default: country = p.Token
        }
    }
    if(p.nextToken() && setTime) {country = p.Token}

    weather := GetWeather(city, country)
    if ((forecastStart+forecastRange) > len(weather.Forecast.Time)) {
        s.ChannelMessageSend(m.ChannelID, "Can't get information in that timerange or for that city.")
        return
    }
    if (len(weather.Forecast.Time) > forecastRange) {
        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Weather forecast for: **%s**", capitalize(city)))
        for i := forecastStart; i < (forecastStart+forecastRange); i++ {
            d := weather.Forecast.Time[i]
            text := fmt.Sprintf("%s there will be %s with a temperature ranging from %s°C to %s°C and a %s from the %s.", days[i], bold(d.Symbol.Name), bold(d.Temperature.Min), bold(d.Temperature.Max), bold(strings.ToLower(d.WindSpeed.Name)), bold(strings.ToLower(d.WindDirection.Name)))
            s.ChannelMessageSend(m.ChannelID, text)
        }
    } else {
        s.ChannelMessageSend(m.ChannelID, "Can't find information about this city.")
        return
    }
}

func jarvis(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if (string(input[len(input)-1][len(input[len(input)-1])-1]) == "?") {
	s.ChannelMessageSend(m.ChannelID, JARVIS_ANSWERS[rand.Intn(len(JARVIS_ANSWERS))])
	}
}

func coin(s *discordgo.Session, m *discordgo.MessageCreate) {
	file, err := os.Open(COIN_FACES_PATHS[rand.Intn(len(COIN_FACES_PATHS))])
	if err != nil { log.Warning(err) }
	s.ChannelFileSend(m.ChannelID, "Coin.png", file)
}

func dice(s *discordgo.Session, m *discordgo.MessageCreate) {
	answers := []string { "1", "2", "3", "4", "5", "6",}
	s.ChannelMessageSend(m.ChannelID, answers[rand.Intn(len(answers))])
}
