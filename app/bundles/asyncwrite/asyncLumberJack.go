package asyncwrite

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"
)

type AsyncW struct {
	dataChan    chan []byte
	closeFinish chan bool
}

func AsyncLumberjack(io *lumberjack.Logger) *AsyncW {
	r := &AsyncW{
		dataChan:    make(chan []byte, 256),
		closeFinish: make(chan bool),
	}
	maxLen := 1024 * 8
	w := io
	go func() {
		defer func() {
			if err := w.Close(); err != nil {
				fmt.Println(err)
			}
			r.closeFinish <- true
			return
		}()
		ticker := time.NewTicker(time.Millisecond * 100)
		defer ticker.Stop()
		var buf bytes.Buffer
		for {
			select {
			case val, ok := <-r.dataChan:
				if !ok {
					if buf.Len() > 0 {
						if _, err := w.Write(buf.Bytes()); err != nil {
							fmt.Println(err)
						}
						buf.Reset()
					}
					return
				}

				buf.Write(val)
				if buf.Len() >= maxLen {
					if _, err := w.Write(buf.Bytes()); err != nil {
						fmt.Println(err)
					}
					buf.Reset()
				}
			case <-ticker.C:
				if buf.Len() > 0 {
					if _, err := w.Write(buf.Bytes()); err != nil {
						fmt.Println(err)
					}
					buf.Reset()
				}
			}
		}
	}()
	return r
}

func AsyncLumberjackBufIo(io *lumberjack.Logger) *AsyncW {
	r := &AsyncW{
		dataChan:    make(chan []byte, 256),
		closeFinish: make(chan bool),
	}
	w := bufio.NewWriterSize(io, 1024*16)
	go func() {
		defer func() {
			if err := w.Flush(); err != nil {
				fmt.Println(err)
			}
			r.closeFinish <- true
		}()
		ticker := time.NewTicker(time.Millisecond * 100)
		defer ticker.Stop()
		for {
			select {
			case val, ok := <-r.dataChan:
				if !ok {
					return
				}
				_, err := w.Write(val)
				if err != nil {
					fmt.Println(err)
				}
			case <-ticker.C:
				err := w.Flush()
				if err != nil {
					fmt.Println(err)
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
