package main

import (
	_ "github.com/lib/pq"
	tgBot "tgtest/telegramBot"
)

func main() {

	tgBot.TelegramBot()

}
