package sqlbundle

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	bundleConfig       = "sqlbundle.yml"
	targetDir          = "target"
	generatedDir       = "generated-sql"
	sqlExt             = ".sql"
	dockerFileName     = "Dockerfile"
	checkPointFileName = "CHECKPOINT"
	dockerTemplate     = `FROM %s:%s

MAINTAINER sqlbundle

COPY generated-sql/*.sql /sql/product/
COPY generated-sql/CHECKPOINT /sql/product/
`
	checkPointTemplate = `start=%s
end=%s
`
)

type CurrentDir struct {
	Files []PathInfo
	Dirs  []PathInfo
	Path  string
}

type PathInfo struct {
	Info os.FileInfo
	Path string
}

type CheckPoint struct {
	Start string
	End   string
}

type BundleConfig struct {
	Base      *BundleBaseConfig `yaml:"base"`
	Build     *BundleBaseConfig `yaml:"build"`
	Revisions []string          `yaml:"revisions,omitempty"`
	Patches   []string          `yaml:"patches,omitempty"`
}

type BundleBaseConfig struct {
	Image      string `yaml:"image,omitempty"`
	Version    string `yaml:"version,omitempty"`
	Classifier string `yaml:"classifier,omitempty"`
}

type SQLBundle struct {
	WorkingDir string
	BundleFile string
	Clean      bool
}

func (b *SQLBundle) Run() error {
	if len(strings.TrimSpace(b.BundleFile)) == 0 {
		b.BundleFile = filepath.Join(b.WorkingDir, bundleConfig)
	}

	config, err := ReadBundle(b.BundleFile)
	if err != nil {
		return err
	}

	_ = os.RemoveAll(targetDir)

	target := filepath.Join(b.WorkingDir, targetDir)
	err = os.MkdirAll(target, 0766)
	if err != nil {
		return err
	}

	sequence := new(int)
	*sequence = 1
	cp := &CheckPoint{
		Start: "",
		End:   "",
	}
	for _, revision := range config.Revisions {
		fmt.Println("[SQLBUNDLE] process version", revision)
		path := filepath.Join(b.WorkingDir, revision)
		cp.Start = cp.End
		err := copyEachVersion(path, target, sequence, cp)
		if err != nil {
			return nil
		}
		fmt.Println(fmt.Sprintf("[SQLBUNDLE] update checkpoint to start = %s and end = %s", cp.Start, cp.End))
		checkPointPath := filepath.Join(target, generatedDir, checkPointFileName)
		checkPointContent := fmt.Sprintf(checkPointTemplate, cp.Start, cp.End)
		err = ioutil.WriteFile(checkPointPath, []byte(checkPointContent), 0644)
		if err != nil {
			return err
		}
	}
	dockerFilePath := filepath.Join(target, dockerFileName)
	dockerContent := fmt.Sprintf(dockerTemplate, config.Base.Image, config.Base.Version)
	err = ioutil.WriteFile(dockerFilePath, []byte(dockerContent), 0644)
	if err != nil {
		return err
	}
	return nil
}

func readCurrentDir(dirPath string) (CurrentDir, error) {
	currentDir := CurrentDir{
		Files: make([]PathInfo, 0),
		Dirs:  make([]PathInfo, 0),
		Path:  dirPath,
	}
	file, err := os.Open(dirPath)
	if err != nil {
		return currentDir, err
	}
	defer func() {
		_ = file.Close()
	}()

	list, _ := file.Readdirnames(0) // 0 to read all files and folders
	for _, name := range list {
		filePath := filepath.Join(dirPath, name)
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return currentDir, err
		}
		pathInfo := PathInfo{
			Path: filePath,
			Info: fileInfo,
		}
		if pathInfo.Info.IsDir() {
			currentDir.Dirs = append(currentDir.Dirs, pathInfo)
		} else {
			currentDir.Files = append(currentDir.Files, pathInfo)
		}
	}
	return currentDir, nil
}

func copyEachVersion(dir, target string, sequence *int, cp *CheckPoint) (error) {
	bundleFile := filepath.Join(dir, bundleConfig)
	generatedSql := filepath.Join(target, generatedDir)
	err := os.MkdirAll(generatedSql, 0766)
	if err != nil {
		return err
	}
	_, err = os.Stat(bundleFile)
	hasPatches := true
	if err != nil {
		if os.IsNotExist(err) {
			hasPatches = false
		} else {
			return err
		}
	}

	revisionDir, err := readCurrentDir(dir)
	if err != nil {
		return err
	}

	err = compileSqlFile(generatedSql, revisionDir, cp, sequence)
	if err != nil {
		return err
	}

	if hasPatches {
		config, err := ReadBundle(bundleFile)
		if err != nil {
			return err
		}

		if len(config.Patches) <= 0 {
			return nil
		}

		for _, patchNumber := range config.Patches {
			fmt.Println("[SQLBUNDLE] process patch", patchNumber)
			patchDir := filepath.Join(dir, patchNumber)
			currentPathDir, err := readCurrentDir(patchDir)
			if err != nil {
				return err
			}

			err = compileSqlFile(generatedSql, currentPathDir, cp, sequence)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func replaceTimestampBySequence(fileName string, sequence int) (string, string) {
	partOfName := strings.Split(fileName, "_")
	partOfName[0] = fmt.Sprintf("%05d", sequence)
	return partOfName[0], strings.Join(partOfName, "_")
}

func compileSqlFile(generatedDir string, currentDir CurrentDir, cp *CheckPoint, sequence *int) error {
	if len(currentDir.Files) > 0 {
		sort.Slice(currentDir.Files, func(i, j int) bool {
			return currentDir.Files[i].Info.ModTime().Unix() < currentDir.Files[j].Info.ModTime().Unix()
		})
		cp.Start = cp.End
		//copy files
		for _, file := range currentDir.Files {
			_, fileName := filepath.Split(file.Path)
			if filepath.Ext(fileName) != sqlExt {
				continue
			}
			sequenceStr, fullName := replaceTimestampBySequence(fileName, *sequence)
			*sequence++
			cp.End = sequenceStr
			err := CopyFile(file.Path, filepath.Join(generatedDir, fullName))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ReadBundle(file string) (BundleConfig, error) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		err = errors.New("configuration file not found")
		return BundleConfig{}, err
	}

	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		err = errors.New(fmt.Sprintf("read application config file get error %v", err))
		return BundleConfig{}, err
	}

	var config BundleConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return BundleConfig{}, err
	}
	return config, nil
}

func CopyFile(src, dst string) error {
	_, fileName := filepath.Split(src)
	fmt.Println(fmt.Sprintf("[SQLBUNDLE] Copying %s to %s", fileName, dst))
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = source.Close()
	}()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = destination.Close()
	}()
	_, err = io.Copy(destination, source)
	return err
}
