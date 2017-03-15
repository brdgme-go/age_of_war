package age_of_war

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGame_Start(t *testing.T) {
	g := &Game{}
	_, err := g.Start(3)
	assert.NoError(t, err)
}
