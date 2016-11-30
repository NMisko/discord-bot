package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
	"github.com/hraban/opus"
	"github.com/xlab/closer"
	"github.com/xlab/pocketsphinx-go/sphinx"
)

var (
	voiceConnections map[string]*discordgo.VoiceConnection = make(map[string]*discordgo.VoiceConnection)
)

const (
	//samplesPerChannel = 512
	sampleRate = 48000
	channels   = 2
	hmm        = "/usr/local/share/pocketsphinx/model/en-us/en-us"
	dict       = "/usr/local/share/pocketsphinx/model/en-us/cmudict-en-us.dict"
	lm         = "/usr/local/share/pocketsphinx/model/en-us/en-us.lm.bin"
	outraw     = "temp"
)

func joinChannel(user *discordgo.User, guild *discordgo.Guild) {
	channel := getCurrentVoiceChannel(user, guild)
	if channel == nil {
		log.WithFields(log.Fields{
			"user":  user.ID,
			"guild": guild.ID,
		}).Warning("Failed to find channel to play sound in")
		return
	}
	vc, err := discord.ChannelVoiceJoin(guild.ID, channel.ID, false, false)
	if err != nil {
		log.Warning("Failed to join channel: ", err)
		return
	}
	voiceConnections[guild.ID] = vc

	listen(vc)
}

func leaveChannel(user *discordgo.User, guild *discordgo.Guild) {
	voiceConnections[guild.ID].Disconnect()
	delete(voiceConnections, guild.ID)
}

type Listener struct {
	inSpeech   bool
	uttStarted bool
	dec        *sphinx.Decoder
}

func listen(vc *discordgo.VoiceConnection) {
	// Init CMUSphinx
	cfg := sphinx.NewConfig(
		sphinx.HMMDirOption(hmm),
		sphinx.DictFileOption(dict),
		sphinx.LMFileOption(lm),
		sphinx.SampleRateOption(sampleRate),
	)
	log.Println("Loading CMU PhocketSphinx.")
	log.Println("This may take a while depending on the size of your model.")
	dec, err := sphinx.NewDecoder(cfg)
	if err != nil {
		closer.Fatalln(err)
	}
	closer.Bind(func() {
		dec.Destroy()
	})
	//inSpeech := false
	uttStarted := false

	if !dec.StartUtt() {
		closer.Fatalln("[ERR] Sphinx failed to start utterance")
	}
	//end init

	opdec, err := opus.NewDecoder(sampleRate, channels)
	if err != nil {
		log.Warning("Decoder creation unsuccessful")
	}

	//var frame_size_ms float32 = 60 // if you don't know, go with 60 ms.
	//frame_size := channels * frame_size_ms * sampleRate / 1000
	//pcm := make([]int16, int(frame_size))
	pcm := make([]int16, 512)

	// NEED TO DECODE 512 SAMPLES AT A TIME
	// Discord Opus: 2 channels (stereo) and a sample rate of 48Khz
	// Need to resample sound to 16khz
	for {
		recv := <-vc.OpusRecv
		//log.Info(recv)
		// log.Info("")
		// log.Info("~~~~~~~~~~~~~~")
		// log.Info(recv)
		// log.Info("-------------")
		// log.Info(recv.PCM)
		// log.Info("")
		_, err := opdec.Decode(recv.Opus, pcm)
		if err != nil {
			log.Info("Couldn't decode Data.")
		}
		// log.Info("++++++++++++++")
		// log.Info(pcm)
		log.Info("==============")
		if _, ok := dec.ProcessRaw(pcm, true, false); !ok {
			log.Warning("Language processing failed")
		}
		if dec.IsInSpeech() {
			log.Info("Still in speech.")
			//inSpeech = true
			if !uttStarted {
				uttStarted = true
				log.Println("Listening..")
			}
		} else if uttStarted {
			log.Info("Speech stopped.")
			// speech -> silence transition, time to start new utterance
			dec.EndUtt()
			uttStarted = false
			hyp, _ := dec.Hypothesis()
			if len(hyp) > 0 {
				log.Info(" > hypothesis: %s", hyp)
			} else {
				log.Info("ah, nothing")
			}
			if !dec.StartUtt() {
				closer.Fatalln("[ERR] Sphinx failed to start utterance")
			}
		} else {
			log.Info("Neither in speech, nor has utt started.")
		}
	}
}
