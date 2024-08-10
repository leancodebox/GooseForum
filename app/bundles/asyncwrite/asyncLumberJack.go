package asyncwrite

import (
	"bytes"
	"context"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

type AsyncW struct {
	io          *lumberjack.Logger
	dataChan    chan []byte
	closeFinish chan bool
}

func AsyncLumberjack(io *lumberjack.Logger) *AsyncW {
	r := &AsyncW{
		io:          io,
		dataChan:    make(chan []byte, 256),
		closeFinish: make(chan bool),
	}
	go func() {
		defer func() {
			r.io.Close()
			r.closeFinish <- true
			return
		}()
		ticker := time.NewTicker(time.Millisecond * 100) // 每0.333秒执行一次批量写入
		defer ticker.Stop()
		var buf bytes.Buffer
		maxLen := 1024 * 8
		for {
			select {
			case val, ok := <-r.dataChan:
				if !ok {
					if buf.Len() > 0 {
						_, err := r.io.Write(buf.Bytes())
						if err != nil {
							fmt.Println(err)
						}
						buf.Reset()
					}
					return
				}

				buf.Write(val)
				if buf.Len() >= maxLen {
					_, err := r.io.Write(buf.Bytes())
					if err != nil {
						fmt.Println(err)
					}
					buf.Reset()
				}
			case <-ticker.C:
				if buf.Len() > 0 {
					_, err := r.io.Write(buf.Bytes())
					if err != nil {
						fmt.Println(err)
					}
					buf.Reset()
				}
			}
		}
	}()
	return r
}

func (itself *AsyncW) Write(p []byte) (n int, err error) {
	data := make([]byte, len(p))
	copy(data, p)
	itself.dataChan <- data
	return len(p), nil
}

func (itself *AsyncW) Stop(ctx context.Context) {
	close(itself.dataChan)
	select {
	case <-ctx.Done():
		return
	case <-itself.closeFinish:
		return
	}
}
