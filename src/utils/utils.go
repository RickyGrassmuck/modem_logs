package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
)

const chunkSize = 64000

func CalculateSpread(values []float64) float64 {
	sort.Float64s(values)
	return values[len(values)-1] - values[0]
}

func DeepCompare(file1, file2 string) (bool, error) {
	// Check file size ...

	f1, err := os.Open(file1)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false, err
	}
	defer f2.Close()

	for {
		b1 := make([]byte, chunkSize)
		_, err1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true, nil
			} else if err1 == io.EOF || err2 == io.EOF {
				return false, nil
			} else {
				return false, fmt.Errorf("Error1: %v\nError2: %v\n", err1, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			return false, nil
		}
	}
}
