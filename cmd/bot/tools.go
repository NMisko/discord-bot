package main

import (
    "net/http"
    "io/ioutil"
    "math/rand"
    "time"
    "strconv"
    "strings"
    "unicode"
    "unicode/utf8"

    "golang.org/x/net/html"

    log "github.com/Sirupsen/logrus"
)

func ReadWebsite(url string) (out *html.Node) {
    resp, err := http.Get(url)

    if err != nil {
            log.WithFields(log.Fields{
                "error": err,
            }).Warning("Failed to GET url: ", url)
            return nil
    }
    defer resp.Body.Close()

    root, err := html.Parse(resp.Body)

    if err != nil {
            log.WithFields(log.Fields{
                "error": err,
            }).Warning("Failed to Parse: ", resp.Body)
            return nil
    }

    return root
}

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

func capitalize(s string) string {
    if s == "" { return s }
    r, n := utf8.DecodeRuneInString(s)
    return string(unicode.ToTitle(r)) + s[n:]
}

func lowercase(s string) string {
    if s == "" { return s }
    r, n := utf8.DecodeRuneInString(s)
    return string(unicode.ToLower(r)) + s[n:]
}

type Parser struct {
    tokenIndex int
    tokens []string
    Token string
}

func (s *Parser) nextToken() bool {
    s.tokenIndex = s.tokenIndex + 1
    if(s.tokenIndex >= len(s.tokens)) {
        return false
    } else {
        s.Token = s.tokens[s.tokenIndex]
    }
    return true
}

func NewParser(list []string) *Parser {
    return &Parser{-1, list, ""}
}
