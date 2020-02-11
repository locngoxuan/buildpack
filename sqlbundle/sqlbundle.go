package sqlbundle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/jhoonb/archivex"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/docker"
	"sort"
	"strings"
)

const (
	bundleConfig       = "sqlbundle.yml"
	targetDirName      = "target"
	generatedDirName   = "generated-sql"
	sqlExt             = ".sql"
	dockerFileName     = "Dockerfile"
	checkPointFileName = "CHECKPOINT"
	dockerTemplate     = `FROM %s:%s

MAINTAINER sqlbundle <sqlbundle@fortna.com>

COPY generated-sql/*.sql /sql/%s/
COPY generated-sql/CHECKPOINT /sql/%s/
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
	WorkingDir  string
	BundleFile  string
	Clean       bool
	Dockerize   bool
	DockerHosts []string
	Version     string
}

var termFd uintptr
var width = 200
var output io.Writer

const endLineN = "\n"
const endLineR = "\r"

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
	target := filepath.Join(b.WorkingDir, targetDirName)
	return os.RemoveAll(target)
}

func (b *SQLBundle) Run(writer io.Writer) error {
	termFd, _ = term.GetFdInfo(os.Stdout)
	output = writer
	if len(strings.TrimSpace(b.BundleFile)) == 0 {
		b.BundleFile = filepath.Join(b.WorkingDir, bundleConfig)
	}

	config, err := ReadBundle(b.BundleFile)
	if err != nil {
		return err
	}

	finalVersion := config.Build.Version
	if len(b.Version) > 0 {
		finalVersion = b.Version
	}

	target := filepath.Join(b.WorkingDir, targetDirName)
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
		checkPointPath := filepath.Join(target, generatedDirName, checkPointFileName)
		checkPointContent := fmt.Sprintf(checkPointTemplate, cp.Start, cp.End)
		err = ioutil.WriteFile(checkPointPath, []byte(checkPointContent), 0644)
		if err != nil {
			return err
		}
	}
	dockerFilePath := filepath.Join(target, dockerFileName)
	dockerContent := fmt.Sprintf(dockerTemplate, config.Base.Image, config.Base.Version, config.Build.Classifier, config.Build.Classifier)
	err = ioutil.WriteFile(dockerFilePath, []byte(dockerContent), 0644)
	if err != nil {
		return err
	}

	bundleInTarget := filepath.Join(target, bundleConfig)
	err = CopyFile(b.BundleFile, bundleInTarget)
	if err != nil {
		return err
	}

	if b.Dockerize {
		if b.DockerHosts == nil || len(b.DockerHosts) == 0 {
			return errors.New("docker hosts is not configured")
		}
		// exec docker build
		client, err := docker.NewClient(b.DockerHosts)
		if err != nil {
			return err
		}

		parts := strings.Split(config.Build.Image, "/")
		finalName := fmt.Sprintf("%s-%s", strings.Join(parts, "-"), finalVersion)
		finalBuild := filepath.Join(target, fmt.Sprintf("%s.tar", finalName))
		tags := []string{fmt.Sprintf("%s:%s", config.Build.Image, finalVersion)}

		//create build context
		tar := new(archivex.TarFile)
		err = tar.Create(finalBuild)
		if err != nil {
			return err
		}
		err = tar.AddAll(filepath.Join(target, generatedDirName), true)
		if err != nil {
			return err
		}

		f, err := os.Open(filepath.Join(target, dockerFileName))
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()
		fileInfo, _ := f.Stat()
		err = tar.Add(dockerFileName, f, fileInfo)
		if err != nil {
			return err
		}
		err = tar.Close()
		if err != nil {
			return err
		}
		response, err := client.BuildImage(finalBuild, tags)
		if err != nil {
			return err
		}
		defer func() {
			_ = response.Body.Close()
		}()
		printHeader(fmt.Sprintf("Building docker image %s:%s", config.Build.Image, finalVersion), endLineN)
		return displayImageBuildLog(response.Body)
	}

	if b.Clean {
		err = os.RemoveAll(target)
		if err != nil {
			return err
		}
	}

	return nil
}

func displayImageBuildLog(in io.Reader) error {
	var dec = json.NewDecoder(in)
	for {
		var jm jsonmessage.JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if jm.Stream == "" {
			continue
		}

		printLineMessage(fmt.Sprintf("%s", jm.Stream), endLineR)
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
	generatedSql := filepath.Join(target, generatedDirName)
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
