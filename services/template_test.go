package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderToString(t *testing.T) {
	tmplMgr, err := NewTemplateService("../utilities/templates")
	assert.NoError(t, err, "Template manager should load without errors")

	data := map[string]interface{}{
		"ResetLink": "http://marketplace.com/user/password-reset/1234",
	}
	output, err := tmplMgr.RenderToString("password_reset.html", data)
	assert.NoError(t, err, "Rendering should not return an error")
	assert.Contains(t, output, "If you did not request a password reset, disregard this email.")
}
