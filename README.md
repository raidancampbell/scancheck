# Scancheck

scancheck is a linter to ensure the `bufio.Scanner` is used correctly.  Currently, it only checks to ensure that callers to a scanner's `Err()` don't occur within the scanner's `for scanner.Scan()` loop.

## Status
This linter errs on the side of false-negatives, and currently doesn't handle the following scenarios
 - a `bufio.Scanner` being returned from a non-`new` function
 - aliased `bufio` package
 - nested loops of different `bufio.Scanner` instances

## Usage

```shell
go install .
# OR
go build .

# navigate to your desired go project to lint
scancheck ./...
```