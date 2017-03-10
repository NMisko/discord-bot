package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
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
	go remind(time.Duration(i)*time.Second, j, m.ChannelID, m.Author.ID)

	s.ChannelMessageSend(m.ChannelID, "Ok.")
}

func remind(d time.Duration, text string, channelid string, authorid string) {
	time.Sleep(d)
	discord.ChannelMessageSend(channelid, fmt.Sprintf("<@%s> %s", authorid, text))
}
