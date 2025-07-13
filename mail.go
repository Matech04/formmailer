package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strconv"

	"github.com/go-gomail/gomail"
)

func sendEmail(receiver, name, email, phone, textarea string, fileHeaders []*multipart.FileHeader) error {

	from := os.Getenv("SMTP_FROM")

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", receiver)
	m.SetHeader("Subject", "Test")
	m.SetBody("text/html", fmt.Sprintf("<p>%s<p/><br/><p>%s<p/><br/><p>%s<p/><br/><p>%s<p/>", name, email, phone, textarea))

	if len(fileHeaders) != 0 {
		for _, fileHeader := range fileHeaders {
			file, err := fileHeader.Open()
			if err != nil {
				log.Printf("Error while opening attachment %s: %v", fileHeader.Filename, err)
				return fmt.Errorf("failed to open attachment %s: %w", fileHeader.Filename, err)
			}
			defer file.Close()

			m.Attach(
				fileHeader.Filename,
				gomail.SetCopyFunc(func(w io.Writer) error {
					_, err := io.Copy(w, file)
					if err != nil {
						return fmt.Errorf("failed to copy file %s: %w", fileHeader.Filename, err)
					}
					return nil
				}),
				gomail.SetHeader(map[string][]string{
					"Content-Type": {fileHeader.Header.Get("Content-Type")},
				}))
		}
	}

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")

	if host == "" || port == "" || username == "" || password == "" {
		return fmt.Errorf("missing SMTP configuration")
	}

	// string to int
	port_int, err := strconv.Atoi(port)
	if err != nil {
		// ... handle error
		return fmt.Errorf("error while converting string port to int: %v", err)
	}

	d := gomail.NewDialer(host, port_int, username, password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error while sending mail %w", err)
	}

	fmt.Println("Email wysłany pomyślnie")

	return nil
}
