package polly

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/aws/smithy-go"

	"github.com/forsyth/speech"
)

// PollySpeakerV2 is a session that converts text to speech, by calls to Speak.
// The language, voice and speech speed can be different on each call.
type PollySpeakerV2 struct {
	creds    *speech.Credentials
	region   string
	config   aws.Config
	format   types.OutputFormat
	sampling string
}

// NewPollyV2 creates a PollySpeakerV2 session with default credentials, region and format.
func NewPollyV2(creds *speech.Credentials, region string, format string, sampleRate int) (*PollySpeakerV2, error) {
	if creds.ClientID == "" || len(creds.Keys) < 1 || creds.Keys[0] == "" {
		return nil, errors.New("missing id or key in credential")
	}
	if format == "ogg" || format == "oggvorbis" {
		format = "ogg_vorbis"
	}
	var sv types.OutputFormat
	form, err := cvtString("output format", format, sv.Values())
	if err != nil {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(creds.ClientID, creds.Keys[0], "")),
	)
	if err != nil {
		return nil, fmt.Errorf("polly session error: %v", err)
	}
	speech := &PollySpeakerV2{
		creds:    creds,
		region:   region,
		config:   cfg,
		format:   form,
		sampling: fmt.Sprint(sampleRate),
	}
	return speech, nil
}

// Speak says something at a given speaking rate ("medium" or "slow") within a locale using a specified voice.
func (s *PollySpeakerV2) Speak(text, speakingRate, locale, voice string) (*speech.Spoken, error) {
	var lc types.LanguageCode
	loc, err := cvtString("locale", locale, lc.Values())
	if err != nil {
		return nil, err
	}
	var vc types.VoiceId
	vID, err := cvtString("voice name", voice, vc.Values())
	if err != nil {
		return nil, err
	}
	svc := polly.NewFromConfig(s.config)
	input := &polly.SynthesizeSpeechInput{
		OutputFormat: s.format,
		SampleRate:   aws.String(s.sampling), // mp3, ogg_vorbis: 8000, 16000, 22050; pcm: 8000, 16000
		Text:         aws.String(to_ssml(text, speakingRate)),
		TextType:     types.TextTypeSsml,
		VoiceId:      vID,
		LanguageCode: loc, // defaults to default locale for given voice
		// SpeechMarkTypes	[]*string	// "sentence", "word", "viseme", "ssml" <mark>
	}
	result, err := svc.SynthesizeSpeech(context.Background(), input)
	if err != nil {
		return nil, err
	}
	spoken := &speech.Spoken{
		Audio:       result.AudioStream,
		TextLen:     int(result.RequestCharacters),
		Format:      string(input.OutputFormat),
		ContentType: aws.ToString(result.ContentType),
	}
	return spoken, nil
}

// DecodeErrorV2 returns the code, message and fault suggestion from an AWS error; it reverts to "general" and err.Error() if it is not an AWS error.
func DecodeErrorV2(err error) (code string, msg string, fault string) {
	var ae smithy.APIError
	if errors.As(err, &ae) {
		return ae.ErrorCode(), ae.ErrorMessage(), ae.ErrorFault().String()
	}
	return "general", err.Error(), ""
}

// cvtString converts a plain string to a value of an AWS api v2 string-enum type,
// returning an error if the value isn't a valid enum value.
func cvtString[T ~string](what string, s string, valid []T) (T, error) {
	for _, v := range valid {
		if T(s) == v {
			return v, nil
		}
	}
	var zero T
	return zero, fmt.Errorf("invalid %s: %s", what, s)
}
