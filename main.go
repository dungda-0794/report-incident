package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type params struct {
	WebhookSettingId string `json:"webhook_setting_id"`
	WebhookEventType string `json:"webhook_event_type"`
	WebhookEvent     struct {
		FromAccountId int    `json:"from_account_id"`
		ToAccountId   int    `json:"to_account_id"`
		RoomId        int    `json:"room_id"`
		MessageId     string `json:"message_id"`
		Body          string `json:"body"`
	} `json:"webhook_event"`
}

func main() {
	viper.SetConfigFile(`infrastructure/config.json`)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Errors: not exists or is wrong json format", err)
	}

	e := echo.New()
	e.POST("/report", report)
	e.Logger.Fatal(e.Start(":" + viper.GetString(`serve.port`)))
}

func report(c echo.Context) error {
	p := &params{}
	if err := c.Bind(p); err != nil {
		return err
	}
	matched, _ := regexp.MatchString(viper.GetString(`regex`), p.WebhookEvent.Body)
	if matched {
		body := strings.Replace(p.WebhookEvent.Body, viper.GetString(`chatwork.remove`), "", 1)
		fmt.Println(time.Now().Format(time.RFC850))
		toChatWork(body)
		return c.JSON(http.StatusOK, nil)
	}

	return c.JSON(http.StatusOK, nil)
}

func toChatWork(body string) error {
	to := viper.GetString(`chatwork.to`)
	room := viper.GetString(`chatwork.room`)
	content := fmt.Sprintf(
		"%s\n%s",
		to,
		body,
	)

	client := &http.Client{}
	endpoint := fmt.Sprintf("https://api.chatwork.com/v2/rooms/%s/messages?body=%s", room, url.QueryEscape(content))
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("X-ChatWorkToken", viper.GetString(`chatwork.token`))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Execute request
	res, _ := client.Do(req)
	defer res.Body.Close()

	return nil
}
