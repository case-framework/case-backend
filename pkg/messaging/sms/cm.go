package sms

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type SMSTo struct {
	Number string `json:"number"`
}

type SMSBody struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type SingleSMS struct {
	AllowedChannels []string `json:"allowedChannels"`
	From            string   `json:"from"`
	To              []SMSTo  `json:"to"`
	Body            SMSBody  `json:"body"`
}

type SMSAuth struct {
	Producttoken string `json:"producttoken"`
}

type SMSSendingReq struct {
	Messages struct {
		Authentication SMSAuth     `json:"authentication"`
		Msg            []SingleSMS `json:"msg"`
	} `json:"messages"`
}

func runSMSsending(to string, message string, from string) error {
	if SmsGatewayConfig == nil || SmsGatewayConfig.URL == "" {
		return errors.New("connection to sms gateway not initialized")
	}

	payload := SMSSendingReq{
		Messages: struct {
			Authentication SMSAuth     `json:"authentication"`
			Msg            []SingleSMS `json:"msg"`
		}{
			Authentication: SMSAuth{
				Producttoken: SmsGatewayConfig.APIKey,
			},
			Msg: []SingleSMS{
				{
					AllowedChannels: []string{"SMS"},
					From:            from,
					To: []SMSTo{
						{
							Number: to,
						},
					},
					Body: SMSBody{
						Type:    "auto",
						Content: message,
					},
				},
			},
		},
	}

	json_data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(SmsGatewayConfig.URL, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		slog.Error("sms gateway returned error", slog.String("status", resp.Status))
		return errors.New("sms gateway returned error")
	}

	var res map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		slog.Error("Error decoding response", slog.String("error", err.Error()))
		return err
	}

	errorCode, ok := res["errorCode"]
	if !ok {
		slog.Error("no error code in response")
		return errors.New("no error code in response")
	}

	errorCodeInt, ok := errorCode.(int)
	if !ok {
		slog.Error("error code is not a number")
		return errors.New("error code is not a number")
	}
	if errorCodeInt != 0 {
		slog.Error("sms gateway returned error", slog.Int("errorCode", int(errorCodeInt)))
		return errors.New("sms gateway returned error")
	}

	return nil
}
