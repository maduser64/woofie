package main

import (
	"os"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	pre := time.Now()
	os.Args = []string{"foo", "--on", "--off", "--ignore"}
	main()
	if time.Now().Sub(pre) < (1900*time.Millisecond) {
		t.Error("Command didn't take long enough to run (syntax?)")
	}
}
