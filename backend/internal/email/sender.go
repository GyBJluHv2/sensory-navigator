// Package email инкапсулирует отправку писем подтверждения регистрации.
//
// Реализация выбирается во время старта по конфигурации:
//   - если SMTP_HOST пустой -> используется LogSender, который пишет письмо
//     в стандартный лог (для разработки и автотестов);
//   - иначе используется SMTPSender, выполняющий реальный SMTP-обмен.
//
// Такой дизайн позволяет защитному курсовому стенду работать без живого SMTP
// и одновременно даёт штатный режим отправки через любой почтовый
// провайдер (Yandex/Gmail/Mailtrap/др.).
package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// Sender — интерфейс отправителя писем (DI).
type Sender interface {
	SendVerificationCode(toEmail, code string) error
}

// LogSender пишет содержимое письма в стандартный лог.
// Используется в DEV-окружении и в тестах.
type LogSender struct{}

func (LogSender) SendVerificationCode(toEmail, code string) error {
	log.Printf("[email-stub] -> %s: код подтверждения = %s", toEmail, code)
	return nil
}

// SMTPSender отправляет письма через SMTP-сервер.
// Поддерживает STARTTLS (Yandex/Gmail/SendGrid/Mailtrap и т. п.).
type SMTPSender struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
	UseTLS   bool
}

func (s SMTPSender) SendVerificationCode(toEmail, code string) error {
	subject := "Код подтверждения регистрации в Сенсорном навигаторе"
	body := fmt.Sprintf(`Здравствуйте!

Чтобы завершить регистрацию в приложении «Сенсорный навигатор»,
введите следующий код подтверждения:

    %s

Код действителен в течение 15 минут.

Если вы не запрашивали регистрацию, просто проигнорируйте это письмо.

— Команда Сенсорного навигатора`, code)

	msg := buildMimeMessage(s.From, toEmail, subject, body)
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)

	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)

	if s.UseTLS {
		return sendWithStartTLS(addr, s.Host, auth, s.From, []string{toEmail}, msg)
	}
	return smtp.SendMail(addr, auth, s.From, []string{toEmail}, msg)
}

// buildMimeMessage формирует RFC 5322-сообщение с UTF-8 кодировкой темы и тела.
func buildMimeMessage(from, to, subject, body string) []byte {
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: =?UTF-8?B?" + base64Encode(subject) + "?=",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: base64",
	}
	return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + base64Encode(body))
}

func base64Encode(s string) string {
	const pad = '='
	const tab = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	src := []byte(s)
	out := make([]byte, 0, ((len(src)+2)/3)*4)
	for i := 0; i < len(src); i += 3 {
		b1 := src[i]
		var b2, b3 byte
		var lim int
		if i+1 < len(src) {
			b2 = src[i+1]
			lim = 1
		}
		if i+2 < len(src) {
			b3 = src[i+2]
			lim = 2
		}
		out = append(out,
			tab[b1>>2],
			tab[((b1&0x03)<<4)|(b2>>4)],
		)
		if lim >= 1 {
			out = append(out, tab[((b2&0x0f)<<2)|(b3>>6)])
		} else {
			out = append(out, byte(pad))
		}
		if lim >= 2 {
			out = append(out, tab[b3&0x3f])
		} else {
			out = append(out, byte(pad))
		}
	}
	return string(out)
}

// sendWithStartTLS подключается к SMTP, выполняет STARTTLS, аутентификацию
// и отправляет письмо. Аналог smtp.SendMail, но с принудительным TLS.
func sendWithStartTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth: %w", err)
			}
		}
	}

	if err := c.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	for _, t := range to {
		if err := c.Rcpt(t); err != nil {
			return fmt.Errorf("smtp rcpt %s: %w", t, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	return c.Quit()
}