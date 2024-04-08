package speech

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/aws/smithy-go"
)

const (
	Region = "eu-west-2" // TO DO: configure it
)

type Credentials struct {
	ClientID string
	Keys     []string
}

// Spoken represents a stream of bytes in Audio, in the given Format.
type Spoken struct {
	Audio       io.ReadCloser // sadly unseekable stream of data
	TextLen     int           // length of spoken text (for accounting)
	Format      string        // "mp3", "ogg_vorbis", "pcm", "json"
	ContentType string
}

// Speech is a session that converts text to speech, by calls to Speak.
// The language, voice and speech speed can be different on each call.
type Speech struct {
	creds    *Credentials
	region   string
	config   aws.Config
	format   types.OutputFormat
	sampling string
}

// New creates a Speech session with default credentials, region and format.
func New(creds *Credentials, region string, format string, sampleRate int) (*Speech, error) {
	if creds.ClientID == "" || len(creds.Keys) < 1 || creds.Keys[0] == "" {
		return nil, errors.New("missing id or key in credential")
	}
	var sv types.OutputFormat
	form, err := cvtString("output format", format, sv.Values())
	if err != nil {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	//awsCredCache := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(creds.ClientID, creds.Keys[0], ""))
	//awsCreds, err := awsCredCache.Retrieve(context.Background())
	//if err != nil {
	//	return nil, fmt.Errorf("new speech session: credentials: %v", err)
	//}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(creds.ClientID, creds.Keys[0], "")),
	)
	if err != nil {
		return nil, err
	}
	speech := &Speech{
		creds:    creds,
		region:   region,
		config:   cfg,
		format:   form,
		sampling: fmt.Sprint(sampleRate),
	}
	return speech, nil
}

func def(x string) *string {
	if x != "" {
		return &x
	}
	return nil
}

// Speak says something at a given speaking rate ("medium" or "slow") within a locale using a specified voice.
// It currently returns the PCM representation.
func (s *Speech) Speak(text, speakingRate, locale, voice string) (*Spoken, error) {
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
	spoken := &Spoken{
		Audio:       result.AudioStream,
		TextLen:     int(result.RequestCharacters),
		Format:      string(input.OutputFormat),
		ContentType: aws.ToString(result.ContentType),
	}
	return spoken, nil
}

// Speak returns a representation of the given text, spoken by the given voice for a locale at a given rate ("medium" or "slow").
// Currently it returns PCM at 16k.
func Speak(text, speakingRate, locale, voice string, creds *Credentials) (*Spoken, error) {
	sesh, err := New(creds, Region, "pcm", 16000)
	if err != nil {
		return nil, err
	}
	return sesh.Speak(text, speakingRate, locale, voice)
}

// to_ssml converts text to SSML (AWS version), replacing reserved characters by XML entities.
func to_ssml(text, speakingRate string) string {
	sb := &strings.Builder{}
	sb.WriteString("<speak><amazon:auto-breaths>")
	sb.WriteString("<prosody rate=\"" + speakingRate + "\">")
	for _, r := range text {
		switch r {
		case '<':
			sb.WriteString("&lt;")
		case '&':
			sb.WriteString("&amp;")
		default:
			sb.WriteRune(r)
		}
	}
	sb.WriteString("</prosody>")
	sb.WriteString("</amazon:auto-breaths>")
	sb.WriteString("</speak>")
	return sb.String()
}

// DecodeError returns the code, message and fault suggestion from an AWS error; it reverts to "general" and err.Error() if it is not an AWS error.
func DecodeError(err error) (code string, msg string, fault string) {
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

// possible codes of interest:

// ErrCodeTextLengthExceededException "TextLengthExceededException"
// ErrCodeInvalidSampleRateException "InvalidSampleRateException"
// ErrCodeInvalidSsmlException "InvalidSsmlException"
// ErrCodeLexiconNotFoundException "LexiconNotFoundException"
// ErrCodeServiceFailureException "ServiceFailureException"
// ErrCodeMarksNotSupportedForFormatException "MarksNotSupportedForFormatException"
// ErrCodeSsmlMarksNotSupportedForTextTypeException "SsmlMarksNotSupportedForTextTypeException"
// ErrCodeLanguageNotSupportedException "LanguageNotSupportedException"
