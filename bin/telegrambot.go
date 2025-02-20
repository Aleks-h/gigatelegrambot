package main

import (
	"bufio"
	"encoding/json"
	"errors"
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

type state string

const (
	Chatting  state = "chatting"
	Weather   state = "weather"
	Crocodile state = "crocodile"
)

type crocstate string

const (
	initCmndNmbr crocstate = "initCmndNmbr"
	initTeamNmbr crocstate = "initTeamNmbr"
	ready        crocstate = "ready"
	start        crocstate = "start"
	answr        crocstate = "answer"
	deflt        crocstate = "default"
)

// Current содержит информацию о текущих погодных условиях
type Current struct {
	Time          string  `json:"time"`
	Temperature2M float64 `json:"temperature_2m"`
	WindSpeed10M  float64 `json:"wind_speed_10m"`
}

type crocConfig struct {
	wrdNmbr  int
	cmndNmbr int
}

var (
	usersList                 = []int64{}
	hourCntBirthDay int       = 7
	currState       state     = Chatting
	currcrocstate   crocstate = deflt
	crConf          crocConfig
)

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
			currState = Weather
			continue
		case currState == Weather:
			currState = Chatting
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
		case req == "/crocodile":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нажмите старт для начала игры")
			Buttn1 := tgbotapi.NewKeyboardButton("Старт")
			Buttn2 := tgbotapi.NewKeyboardButton("Отмена")
			row1 := tgbotapi.NewKeyboardButtonRow(Buttn1)
			row2 := tgbotapi.NewKeyboardButtonRow(Buttn2)
			numericKeyboard := tgbotapi.NewReplyKeyboard()
			numericKeyboard.Keyboard = append(numericKeyboard.Keyboard, row1)
			numericKeyboard.Keyboard = append(numericKeyboard.Keyboard, row2)
			msg.ReplyMarkup = numericKeyboard
			bot.Send(msg)
			currState = Crocodile
			continue
		case currState == Crocodile:
			switch {
			case currcrocstate == deflt && req == "Старт":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите количество команд")
				numericLine := tgbotapi.NewKeyboardButtonRow()
				for i := 0; i < 10; i++ {
					Buttn := tgbotapi.NewKeyboardButton(strconv.Itoa(i + 1))
					numericLine = append(numericLine, Buttn)
				}
				cnclbtn := tgbotapi.NewKeyboardButton("Отмена")
				cnclline := tgbotapi.NewKeyboardButtonRow(cnclbtn)
				numericKeyboard := tgbotapi.NewReplyKeyboard()
				numericKeyboard.Keyboard = append(numericKeyboard.Keyboard, numericLine)
				numericKeyboard.Keyboard = append(numericKeyboard.Keyboard, cnclline)
				msg.ReplyMarkup = numericKeyboard
				bot.Send(msg)
				currcrocstate = initCmndNmbr
				continue
			case req == "Отмена":
				answer = "Хорошо, сыграем в другой раз"
				currcrocstate = deflt
				currState = Chatting
			case currcrocstate == initCmndNmbr:
				cmndN, err := strconv.Atoi(req)
				if err != nil {
					panic(err)
				}
				var msg tgbotapi.MessageConfig
				if cmndN > 10 {
					crConf.cmndNmbr = 10
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Нельзя выбрать количество команд больше 10. Установлено значение 10.\nВыберите количество слов")
				} else {
					crConf.cmndNmbr = cmndN
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите количество слов")
				}

				number := 0
				numericKeyboard := tgbotapi.NewReplyKeyboard()
				for j := 0; j < 4; j++ {
					numericLine := tgbotapi.NewKeyboardButtonRow()
					for i := 0; i < 6; i++ {
						Buttn := tgbotapi.NewKeyboardButton(strconv.Itoa(number + 1))
						numericLine = append(numericLine, Buttn)
						number++
					}
					numericKeyboard.Keyboard = append(numericKeyboard.Keyboard, numericLine)
				}
				cnclbtn := tgbotapi.NewKeyboardButton("Отмена")
				cnclline := tgbotapi.NewKeyboardButtonRow(cnclbtn)
				numericKeyboard.Keyboard = append(numericKeyboard.Keyboard, cnclline)
				msg.ReplyMarkup = numericKeyboard
				bot.Send(msg)
				currcrocstate = initTeamNmbr
				continue

			case currcrocstate == initTeamNmbr:
				//	var msg tgbotapi.MessageConfig
				wrdN, err := strconv.Atoi(req)
				if err != nil {
					panic(err)
				}
				//	var msg tgbotapi.MessageConfig
				var msg tgbotapi.MessageConfig
				if wrdN > 24 {
					crConf.wrdNmbr = 24
					str := fmt.Sprintf("Нельзя выбрать количество слов больше 24. Установлено значение 24.\nКоличество команд: %d", crConf.cmndNmbr)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, str)
				} else {
					crConf.wrdNmbr = wrdN
					str := fmt.Sprintf("Количество команд: %d\nКоличество слов: %d", crConf.cmndNmbr, crConf.wrdNmbr)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, str)
				}
				Buttn1 := tgbotapi.NewKeyboardButton("Далее")
				Buttn2 := tgbotapi.NewKeyboardButton("Отмена")
				row1 := tgbotapi.NewKeyboardButtonRow(Buttn1)
				row2 := tgbotapi.NewKeyboardButtonRow(Buttn2)
				numericKeyboard := tgbotapi.NewReplyKeyboard()
				numericKeyboard.Keyboard = append(numericKeyboard.Keyboard, row1)
				numericKeyboard.Keyboard = append(numericKeyboard.Keyboard, row2)
				msg.ReplyMarkup = numericKeyboard
				bot.Send(msg)
				req := fmt.Sprintf("http://127.0.0.1:8091/init?numberOfWords=%d&numberOfTeams=%d",
					crConf.wrdNmbr, crConf.cmndNmbr)
				_, err = sendToCrocodile(req)
				if err != nil {
					answer = err.Error()
					break
				}
				currcrocstate = ready
				continue

			case currcrocstate == ready:
				req = "http://127.0.0.1:8091/ready"
				ans, err := sendToCrocodile(req)
				if err != nil {
					answer = err.Error()
					break
				}
				msgtosnd := string(ans[:])
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgtosnd)
				Buttn1 := tgbotapi.NewKeyboardButton("Старт")
				row1 := tgbotapi.NewKeyboardButtonRow(Buttn1)
				numericKeyboard := tgbotapi.NewReplyKeyboard()
				numericKeyboard.Keyboard = append(numericKeyboard.Keyboard, row1)
				msg.ReplyMarkup = numericKeyboard
				_, err = bot.Send(msg)
				if err != nil {
					panic(err)
				}
				currcrocstate = start
				continue
			case currcrocstate == start:
				req = "http://127.0.0.1:8091/start"
				ans, err := sendToCrocodile(req)
				if err != nil {
					answer = err.Error()
					break
				}
				msgtosnd := string(ans[:])
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgtosnd)
				Buttn1 := tgbotapi.NewKeyboardButton("+")
				Buttn2 := tgbotapi.NewKeyboardButton("-")
				row1 := tgbotapi.NewKeyboardButtonRow(Buttn1, Buttn2)
				numericKeyboard := tgbotapi.NewReplyKeyboard()
				numericKeyboard.Keyboard = append(numericKeyboard.Keyboard, row1)
				msg.ReplyMarkup = numericKeyboard
				_, err = bot.Send(msg)
				if err != nil {
					panic(err)
				}
				currcrocstate = answr
				continue
			case currcrocstate == answr:
				var ans string
				if answer == "+" {
					ans = "right"
				} else if answer == "-" {
					ans = "skip"
				}
				req = fmt.Sprintf("http://127.0.0.1:8091/answer?value=%s", ans)
				_, err := sendToCrocodile(req)
				if err != nil {
					answer = err.Error()
					break
				}
				currcrocstate = start
				continue
			}
		//	var result Current
		//	err3 := json.Unmarshal(body, &result)
		//	if err3 != nil {
		//		return Current{}, err
		//	}
		//	return result, nil
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

func sendToCrocodile(strtosend string) ([]byte, error) {
	client := http.Client{}
	response, err := client.Get(strtosend)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err := errors.New(fmt.Sprintf("Получен код возврата %d", response.StatusCode))
		return nil, err
	}
	body, err2 := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err2 != nil {
		return nil, err2
		//return Current{}, err
	}
	fmt.Println(body)
	return body, nil
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
