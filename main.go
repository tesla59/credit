package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-ping/ping"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/ini.v1"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	var (
		msg   tgbotapi.MessageConfig
		user  User
		reply string
	)

	// Reading from config.ini
	config, err := ini.Load("config.ini")
	Check(err)
	token := config.Section("").Key("token").String()

	flushMode, err := config.Section("").Key("flushMode").Bool()
	Check(err)

	if token == "" {
		panic("Please Enter Token")
	}

	// Bot and DB configs
	bot, err := tgbotapi.NewBotAPI(token)
	Check(err)
	db, err := gorm.Open(sqlite.Open("main.db"), &gorm.Config{})
	Check(err)

	// Set schema for db
	_ = db.AutoMigrate(&User{})

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	Check(err)

	for update := range updates {

		// Ignore all old messages
		if flushMode {
			continue
		}

		// This variable states if the message is a reply to something or not
		// Since attempting to perform operaion on ReplyTo variable are considered nill pointer dereferencing if the message wasn't actually a reply
		// Will find better way to implement later
		msgIsAReply := func() bool {
			if update.Message.ReplyToMessage == nil {
				return false
			} else {
				return true
			}
		}()

		// ignore any non-Message Updates
		if update.Message == nil {
			continue
		}

		if msgIsAReply && (update.Message.Text == "+" || update.Message.Text == "-") && (update.Message.From.ID != update.Message.ReplyToMessage.From.ID) && !update.Message.ReplyToMessage.From.IsBot {
			// Check if UserID already exist in db
			result := db.Where(&User{UserID: update.Message.ReplyToMessage.From.ID, ChatID: update.Message.Chat.ID}).First(&user)
			if result.Error != nil {
				// Adding entry if user doesn't exist
				db.Create(&User{
					UserID:    update.Message.ReplyToMessage.From.ID,
					ChatID:    update.Message.Chat.ID,
					FirstName: update.Message.ReplyToMessage.From.FirstName,
					LastName:  update.Message.ReplyToMessage.From.LastName,
					UserName:  update.Message.ReplyToMessage.From.UserName,
					Credit:    0,
				})
			}

			db.Where(&User{UserID: update.Message.ReplyToMessage.From.ID, ChatID: update.Message.Chat.ID}).First(&user)
			if update.Message.Text == "+" {
				user.Credit += 20
				reply = Markup("+20") + " Credit, Citizen!\nYou have " + Markup(fmt.Sprint(user.Credit)) + " points"
			} else if update.Message.Text == "-" {
				user.Credit -= 20
				reply = Markup("-20") + " Credit, Citizen!\nYou have " + Markup(fmt.Sprint(user.Credit)) + " points."
			}
			db.Save(&user)
			user = User{}

			msg = tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			msg.ParseMode = tgbotapi.ModeHTML
			msg.ReplyToMessageID = update.Message.ReplyToMessage.MessageID
		}

		// Commands
		if update.Message.IsCommand() && strings.Contains(update.Message.Text, "@") == strings.Contains(update.Message.Text, bot.Self.UserName) {
			if update.Message.Command() == "ping" {
				pinger, err := ping.NewPinger("www.google.com")
				Check(err)
				pinger.Count = 5
				err = pinger.Run() // Blocks until finished.
				Check(err)

				stats := pinger.Statistics()
				text := "Result: " + Markup(fmt.Sprint(stats.AvgRtt))

				msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
				msg.ParseMode = tgbotapi.ModeHTML
				msg.ReplyToMessageID = update.Message.MessageID
			}
		}
		// Send the final Message
		bot.Send(msg)
		// Flush the last message
		msg = tgbotapi.MessageConfig{}
	}
}
