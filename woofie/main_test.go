package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	*logDest = "foo.log"
	initlog()
	logger.Println("Hi there!")
	f, err := os.Open("foo.log")
	if err != nil { t.Error(err) }
	f.Close()
	os.Remove("foo.log")
}
