package testdata

import (
	"bufio"
	"io"
)

func correctErrorScanner(reader io.Reader) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		_ = scanner.Bytes()

		// this is incorrect behavior: if scanner.Scan() returns false, scanner.Err() should be checked.
		// meaning that scanner.Err() should only be checked outside the loop.
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

}
