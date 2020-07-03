package generator

import (
	"fmt"
	"go/types"
	"golang.org/x/tools/go/packages"
)

type Target struct {
	Pkg *packages.Package
	Interface types.Interface
}


func FromPath(path string) error {
	// load all the target interface values, and their output
	targets, err := identifyTargets()
	if err != nil {
		return err
	}

	fmt.Println(targets)
	return nil
}

func identifyTargets() ([]Target, error) {
	return nil, nil
}
