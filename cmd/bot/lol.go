package main

import (
	"fmt"
	"strings"
	"regexp"

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
	var htmlData string
	if (region == "") {
		htmlData = DWebsite(fmt.Sprintf("http://op.gg/summoner/userName=%s", summonername));
	} else {
    	htmlData = DWebsite(fmt.Sprintf("http://%s.op.gg/summoner/userName=%s", region, summonername));
	}
	var wins string
	var losses string
	var winratio string
	var rank string
	var rankImage string
	var lp string


    userExistsRegexp := regexp.MustCompile("This summoner is not registered at")
    userExists :=  userExistsRegexp.FindAllString(htmlData, -1)
    if (len(userExists) > 0) {
        log.Warning("User ", summonername, " not found.")
        return Summoner{"","","","","", ""}
    }

    tierBoxRegexp := regexp.MustCompile("(<div class=\"TierBox Box).*(/div).*(/div)")
    tierBox := tierBoxRegexp.FindAllString(htmlData, -1)

    numbers := regexp.MustCompile("[0-9]+")
    winsRegexp := regexp.MustCompile("(wins\">).*W(</span)")
    winsbox := winsRegexp.FindAllString(htmlData, -1)
	if (len(winsbox) > 0) {
		wins = numbers.FindAllString(winsbox[0], -1)[0]
    } else {
		wins = ""
	}
    lossesRegexp := regexp.MustCompile("(losses\").*L(</span)")
    lossesbox := lossesRegexp.FindAllString(htmlData, -1)

	if (len(lossesbox) > 0) {
		losses = numbers.FindAllString(lossesbox[0], -1)[0]
    } else {
		losses = ""
	}
    winratioRegexp := regexp.MustCompile("(winratio).*(</span)")
    winratiobox := winratioRegexp.FindAllString(htmlData, -1)

	if (len(winratiobox) > 0) {
		winratio = numbers.FindAllString(winratiobox[0], -1)[0]
    } else {
		winratio = ""
	}
    rankImageRegexp := regexp.MustCompile("sk2.op.gg/images/medals/.*.png")

	var rankImagearr []string
	if (len(tierBox) > 0) {
		rankImagearr = rankImageRegexp.FindAllString(tierBox[0], -1)
		if (len(rankImagearr) > 0) {
			rankImage = rankImagearr[0]
		} else {
			rankImage = ""
		}
	} else {
		rankImage = ""
	}

    rankRegexp := regexp.MustCompile("((Bronze|Silver|Gold|Platinum|Diamond) [1-5])|(Unranked|Master|Challenger)")

	var rankarr []string
	if (len(tierBox) > 0) {
		rankarr = rankRegexp.FindAllString(tierBox[0], -1)

		if (len(rankarr) > 0) {
			rank = rankarr[0]
		} else {
			rank = ""
		}
	} else {
		rank = ""
	}

	lpRegexp := regexp.MustCompile("[0-9]* LP")
	lpbox := lpRegexp.FindAllString(htmlData, -1)
	log.Info(lpbox)
	if (len(lpbox) > 0) {
		lp = numbers.FindAllString(lpbox[0], -1)[0]
	} else {
		lp = ""
	}
	log.Info(lp)

	winrarr := []string{winratio,"%"}
	log.Info(Summoner{rank, rankImage, wins, losses, strings.Join(winrarr, ""), lp})
    return Summoner{rank, rankImage, wins, losses, strings.Join(winrarr, ""), lp}
}
