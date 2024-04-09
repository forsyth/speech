package polly

import (
	"io"
	"os"
	"testing"

	"github.com/forsyth/speech"
)

func getenv(t *testing.T, name string) string {
	s := os.Getenv(name)
	if s == "" {
		t.Fatalf("need value for environment variable %s", name)
	}
	return s
}

func TestSpeech(t *testing.T) {
	creds := &speech.Credentials{
		ClientID: getenv(t, "SPEECH_ID"),
		Keys:     []string{getenv(t, "SPEECH_KEY")},
	}
	region := getenv(t, "SPEECH_REGION")
	formats := []string{"pcm", "mp3", "ogg"}
	type aPolly struct {
		name string
		mk   func(*speech.Credentials, string, string, int) (speech.Speaker, error)
	}
	pollys := []aPolly{
		aPolly{"polly v1", func(cred *speech.Credentials, r string, f string, s int) (speech.Speaker, error) {
			return NewPollyV1(cred, r, f, s)
		},
		},
		aPolly{"polly v2", func(cred *speech.Credentials, r string, f string, s int) (speech.Speaker, error) {
			return NewPollyV2(cred, r, f, s)
		},
		},
	}
	for _, p := range pollys {
		t.Run(p.name, func(t *testing.T) {
			for _, f := range formats {
				speaker, err := p.mk(creds, region, f, 16000)
				if err != nil {
					t.Errorf("new speaker %s/%s: got error %v", p.name, f, err)
					return
				}
				speak := func(text, rate, locale, voice string) {
					r, err := speaker.Speak(text, rate, locale, voice)
					if err != nil {
						t.Errorf("mis-spoke: want no error; get %v", err)
						return
					}
					defer r.Audio.Close()
					bytes, err := io.ReadAll(r.Audio)
					if err != nil {
						t.Errorf("failed to read audio: %#v", err)
					} else {
						t.Logf("%v spoke %v characters of %v text in %v format, length %v bytes\n", voice, r.TextLen, locale, r.Format, len(bytes))
					}
				}
				speak("bonjour, comment ça va?", "medium", "fr-FR", "Mathieu")
				speak("bonjour, comment ça va?", "medium", "fr-FR", "Lea")
				speak("quando sono stato in Firenza, due anni fa", "medium", "it-IT", "Carla")
			}
		})
	}
}
