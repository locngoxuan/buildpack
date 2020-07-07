package common

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const AlpineImage = "alpine:3.12.0"

func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

type CreateDirOption struct {
	WorkDir       string
	RelativePath  string
	AbsPath       string
	SkipContainer bool
	Perm          os.FileMode
}

type DeleteDirOption struct {
	WorkDir       string
	RelativePath  string
	AbsPath       string
	SkipContainer bool
}

func CreateDirInContainer(workDir, folderName string) error {
	dockerHost, err := CheckDockerHostConnection()
	if err != nil {
		return errors.New(fmt.Sprintf("can not connect to docker host: %s", err.Error()))
	}
	dockerCommandArg := make([]string, 0)
	dockerCommandArg = append(dockerCommandArg, "-H", dockerHost)
	dockerCommandArg = append(dockerCommandArg, "run", "--rm")

	image := AlpineImage
	dockerCommandArg = append(dockerCommandArg, "--workdir", "/working")
	dockerCommandArg = append(dockerCommandArg, "-v", fmt.Sprintf("%s:/working", workDir))
	dockerCommandArg = append(dockerCommandArg, image)
	dockerCommandArg = append(dockerCommandArg, "mkdir", "-p", folderName)
	//PrintInfo("working dir %s", workDir)
	//PrintInfo("docker %s", strings.Join(dockerCommandArg, " "))
	dockerCmd := exec.Command("docker", dockerCommandArg...)
	dockerCmd.Stdout = ioutil.Discard
	dockerCmd.Stderr = ioutil.Discard
	return dockerCmd.Run()
}

func DeleteDirOnContainer(workDir, folderName string) error {
	dockerHost, err := CheckDockerHostConnection()
	if err != nil {
		return errors.New(fmt.Sprintf("can not connect to docker host: %s", err.Error()))
	}
	dockerCommandArg := make([]string, 0)
	dockerCommandArg = append(dockerCommandArg, "-H", dockerHost)
	dockerCommandArg = append(dockerCommandArg, "run", "--rm")

	image := AlpineImage
	dockerCommandArg = append(dockerCommandArg, "--workdir", "/working")
	dockerCommandArg = append(dockerCommandArg, "-v", fmt.Sprintf("%s:/working", workDir))
	dockerCommandArg = append(dockerCommandArg, image)
	dockerCommandArg = append(dockerCommandArg, "rm", "-rf", folderName)
	//PrintInfo("working dir %s", workDir)
	//PrintInfo("docker %s", strings.Join(dockerCommandArg, " "))
	dockerCmd := exec.Command("docker", dockerCommandArg...)
	dockerCmd.Stdout = ioutil.Discard
	dockerCmd.Stderr = ioutil.Discard
	return dockerCmd.Run()
}

func CreateDir(opt CreateDirOption) error {
	if opt.SkipContainer {
		return os.MkdirAll(opt.AbsPath, opt.Perm)
	}
	return CreateDirInContainer(opt.WorkDir, opt.RelativePath)
}

func DeleteDir(option DeleteDirOption) error {
	if option.SkipContainer {
		return os.RemoveAll(option.AbsPath)
	}
	return DeleteDirOnContainer(option.WorkDir, option.RelativePath)
}

func SumContentMD5(file string) (string, error) {
	hasher := md5.New()
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
