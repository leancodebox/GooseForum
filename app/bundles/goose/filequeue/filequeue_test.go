package filequeue

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/goose/fileopt"
	"github.com/leancodebox/GooseForum/app/bundles/goose/jsonopt"
	"io"
	"sync"
	"testing"
)

func TestCheckQueueData(t *testing.T) {
	// 64 + 128+ 128 = 256 + 64 = 320
	// 280 - 128 -128 =  24
	data, err := fileopt.FileGetContents("./storage/queue/1_000_000_000.q")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(len(data))
	fmt.Println(data[0:63])
	for i := 0; i < (len(data)-64)/128+1; i++ {
		start := 64 + i*128
		end := start + 128
		fmt.Println(data[start:end])
	}
}

type TestUnitData struct {
	Valid bool   `json:"valid"`
	Id    int64  `json:"id"`
	Data  string `json:"data"`
}

type Queue interface {
	Push(data string) error
	Pop() (string, error)
	Len() int64
	Vacuum() error
}

func TestQueue(t *testing.T) {
	maxTest := 1_000_000
	stopNum := maxTest / 10
	strData := "字字字字字"
	var q Queue
	q, err := NewDefaultFileQueue("./storage/queue")
	if err != nil {
		t.Error(err)
	}
	err = q.Vacuum()
	if err != nil {
		t.Error(err)
	}
	t.Log("插入足够的内容到队列中")
	for i := 1; i <= maxTest; i++ {
		err = q.Push(jsonopt.Encode(TestUnitData{true, int64(i), strData}))
		if err != nil {
			t.Error(err)
		}
	}
	t.Log("顺序取出10条数据")

	for i := 1; i <= 10; i++ {
		_, popErr := q.Pop()
		if popErr != nil {
			t.Error(popErr)
			break
		}
	}
	t.Log("Vacuum")
	err = q.Vacuum()
	if err != nil {
		t.Error(err)
	}
	t.Log("开始读取队列所有内容")
	wg := sync.WaitGroup{}
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n := 0
			for {
				data, popErr := q.Pop()
				if popErr != nil {
					if popErr == io.EOF {
						break
					} else {
						t.Error(popErr)
						break
					}
				}
				n += 1
				if n%stopNum == 0 {
					t.Log(`n%`+toString(stopNum), data)
				}
				res := jsonopt.Decode[TestUnitData](data)
				if res.Data != strData {
					t.Errorf("数据波动 %v", res.Data)
				}
			}
		}()
	}
	wg.Wait()
	err = q.Vacuum()
	if err != nil {
		t.Error(err)
	}
}

func TestQueue4bigData(t *testing.T) {
	maxTest := 1_000_000
	stopNum := maxTest / 10
	var q Queue
	q, err := NewFileQueue("./storage/bigBlockQueue", 1024)
	var longStr string
	for i := 1; i <= 100; i++ {
		longStr += "字"
	}
	if err != nil {
		t.Error(err)
	}
	err = q.Vacuum()
	if err != nil {
		t.Error(err)
	}
	for i := 1; i <= maxTest; i++ {
		err = q.Push(jsonopt.Encode(TestUnitData{true, int64(i), longStr}))
		if err != nil {
			t.Error(err)
		}
	}
	n := 0
	for {
		data, popErr := q.Pop()
		if popErr != nil {
			if popErr == io.EOF {
				break
			} else {
				t.Error(popErr)
				break
			}
		}
		n += 1
		if n%stopNum == 0 {
			t.Log(`n%`+fmt.Sprintf("%v", stopNum), data)
		}
		res := jsonopt.Decode[TestUnitData](data)
		if res.Data != longStr {
			t.Errorf("数据波动 %v", res.Data)
		}
	}
	err = q.Vacuum()
	if err != nil {
		t.Error(err)
	}
}

func toString(p any) string {
	return fmt.Sprintf("%v", p)
}
