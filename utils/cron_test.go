package utils

import (
	"testing"
	"time"

	"gopkg.in/go-playground/assert.v1"
)

// Tests that un-register works as expected.
func TestCron(t *testing.T) {
	prov := NewCron()
	called := 0
	var id int
	id, _ = prov.AddFunc("@every 1s", func() {
		called++
		if 2 == called {
			prov.RemoveFunc(id)
		}
	})

	time.Sleep(4 * time.Second)
	assert.Equal(t, 2, called)
}
