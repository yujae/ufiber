package usecase

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/mail"
	"net/smtp"
)

type Mail struct {
	fromName string
	fromMail string
	toName   string
	toMail   string
	subject  string
	body     string
}

func SendMail(c *fiber.Ctx, m Mail) error {
	host := c.Locals("Mail.Host").(string)
	port := c.Locals("Mail.Port").(string)
	ID := c.Locals("Mail.ID").(string)
	PW := c.Locals("Mail.PW").(string)
	fromName := m.fromName
	fromMail := m.fromMail
	toName := m.toName
	toMail := m.toMail
	subject := m.subject
	body := m.body

	to := mail.Address{Name: toName, Address: toMail}
	from := mail.Address{Name: fromName, Address: fromMail}

	header := make(map[string]string)
	header["To"] = to.String()
	header["From"] = from.String()
	header["Subject"] = subject
	header["Content-Type"] = `text/html; charset="UTF-8"`
	msg := ""
	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body
	bMsg := []byte(msg)

	auth := smtp.PlainAuth("", ID, PW, host)
	addr := host + ":" + port

	err := smtp.SendMail(addr, auth, fromMail, []string{toMail}, bMsg)
	if err != nil {
		return nil
	}
	return nil
}
