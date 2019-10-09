package util

import "os"

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

		if n == 0 {
			break
		}

		message = message + string(buffer[0:n])
	}

	return num, message, nil
}