// config.go
package main

// Config przechowuje ustawienia aplikacji.
type Config struct {
	AllowedImageTypes map[string]bool
	AllowedEmails     []string
	ErrorMessages     map[string]map[string]string
	CORS              []string
}

// DefaultConfig zwraca domyślną konfigurację aplikacji.
func DefaultConfig() Config {
	return Config{
		AllowedImageTypes: map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/webp": true,
		},
		AllowedEmails: []string{
			"mateusz.chodacki@gmail.com",
			"office@gmtechnics.com",
			"customsteel.electroworks@gmail.com",
			"test@purelymail.com",
		},
		ErrorMessages: map[string]map[string]string{
			"en": {
				"method_not_supported": "Method not supported",
				"wrong_receiver":       "Wrong receiver",
				"name_required":        "Name is required",
				"email_required":       "Email is required",
				"email_incorrect":      "Email is incorrect",
				"phone_required":       "Phone number is incorrect",
				"message_too_long":     "Message is too long",
				"too_many_files":       "Too many files",
				"file_size":            "File is too big",
				"file_cant_open":       "File couldn't be opened",
				"file_cant_read":       "File couldn't be read",
				"file_too_small":       "File is too small to be an image",
				"unknown_file_type":    "Unknown file type",
				"unsupported_file":     "Unsupported file type",
				"email_fail":           "Error while trying to send email",
				"form_success":         "Form sent successfully",
			},
			"pl": {
				"method_not_supported": "Metoda nie wspierana",
				"wrong_receiver":       "Nieprawidłowy odbiorca",
				"name_required":        "Imię jest wymagane",
				"email_required":       "Email jest wymagany",
				"email_incorrect":      "Email jest nieprawidłowy",
				"phone_required":       "Numer telefonu jest nieprawidłowy",
				"message_too_long":     "Wiadomość jest za długa",
				"too_many_files":       "Zbyt wiele plików",
				"file_size":            "Plik jest za duży",
				"file_cant_open":       "Nie można otworzyć pliku",
				"file_cant_read":       "Nie można odczytać pliku",
				"file_too_small":       "Plik jest za mały, aby być obrazem",
				"unknown_file_type":    "Nieznany typ pliku",
				"unsupported_file":     "Nieobsługiwany typ pliku",
				"email_fail":           "Błąd podczas próby wysyłania emaila",
				"form_success":         "Formularz wysłany pomyślnie",
			},
			"de": {
				"method_not_supported": "Methode wird nicht unterstützt",
				"wrong_receiver":       "Falscher Empfänger",
				"name_required":        "Name ist erforderlich",
				"email_required":       "E-Mail ist erforderlich",
				"email_incorrect":      "E-Mail ist ungültig",
				"phone_required":       "Telefonnummer ist ungültig",
				"message_too_long":     "Nachricht ist zu lang",
				"too_many_files":       "Zu viel Datei",
				"file_size":            "Datei ist zu groß",
				"file_cant_open":       "Datei konnte nicht geöffnet werden",
				"file_cant_read":       "Datei konnte nicht gelesen werden",
				"file_too_small":       "Datei ist zu klein, um ein Bild zu sein",
				"unknown_file_type":    "Unbekannter Dateityp",
				"unsupported_file":     "Nicht unterstützter Dateityp",
				"email_fail":           "Fehler beim Versuch, eine E-Mail zu senden",
				"form_success":         "Formular erfolgreich gesendet",
			},
		},
		CORS: []string{
			"http://localhost:4321",
			"https://www.custom-steel.eu",
		},
	}

}
