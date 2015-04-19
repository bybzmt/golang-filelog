package flog

import (
	"testing"
)

func Test_log(t *testing.T) {
	l := New(LOG_LOCAL0|LOG_INFO, "test")

	err := l.Debug("Debug 这个不显示")
	if err != nil {
		t.Errorf("Expect:nil, get:%v", err)
	}

	l.Info("Info 这个应该显示")
	l.Notice("Notice 这个应该显示")
	l.Notice("Warning 这个应该显示")
}
