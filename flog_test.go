package flog

import (
	"testing"
)

func Test_log(t *testing.T) {
	l := New(LOG_DEBUG, "test")

	err := l.Debug("debug")
	if err != nil {
		t.Errorf("Expect:nil, get:%v", err)
	}
}
