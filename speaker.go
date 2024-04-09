// Package speech provides an interface, Speaker, that allows an application to
// access the basic function of a text-to-speech service, hiding the underlying
// provider, allowing migration (eg, from Bing to Polly).
// Call the provider-specific New function (eg, polly.NewPollyV1, polly.NewPollyV2)
// and assign the result to a Speaker variable.
package speech

import (
	"io"
)

// Credentials represents a common form of service credentials:
// a user ID string and then one or more keys to identify and authorise the service.
type Credentials struct {

	// ClientID is whatever is conventional as the ID finally presented to the service.
	ClientID string

	// Keys contains one or more service access keys.
	Keys []string
}

// Spoken represents a stream of bytes in Audio, in the given Format.
type Spoken struct {

	// Audio produces the stream of audio, sadly unseekable.
	Audio io.ReadCloser

	// TextLen is the length of the spoken text's recording format in bytes.
	TextLen int

	// Format names the recording format ("mp3", "ogg_vorbis", "pcm", "json").
	Format string

	// ContentType is the HTTP "Content-Type" header's value for the recording format.
	ContentType string
}

// Speaker represents a session on a text-to-speech engine.
// The language, voice and speaking rate can be different on each call.
type Speaker interface {
	// Speak says something at a given speaking rate ("medium or "slow") within a locale, using a specified voice.
	Speak(text, speakingRate, locale, voice string) (*Spoken, error)
}
