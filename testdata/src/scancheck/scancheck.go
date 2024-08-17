package scancheck

import (
	alias "bufio"
	"io"
)

func incorrectErrorScanner(reader io.Reader) {
	scanner := alias.NewScanner(reader)

	for scanner.Scan() {
		_ = scanner.Bytes()

		// this is incorrect behavior: if scanner.Scan() returns false, scanner.Err() should be checked.
		// meaning that scanner.Err() should only be checked outside the loop.
		if err := scanner.Err(); err != nil { // want "scanner.Err\\(\\) called inside a Scan\\(\\) loop"
			panic(err)
		}
	}
}

func bufioRawScanner(reader io.Reader) {
	var scanner = alias.Scanner{}

	for scanner.Scan() {
		_ = scanner.Bytes()
		if err := scanner.Err(); err != nil { // want "scanner.Err\\(\\) called inside a Scan\\(\\) loop"
			panic(err)
		}
	}
}

func bufioRawNewScanner(reader io.Reader) {
	scanner := new(alias.Scanner)

	for scanner.Scan() {
		_ = scanner.Bytes()
		if err := scanner.Err(); err != nil { // want "scanner.Err\\(\\) called inside a Scan\\(\\) loop"
			panic(err)
		}
	}
}

func multipleAssignment(reader io.Reader) {
	_, scanner := alias.NewReader(reader), alias.NewScanner(reader)

	for scanner.Scan() {
		_ = scanner.Bytes()
		if err := scanner.Err(); err != nil { // want "scanner.Err\\(\\) called inside a Scan\\(\\) loop"
			panic(err)
		}
	}
}

func unrelatedBufioScanner(reader io.Reader) {
	x := func(_ alias.Scanner) *notABufioScanner {
		return newNotBufioScanner()
	}
	scanner := x(alias.Scanner{})

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}

}

func correctErrorScanner(reader io.Reader) {
	scanner := alias.NewScanner(reader)

	for scanner.Scan() {
		_ = scanner.Bytes()

		// this is incorrect behavior: if scanner.Scan() returns false, scanner.Err() should be checked.
		// meaning that scanner.Err() should only be checked outside the loop.
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func hasNoScanner() {}

func scannerIsNotScanned(reader io.Reader) {
	scanner := alias.NewScanner(reader)
	_ = scanner.Bytes()
}

func scannerScannedOutsideForLoop(reader io.Reader) {
	scanner := alias.NewScanner(reader)
	_ = scanner.Scan()
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	_ = scanner.Bytes()
}

func bufioNotScanner(reader io.Reader) {
	r := alias.NewReader(reader)

	for _, err := r.ReadByte(); err != nil; {
	}
}

func scannerNotBufio(reader io.Reader) {
	sg := scannerGenerator{}
	scanner := sg.NewScanner()

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}
}

func scannerShadowingBufio(reader io.Reader) {
	bufio := scannerGenerator{}
	scanner := bufio.NewScanner()

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}
}

func scannerAlmostShadowingBufio(reader io.Reader) {
	bufio := scannerGenerator{}
	scanner := bufio.NewScannerWithDifferentName()

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}
}

func scanNotScanner(reader io.Reader) {
	b := new(boolScanner)
	for b.Scan() {
		if err := newNotBufioScanner().Err(); err != nil {
			panic(err)
		}
	}
}

func newNotBufioScanner() *notABufioScanner {
	return new(notABufioScanner)
}

type notABufioScanner struct{}

func (n notABufioScanner) Scan() bool {
	return true
}

func (n notABufioScanner) Err() error {
	return nil
}

type scannerGenerator struct{}

func (s scannerGenerator) NewScanner() *notABufioScanner {
	return newNotBufioScanner()
}

func (s scannerGenerator) NewScannerWithDifferentName() *notABufioScanner {
	return newNotBufioScanner()
}

type boolScanner struct{}

func (b boolScanner) Scan() bool {
	return false
}
