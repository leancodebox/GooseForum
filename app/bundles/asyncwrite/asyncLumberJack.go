package asyncwrite

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type AsyncW struct {
	dataChan    chan []byte
	closeFinish chan bool
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
