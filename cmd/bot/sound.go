package main

/** Structure:
sound := createSound(pathToSoundFile)
go enqueueSound(user, guild, sound, title)
	| enqueuePlay (creates "play" out of sound) <--
		| queues[play.GuildID] <- play           |
		 or                                      |
	 	| playSound(play)                        |  		 <--
			Play(play)                           |             |
			if looping ---------------------------    		     |
			if there's another "play" in the queue, play it ---|
**/

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

var (
	// Replace all this with struct
	// Map of Guild id's to *Play channels, used for queuing and rate-limiting guilds
	queues     map[string]chan *Play = make(map[string]chan *Play)
	songs      map[string]string     = make(map[string]string)
	skips      map[string]bool       = make(map[string]bool)
	loops      map[string]bool       = make(map[string]bool)
	autos      map[string]bool       = make(map[string]bool)
	lastPlayed map[string]*Play      = make(map[string]*Play)
	current    map[string]string     = make(map[string]string)

	// Owner
	OWNER string

	// Sound encoding settings
	BITRATE        = 128
	MAX_QUEUE_SIZE = 6

	//Shard (or -1)
	SHARDS []string = make([]string, 0)
)

// Play represents an individual use of the !airhorn command
type Play struct {
	GuildID   string
	ChannelID string
	UserID    string
	Sound     *Sound
	Title     string
	Url       string
}

// Sound represents a sound clip
type Sound struct {
	Name string

	// Channel used for the encoder routine
	encodeChan chan []int16

	// Buffer to store encoded PCM packets
	buffer [][]byte
}

func createSound(Name string) *Sound {
	return &Sound{
		Name:       Name,
		encodeChan: make(chan []int16, 10),
		buffer:     make([][]byte, 0),
	}
}

// Enqueues a play into the ratelimit/buffer guild queue
func enqueueSound(userID string, guildID string, sound *Sound, title string, url string) {
	skips[guildID] = false

	guild, _ := discord.State.Guild(guildID)
	member, err := discord.State.Member(guildID, userID)
	if err != nil {
		return
	}

	user := member.User

	// Grab the users voice channel
	channel := getCurrentVoiceChannel(user, guild)
	if channel == nil {
		log.WithFields(log.Fields{
			"user":  userID,
			"guild": guildID,
		}).Warning("Failed to find channel to play sound in")
		return
	}

	play := &Play{
		GuildID:   guildID,
		ChannelID: channel.ID,
		UserID:    userID,
		Sound:     sound,
		Title:     title,
		Url:       url,
	}

	lastPlayed[guildID] = play

	if autos[play.GuildID] {
		go autoplay(play.GuildID)
	}

	enqueuePlay(play)
}

func enqueuePlay(play *Play) {
	// Check if we already have a connection to this guild
	_, exists := queues[play.GuildID]

	if exists {
		if len(queues[play.GuildID]) < MAX_QUEUE_SIZE {
			queues[play.GuildID] <- play
		}
	} else {
		queues[play.GuildID] = make(chan *Play, MAX_QUEUE_SIZE)
		playSound(play, nil)
	}
}

// Play a sound
func playSound(play *Play, vc *discordgo.VoiceConnection) (err error) {
	log.WithFields(log.Fields{
		"play": play,
	}).Info("Playing sound")

	if vc == nil {
		vc, err = discord.ChannelVoiceJoin(play.GuildID, play.ChannelID, false, false)
		// vc.Receive = false
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Failed to play sound")
			delete(queues, play.GuildID)
			return err
		}
	}
	// If we need to change channels, do that now
	if vc.ChannelID != play.ChannelID {
		vc.ChangeChannel(play.ChannelID, false, false)
		time.Sleep(time.Millisecond * 125)
	}
	// Sleep for a specified amount of time before playing the sound
	time.Sleep(time.Millisecond * 500)

	if autos[play.GuildID] {
		go autoplay(play.GuildID)
	}

	// Play the sound
	play.Sound.Play(vc)
	youtubeQueues[play.GuildID].remove(play.Title)

	// loop
	if loops[play.GuildID] && !skips[play.GuildID] {
		go enqueuePlay(play)
	}

	if skips[play.GuildID] {
		skips[play.GuildID] = false
	} else if loops[play.GuildID] {
		youtubeQueues[play.GuildID].enqueue(play.Title)
	}

	time.Sleep(time.Millisecond * time.Duration(500))

	// If there is another song in the queue, recurse and play that
	if len(queues[play.GuildID]) > 0 {
		play := <-queues[play.GuildID]
		playSound(play, vc)
		return nil
	}

	// If the queue is empty, delete it
	time.Sleep(time.Millisecond * time.Duration(500))
	delete(queues, play.GuildID)

	vc.Disconnect()
	return nil
}

// Plays this sound over the specified VoiceConnection
func (s *Sound) Play(vc *discordgo.VoiceConnection) {
	vc.Speaking(true)
	defer vc.Speaking(false)

	// t := time.NewTicker(time.Duration(1) * time.Second)
	// go s.control(t)

	for _, buff := range s.buffer {
		vc.OpusSend <- buff
		if skips[vc.GuildID] == true {
			//skips[vc.GuildID] = false
			break
		}
	}
}

// Encode reads data from ffmpeg and encodes it using gopus
func (s *Sound) Encode() {
	encoder, err := gopus.NewEncoder(48000, 2, gopus.Audio)
	if err != nil {
		fmt.Println("NewEncoder Error:", err)
		return
	}

	encoder.SetBitrate(BITRATE * 1000)
	encoder.SetApplication(gopus.Audio)

	for {
		pcm, ok := <-s.encodeChan
		if !ok {
			// if chan closed, exit
			return
		}

		// try encoding pcm frame with Opus
		opus, err := encoder.Encode(pcm, 960, 960*2*2)
		if err != nil {
			fmt.Println("Encoding Error:", err)
			return
		}

		// Append the PCM frame to our buffer
		s.buffer = append(s.buffer, opus)
	}
}

// Load attempts to load and encode a sound file from disk
func (s *Sound) Load(path string) error {
	s.encodeChan = make(chan []int16, 10)
	defer close(s.encodeChan)
	go s.Encode()

	newpath := fmt.Sprintf("%sout.m4a", path) //overwriting file doesn't work to decrease volume
	if _, err := exec.Command("ffmpeg", "-y", "-i", path, "-af", "volume=0.1", newpath).Output(); err != nil {
		fmt.Println("Error decreasing volume:", err)
	}

	ffmpeg := exec.Command("ffmpeg", "-i", newpath, "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")
	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		fmt.Println("StdoutPipe Error:", err)
		return err
	}

	err = ffmpeg.Start()
	if err != nil {
		fmt.Println("RunStart Error:", err)
		return err
	}

	for {
		// read data from ffmpeg stdout
		InBuf := make([]int16, 960*2)
		err = binary.Read(stdout, binary.LittleEndian, &InBuf)

		// If this is the end of the file, just return
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			if err = os.Remove(newpath); err != nil {
				log.Warning("Failed removal of sound file with changed volume: ", err)
			}
			return nil
		}

		if err != nil {
			fmt.Println("error reading from ffmpeg stdout :", err)
			return err
		}

		// write pcm data to the encodeChan
		s.encodeChan <- InBuf
	}

	return nil
}

// func (s *Sound) control(t *time.Ticker) {
//     for {
//         <- t.C
//
//     }
// }

// Attempts to find the current users voice channel inside a given guild
func getCurrentVoiceChannel(user *discordgo.User, guild *discordgo.Guild) *discordgo.Channel {
	for _, vs := range guild.VoiceStates {
		if vs.UserID == user.ID {
			channel, _ := discord.State.Channel(vs.ChannelID)
			return channel
		}
	}
	return nil
}

func next(GuildID string) {
	skips[GuildID] = true
}

func loop(GuildID string) {
	loops[GuildID] = !loops[GuildID]
}

func auto(GuildID string) {
	autos[GuildID] = !autos[GuildID]
	if autos[GuildID] {
		go autoplay(GuildID)
	}
}

type YoutubeRelatedVideos struct {
	Items []struct {
		Id struct {
			VideoId string `json:"videoId"`
		} `json:"id"`
	} `json:"items"`
}

func autoplay(GuildID string) {
	// check if auto playing is necessary (queue is smaller than 3)
	// then get related video to last video in queue and queue it
	if lastPlayed[GuildID] != nil && (youtubeDownloading[GuildID] == nil || youtubeDownloading[GuildID].length() == 0) && (youtubeQueues[GuildID] == nil || youtubeQueues[GuildID].length() < 3) {
		play := lastPlayed[GuildID]

		id := strings.Split(play.Url, "watch?v=")[1]
		template := "https://www.googleapis.com/youtube/v3/search?part=snippet&relatedToVideoId=%s&type=video&key=AIzaSyBDsvj-LjQzjOk1yBLeKwKxjYj5TIjpk1g"

		url := fmt.Sprintf(template, id)
		resp, err := http.Get(url)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Warning("Failed to GET Youtubevideo suggested videos: ")
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		var y YoutubeRelatedVideos
		json.Unmarshal(body, &y)

		// Random index, to avoid circling
		newId := y.Items[rand.Intn(3)].Id.VideoId
		log.Info("Autoplaying " + string(newId))

		queueYoutube([]string{fmt.Sprintf("https://www.youtube.com/watch?v=%s", newId)}, nil, nil, GuildID, play.UserID)
	}
}
