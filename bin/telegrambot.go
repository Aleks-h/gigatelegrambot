package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	//tgbotapi "github.com/Syfaro/telegram-bot-api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gigaapi "github.com/saintbyte/gigachat_api"
)

// Current содержит информацию о текущих погодных условиях
type Current struct {
	Time          string  `json:"time"`
	Temperature2M float64 `json:"temperature_2m"`
	WindSpeed10M  float64 `json:"wind_speed_10m"`
}

var usersList = []int64{}
var hourCntBirthDay int = 7
var weateherMod bool = false

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
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if !isIdOnList(update.Message.Chat.ID) {
			usersList = append(usersList, update.Message.Chat.ID)
		}
		var answer string
		req := update.Message.Text
		switch {
		case weateherMod:
			weateherMod = false
			fmt.Println(req)
			regionsMap := confReader()
			region, valid := regionsMap[req]
			if !valid {
				continue
			}
			weather, err := GetWeather(region)
			if err != nil {
				continue
			}
			answer = "Погода сейчас: " + strconv.FormatFloat(weather.Temperature2M, 'f', -1, 64) + " градус(а)"
		case req == "/weather":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите город:")
			regionsMap := confReader()
			regioncount := float64(len(regionsMap))
			rowcount := math.Ceil(regioncount / 3)
			numericKeyboard := tgbotapi.NewReplyKeyboard()
			numericLine := tgbotapi.NewKeyboardButtonRow()

			var regionsNames []string
			for rus, _ := range regionsMap {
				regionsNames = append(regionsNames, rus)
			}
			number := 0
			for i := 0; i < int(rowcount); i++ {
				for i := 0; i < 3; i++ {
					if number >= len(regionsNames) {
						break
					}
					Buttn := tgbotapi.NewKeyboardButton(regionsNames[number])
					numericLine = append(numericLine, Buttn)
					number++
				}
				numericKeyboard.Keyboard = append(numericKeyboard.Keyboard, numericLine)
			}
			msg.ReplyMarkup = numericKeyboard
			bot.Send(msg)
			weateherMod = true
			continue
		case req == "/crocodile":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нажмите старт для начала игры")
			Buttn1 := tgbotapi.NewKeyboardButton("Старт")
			Buttn2 := tgbotapi.NewKeyboardButton("Отмена")
			row := tgbotapi.NewKeyboardButtonRow(Buttn1, Buttn2)
			tgbotapi.NewKeyboardButton
			numericKeyboard := tgbotapi.
			numericKeyboard.Keyboard = append(numericKeyboard.Keyboard,Buttn1)
			msg.ReplyMarkup = numericKeyboard
			bot.Send(msg)
			weateherMod = true
			continue
		default:
			chat := gigaapi.NewGigachat()
			var err error
			answer, err = chat.Ask(req)
			if err != nil {
				panic(err)
			}
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, answer)
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, err := bot.Send(msg)
		if err != nil {
			panic(err)
		}
	}
}

func confReader() map[string]string {
	byteValue, err := os.ReadFile("./regions.json")
	if err != nil {
		panic(err)
	}
	regions := make(map[string]string)
	err = json.Unmarshal([]byte(byteValue), &regions)
	if err != nil {
		panic(err)
	}
	return regions
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
		datetime := time.Now()
		if datetime.Hour() == 10 && datetime.Minute() == 00 {
			_, m, d := time.Now().Date()
			date := m.String() + " " + strconv.Itoa(d)
			compliment := "придумай для Алёны приятные слова на " + date
			println(compliment)
			sendToGigaChat(compliment)
		}
		if datetime.Hour() == 11 && datetime.Minute() == 00 {
			y, m, d := time.Now().Date()
			date := m.String() + " " + strconv.Itoa(d) + " " + strconv.Itoa(y)
			prediction := "придумай астрологический прогноз для Стрельца " + date
			println(prediction)
			sendToGigaChat(prediction)
		}
		if datetime.Month().String() == `December` && datetime.Day() == 1 {
			if datetime.Hour() == hourCntBirthDay {
				congrats := "Поздравь Алёну с днем рождения"
				sendToGigaChat(congrats)
				hourCntBirthDay++
			}
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

func createMenu() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		panic(err)
	}

	cmdCfg := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{
			Command:     "weather",
			Description: "Информация о погоде",
		},
		tgbotapi.BotCommand{
			Command:     "crocodile",
			Description: "сыграть в крокодила",
		},
		tgbotapi.BotCommand{
			Command:     "info",
			Description: "Информация",
		},
	)
	bot.Send(cmdCfg)
}

// Вызываем бота
func main() {
	createMenu()
	readUsersList()
	go timerEvent()
	go cmndLineInteract()
	telegramBot()
}

func GetWeather(region string) (Current, error) {
	url := fmt.Sprintf("http://127.0.0.1:8090/currentWeather?region=%s", region)
	client := http.Client{}
	response, err := client.Get(url)

	if err != nil {
		return Current{}, err
	}
	if response.StatusCode != http.StatusOK {
		return Current{}, err
	}
	body, err2 := io.ReadAll(response.Body)
	if err2 != nil {
		return Current{}, err
	}
	defer response.Body.Close()
	var result Current
	err3 := json.Unmarshal(body, &result)
	if err3 != nil {
		return Current{}, err
	}
	return result, nil
}
