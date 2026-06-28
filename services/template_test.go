package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderToString(t *testing.T) {
	templates, err := loadTemplates("../utilities/templates")
	assert.NoError(t, err, "Loading templates should not return an error")

	tmplMgr := &templateService{templates}

	data := map[string]interface{}{
		"ResetLink": "http://marketplace.com/user/password-reset/1234",
	}
	output, err := tmplMgr.RenderHtmlToString(EmailPasswordReset, data)
	assert.NoError(t, err, "Rendering should not return an error")
	assert.Contains(t, output, "If you did not request a password reset, disregard this email.")
}
