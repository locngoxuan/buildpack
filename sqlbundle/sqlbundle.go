package sqlbundle

import (
	"errors"
	"fmt"
	"github.com/docker/docker/pkg/term"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	sqlBundleFile      = "Sqlbundlefile"
	TargetDirName      = "target"
	GeneratedDirName   = "generated-sql"
	ExtSql             = ".sql"
	CheckPointFileName = "CHECKPOINT"

	checkPointTemplate = `start=%s
end=%s`
)

type CurrentDir struct {
	Files []PathInfo
	Dirs  []PathInfo
	Path  string
}

type PathInfo struct {
	Info os.FileInfo
	Path string
	Name string
}

func FileConfig() string {
	return sqlBundleFile
}

func (p *PathInfo) prefixTimeStamp() int64 {
	parts := strings.Split(p.Name, "_")
	if len(parts) == 0 {
		return p.Info.ModTime().Unix()
	}
	v, err := strconv.Atoi(parts[0])
	if err != nil {
		return p.Info.ModTime().Unix()
	}
	return int64(v)
}

type CheckPoint struct {
	Start string
	End   string
}

type BundleConfig struct {
	Target    string   `yaml:"target,omitempty"`
	Revisions []string `yaml:"revisions,omitempty"`
	Patches   []string `yaml:"patches,omitempty"`
}

type SQLBundle struct {
	WorkingDir string
	BundleFile string
	Clean      bool
	Version    string
}

var termFd uintptr
var width = 200
var output io.Writer
var targets map[string]int

var sqlNamePrefix = 0

const endLineN = "\n"

func init() {
	targets = make(map[string]int)
	targets["product"] = 1
	targets["project"] = 2
}

func targetPrefix(classifier string) int {
	v, ok := targets[classifier]
	if !ok {
		return 0
	}
	return v
}

func printHeader(msg, end string) {
	ws, err := term.GetWinsize(termFd)
	if err == nil {
		width = int(ws.Width) / 2
	}

	if output != nil {
		fmt.Println("[INFO]")
		_, _ = fmt.Fprintf(output, fmt.Sprintf("[INFO] %s%s", strings.Repeat("-", width), end))
		_, _ = fmt.Fprintf(output, fmt.Sprintf("[INFO] %s%s", msg, end))
		_, _ = fmt.Fprintf(output, fmt.Sprintf("[INFO] %s%s", strings.Repeat("-", width), end))
		fmt.Println("[INFO]")
	}
}

func printLineMessage(msg, end string) {
	if output != nil {
		_, _ = fmt.Fprintf(output, fmt.Sprintf("[INFO] %s%s", msg, end))
	}
}

func (b *SQLBundle) RunClean() error {
	target := filepath.Join(b.WorkingDir, TargetDirName)
	return os.RemoveAll(target)
}

func (b *SQLBundle) Run(writer io.Writer) error {
	termFd, _ = term.GetFdInfo(os.Stdout)
	output = writer
	if len(strings.TrimSpace(b.BundleFile)) == 0 {
		b.BundleFile = filepath.Join(b.WorkingDir, FileConfig())
	}

	config, err := ReadBundle(b.BundleFile)
	if err != nil {
		return err
	}

	_, ok := targets[config.Target]
	if !ok {
		return errors.New("target " + config.Target + " is not supported")
	}

	sqlNamePrefix = targetPrefix(config.Target)
	finalVersion := ""
	if len(config.Revisions) > 0 {
		finalVersion = config.Revisions[len(config.Revisions)-1]
	}
	if len(b.Version) > 0 {
		finalVersion = b.Version
	}

	if finalVersion == "" {
		return errors.New("version is not specified")
	}

	target := filepath.Join(b.WorkingDir, TargetDirName)
	_ = os.RemoveAll(target)
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
		printHeader(fmt.Sprintf("Process version %s", revision), endLineN)
		path := filepath.Join(b.WorkingDir, revision)
		err := copyEachVersion(path, target, sequence, cp)
		if err != nil {
			return nil
		}
		printLineMessage(fmt.Sprintf("Update checkpoint : start = %s, end = %s", cp.Start, cp.End), endLineN)
		checkPointPath := filepath.Join(target, GeneratedDirName, CheckPointFileName)
		checkPointContent := fmt.Sprintf(checkPointTemplate, cp.Start, cp.End)
		err = ioutil.WriteFile(checkPointPath, []byte(checkPointContent), 0644)
		if err != nil {
			return err
		}
	}

	bundleInTarget := filepath.Join(target, FileConfig())
	err = CopyFile(b.BundleFile, bundleInTarget)
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
			Name: name,
		}
		if pathInfo.Info.IsDir() {
			currentDir.Dirs = append(currentDir.Dirs, pathInfo)
		} else {
			if filepath.Ext(filePath) == ExtSql {
				currentDir.Files = append(currentDir.Files, pathInfo)
			}
		}
	}
	return currentDir, nil
}

func copyEachVersion(dir, target string, sequence *int, cp *CheckPoint) (error) {
	bundleFile := filepath.Join(dir, FileConfig())
	generatedSql := filepath.Join(target, GeneratedDirName)
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
			printHeader(fmt.Sprintf("Process patch %s", patchNumber), endLineN)
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
	partOfName[0] = fmt.Sprintf("%d%05d", sqlNamePrefix, sequence)
	return partOfName[0], strings.Join(partOfName, "_")
}

func compileSqlFile(generatedDir string, currentDir CurrentDir, cp *CheckPoint, sequence *int) error {
	if len(currentDir.Files) > 0 {
		sort.Slice(currentDir.Files, func(i, j int) bool {
			return currentDir.Files[i].prefixTimeStamp() < currentDir.Files[j].prefixTimeStamp()
		})
		cp.Start = cp.End
		//copy files
		for _, file := range currentDir.Files {
			_, fileName := filepath.Split(file.Path)
			if filepath.Ext(fileName) != ExtSql {
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
		err = errors.New(file + " is not found")
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
	printLineMessage(fmt.Sprintf("Copying %s to %s", fileName, dst), endLineN)
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
