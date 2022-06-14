package web

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestChooseLoggerFilenameTestable(t *testing.T) {
	{
		mytime, err := time.Parse("Jan 2 15:04:05 2006 MST", "Jan 2 15:04:05 2006 MST")
		assert.Nil(t, err)
		got := chooseLoggerFilenameTestable("foo.log", mytime, 123)
		expected := "foo.06-01-02.123.log"
		assert.Equal(t, expected, got)
	}
}
func TestChooseLoggerFilenameLegacyTestable(t *testing.T) {
	{
		mytime, err := time.Parse("Jan 2 15:04:05 2006 MST", "Jan 2 15:04:05 2006 MST")
		assert.Nil(t, err)
		got := chooseLoggerFilenameLegacyTestable("foo.log", mytime)
		expected := "foo.06-01-02.log"
		assert.Equal(t, expected, got)
	}
}
