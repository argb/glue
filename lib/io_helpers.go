package lib

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func Create(filename string) *os.File {

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	return file
}

func WriteString(filename string, str string) (int, error) {
	f := Create(filename)
	defer f.Close()

	//n, err := fmt.Fprintf(f, str)
	n, err := io.WriteString(f, str)

	fmt.Printf("wrote %d bytes to %s\n", n, filename)

	return n, err
}

func WriteStr(filename string, str string) (int, error){
	f := Create(filename)
	defer f.Close()

	w := bufio.NewWriter(f)

	n, err := w.WriteString(str)
	w.Flush()

	fmt.Printf("wrote %d bytes to %s\n", n, filename)
	return n, err
}

