package generator

import (
	"io/ioutil"
	"path"

	"golang.org/x/tools/go/packages"
)

func Run(sourcePackagePath string, sourceFileBasename string, target string) error {
	ps, err := PackagesFromPath(sourcePackagePath)

	parsed, err := Process(ps, target)
	if err != nil {
		return err
	}

	buf, err := Generate(sourcePackagePath, parsed)
	if err != nil {
		return err
	}

	targetPath := path.Join(sourcePackagePath, toTargetName(sourceFileBasename))
	err = ioutil.WriteFile(targetPath, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func PackagesFromPath(wd string) ([]*packages.Package, error) {
	return packages.Load(&packages.Config{
		Mode: packages.NeedTypes |
			packages.NeedDeps |
			packages.NeedImports,
	}, wd)
}

