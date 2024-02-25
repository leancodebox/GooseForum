package asyncwrite

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

type AsyncW struct {
	io       *lumberjack.Logger
	dataChan chan *[]byte
	done     chan bool
}

func AsyncLumberjack(io *lumberjack.Logger) *AsyncW {
	r := AsyncW{
		io:       io,
		dataChan: make(chan *[]byte, 1024),
		done:     make(chan bool),
	}
	go func() {
		ticker := time.NewTicker(time.Millisecond * 333) // 每0.333秒执行一次批量写入
		defer ticker.Stop()
		var data []byte
		for {
			select {
			case val := <-r.dataChan:
				data = append(data, *val...)
				if len(data) >= 100 {
					r.io.Write(data)
					data = []byte{}
				}
			case <-ticker.C:
				if len(data) > 0 {
					r.io.Write(data)
					data = []byte{}
				}
			case <-r.done:
				if len(data) > 0 {
					r.io.Write(data)
					data = []byte{}
				}
				return
			}
		}
	}()
	return &r
}

func (itself *AsyncW) Write(p []byte) (n int, err error) {
	data := make([]byte, len(p))
	copy(data, p)
	itself.dataChan <- &data
	return 0, nil
}

func (itself *AsyncW) Stop() {
	itself.done <- true
	itself.io.Close()
	close(itself.dataChan)
	close(itself.done)
}
