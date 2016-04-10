package main

import (
    "encoding/binary"
    "fmt"
    "io"
    "os/exec"
    "time"

    log "github.com/Sirupsen/logrus"
    "github.com/bwmarrin/discordgo"
    "github.com/layeh/gopus"
    redis "gopkg.in/redis.v3"
)

var (
    // Map of Guild id's to *Play channels, used for queuing and rate-limiting guilds
    queues map[string]chan *Play = make(map[string]chan *Play)

    // Sound attributes
    AIRHORN_SOUND_RANGE = 0
    KHALED_SOUND_RANGE  = 0
    CENA_SOUND_RANGE    = 0
    ETHAN_SOUND_RANGE   = 0

    // Sound encoding settings
    BITRATE        = 128
    MAX_QUEUE_SIZE = 6

    // Sound Types
    TYPE_AIRHORN = 0
    TYPE_KHALED  = 1
    TYPE_CENA    = 2
    TYPE_ETHAN   = 3

    // Redis client connection (used for stats)
    rcli *redis.Client

    // Owner
    OWNER string

    // Shard (or -1)
    SHARDS []string = make([]string, 0)

    // Commands
    COMMANDS []string = []string{
    /*"!airhorn",
    "!anotha",
    "!anothaone",
    "!cena",
    "!johncena",
    "!eb",
    "!ethanbradberry",
    "!h3h3",
    "!nycto",*/
    }
)

// Play represents an individual use of the !airhorn command
type Play struct {
	GuildID   string
	ChannelID string
	UserID    string
	Sound     *Sound

	// If true, this was a forced play using a specific airhorn sound name
	Forced bool

	// If true, we need to appreciate this value
	Khaled bool
}

// Sound represents a sound clip
type Sound struct {
	Name string

	// Weight adjust how likely it is this song will play, higher = more likely
	Weight int

	// Delay (in milliseconds) for the bot to wait before sending the disconnect request
	PartDelay int

	// Sound Type
	Type int

	// Channel used for the encoder routine
	encodeChan chan []int16

	// Buffer to store encoded PCM packets
	buffer [][]byte
}

func createSound(Name string, Weight int, PartDelay int, Type int) *Sound {
	return &Sound{
		Name:       Name,
		Weight:     Weight,
		PartDelay:  PartDelay,
		Type:       Type,
		encodeChan: make(chan []int16, 10),
		buffer:     make([][]byte, 0),
	}
}

// Array of all the sounds we have
var AIRHORNS []*Sound = []*Sound{
	createSound("default", 1000, 250, TYPE_AIRHORN),
	createSound("reverb", 800, 250, TYPE_AIRHORN),
	createSound("spam", 800, 0, TYPE_AIRHORN),
	createSound("tripletap", 800, 250, TYPE_AIRHORN),
	createSound("fourtap", 800, 250, TYPE_AIRHORN),
	createSound("distant", 500, 250, TYPE_AIRHORN),
	createSound("echo", 500, 250, TYPE_AIRHORN),
	createSound("clownfull", 250, 250, TYPE_AIRHORN),
	createSound("clownshort", 250, 250, TYPE_AIRHORN),
	createSound("clownspam", 250, 0, TYPE_AIRHORN),
	createSound("horn_highfartlong", 200, 250, TYPE_AIRHORN),
	createSound("horn_highfartshort", 200, 250, TYPE_AIRHORN),
	createSound("midshort", 100, 250, TYPE_AIRHORN),
	createSound("truck", 10, 250, TYPE_AIRHORN),
}

var KHALED []*Sound = []*Sound{
	createSound("one", 1, 250, TYPE_KHALED),
	createSound("one_classic", 1, 250, TYPE_KHALED),
	createSound("one_echo", 1, 250, TYPE_KHALED),
}

var CENA []*Sound = []*Sound{
	createSound("airhorn", 1, 250, TYPE_CENA),
	createSound("echo", 1, 250, TYPE_CENA),
	createSound("full", 1, 250, TYPE_CENA),
	createSound("jc", 1, 250, TYPE_CENA),
	createSound("nameis", 1, 250, TYPE_CENA),
	createSound("spam", 1, 250, TYPE_CENA),
}

var ETHAN []*Sound = []*Sound{
	createSound("areyou_classic", 100, 250, TYPE_ETHAN),
	createSound("areyou_condensed", 100, 250, TYPE_ETHAN),
	createSound("areyou_crazy", 100, 250, TYPE_ETHAN),
	createSound("areyou_ethan", 100, 250, TYPE_ETHAN),
	createSound("classic", 100, 250, TYPE_ETHAN),
	createSound("echo", 100, 250, TYPE_ETHAN),
	createSound("high", 100, 250, TYPE_ETHAN),
	createSound("slowandlow", 100, 250, TYPE_ETHAN),
	createSound("cuts", 30, 250, TYPE_ETHAN),
	createSound("beat", 30, 250, TYPE_ETHAN),
	createSound("sodiepop", 1, 250, TYPE_ETHAN),
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
func (s *Sound) Load() error {
	s.encodeChan = make(chan []int16, 10)
	defer close(s.encodeChan)
	go s.Encode()

	var path string
	if s.Type == TYPE_AIRHORN {
		path = fmt.Sprintf("audio/airhorn_%v.wav", s.Name)
	} else if s.Type == TYPE_KHALED {
		path = fmt.Sprintf("audio/another_%v.wav", s.Name)
	} else if s.Type == TYPE_CENA {
		path = fmt.Sprintf("audio/jc_%v.wav", s.Name)
	} else if s.Type == TYPE_ETHAN {
		path = fmt.Sprintf("audio/ethan_%v.wav", s.Name)
	}

	ffmpeg := exec.Command("ffmpeg", "-i", path, "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")
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
			return nil
		}

		if err != nil {
			fmt.Println("error reading from ffmpeg stdout :", err)
			return err
		}

		// write pcm data to the encodeChan
		s.encodeChan <- InBuf
	}
}

// Plays this sound over the specified VoiceConnection
func (s *Sound) Play(vc *discordgo.VoiceConnection) {
	vc.Speaking(true)
	defer vc.Speaking(false)

	for _, buff := range s.buffer {
		vc.OpusSend <- buff
	}
}

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

// Whether a guild id is in this shard
func shardContains(guildid string) bool {
	if len(SHARDS) != 0 {
		ok := false
		for _, shard := range SHARDS {
			if len(guildid) >= 5 && string(guildid[len(guildid)-5]) == shard {
				ok = true
				break
			}
		}
		return ok
	}
	return true
}

// Returns a random sound
func getRandomSound(stype int) *Sound {
	var i int

	if stype == TYPE_AIRHORN {
		number := randomRange(0, AIRHORN_SOUND_RANGE)

		for _, item := range AIRHORNS {
			i += item.Weight

			if number < i {
				return item
			}
		}
	} else if stype == TYPE_KHALED {
		number := randomRange(0, KHALED_SOUND_RANGE)

		for _, item := range KHALED {
			i += item.Weight

			if number < i {
				return item
			}
		}
	} else if stype == TYPE_CENA {
		number := randomRange(0, CENA_SOUND_RANGE)

		for _, item := range CENA {
			i += item.Weight

			if number < i {
				return item
			}
		}
	} else if stype == TYPE_ETHAN {
		number := randomRange(0, ETHAN_SOUND_RANGE)

		for _, item := range ETHAN {
			i += item.Weight

			if number < i {
				return item
			}
		}
	}

	return nil
}

// Enqueues a play into the ratelimit/buffer guild queue
func enqueuePlay(user *discordgo.User, guild *discordgo.Guild, sound *Sound, khaled bool, stype int) {
	// Grab the users voice channel
	channel := getCurrentVoiceChannel(user, guild)
	if channel == nil {
		log.WithFields(log.Fields{
			"user":  user.ID,
			"guild": guild.ID,
		}).Warning("Failed to find channel to play sound in")
		return
	}

	var forced bool = true
	if sound == nil {
		forced = false
		sound = getRandomSound(stype)
	}

	play := &Play{
		GuildID:   guild.ID,
		ChannelID: channel.ID,
		UserID:    user.ID,
		Sound:     sound,
		Forced:    forced,
		Khaled:    khaled,
	}

	// Check if we already have a connection to this guild
	_, exists := queues[guild.ID]

	if exists {
		if len(queues[guild.ID]) < MAX_QUEUE_SIZE {
			queues[guild.ID] <- play
		}
	} else {
		queues[guild.ID] = make(chan *Play, MAX_QUEUE_SIZE)
		playSound(play, nil)
	}
}

func trackSoundStats(play *Play) {
	if rcli == nil {
		return
	}

	_, err := rcli.Pipelined(func(pipe *redis.Pipeline) error {
		var baseChar string

		if play.Forced {
			baseChar = "f"
		} else {
			baseChar = "a"
		}

		base := fmt.Sprintf("airhorn:%s", baseChar)
		pipe.Incr("airhorn:total")
		pipe.Incr(fmt.Sprintf("%s:total", base))
		pipe.Incr(fmt.Sprintf("%s:sound:%s", base, play.Sound.Name))
		pipe.Incr(fmt.Sprintf("%s:user:%s:sound:%s", base, play.UserID, play.Sound.Name))
		pipe.Incr(fmt.Sprintf("%s:guild:%s:sound:%s", base, play.GuildID, play.Sound.Name))
		pipe.Incr(fmt.Sprintf("%s:guild:%s:chan:%s:sound:%s", base, play.GuildID, play.ChannelID, play.Sound.Name))
		pipe.SAdd(fmt.Sprintf("%s:users", base), play.UserID)
		pipe.SAdd(fmt.Sprintf("%s:guilds", base), play.GuildID)
		pipe.SAdd(fmt.Sprintf("%s:channels", base), play.ChannelID)
		return nil
	})

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warning("Failed to track stats in redis")
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
	time.Sleep(time.Millisecond * 32)

	// If we're appreciating this sound, lets play some DJ KHALLLLLEEEEDDDD
	if play.Khaled {
		dj := getRandomSound(TYPE_KHALED)
		dj.Play(vc)
	}

	// Track stats for this play in redis
	go trackSoundStats(play)

	// Play the sound
	play.Sound.Play(vc)

	// If there is another song in the queue, recurse and play that
	if len(queues[play.GuildID]) > 0 {
		play := <-queues[play.GuildID]
		playSound(play, vc)
		return nil
	}

	// If the queue is empty, delete it
	time.Sleep(time.Millisecond * time.Duration(play.Sound.PartDelay))
	delete(queues, play.GuildID)
	vc.Disconnect()
	return nil
}
