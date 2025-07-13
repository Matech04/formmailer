package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/mail"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"github.com/microcosm-cc/bluemonday"
)

func sendResponse(w http.ResponseWriter, errorCode, lang string, statusCode int, config Config) {

	message := config.ErrorMessages[lang][errorCode]
	if message == "" {
		message = config.ErrorMessages["en"][errorCode]
		if message == "" {
			message = "Error message not found"
		}
	}

	var status string
	if statusCode == http.StatusOK || statusCode == http.StatusCreated || statusCode == http.StatusAccepted {
		status = "success"
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

func formHandler(w http.ResponseWriter, r *http.Request, config Config) {

	p := bluemonday.UGCPolicy()

	if r.Method != http.MethodPost {
		sendResponse(w, "method_not_supported", "en", http.StatusMethodNotAllowed, config)
		return
	}

	lang := p.Sanitize(r.FormValue("lang"))
	if _, ok := config.ErrorMessages[lang]; !ok {
		lang = "en"
	}

	r.ParseMultipartForm(15 << 20)

	receiver := p.Sanitize(r.FormValue("receiver"))
	validReceiver := false

	for _, allowedEmail := range config.AllowedEmails {
		if receiver == allowedEmail {
			validReceiver = true
			break
		}
	}

	if !validReceiver {
		sendResponse(w, "wrong_receiver", lang, http.StatusBadRequest, config)
		return
	}

	name := p.Sanitize(r.FormValue("name"))
	if name == "" {
		sendResponse(w, "name_required", lang, http.StatusBadRequest, config)
		return

	}
	email := p.Sanitize(r.FormValue("email"))
	if email == "" {
		sendResponse(w, "email_required", lang, http.StatusBadRequest, config)
		return
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		sendResponse(w, "email_incorrect", lang, http.StatusBadRequest, config)
		return
	}

	phone := p.Sanitize(r.FormValue("phone"))
	if phone == "" {
		sendResponse(w, "phone_required", lang, http.StatusBadRequest, config)
		return
	}

	textarea := p.Sanitize(r.FormValue("textarea"))
	if len(textarea) > 1000 {
		sendResponse(w, "message_too_long", lang, http.StatusBadRequest, config)
		return
	}

	const maxFileSize = 10 << 20

	files := r.MultipartForm.File["images"]
	log.Printf("Recived %d files", len(files))
	var validatedFileHeaders []*multipart.FileHeader

	if len(files) > 3 {
		sendResponse(w, "too_many_files", lang, http.StatusBadRequest, config)
		return
	}

	if len(files) > 0 {

		for _, fileHeader := range files {

			if fileHeader.Size > maxFileSize {
				sendResponse(w, "file_size", lang, http.StatusRequestEntityTooLarge, config)
				return
			}

			file, err := fileHeader.Open()
			if err != nil {
				sendResponse(w, "file_cant_open", lang, http.StatusInternalServerError, config)
				return
			}

			defer file.Close()

			buffer := make([]byte, 261)
			n, err := file.Read(buffer)
			if err != nil {
				sendResponse(w, "file_cant_read", lang, http.StatusBadRequest, config)
				return
			}
			if n < 261 {
				sendResponse(w, "file_too_small", lang, http.StatusBadRequest, config)
				return
			}

			kind, err := filetype.Match(buffer[:n])
			if err != nil || kind == types.Unknown {
				sendResponse(w, "unknown_file_type", lang, http.StatusBadRequest, config)
				return
			}

			if !config.AllowedImageTypes[kind.MIME.Value] {
				sendResponse(w, "unsupported_file", lang, http.StatusBadRequest, config)
				return
			}

			// Validation passed adding file to array
			validatedFileHeaders = append(validatedFileHeaders, fileHeader)
			log.Printf("File %s is valid (%s)", fileHeader.Filename, kind.MIME.Value)

		}
	}

	err = sendEmail(receiver, name, email, phone, textarea, validatedFileHeaders)
	if err != nil {
		fmt.Printf("%s", err)
		sendResponse(w, "email_fail", lang, http.StatusInternalServerError, config)
		return
	}

	sendResponse(w, "form_success", lang, http.StatusOK, config)

}
