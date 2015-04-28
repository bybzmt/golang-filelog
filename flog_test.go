package flog

import (
	"testing"
)

func Test_log(t *testing.T) {
	l, err := New("", "local0:Notice", "test")
	if err != nil {
		t.Errorf("Expect:nil, get:%v", err)
		return
	}
	defer l.Close()

	err = l.Debug("Debug 这个不显示")
	if err != nil {
		t.Errorf("Expect:nil, get:%v", err)
	}

	l.Info("Info 这个不显示")
	l.Notice("Notice 这个应该显示")
	l.Notice("Warning 这个应该显示")
}

func Test_log2(t *testing.T) {
	l, err := New("", "", "test")
	if err != nil {
		t.Errorf("Expect:nil, get:%v", err)
		return
	}
	defer l.Close()

	err = l.Debug("Debug 这个不显示")
	if err != nil {
		t.Errorf("Expect:nil, get:%v", err)
	}

	l.Info("Info 这个应该显示")
	l.Notice("Notice 这个应该显示")
	l.Notice("Warning 这个应该显示")
}
