package util

import (
	"fmt"
	"os"
)

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	return read, write, nil
}

func ReadPipe(reader *os.File) (int, string, error) {
	buffer := make([]byte, 1024)

	message := ""
	num := 0
	for {
		n, err := reader.Read(buffer)
		num += n
		if err != nil {
			return num, message, err
		}

		fmt.Printf("buffer:%s, num:%d\n", string(buffer[0:n]), num)
		message = message + string(buffer[0:n])
		if n < 1024 {
			break
		}
	}

	return num, message, nil
}

type MyContainerError struct {
	Message string
}

func (err *MyContainerError) Error() string  {
	return err.Message
}