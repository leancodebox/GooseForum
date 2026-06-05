// Package lineopt provides helpers for processing text files line by line.
package lineopt

import (
	"bufio"
	"errors"
	"os"
)

// ReadLine opens filePath and calls action for each scanned line.
func ReadLine(filePath string, action func(item string)) error {
	f, errF := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if errF != nil {
		return errF
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		action(line)
	}
	return errors.Join(scanner.Err(), f.Close())
}
