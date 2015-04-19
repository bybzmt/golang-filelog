package flog

import (
	"fmt"
	"os"
	"io"
	"strings"
	"sync"
	"time"
)

type Priority int

const severityMask = 0x07
const facilityMask = 0xf8

const (
	LOG_EMERG Priority = iota
	LOG_ALERT
	LOG_CRIT
	LOG_ERR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
)

const (
	LOG_KERN Priority = iota << 3
	LOG_USER
	LOG_MAIL
	LOG_DAEMON
	LOG_AUTH
	LOG_SYSLOG
	LOG_LPR
	LOG_NEWS
	LOG_UUCP
	LOG_CRON
	LOG_AUTHPRIV
	LOG_FTP
	_ // unused
	_ // unused
	_ // unused
	_ // unused
	LOG_LOCAL0
	LOG_LOCAL1
	LOG_LOCAL2
	LOG_LOCAL3
	LOG_LOCAL4
	LOG_LOCAL5
	LOG_LOCAL6
	LOG_LOCAL7
)

type Writer interface {
	Alert(m string) (err error)
	Close() error
	Crit(m string) (err error)
	Debug(m string) (err error)
	Emerg(m string) (err error)
	Err(m string) (err error)
	Info(m string) (err error)
	Notice(m string) (err error)
	Warning(m string) (err error)
	Write(b []byte) (int, error)
}

type Flog struct {
	priority Priority
	filter Priority
	tag      string
	mu   sync.Mutex
	w io.WriteCloser
	stamp string
	now chan time.Time
	ticker *time.Ticker
}

func New(priority Priority, tag string) *Flog {
	return new(Flog).Init(os.Stderr, priority, (priority & severityMask), tag)
}

func NewFile(filename string, priority Priority, tag string) (w *Flog, err error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_SYNC, 0666)
	if err != nil {
		return nil, err
	}

	w = new(Flog).Init(f, priority, (priority & severityMask), tag)
	return w, nil
}

func (l *Flog) Init(w io.WriteCloser, priority, filter Priority, tag string) *Flog {
	l.w = w
	l.priority = priority
	l.filter = filter
	l.tag = tag
	return l
}

func (w *Flog) SetTag(tag string) {
	w.tag = tag
}

func (w *Flog) SetPriority(priority, filter Priority) {
	w.priority = priority
	w.filter = filter
}

// 设置log记录的时间变化频率
// 这个设置只有在你的log调用频率过高系统时间调用
// 己变成瓶颈了才有意义。
// （^_^~..表示不知道什么情况才会有如此大量的log)
func (w *Flog) SetTick(precision time.Duration) {
	if w.ticker != nil {
		w.ticker.Stop()
		w.ticker = nil
	}

	if (precision < 1) {
		return
	}

	w.ticker = time.NewTicker(precision)

	//己知bug：
	//由于time.Ticker的chan无法闭所以这个go程永远不会停止
	//所以这个SetTick设过N次就会有N-1个永远阻塞的go程！
	go func() {
		for now := range w.ticker.C {
			w.stamp = now.Format(time.Stamp)
		}
	}()
}

func (w *Flog) Write(b []byte) (int, error) {
	return w.writeAndRetry(w.priority, string(b))
}

func (w *Flog) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.Close()
}

func (w *Flog) Emerg(m string) (err error) {
	_, err = w.writeAndRetry(LOG_EMERG, m)
	return err
}

func (w *Flog) Alert(m string) (err error) {
	_, err = w.writeAndRetry(LOG_ALERT, m)
	return err
}

func (w *Flog) Crit(m string) (err error) {
	_, err = w.writeAndRetry(LOG_CRIT, m)
	return err
}

func (w *Flog) Err(m string) (err error) {
	_, err = w.writeAndRetry(LOG_ERR, m)
	return err
}

func (w *Flog) Warning(m string) (err error) {
	_, err = w.writeAndRetry(LOG_WARNING, m)
	return err
}

func (w *Flog) Notice(m string) (err error) {
	_, err = w.writeAndRetry(LOG_NOTICE, m)
	return err
}

func (w *Flog) Info(m string) (err error) {
	_, err = w.writeAndRetry(LOG_INFO, m)
	return err
}

func (w *Flog) Debug(m string) (err error) {
	_, err = w.writeAndRetry(LOG_DEBUG, m)
	return err
}

func (w *Flog) writeAndRetry(p Priority, s string) (int, error) {
	tp := p & severityMask
	if w.filter > tp {
		return 0, nil
	}

	pr := (w.priority & facilityMask) | tp

	w.mu.Lock()
	defer w.mu.Unlock()

	return w.write(pr, s)
}

func (w *Flog) write(p Priority, msg string) (int, error) {
	nl := ""
	if !strings.HasSuffix(msg, "\n") {
		nl = "\n"
	}

	if w.ticker == nil {
		w.stamp = time.Now().Format(time.Stamp)
	}

	_, err := fmt.Fprintf(w.w, "<%d>%s %s[%d]: %s%s", p, w.stamp, w.tag, os.Getpid(), msg, nl)
	if err != nil {
		return 0, err
	}

	return len(msg), nil
}



