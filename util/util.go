package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

var tgAlerter TgAlerter

type TgAlerter struct {
	BotId  string
	ChatId string
}

func InitTgAlerter(cfg *AlertConfig) {
	tgAlerter = TgAlerter{
		BotId:  cfg.TelegramBotId,
		ChatId: cfg.TelegramChatId,
	}
}

func SendTelegramMessage(msg string) {
	if tgAlerter.BotId == "" || tgAlerter.ChatId == "" || msg == "" {
		return
	}

	endPoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tgAlerter.BotId)
	formData := url.Values{
		"chat_id":    {tgAlerter.ChatId},
		"parse_mode": {"html"},
		"text":       {msg},
	}
	Logger.Infof("send tg message, bot_id=%s, chat_id=%s, msg=%s", tgAlerter.BotId, tgAlerter.ChatId, msg)
	res, err := http.PostForm(endPoint, formData)
	if err != nil {
		Logger.Errorf("send telegram message error, bot_id=%s, chat_id=%s, msg=%s, err=%s", tgAlerter.BotId, tgAlerter.ChatId, msg, err.Error())
		return
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		Logger.Errorf("read http response error, err=%s", err.Error())
		return
	}
	Logger.Infof("tg response: %s", string(bodyBytes))
}
