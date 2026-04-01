package mailer

import (
	"01cloud-api/api/notifications"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func SendEmail(email, subject, body, link, caption string) notifications.SendEmailMessage {
	return SendMultipleEmail([]string{email}, subject, body, link, caption)
}
func SendMultipleEmail(emails []string, subject, body, link, caption string) notifications.SendEmailMessage {
	return notifications.SendEmailMessage{
		Body:    body,
		User:    emails,
		Subject: subject,
		Link:    link,
		Caption: caption,
	}
}

var mp = map[string]string{}

func GetMessageBody(template string) string {
	path := os.Getenv("EMAIL_TEMPLATE_PATH")
	effectivePath := path + template + ".tpl"
	if v, ok := mp[template]; ok {
		return v
	}
	body, err := os.ReadFile(effectivePath)
	if err == nil {
		mp[template] = string(body)
		return string(body)
	}
	if d, ok := EMAIL_DATA[template]; ok {
		mp[template] = d
		return d
	}
	logrus.Errorf("email template not found :: %s", template)
	return ""
}

func SetMessageBody(tplBody string, template string) error {
	path := os.Getenv("EMAIL_TEMPLATE_PATH")
	effectivePath := path + template + ".tpl"
	err := os.WriteFile(effectivePath, []byte(tplBody), 0644)
	if err != nil {
		logrus.Error("email template write error :: ", err)
		return err
	}
	mp[template] = tplBody
	return nil
}

func replaceMessageBody(body string, mappedValue map[string]string) string {
	for key, value := range mappedValue {
		body = strings.ReplaceAll(body, key, value)
	}
	return body
}
