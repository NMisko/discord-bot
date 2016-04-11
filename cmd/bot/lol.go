package main

import (
	"fmt"
	"regexp"

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
    if (ok) {lp = scrape.Text(node)} else {log.Info("Could not find winratio.")}

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
