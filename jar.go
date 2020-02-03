package buildpack

import (
	"archive/zip"
	"errors"
	"io/ioutil"
	"path/filepath"
)

const pomFile = "pom.xml"

func ReadPomFromJar(jarFile string) ([]byte, error) {
	r, err := zip.OpenReader(jarFile)

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = r.Close()
	}()

	for _, f := range r.File {
		_, fileName := filepath.Split(f.Name)
		if fileName == pomFile {
			rc, err := f.Open()

			if err != nil {
				return nil, err
			}

			bytes, err := ioutil.ReadAll(rc)
			if err != nil {
				return nil, err
			}

			if len(bytes) == 0 {
				return nil, errors.New("")
			}
			return bytes, nil
		}
	}

	return nil, errors.New("")
}
