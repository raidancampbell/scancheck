# Scancheck

scancheck is a linter to ensure the `bufio.Scanner` is used correctly.  Currently, it only checks to ensure that callers to a scanner's `Err()` don't occur within the scanner's `for scanner.Scan()` loop.

## Usage

```shell
go install .
# OR
go build .

# navigate to your desired go project to lint
scancheck ./...
```