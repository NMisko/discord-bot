
/*	This file contains all high level commands of the bot. 


package main
import (
    "fmt"
    "math/rand"
    "github.com/bwmarrin/discordgo"
    "strings"
    "strconv"
    "os"
    "os/exec"
    "io/ioutil"
    "image"
    "image/jpeg"
    "image/png"
    "bytes"
    "regexp"

    log "github.com/Sirupsen/logrus"
    yt "github.com/kkdai/youtube"
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
    PATH_TO_CLASSIFY_EXEC = "tensorflow/imagenet/classify_image.py"
)

/* 	Sends messages with information about the given players LoL rank.
	Takes input of form: <Username> <Region>
*/
func elo(input []string, s *discordgo.Session, m *discordgo.MessageCreate, riotkey string) {
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
    summoner := GetSummoner(name, region, riotkey)
    if (summoner.ID == 0) {
        s.ChannelMessageSend(m.ChannelID, "Could not find player.")
        return
    }
    league := GetLeague(strconv.Itoa(summoner.ID), region, riotkey)
    if(league.Tier != "Unranked") {
        entry := league.Entry[0]
        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("**%s %s** %sLP", capitalize(strings.ToLower(league.Tier)), entry.Division, strconv.Itoa(entry.LP)))
        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Wins: **%s** Losses: **%s**", strconv.Itoa(entry.Wins), strconv.Itoa(entry.Losses)))
    } else {
        s.ChannelMessageSend(m.ChannelID, "**Unranked**")
    }
    switch name {
        case "uznick": s.ChannelMessageSend(m.ChannelID, "But deserves Challenjour, Kappa.")
        case "mehdid": s.ChannelMessageSend(m.ChannelID, "Also, best Amumu EUW.")
        case "flakelol": s.ChannelMessageSend(m.ChannelID, "Also, best Shen EUW.")
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

/* 	Answers the question 8-Ball Style randomly.
*/
func jarvis(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	if (string(input[len(input)-1][len(input[len(input)-1])-1]) == "?") {
	s.ChannelMessageSend(m.ChannelID, JARVIS_ANSWERS[rand.Intn(len(JARVIS_ANSWERS))])
	}
}

/*	Uploads a random coin face.
*/
func coin(s *discordgo.Session, m *discordgo.MessageCreate) {
	file, err := os.Open(COIN_FACES_PATHS[rand.Intn(len(COIN_FACES_PATHS))])
	if err != nil { log.Warning(err) }
	s.ChannelFileSend(m.ChannelID, "Coin.png", file)
}

/*	Sends a message with the result of a six-sided dice roll.
*/
func dice(s *discordgo.Session, m *discordgo.MessageCreate) {
	answers := []string { "1", "2", "3", "4", "5", "6",}
	s.ChannelMessageSend(m.ChannelID, answers[rand.Intn(len(answers))])
}

/*	Downloads the linked picture and categorizes it with an ANN. Sends messages containing the different possible results and their probabilities.
*/
func classifyImage(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
    local := fmt.Sprintf("temp/%s.jpg", m.ID)
    localConverted := fmt.Sprintf("temp/%sC.jpg", m.ID)
    image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

    p := NewParser(input)
    if (!p.nextToken()) {return}

    log.Info("Starting classification of ", input)

    //Download file into local (hopefully unique file)
    err := DownloadFile(p.Token, local)
    if(err != nil) {
        log.Warning(err)
        return
    }

    //Read local file again
    imgBuffer, err := ioutil.ReadFile(local)
    if err != nil {
            log.Warning("Can't read local file")
            return
    }

    reader := bytes.NewReader(imgBuffer)

    //Decode into image type
    img, _, err := image.Decode(reader)
	if err != nil {
		log.Println(err)
		return
	}

    //Create file that'll contain converted image
    out, err := os.Create(localConverted)
	if err != nil {
		log.Println(err)
		return
	}
	defer out.Close()

    //Convert image to jpeg
    err = jpeg.Encode(out, img, nil)
    if err != nil {
        log.Println(err)
        return
    }

    //Classify image
    cmd := exec.Command("python", PATH_TO_CLASSIFY_EXEC, "--image_file", localConverted)
    response, err := cmd.Output()

    log.Info(string(response))
    log.Warning(err)
    s.ChannelMessageSend(m.ChannelID, "Results: ")

    message := ""
    nameRegex := regexp.MustCompile(`([[a-zA-Z]|,|\s)*[a-zA-Z]`)
    scoreRegex := regexp.MustCompile("[0-9][0-9]")

    lines := strings.Split(string(response), "\n")
    for i := 0; i < len(lines)-1; i++ {
        name := nameRegex.FindString(lines[i])
        names := strings.Split(name, ",")
        for j := 0; j < len(names); j++ {
                if(j < len(names)-1) {message += fmt.Sprintf("%s,", bold(capitalize(strings.Trim(names[j], " "))))
                } else {message += bold(capitalize(strings.Trim(names[j], " ")))}
        }
        score := scoreRegex.FindString(lines[i])
        message += fmt.Sprintf(" (%s%%)\n", score)
    }
    s.ChannelMessageSend(m.ChannelID, message)

    //delete local images
    err = os.Remove(local)
    if err != nil {log.Warning(err)}
    err = os.Remove(localConverted)
    if err != nil {log.Warning(err)}

    log.Info("Finished classification.")
}

/* Queues up a Youtube video, whose sound is played in the voice channel the command caller is in. Downloads the entire Youtube video locally, which might take a while, based on the internet connection.
*/
func queueYoutube(input []string, s *discordgo.Session, m *discordgo.MessageCreate, g *discordgo.Guild) {
    if (len(input) < 1) {
        s.ChannelMessageSend(m.ChannelID, "Usage: !play <link>")
        return
    }
    link := input[0]
    log.Info("Downloading ", link)

    y := yt.NewYoutube()
    err := y.DecodeURL(link)
    if err != nil {
        message := fmt.Sprintf("Invalid link: %s", err)
        s.ChannelMessageSend(m.ChannelID, message)
        return
    }
    y.StartDownload("./temp/")
    title := y.StreamList[0]["title"]
    file := fmt.Sprintf("./temp/%s.mp4", title)

    sound := createSound(link)
    sound.Load(file)

    log.Info("Enqueuing video...")
    go enqueuePlay(m.Author, g, sound)
    log.Info("Done.")

    log.Info("Trying to remove ", file)
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
