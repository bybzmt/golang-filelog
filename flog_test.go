package flog

import (
	"testing"
	"time"
)

func Test_log(t *testing.T) {
	l := New(LOG_DEBUG, "test")

	err := l.Debug("debug")
	if err != nil {
		t.Errorf("Expect:nil, get:%v", err)
	}

	//己知bug：
	//由于time.Ticker的chan无法闭所以这个go程永远不会停止
	//所以这个SetTick设过N次就会有N-1个永远阻塞的go程！
	l.SetTick(100 * time.Millisecond)
	l.Debug("debug")
	l.Debug("debug")
}
