package main

import (
	"github.com/raidancampbell/scancheck/pkg/scancheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(scancheck.Analyzer)
}
