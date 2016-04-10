package main

import (
    "net/http"
    "io/ioutil"
    "math/rand"
    "time"
    "strconv"
    "strings"

    log "github.com/Sirupsen/logrus"
)

func DWebsite(url string) (out string){
    //client := &http.Client{}
    // resp, err := client.Get("url")
    resp, err := http.Get(url)

    if err != nil {
            log.WithFields(log.Fields{
                "error": err,
            }).Warning("Failed to GET url.")
            return ""
    }

    defer resp.Body.Close()

    htmlData, err := ioutil.ReadAll(resp.Body)

    if err != nil {
            log.WithFields(log.Fields{
                "error": err,
            }).Warning("Failed to read response body.")
            return ""
    }

    return string(htmlData)
}

// Returns a random integer between min and max
func randomRange(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max-min) + min
}

func convert(b []byte) string {
	s := make([]string, len(b))
	for i := range b {
		s[i] = strconv.Itoa(int(b[i]))
	}
	return strings.Join(s, ",")
}

func scontains(key string, options ...string) bool {
	for _, item := range options {
		if item == key {
			return true
		}
	}
	return false
}
