package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAPP(t *testing.T) {
	assert.Equal(t, GetAPP().Instance.ID, app.Instance.ID)
	assert.Equal(t, GetAPP().Instance.ID, app.Instance.ID)
}
