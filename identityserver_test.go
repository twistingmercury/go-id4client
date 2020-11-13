package id4client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/twistingmercury/go-id4client"
)

func TestIdentityConfig(t *testing.T) {
	var err error
	c := id4client.IdentityConfig{
		BaseURL:        "[ID4 BASE URL]",
		ID:             "[APP ID]",
		Secret:         "[APP SECRET]",
		IntrospectPath: "/connect/instrospect",
		TokenPath:      "/connect/token",
		ServiceName:    "[APP NAME]",
		ServiceVersion: "[APP VERSION]",
		CommitHash:     "[GIT COMMIT HASH",
	}

	err = id4client.InitConfig(c)
	assert.NoError(t, err)

	c.CommitHash = ""

	err = id4client.InitConfig(c)
	assert.NoError(t, err)
}
