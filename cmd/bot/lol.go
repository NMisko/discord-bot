package main

import (
	"fmt"
	"regexp"
	"net/http"
	"encoding/json"
	"io/ioutil"


	"golang.org/x/net/html"
	"github.com/yhat/scrape"
	log "github.com/Sirupsen/logrus"
)

type Summoner struct {
    rank string
    rankImage string
    wins string
    losses string
    winratio string
	lp string
}

func GetSummonerElo(summonername string, region string) Summoner {
    var rank, rankImage, wins, losses, winratio, lp string = "","","","","",""
    var root *html.Node

    if (region == "") {
        root = ReadWebsite(fmt.Sprintf("http://op.gg/summoner/userName=%s", summonername));
    } else {
        root = ReadWebsite(fmt.Sprintf("http://%s.op.gg/summoner/userName=%s", region, summonername));
    }

    node, ok := scrape.Find(root, scrape.ByClass("TierRank"))
    if (ok) {rank = scrape.Text(node)} else {log.Info("Could not find rank.")}

	node, ok = scrape.Find(root, scrape.ByClass("Medal"))
    if (ok) {
        node, ok = scrape.Find(node, scrape.ByClass("Image"))
        if (ok) {
            rankImage = scrape.Attr(node, "src")
        } else {log.Info("Could not find rankImage. 02")}
    } else {log.Info("Could not find rankImage. 01")}

    node, ok = scrape.Find(root, scrape.ByClass("wins"))
    if (ok) {wins = scrape.Text(node)} else {log.Info("Could not find wins.")}

    node, ok = scrape.Find(root, scrape.ByClass("losses"))
    if (ok) {losses = scrape.Text(node)} else {log.Info("Could not find losses.")}

    node, ok = scrape.Find(root, scrape.ByClass("winratio"))
    if (ok) {winratio = scrape.Text(node)} else {log.Info("Could not find winratio.")}

    node, ok = scrape.Find(root, scrape.ByClass("LeaguePoints"))
    if (ok) {lp = scrape.Text(node)} else {log.Info("Could not find lp.")}

    numbers := regexp.MustCompile("[0-9]+")

    winsarr := numbers.FindAllString(wins,-1)
    lossesarr := numbers.FindAllString(losses,-1)
    winratioarr := numbers.FindAllString(winratio,-1)
    lparr := numbers.FindAllString(lp,-1)

	if (len(winsarr) > 0) {wins = winsarr[0]}
	if (len(lossesarr) > 0) {losses = lossesarr[0]}
	if (len(winratioarr) > 0) {winratio = winratioarr[0]}
	if (len(lparr) > 0) {lp = lparr[0]}

    summoner := Summoner{rank, rankImage, wins, losses, winratio, lp}
    log.Info(summoner)
    return summoner
}

type RiotSummoner struct {
    Name string `json:"name"`
    ID int `json:"id"`
}


func GetSummoner(summoner string, region string, key string) RiotSummoner {
    jsonMessage := riotApiCall(fmt.Sprintf("/api/lol/%s/v1.4/summoner/by-name/%s", region, summoner), region, key)
    var w map[string]RiotSummoner
    json.Unmarshal(jsonMessage, &w)
    log.Info(w)

    return w[lowercase(summoner)]
}

type RiotLeague struct {
    Tier string `json:"tier"`
    Name string `json:"name"`
    Entry []struct {
        LP int `json:"leaguePoints"`
        Division string `json:"division"`
        Wins int `json:"wins"`
        Losses int `json:"losses"`
    } `json:"entries"`
}

func GetLeague(summonerid string, region string, key string) RiotLeague {
	jsonMessage := riotApiCall(fmt.Sprintf("/api/lol/%s/v2.5/league/by-summoner/%s/entry", region, summonerid), region, key)

	var w map[string][]RiotLeague
	json.Unmarshal(jsonMessage, &w)
	log.Info(w)

	if(len(w[summonerid]) < 1) {
		out := RiotLeague{Tier: "Unranked"}
		return out
	}
    return w[summonerid][0]
}

func riotApiCall(call string, region string, key string) []byte {
	url := fmt.Sprintf("https://%s.api.pvp.net%s?api_key=%s", region, call, key)
	log.Info(url)
	resp, err := http.Get(url)

	if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Warning("Failed to GET Summoner: ")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return body
}
