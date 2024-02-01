package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
	_ "go.uber.org/automaxprocs"
)

var logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.InfoLevel)
var Config appConfig

type appConfig struct {
	Token string `mapstructure:token`
	Debug string `mapstructure:debug`
	Port  string `mapstructure:port`
}

type inputStruct struct {
	Channels   []string `json:"channels"`
	Message    string   `json:"message"`
	Topic      string   `json:"topic,omitempty"`
	Attachment string
}

func (config appConfig) Validate() error {
	return validation.ValidateStruct(&config,
		validation.Field(&config.Token, validation.Required),
		validation.Field(&config.Port, validation.Required),
	)
}

func LoadConfig() error {
	v := viper.New()
	v.SetEnvPrefix("slackgw")
	v.AutomaticEnv()
	v.BindEnv("debug")
	v.BindEnv("token")
	v.BindEnv("port")
	v.SetDefault("port", 8080)

	if err := v.Unmarshal(&Config); err != nil {
		return err
	}

	return Config.Validate()
}

func returnMessage(messageFormat string, origMessage string, origTopic string) string {
	logmsg := fmt.Sprintf("format: %s, msg: %s, topic: %s", messageFormat, origMessage, origTopic)
	logger.Debug().Msg(logmsg)
	var retMsg string
	if origTopic != "" {
		retMsg = fmt.Sprintf("%s - %s", origTopic, origMessage)
	} else {
		retMsg = fmt.Sprintf("%s", origMessage)
	}
	logger.Debug().Msg(retMsg)
	if messageFormat == "plaintext" {
		return fmt.Sprintf("```%s\n```", retMsg)
	}
	return retMsg
}

func defaultRoots(rw http.ResponseWriter, request *http.Request) {
	validEndpoint := false
	var msgFormat string
	if request.URL.Path == "/" {
		validEndpoint = true
		msgFormat = "markdown"
	}

	if request.URL.Path == "/raw" {
		validEndpoint = true
		msgFormat = "plaintext"
	}

	if !validEndpoint {
		http.Error(rw, "404 not found.", http.StatusNotFound)
		return
	}

	switch request.Method {
	case "POST":
		decoder := json.NewDecoder(request.Body)

		var t inputStruct
		err := decoder.Decode(&t)

		if err != nil {
			logMsg := fmt.Sprintf("slack-gw ERROR: failed to decode :%s", request.Body)
			logger.Error().Msg(logMsg)
		}

		outMessage := returnMessage(msgFormat, t.Message, t.Topic)
		for _, channelID := range t.Channels {
			go postMessage(msgFormat, channelID, outMessage)
		}
	default:
		fmt.Fprintf(rw, "Sorry, only the POST method is supported.")
	}
}

func postMessage(messageFormat string, channelID string, inputMessage string) {
	api := slack.New(Config.Token)
	var err error
	var timestamp string

	if messageFormat == "markdown" {
		textBlockObject := slack.NewTextBlockObject("mrkdwn", inputMessage, false, false)
		sectionBlock := slack.NewSectionBlock(textBlockObject, nil, nil)
		blocks := []slack.Block{sectionBlock}
		blockOptions := slack.MsgOptionBlocks(blocks...)
		channelID, timestamp, err = api.PostMessage(channelID, blockOptions)
	} else {
		channelID, timestamp, err = api.PostMessage(channelID, slack.MsgOptionText(inputMessage, false))
	}

	if err != nil {
		logMsg := fmt.Sprintf("ERROR: unable to send message - %s", err)
		logger.Error().Msg(logMsg)
		return
	}
	logMsg := fmt.Sprintf("Message successfully sent to channel %s at %s", channelID, timestamp)
	logger.Info().Msg(logMsg)
}

func main() {
	LoadConfig()
	if Config.Token == "" {
		logger.Error().Msg("Slack token not set")
		os.Exit(1)
	}

	if Config.Debug != "" {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.DebugLevel)
		logger.Debug().Msg("Debugging enabled!")
	}

	http.HandleFunc("/", defaultRoots)
	gwPort := fmt.Sprintf(":%s", Config.Port)
	logMsg := fmt.Sprintf("Starting server on %s", gwPort)
	logger.Info().Msg(logMsg)
	if err := http.ListenAndServe(gwPort, nil); err != nil {
		logger.Fatal().Err(err)
	}
}
