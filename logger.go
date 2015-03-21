package borm

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fatih/color"
)

type logger struct {
	on  bool
	end chan error
}

func logit(log bool, meth string, buckets []string, id string, i ...interface{}) *logger {
	l := &logger{on: log, end: make(chan error)}
	if !log {
		return l
	}
	go func() {
		start := time.Now().UnixNano()
		var data string
		if len(i) > 0 {
			switch i[0].(type) {
			case mod:
				d, _ := json.Marshal(i[0])
				data = string(d)
			case string:
				data = i[0].(string)
			case []string:
				data = strings.Join(i[0].([]string), ",")
			case []byte:
				data = string(i[0].([]byte))
			}
		}
		l.out(start, meth, buckets, id, data, <-l.end)
	}()
	return l
}

func (l *logger) done(err error) error {
	if l.on {
		l.end <- err
	}
	return err
}

func (l *logger) out(start int64, meth string, buckets []string, id string, i string, err error) {
	dif := float64((time.Now().UnixNano()-start)/int64(time.Millisecond)) / 1000
	msg := fmt.Sprintf("BOLTDB[%.3fs]: %s %s:%s %s", dif, meth, strings.Join(buckets, "/"), id, i)
	if err != nil {
		msg = fmt.Sprintf("%s Eror: %s", msg, err.Error())
		log.Println(color.RedString(msg))
	} else {
		log.Println(color.GreenString(msg))
	}
}
