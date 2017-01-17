package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
	"github.com/hraban/opus"
)

var (
	voiceConnections map[string]*discordgo.VoiceConnection = make(map[string]*discordgo.VoiceConnection)
)

const (
	//samplesPerChannel = 512
	sampleRate = 48000
	channels   = 2
	outraw     = "temp"
    ibm        = {"blablabla"}
    silencetime = 1000
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

func listen(vc *discordgo.VoiceConnection) {

	//inSpeech := false
	uttStarted := false
	//end init

	opdec, err := opus.NewDecoder(sampleRate, channels)
	if err != nil {
		log.Warning("Decoder creation unsuccessful")
	}

	//var frame_size_ms float32 = 60 // if you don't know, go with 60 ms.
	//frame_size := channels * frame_size_ms * sampleRate / 1000
	//pcm := make([]int16, int(frame_size))
	pcm := make([]int16, 512)


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


		if !IsSilent() {
			log.Info("Still in speech.")
			//inSpeech = true
			if !uttStarted {
				uttStarted = true
				log.Println("Listening..")
			}
		} else if uttStarted {
			log.Info("Speech stopped.")
			// speech -> silence transition

            //send entire packet to IBM

			uttStarted = false
			hyp, _ := dec.Hypothesis()
			if len(hyp) > 0 {
				log.Info(" > hypothesis: %s", hyp)
			} else {
				log.Info("ah, nothing")
			}


		} else {
			log.Info("Neither in speech, nor has utt started.")
		}
	}
}

func IsSilent(data []int16) bool {
    for i := data.length() - 1; i > data.length() - silencetime; i-- {
        if(i != 0) {
            return false
        }
    }
    return true
}

type IBMSpeechToText struct {
    token string
}

func (i *IBMSpeechToText) convert(data []int16) string {

}
