package utilities

import (
	"bytes"
	"fmt"
	"text/template"
)

// GeneratePasswordResetEmail emails a link to reset a user's password
func GeneratePasswordResetEmail(recipient string, code string) (string, error) {
	data := map[string]string{
		"ResetLink": fmt.Sprintf("%s/users/password-reset/confirm?email=%s&code=%s", getEnv("BASE_URL"), recipient, code),
	}
	tmpl, err := template.ParseFiles("utilities/templates/password_reset.html")
	if err != nil {
		return "", err
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, data)
	if err != nil {
		return "", err
	}

	return body.String(), nil
}
