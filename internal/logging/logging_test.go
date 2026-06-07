package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupLogging(t *testing.T) {
	assert.NotNil(t, SetupLogging())
}
