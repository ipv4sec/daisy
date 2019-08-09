package main

import (
	"flag"
)

var (
	sourceURL string
	targetURL string
)
func init() {
	flag.StringVar(&sourceURL, "s", "redis://localhost:6379/0", "Source")
	flag.StringVar(&targetURL, "t", "redis://user:passwd@localhost:6379/0", "Target")
}
func main() {
	flag.Parse()
}
