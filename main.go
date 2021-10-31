package main

import (
	"fmt"
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/ini.v1"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	User_id    int `gorm:"primaryKey"`
	First_name string
	Last_name  string
	Username   string
	Credit     int
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	var (
		msg tgbotapi.MessageConfig
		user User
		reply string
		flushMode bool = false
	)

	// Reading from cnofig.ini
	config, err := ini.Load("config.ini")
	Check(err)
	token := config.Section("").Key("token").String()
	if token == ""{
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
	// flushMode = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	Check(err)

	for update := range updates {
		// ignore any non-Message Updates + any non replied message
		if update.Message == nil || update.Message.ReplyToMessage == nil {
			continue
		}

		// Ignore all old messages
		if flushMode { 
			continue
		}

		if (update.Message.Text == "+" || update.Message.Text == "-") && (update.Message.From.ID != update.Message.ReplyToMessage.From.ID) && !update.Message.ReplyToMessage.From.IsBot {
			// Check if user_id already exist in db
			result := db.First(&user, update.Message.ReplyToMessage.From.ID);
			if result.Error != nil {
				// Adding entry if user doesnt exist
				db.Create(&User{
					User_id:    update.Message.ReplyToMessage.From.ID,
					First_name: update.Message.ReplyToMessage.From.FirstName,
					Last_name:  update.Message.ReplyToMessage.From.LastName,
					Username:   update.Message.ReplyToMessage.From.UserName,
					Credit:     0,
				})
			}

			if update.Message.Text == "+" {
				_ = db.First(&user,update.Message.ReplyToMessage.From.ID)
				user.Credit += 20
				db.Save(&user)
				reply = "<code>+20</code> Credit, Citizen!\nYou have <code>" + fmt.Sprint(user.Credit) + "</code> points."
				user = User{}
			}else if update.Message.Text == "-" {
				_ = db.First(&user,update.Message.ReplyToMessage.From.ID)
				user.Credit -= 20
				db.Save(&user)
				reply = "<code>-20</code> Credit, Citizen!\nYou have <code>" + fmt.Sprint(user.Credit) + "</code> points."
				user = User{}
			}
			msg = tgbotapi.NewMessage(update.Message.Chat.ID,reply)
			msg.ParseMode=tgbotapi.ModeHTML
			msg.ReplyToMessageID = update.Message.ReplyToMessage.MessageID
			bot.Send(msg)
		}

	}
}
