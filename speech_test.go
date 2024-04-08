package speech

import (
	"io"
	"os"
	"testing"
)

func getenv(name string) string {
	s := os.Getenv(name)
	if s == "" {
		return "$" + name
	}
	return s
}

func TestSpeech(t *testing.T) {
	creds := &Credentials{
		ClientID: getenv("AWS_ACCESS_KEY_ID"),
		Keys:     []string{getenv("AWS_SECRET_ACCESS_KEY")},
	}
	t.Run("Speak", func(t *testing.T) {
		speak := func(text, rate, locale, voice string) {
			r, err := Speak(text, rate, locale, voice, creds)
			if err != nil {
				code, _, fault := DecodeError(err)
				if code != "general" {
					t.Errorf("mis-spoke: want no error; got %v (code: %q fault: %q)", err, code, fault)
				} else {
					t.Errorf("mis-spoke: want no error; get %v", err)
				}
				return
			}
			defer r.Audio.Close()
			bytes, err := io.ReadAll(r.Audio)
			if err != nil {
				t.Errorf("failed to read audio: want no error; got %v", err)
			} else {
				t.Logf("%v spoke %v characters of %v text in %v format, length %v bytes\n", voice, r.TextLen, locale, r.Format, len(bytes))
			}
		}
		speak("bonjour, comment ça va?", "medium", "fr-FR", "Mathieu")
		speak("bonjour, comment ça va?", "medium", "fr-FR", "Lea")
		speak("quando sono stato in Firenza, due anni fa", "medium", "it-IT", "Carla")
	})
	t.Run("Speech", func(t *testing.T) {
		speech, err := New(creds, Region, "mp3", 8000)
		if err != nil {
			t.Errorf("new speech: got error %v", err)
			return
		}
		speak := func(text, rate, locale, voice string) {
			r, err := speech.Speak(text, rate, locale, voice)
			if err != nil {
				code, _, fault := DecodeError(err)
				if code != "general" {
					t.Errorf("mis-spoke: want no error; got %v (code: %q fault: %q)", err, code, fault)
				} else {
					t.Errorf("mis-spoke: want no error; get %v", err)
				}
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
	})
}
