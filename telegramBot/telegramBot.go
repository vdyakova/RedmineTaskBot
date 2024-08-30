package telegramBot

import (
	"fmt"
	"github.com/mymmrac/telego"
	"os"
	authorization "tgtest/authorizationRedmine"
)

var loginRD = " "
var passwordRD = " "

func TelegramBot() {
	var userFirstName string
	var userLastName string
	var userUsername string
	botToken := " "
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	info, _ := bot.GetWebhookInfo()
	fmt.Printf("Webhook Info: %+v\n", info)
	updates, _ := bot.UpdatesViaLongPolling(nil)
	defer bot.StopLongPolling()
	for update := range updates {
		if update.Message != nil {
			chatId := telego.ChatID{ID: update.Message.Chat.ID}
			fmt.Println("Username", chatId.ID)
			if update.Message.From != nil {
				userFirstName = update.Message.From.FirstName
				userLastName  = update.Message.From.LastName
				userUsername  = update.Message.From.Username
				fmt.Printf("Username: %s %s (@%s)\n", userFirstName, userLastName, userUsername)
			}
			// авторизация в Redmine и установление личности для которого нужно узнать информацию о задачах
			ch := make(chan string)
			go authorization.AuthorizationRedmine(loginRD, passwordRD, userFirstName, ch)

			message := update.Message
			switch message.Text {

			case "/start":
				sendMessage(bot, chatId, "Hello from your bot! With me you can find out your work tasks!")

			case "/mytask":
				go func() {
					for message := range ch {
						sendMessage(bot, chatId, message)
					}
				}()

			case "/help":
				sendMessage(bot, chatId, " With me you can find out your work tasks! Commands - /start, /mytask, /help")

			default:
				sendMessage(bot, chatId, "Unknown command. Use /help to see available commands.")
			}
		}
	}
}

func sendMessage(bot *telego.Bot, chatId telego.ChatID, text string) {
	_, err := bot.SendMessage(&telego.SendMessageParams{
		ChatID: chatId,
		Text:   text,
	})
	if err != nil {
		fmt.Println(err)
	}
}
