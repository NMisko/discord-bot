package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

/* Struct used to unmarshal the data gotten from the RiotGames API
 */
type RiotSummoner struct {
	Name   string `json:"name"`
	ID     int    `json:"id"`
	Status struct {
		Code int `json:"status_code"`
	} `json:"status"`
}

/* Gets information about a summoner from the RiotGames API. */
func GetSummoner(summoner string, region string, key string) RiotSummoner {
	jsonMessage := riotApiCall(fmt.Sprintf("/lol/summoner/v3/summoners/by-name/%s", summoner), region, key)
	var w RiotSummoner
	json.Unmarshal(jsonMessage, &w)

	return w
}

/* 	Contains data about a league tier.
Used to unmarshall the data gotten from the RiotGames API
*/
type RiotLeagues struct {
	Type string `json:"queueType"`
	Rank string `json:"rank"`
	Tier string `json:"tier"`
}

/* Returns the league of the given summoner. Use 'GetSummoner' to get the summonerid of a summoner (it's not the username).
 */
func GetLeague(summonerid string, region string, key string) []RiotLeagues {
	jsonMessage := riotApiCall(fmt.Sprintf("/lol/league/v3/positions/by-summoner/%s", summonerid), region, key)

	var w []RiotLeagues
	json.Unmarshal(jsonMessage, &w)

	if len(w) < 1 {
		return []RiotLeagues{}
	}
	return w
}

/*	Generic method for a call to the RiotGames API.
 */
func riotApiCall(call string, region string, key string) []byte {
	url := fmt.Sprintf("https://%s.api.riotgames.com%s?api_key=%s", region, call, key)
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
