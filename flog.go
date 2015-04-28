package flog

import (
	"fmt"
	"os"
	"io"
	"strings"
	"regexp"
	"sync"
	"time"
	"log/syslog"
	"net/url"
	"errors"
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
	noclose bool
}

func New(filename, priority, tag string) (Writer, error) {
	_p := log_level(priority)
	if _p == 0 {
		return nil, errors.New("Priority Error")
	}

	switch filename {
	case "" : fallthrough
	case "<stderr>" :
		l := new(Flog)
		l.w = os.Stderr
		l.priority = _p
		l.filter = (_p & severityMask)
		l.tag = tag
		l.noclose = true
		return l, nil
	case "<syslog>" :
		return Dial("", "", _p, tag)
	default:
		ok, _ := regexp.MatchString(`^\w\+`, filename)
		if ok {
			u, err := url.Parse(filename)
			if err != nil {
				return nil, err
			}
			return Dial(u.Scheme, u.Host, _p, tag)
		} else {
			return File(filename, _p, tag)
		}
	}
}

func Dial(network, raddr string, priority Priority, tag string) (*syslog.Writer, error) {
	return syslog.Dial(network, raddr, syslog.Priority(priority), tag)
}

func File(filename string, priority Priority, tag string) (w *Flog, err error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_SYNC, 0666)
	if err != nil {
		return nil, err
	}

	l := new(Flog)
	l.w = f
	l.priority = priority
	l.filter = (priority & severityMask)
	l.tag = tag
	return l, nil
}

func (l *Flog) Init(file string, w io.WriteCloser, priority, filter Priority, tag string) *Flog {

	return l
}

func (w *Flog) SetTag(tag string) {
	w.tag = tag
}

func (w *Flog) SetPriority(priority, filter Priority) {
	w.priority = priority
	w.filter = filter
}

func (w *Flog) Write(b []byte) (int, error) {
	return w.writeAndRetry(w.priority, string(b))
}

func (w *Flog) Close() error {
	if w.noclose {
		return nil
	}

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
	if w.filter < tp {
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

	t1 := time.Now().Format(time.Stamp)

	_, err := fmt.Fprintf(w.w, "<%d>%s %s[%d]: %s%s", p, t1, w.tag, os.Getpid(), msg, nl)
	if err != nil {
		return 0, err
	}

	return len(msg), nil
}

func log_level(level string) Priority {
	level = strings.ToUpper(level)
	sp := strings.SplitN(level, ":", 2)
	if len(sp) < 2 {
		sp = append(sp, "")
		sp[0], sp[1] = sp[1], sp[0]
	}

	var out Priority

	switch sp[0] {
	case "KERN" : out = LOG_KERN
	case "USER" : out = LOG_USER
	case "MAIL" : out = LOG_MAIL
	case "DAEMON" : out = LOG_DAEMON
	case "AUTH" : out = LOG_AUTH
	case "SYSLOG" : out = LOG_SYSLOG
	case "LPR" : out = LOG_LPR
	case "NEWS" : out = LOG_NEWS
	case "UUCP" : out = LOG_UUCP
	case "CRON" : out = LOG_CRON
	case "AUTHPRIV" : out = LOG_AUTHPRIV
	case "FTP" : out = LOG_FTP
	case "" : fallthrough
	case "LOCAL0" : out = LOG_LOCAL0
	case "LOCAL1" : out = LOG_LOCAL1
	case "LOCAL2" : out = LOG_LOCAL2
	case "LOCAL3" : out = LOG_LOCAL3
	case "LOCAL4" : out = LOG_LOCAL4
	case "LOCAL5" : out = LOG_LOCAL5
	case "LOCAL6" : out = LOG_LOCAL6
	case "LOCAL7" : out = LOG_LOCAL7
	default: return 0
	}

	switch sp[1] {
	case "EMERG" : out |= LOG_EMERG
	case "ALERT" : out |= LOG_ALERT
	case "CRIT" : out |= LOG_CRIT
	case "ERR" : out |= LOG_ERR
	case "WARNING" : out |= LOG_WARNING
	case "NOTICE" : out |= LOG_NOTICE
	case "" : fallthrough
	case "INFO" : out |= LOG_INFO
	case "DEBUG" : out |= LOG_DEBUG
	default: return 0
	}

	return out
}

