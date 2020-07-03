package test

import (
	"bytes"
	"os/exec"
	"testing"
)

func TestCanCompileSimpleIntoValidCode(t *testing.T) {

	cmd := exec.Command("/usr/local/bin/go", "generate", "./fixtures/simple/...")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("could not generate:\nSTDOUT:\n%s\n\nSTDERR:\n%s\n\n", stdout.String(), stderr.String())
	}

	//cmd := exec.Command("/usr/local/bin/go", "test", "./fixtures/simple")
	//
	//var stdout bytes.Buffer
	//var stderr bytes.Buffer
	//cmd.Stdout = &stdout
	//cmd.Stderr = &stderr
	//
	//require.NoError(t, cmd.Run())
}
