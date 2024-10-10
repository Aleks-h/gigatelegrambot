package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
	gigaapi "github.com/saintbyte/gigachat_api"
)

var usersList = []int64{}

func telegramBot() {

	//Создаем бота
	//bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	tgbotapi.NewInlineKeyboardMarkup((tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("text", "text"))))
	if err != nil {
		panic(err)
	}
	//Устанавливаем время обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//Получаем обновления от бота
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if !isIdOnList(update.Message.Chat.ID) {
			usersList = append(usersList, update.Message.Chat.ID)
		}
		chat := gigaapi.NewGigachat()
		answer, err := chat.Ask(update.Message.Text)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, answer)
		bot.Send(msg)
		if err != nil {
			panic(err)
		}
	}
}

func isIdOnList(userId int64) bool {
	for _, id := range usersList {
		if id == userId {
			return true
		}
	}
	return false
}

func sendToGigaChat(msg string) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		panic(err)
	}
	chat := gigaapi.NewGigachat()
	answer, err := chat.Ask(msg)
	for _, user := range usersList {
		msg := tgbotapi.NewMessage(user, answer)
		bot.Send(msg)

		if err != nil {
			panic(err)
		}
	}
}

func timerEvent() {
	for {
		hour, min, _ := time.Now().Clock()
		if hour == 10 && min == 00 {
			_, m, d := time.Now().Date()
			date := m.String() + " " + strconv.Itoa(d)
			compliment := "придумай для Алёны приятные слова на " + date
			println(compliment)
			sendToGigaChat(compliment)
		}
		if hour == 11 && min == 00 {
			y, m, d := time.Now().Date()
			date := m.String() + " " + strconv.Itoa(d) + " " + strconv.Itoa(y)
			prediction := "придумай астрологический прогноз для Стрельца " + date
			println(prediction)
			sendToGigaChat(prediction)
		}
		time.Sleep(time.Minute)
	}
}

func readUsersList() {
	file, err := os.Open("userslist")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		id, err := strconv.ParseInt(scanner.Text(), 10, 64)
		if err != nil {
			panic(err)
		}
		fmt.Println(id)
		usersList = append(usersList, id)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func exitFromProgram() {
	file, err := os.Create("userslist")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, user := range usersList {
		file.WriteString(strconv.FormatInt(user, 10))
		file.WriteString("\n")
	}
	os.Exit(0)
}

func cmndLineInteract() {
	var cmnd string
	for {
		fmt.Println("Введите 1 для корректного завершения программы")
		fmt.Scan(&cmnd)
		switch cmnd {
		case "1":
			exitFromProgram()
		}

	}
}

// Вызываем бота
func main() {
	readUsersList()
	go timerEvent()
	go cmndLineInteract()
	telegramBot()
}
