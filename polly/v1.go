package polly

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"

	"github.com/forsyth/speech"
)

// SpeakerV1 is a session that converts text to speech, by calls to Speak.
// The language, voice and speech speed can be different on each call.
type SpeakerV1 struct {
	creds    *speech.Credentials
	region   string
	session  *session.Session
	format   string
	sampling string
}

// NewPollyV1 creates a Speaker session with default credentials, region and format,
// for AWS API v1.
func NewPollyV1(creds *speech.Credentials, region string, format string, sampleRate int) (*SpeakerV1, error) {
	if creds.ClientID == "" || len(creds.Keys) < 1 || creds.Keys[0] == "" {
		return nil, errors.New("missing id or key in credential")
	}
	awscreds := credentials.NewStaticCredentials(creds.ClientID, creds.Keys[0], "")
	sesh, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: awscreds,
	})
	if err != nil {
		return nil, err
	}
	speech := &SpeakerV1{
		creds:    creds,
		region:   region,
		session:  sesh,
		format:   oggly(format),
		sampling: fmt.Sprint(sampleRate),
	}
	return speech, nil
}

// Speak says something at a given rate within a locale using a specified voice.
// The rate is either "medium" or "slow".
// It currently returns the PCM representation.
func (s *SpeakerV1) Speak(text, rate, locale, voice string) (*speech.Spoken, error) {
	svc := polly.New(s.session)
	input := &polly.SynthesizeSpeechInput{
		OutputFormat: aws.String(s.format),   // TO DO: mp3, ogg_vorbis or pcm ("json" for speech marks)
		SampleRate:   aws.String(s.sampling), // mp3, ogg_vorbis: 8000, 16000, 22050; pcm: 8000, 16000
		Text:         aws.String(toSSML(text, rate)),
		TextType:     aws.String("ssml"),
		VoiceId:      &voice,
		LanguageCode: def(locale), // defaults to default locale for given voice
		// SpeakerV1MarkTypes	[]*string	// "sentence", "word", "viseme", "ssml" <mark>
	}
	result, err := svc.SynthesizeSpeech(input)
	if err != nil {
		return nil, err
	}
	spoken := &speech.Spoken{
		Audio:   result.AudioStream,
		TextLen: int(*result.RequestCharacters),
		Format:  *input.OutputFormat,
	}
	return spoken, nil
}

// DecodeError returns the code and message from an AWS error; it reverts to "general" and err.Error() if it is not an AWS error.
// In practice, at least for Polly, the break-down is rarely more informative than the plain Error() text.
func DecodeErrorV1(err error) (string, string) {
	if aerr, ok := err.(awserr.Error); ok {
		return aerr.Code(), aerr.Message()
	}
	return "general", err.Error()
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
