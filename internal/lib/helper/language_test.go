package helper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetIndefiniteArticle(t *testing.T) {
	assert.Equal(t, "a", GetIndefiniteArticle("Grammar Nazi"))
	assert.Equal(t, "an", GetIndefiniteArticle("ICE 3"))
	assert.Equal(t, "a", GetIndefiniteArticle("university"))
	assert.Equal(t, "a", GetIndefiniteArticle("Human"))
	assert.Equal(t, "a", GetIndefiniteArticle("Dancer"))

}
