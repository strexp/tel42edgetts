package tts

import (
	"strings"

	"github.com/wujunwei928/edge-tts-go/edge_tts"
)

func ResolveVoice(lang, voiceName string) string {
	if voiceName == "" {
		if lang == "zh-CN" {
			return "zh-CN-XiaoyiNeural"
		}
		if lang == "en-US" {
			return "en-US-GuyNeural"
		}
	}
	if strings.HasSuffix(voiceName, "Neural") {
		return voiceName
	}
	return lang + "-" + voiceName + "Neural"
}

func Download(text, voice string) ([]byte, error) {
	connOptions := []edge_tts.CommunicateOption{
		edge_tts.SetVoice(voice),
		edge_tts.SetReceiveTimeout(20),
	}
	conn, err := edge_tts.NewCommunicate(text, connOptions...)
	if err != nil {
		return nil, err
	}
	return conn.Stream()
}
