package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	log "github.com/Sirupsen/logrus"
)

var (
	reminders [20]Reminder
)

type Reminder struct {
	Moment    time.Time
	Text      string
	ChannelID string
	AuthorID  string
}

func remindme(input []string, s *discordgo.Session, m *discordgo.MessageCreate) {
	var i int
	var j string
	if len(input) < 2 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("usage: !remindme <seconds> <message> or !rm <seconds> <message>"))
		return
	}
	_, err := fmt.Sscanf(input[0], "%d", &i)
	j = strings.Join(input[1:], " ")

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("usage: !remindme <seconds> <message> or !rm <seconds> <message>"))
		return
	}

	addReminder(time.Duration(i)*time.Second, j, m.ChannelID, m.Author.ID)
	s.ChannelMessageSend(m.ChannelID, "Ok.")
}

func reminderService(t time.Ticker) {
	for {
		<-t.C
		for i, r := range reminders {
			if !r.Moment.IsZero() {
				if time.Now().After(r.Moment) {
					discord.ChannelMessageSend(r.ChannelID, fmt.Sprintf("<@%s> %s", r.AuthorID, r.Text))
					reminders[i] = Reminder{}
				}
			}
		}
	}
}

func addReminder(d time.Duration, text string, channelid string, authorid string) {
	addedreminder := false
	log.Info("adding reminder ", d, " : ", text)
	future := time.Now().Add(d)
	log.Info(future)
	for i, r := range reminders {
		if r.Moment.IsZero() {
			reminders[i] = Reminder{future, text, channelid, authorid}
			addedreminder = true
			break
		}
	}
	if !addedreminder {
		discord.ChannelMessageSend(channelid, fmt.Sprintf("Too many reminders at once. NotLikeThis"))
	}
}
