package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/slack-go/slack"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

var lineBot *linebot.Client
var slackBot *slack.Client
var slackChannelId string

func main() {
	var err error
	cfg := getConfig()
	slackChannelId = cfg.SlackChannelId
	lineBot, err = linebot.New(cfg.LineChannelSecret, cfg.LineChannelAccessToken)
	if err != nil {
		log.Println("Line Bot:", lineBot, " err:", err)
	} else {
		log.Println("Line Bot: OK")
	}
	slackBot = slack.New(cfg.SlackToken, slack.OptionDebug(true))
	params := slack.GetConversationsParameters{}

	groups, _, err := slackBot.GetConversations(&params)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	for _, group := range groups {
		fmt.Printf("ID: %s, Name: %s\n", group.ID, group.Name)
	}

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)
	e.POST("line_callback", lineCallbackHandler)
	// Start server
	e.Logger.Fatal(e.Start(":3000"))
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
func lineCallbackHandler(c echo.Context) error {
	events, err := lineBot.ParseRequest(c.Request())
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			return c.String(http.StatusBadRequest, "Bad Request")
		} else {
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}
	}
	for _, event := range events {
		userID := event.Source.UserID
		groupID := event.Source.GroupID
		sender, _ := lineBot.GetGroupMemberProfile(groupID, userID).Do()
		group, _ := lineBot.GetGroupSummary(groupID).Do()
		senderName := "[" + group.GroupName + "] " + sender.DisplayName + ":"
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			default:
				log.Printf("%+v", message)
			}
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				_, _, err = slackBot.PostMessage(
					slackChannelId,
					slack.MsgOptionText(senderName+"\n"+message.Text, false),
					slack.MsgOptionAsUser(true), // Add this if you want that the bot would post message as a user, otherwise it will send response using the default slackbot
				)
				if err != nil {
					log.Print(err)
				}

			case *linebot.StickerMessage:
				url := fmt.Sprintf("https://stickershop.line-scdn.net/stickershop/v1/sticker/%s/android/sticker.png", message.StickerID)
				attachment := slack.Attachment{}
				attachment.Text = "[" + group.GroupName + "] " + sender.DisplayName + ":"
				attachment.ImageURL = url
				_, _, err = slackBot.PostMessage(
					slackChannelId,
					slack.MsgOptionAttachments(attachment),
					slack.MsgOptionAsUser(true), // Add this if you want that the bot would post message as a user, otherwise it will send response using the default slackbot
				)
				if err != nil {
					log.Print(err)
				}
			case *linebot.ImageMessage:
				content, err := lineBot.GetMessageContent(message.ID).Do()
				if err != nil {
					log.Print(err)
				}
				defer content.Content.Close()
				var contentData = make([]byte, 1024*1024) //1MB
				_, err = content.Content.Read(contentData)
				if err != nil {
					log.Print(err)
				}
				file := slack.FileUploadParameters{
					InitialComment: senderName,
					Content:        string(contentData),
					Channels:       []string{slackChannelId},
				}
				_, err = slackBot.UploadFile(file)
				if err != nil {
					log.Print(err)
				}
			case *linebot.FileMessage:
				content, err := lineBot.GetMessageContent(message.ID).Do()
				if err != nil {
					log.Print(err)
				}
				defer content.Content.Close()
				var contentData = make([]byte, content.ContentLength)
				_, err = content.Content.Read(contentData)
				if err != nil {
					log.Print(err)
				}
				file := slack.FileUploadParameters{
					Filename:       message.FileName,
					Filetype:       content.ContentType,
					InitialComment: senderName,
					Content:        string(contentData),
					Channels:       []string{slackChannelId},
				}
				_, err = slackBot.UploadFile(file)
				if err != nil {
					log.Print(err)
				}
			}

		}
	}

	return c.String(http.StatusOK, "ok")
}
