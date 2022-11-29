package proto_format

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"
)

func TestFormat(t *testing.T) {

	isRoot := true

	_ = filepath.WalkDir("./pb", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if !isRoot {
				return filepath.SkipDir
			} else {
				isRoot = false
				return nil
			}
		}

		if filepath.Ext(d.Name()) != ".proto" {
			return nil
		}
		Format(fmt.Sprintf("./pb/%s", d.Name()))
		return nil
	})
}

func preprocessSchemaFolder(schemaDir string) []string {
	isRoot := true
	var names []string
	_ = filepath.WalkDir(schemaDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if !isRoot {
				return filepath.SkipDir
			} else {
				isRoot = false
				return nil
			}
		}

		if filepath.Ext(d.Name()) != ".proto" {
			return nil
		}

		names = append(names, d.Name())

		return nil
	})
	return names
}
