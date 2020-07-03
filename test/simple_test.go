package test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"github.plaid.com/plaid/typedmiddleware/generator"
)

func TestCanCompileSimpleIntoValidCode(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	require.NoError(t, generator.Run(
		fmt.Sprintf("%s/src/github.plaid.com/plaid/typedmiddleware/fixtures/simple", gopath),
		"simple.go",
		"SimpleMiddleware",
	))
}

func TestCanCompileSimpleIntoValidCodeFunctional(t *testing.T) {
	cmd := exec.Command("/usr/local/bin/go", "generate", "../fixtures/simple")
	mustRunCmd(t, cmd, "could not generate")

	testCmd := exec.Command("/usr/local/bin/go", "test", "-count=1", "../fixtures/simple")
	mustRunCmd(t, testCmd, "tests failed")
}

func mustRunCmd(t *testing.T, cmd *exec.Cmd, msg string)  {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("%s:\nSTDOUT:\n%s\n\nSTDERR:\n%s\n\n", msg, stdout.String(), stderr.String())
	}
}
