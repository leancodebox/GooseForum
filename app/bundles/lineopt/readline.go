package lineopt

import (
	"bufio"
	"os"
)

func ReadLine(filePath string, action func(item string)) error {
	f, errF := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if errF != nil {
		return errF
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		action(line)
	}
	return nil
}
