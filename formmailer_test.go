package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/h2non/filetype"
)

func TestFormHandler(t *testing.T) {
	// Set up environment variables for SMTP (mock values)
	os.Setenv("SMTP_HOST", "smtp.example.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USERNAME", "testuser")
	os.Setenv("SMTP_PASSWORD", "testpass")

	tests := []struct {
		name           string
		method         string
		formData       map[string]string
		files          map[string][]byte
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not supported",
		},
		{
			name:   "Missing receiver",
			method: http.MethodPost,
			formData: map[string]string{
				"name":     "John Doe",
				"email":    "john@example.com",
				"phone":    "123456789",
				"textarea": "Test message",
				"lang":     "en",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Wrong receiver",
		},
		{
			name:   "Invalid email",
			method: http.MethodPost,
			formData: map[string]string{
				"receiver": "test@purelymail.com",
				"name":     "John Doe",
				"email":    "invalid-email",
				"phone":    "123456789",
				"textarea": "Test message",
				"lang":     "en",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email is incorrect",
		},
		{
			name:   "Valid form submission without files",
			method: http.MethodPost,
			formData: map[string]string{
				"receiver": "test@purelymail.com",
				"name":     "John Doe",
				"email":    "john@example.com",
				"phone":    "123456789",
				"textarea": "Test message",
				"lang":     "en",
			},
			expectedStatus: http.StatusOK,
			expectedError:  "Form sent successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			for key, value := range tt.formData {
				writer.WriteField(key, value)
			}
			writer.Close()

			req, err := http.NewRequest(tt.method, "/upload", body)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			formHandler(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check response body
			var response map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Errorf("Failed to decode response: %v", err)
			}

			if status := "success"; tt.expectedStatus >= 200 && tt.expectedStatus < 300 {
				if response[status] != tt.expectedError {
					t.Errorf("Expected success message %q, got %q", tt.expectedError, response[status])
				}
			} else {
				if response["error"] != tt.expectedError {
					t.Errorf("Expected error message %q, got %q", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestFileValidation(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name           string
		fileContent    []byte
		filename       string
		expectedStatus bool
	}{
		{
			name:           "Valid JPEG",
			fileContent:    []byte{0xff, 0xd8, 0xff}, // JPEG header
			filename:       "test.jpg",
			expectedStatus: true,
		},
		{
			name:           "Invalid file type",
			fileContent:    []byte{0x00, 0x00, 0x00}, // Invalid header
			filename:       "test.txt",
			expectedStatus: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a multipart form
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("images", tt.filename)
			if err != nil {
				t.Fatalf("Failed to create form file: %v", err)
			}
			part.Write(tt.fileContent)
			writer.Close()

			req, err := http.NewRequest(http.MethodPost, "/upload", body)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Parse the form
			err = req.ParseMultipartForm(15 << 20)
			if err != nil {
				t.Fatalf("Failed to parse multipart form: %v", err)
			}

			files := req.MultipartForm.File["images"]
			if len(files) == 0 {
				t.Fatal("No files found in form")
			}

			file, err := files[0].Open()
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer file.Close()

			buffer := make([]byte, 261)
			n, err := file.Read(buffer)
			if err != nil && err != io.EOF {
				t.Fatalf("Failed to read file: %v", err)
			}

			kind, err := filetype.Match(buffer[:n])
			if tt.expectedStatus {
				if err != nil || !config.AllowedImageTypes[kind.MIME.Value] {
					t.Errorf("Expected valid file type, got error %v or invalid type %v", err, kind)
				}
			} else {
				if err == nil && config.AllowedImageTypes[kind.MIME.Value] {
					t.Errorf("Expected invalid file type, but validation passed")
				}
			}
		})
	}
}

func TestSendResponse(t *testing.T) {
	tests := []struct {
		name           string
		errorCode      string
		lang           string
		statusCode     int
		expectedStatus string
	}{
		{
			name:           "Success response",
			errorCode:      "form_success",
			lang:           "en",
			statusCode:     http.StatusOK,
			expectedStatus: "success",
		},
		{
			name:           "Error response",
			errorCode:      "email_fail",
			lang:           "en",
			statusCode:     http.StatusInternalServerError,
			expectedStatus: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			sendResponse(rr, tt.errorCode, tt.lang, tt.statusCode)

			if rr.Code != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, rr.Code)
			}

			var response map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Errorf("Failed to decode response: %v", err)
			}

			if _, ok := response[tt.expectedStatus]; !ok {
				t.Errorf("Expected status %q in response, got %v", tt.expectedStatus, response)
			}
		})
	}
}
