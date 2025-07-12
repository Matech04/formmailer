package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/mail"

	"github.com/go-gomail/gomail"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/cors"
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func sendResponse(w http.ResponseWriter, message string, statusCode int) {
	var status string
	if statusCode == http.StatusOK || statusCode == http.StatusCreated || statusCode == http.StatusAccepted {
		status = "succes"
	} else {
		status = "error"
	}
	response := map[string]string{status: message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
	log.Println(message)
}

func sendEmail(name, email, phone, textarea string, fileHeaders []*multipart.FileHeader) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "m@ch.pl")
	m.SetHeader("To", "mateusz.chodacki@gmail.com")
	m.SetHeader("Subject", "Test")
	m.SetBody("text/html", fmt.Sprintf("<p>%s<p/><br/><p>%s<p/><br/><p>%s<p/><br/><p>%s<p/>", name, email, phone, textarea))

	for _, filefileHeader := range fileHeaders {
		file, err := filefileHeader.Open()
		if err != nil {
			log.Printf("Error while opening attachment %s: %v", filefileHeader.Filename, err)
			break
		}
		defer file.Close()
		m.Attach("image.jpg", bytes.NewReader(file), gomail.SetHeader(map[string][]string{
			"Content-ID": {"<image123>"}, // Content-ID do użycia w HTML
		}))

	}
}

func formHandler(w http.ResponseWriter, r *http.Request) {

	p := bluemonday.UGCPolicy()

	if r.Method != http.MethodPost {
		sendResponse(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(15 << 20)

	name := p.Sanitize(r.FormValue("name"))
	if name == "" {
		sendResponse(w, "Name is required", http.StatusBadRequest)
		return

	}
	email := p.Sanitize(r.FormValue("email"))
	if email == "" {
		sendResponse(w, "Email is required", http.StatusBadRequest)
		return
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		sendResponse(w, "Email is not correct", http.StatusBadRequest)
		return
	}

	phone := p.Sanitize(r.FormValue("phone"))
	if phone == "" {
		sendResponse(w, "Phone is required", http.StatusBadRequest)
		return
	}

	//Max znaki
	//textarea := p.Sanitize(r.FormValue("textarea"))

	const maxFileSize = 10 << 20

	files := r.MultipartForm.File["images"]
	log.Printf("Recived %d files", len(files))
	var validatedFileHeaders []*multipart.FileHeader

	for _, fileHeader := range files {

		if fileHeader.Size > maxFileSize {
			sendResponse(w, fmt.Sprintf("File %s is too big", fileHeader.Filename), http.StatusRequestEntityTooLarge)
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			sendResponse(w, "Couldn't open file", http.StatusInternalServerError)
			return
		}

		buffer := make([]byte, 261)
		n, err := file.Read(buffer)
		if err != nil {
			sendResponse(w, "Trouble with reading file", http.StatusBadRequest)
			return
		}
		if n < 261 {
			sendResponse(w, "File is too small to be a photo", http.StatusBadRequest)
			return
		}

		kind, err := filetype.Match(buffer[:n])
		if err != nil || kind == types.Unknown {
			sendResponse(w, "Unknown file type", http.StatusBadRequest)
			return
		}

		if !allowedImageTypes[kind.MIME.Value] {
			sendResponse(w, "Unsupported file type", http.StatusBadRequest)
			return
		}

		// Validation passed adding file to array
		validatedFileHeaders = append(validatedFileHeaders, fileHeader)
		log.Printf("File %s is valid (%s)", fileHeader.Filename, kind.MIME.Value)
		file.Close()

	}

	sendResponse(w, "Form sent with success", http.StatusOK)

}
func main() {

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4321"},
		AllowedMethods:   []string{"POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
	})

	http.HandleFunc("/upload", formHandler)

	handler := c.Handler(http.DefaultServeMux)

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
