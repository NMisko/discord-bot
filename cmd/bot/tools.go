package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html"

	log "github.com/Sirupsen/logrus"
)

func stripChars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return r
		}
		return -1
	}, str)
}

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

func DownloadFile(url string, local string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	file, err := os.Create(local)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}

func DWebsite(url string) (out string) {
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

func contains(key string, options []string) bool {
	for _, item := range options {
		if item == key {
			return true
		}
	}
	return false
}

func icontains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[n:]
}

func lowercase(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

func bold(s string) string {
	if s != "" {
		return fmt.Sprintf("**%s**", s)
	} else {
		return "[no data]"
	}
}

type Parser struct {
	tokenIndex int
	tokens     []string
	Token      string
}

func (s *Parser) nextToken() bool {
	s.tokenIndex = s.tokenIndex + 1
	if s.tokenIndex >= len(s.tokens) {
		return false
	} else {
		s.Token = s.tokens[s.tokenIndex]
	}
	return true
}

func NewParser(list []string) *Parser {
	return &Parser{-1, list, ""}
}

type StringQueue struct {
	stack []string
	head  int //index of the first element
	tail  int //index of the last element
	size  int //total size
	mut   sync.Mutex
}

func newStringQueue(size int) *StringQueue {
	q := StringQueue{}
	q.stack = make([]string, size)
	q.head = 0
	q.tail = -1
	q.size = size
	return &q
}

//need to implement what happens if queue full
func (q *StringQueue) enqueue(s string) error {
	q.mut.Lock()
	//Move queue back to front
	if q.tail == q.size-1 {
		var newstack = make([]string, q.size)
		j := 0
		for i := q.head; i <= q.tail; i++ {
			newstack[j] = q.stack[i]
			j++
		}
		q.tail = q.tail - q.head
		q.head = 0
		q.stack = newstack
	}
	if q.tail == q.size-1 {
		q.mut.Unlock()
		return errors.New("Queue full. Cannot enqueue anymore.")
	} else {
		q.stack[q.tail+1] = s
		q.tail++
		q.mut.Unlock()
		return nil
	}
}

func (q *StringQueue) peek() string {
	q.mut.Lock()
	out := q.stack[q.head]

	q.mut.Unlock()
	return out
}

func (q *StringQueue) dequeue() (string, error) {
	q.mut.Lock()
	if q.tail < q.head {

		q.mut.Unlock()
		return "", errors.New("Queue already empty. Cannot dequeue.")
	} else {
		out := q.stack[q.head]
		q.stack[q.head] = ""
		q.head++

		q.mut.Unlock()
		return out, nil
	}
}

//removes first occurrence of title
func (q *StringQueue) remove(s string) {
	q.mut.Lock()
	for i := q.head; i <= q.tail; i++ {
		if q.stack[i] == s {
			for j := i; j <= q.tail; j++ {
				if j+1 < q.tail-q.head+1 && q.stack[j+1] != "" {
					q.stack[j] = q.stack[j+1]
				}
			}
			q.tail = q.tail - 1
			q.mut.Unlock()
			return
		}
	}
	q.mut.Unlock()
}

func (q *StringQueue) length() int {
	q.mut.Lock()
	out := q.tail - q.head + 1
	q.mut.Unlock()

	return out
}

func (q *StringQueue) toArray() []string {
	q.mut.Lock()
	out := q.stack[q.head:(q.tail + 1)]
	q.mut.Unlock()

	return out
}
