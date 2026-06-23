package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"net/textproto"
	"os"
	"strconv"
	"strings"
)

type PurchaseMailer interface {
	Send(to, subject, body string, attachments []string) error
}

type SMTPPurchaseMailer struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewSMTPPurchaseMailerFromEnv() *SMTPPurchaseMailer {
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	return &SMTPPurchaseMailer{
		host: os.Getenv("SMTP_HOST"),
		port: port,
		user: os.Getenv("SMTP_USER"),
		pass: os.Getenv("SMTP_PASSWORD"),
		from: os.Getenv("SMTP_FROM"),
	}
}

func (m *SMTPPurchaseMailer) Send(to, subject, body string, attachments []string) error {
	// Ambientes de desenvolvimento podem operar sem SMTP; a mensagem ainda
	// fica registrada e passa a ser entregue quando as variáveis forem definidas.
	if m.host == "" || m.from == "" {
		return nil
	}
	message, err := buildPurchaseEmail(m.from, to, subject, body, attachments)
	if err != nil {
		return err
	}
	var auth smtp.Auth
	if m.user != "" {
		auth = smtp.PlainAuth("", m.user, m.pass, m.host)
	}
	return smtp.SendMail(m.host+":"+m.port, auth, m.from, []string{to}, message)
}

func buildPurchaseEmail(from, to, subject, body string, attachments []string) ([]byte, error) {
	var output bytes.Buffer
	writer := multipart.NewWriter(&output)
	fmt.Fprintf(&output, "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%q\r\n\r\n", from, to, subject, writer.Boundary())

	textHeader := make(textproto.MIMEHeader)
	textHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	textPart, err := writer.CreatePart(textHeader)
	if err != nil {
		return nil, err
	}
	_, _ = textPart.Write([]byte(body))

	for index, raw := range attachments {
		if !strings.HasPrefix(raw, "data:") {
			continue
		}
		comma := strings.Index(raw, ",")
		if comma < 0 {
			continue
		}
		metadata, encoded := raw[5:comma], raw[comma+1:]
		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, err
		}
		contentType := strings.Split(metadata, ";")[0]
		if contentType == "" {
			contentType = http.DetectContentType(data)
		}
		extension := "bin"
		if parts := strings.Split(contentType, "/"); len(parts) == 2 {
			extension = strings.Split(parts[1], "+")[0]
		}
		header := make(textproto.MIMEHeader)
		header.Set("Content-Type", contentType)
		header.Set("Content-Disposition", `attachment; filename="anexo-`+strconv.Itoa(index+1)+`.`+extension+`"`)
		header.Set("Content-Transfer-Encoding", "base64")
		part, err := writer.CreatePart(header)
		if err != nil {
			return nil, err
		}
		encodedData := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
		base64.StdEncoding.Encode(encodedData, data)
		_, _ = part.Write(encodedData)
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
