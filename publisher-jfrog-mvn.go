package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	publishDir = "publishes"
)

type PublisherJfrogMVN struct {
	uploads []JfrogMVNUpload
}

type JfrogMVNUpload struct {
	Destination string
	Source      string
	Version     string
	Username    string
	Password    string
	CheckSum
}

type CheckSum struct {
	SHA256 string
	MD5    string
}

type POM struct {
	XMLName    xml.Name `xml:"project"`
	GroupId    string   `xml:"groupId"`
	ArtifactId string   `xml:"artifactId"`
	Classifier string   `xml:"packaging"`
}

func (p *PublisherJfrogMVN) WriteConfig(bp BuildPack, opt BuildPackModuleConfig) error {
	return nil
}

func (p *PublisherJfrogMVN) CreateContext(bp BuildPack, rtOpt BuildPackModuleRuntimeParams) (PublishContext, error) {
	ctx := newPublishContext(rtOpt.Name, rtOpt.Path)
	ctx.BuildPack = bp
	ctx.BuildPackModuleRuntimeParams = rtOpt

	pwd, err := filepath.Abs(bp.getModuleWorkingDir(rtOpt.Path))
	if err != nil {
		return ctx, err
	}
	configFile := filepath.Join(pwd, pomFile)
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err = errors.New("configuration file not found")
		return ctx, err
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		err = errors.New(fmt.Sprintf("read application config file get error %v", err))
		return ctx, err
	}

	var pomProject POM
	err = xml.Unmarshal(yamlFile, &pomProject)
	if err != nil {
		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
		return ctx, err
	}
	addPOMToContext(ctx, pomProject)
	return ctx, nil
}

func getPOMFromContext(ctx PublishContext) POM {
	v, err := ctx.Get("pom")
	if err != nil {
		buildError(*ctx.BuildPack.Error("", err))
	}
	return v.(POM)
}

func addPOMToContext(ctx PublishContext, pom POM) {
	ctx.Add("pom", pom)
}

func (p *PublisherJfrogMVN) Verify(ctx PublishContext) error {
	return nil
}

func (p *PublisherJfrogMVN) Pre(ctx PublishContext) error {
	rtModule := ctx.BuildPackModuleRuntimeParams
	pom := getPOMFromContext(ctx)
	version := ctx.RuntimeParams.VersionRuntimeParams.version(labelSnapshot, 0)

	uploadParams := make([]JfrogMVNUpload, 0)
	//generic information for upload
	artifact := ctx.RuntimeParams.URL
	repository := ctx.RuntimeParams.Repository.Release
	if !ctx.RuntimeParams.Release {
		repository = ctx.RuntimeParams.Repository.Snapshot
	}

	args := strings.Split(pom.GroupId, ".")
	args = append(args, pom.ArtifactId)
	args = append(args, version)
	modulePath := strings.Join(args, "/")
	// end generic information

	//copy flattened pom
	pomSrc := ctx.buildPathOnRoot(rtModule.Path, pomFlattened)
	pomName := fmt.Sprintf("%s-%s.pom", pom.ArtifactId, version)
	pomPublished := ctx.buildPathOnRoot(publishDir, pomName)
	err := copyFile(pomSrc, pomPublished)
	if err != nil {
		return err
	}

	pomSumSHA256, err := checksumSHA256(pomPublished)
	if err != nil {
		return err
	}
	pomSumMD5, err := checksumMD5(pomPublished)
	if err != nil {
		return err
	}
	pomParam := JfrogMVNUpload{
		Destination: fmt.Sprintf("%s/%s/%s/%s", artifact, repository, modulePath, pomName),
		Source:      pomPublished,
		Version:     version,
		Username:    ctx.RuntimeParams.Username,
		Password:    ctx.RuntimeParams.Password,
		CheckSum: CheckSum{
			SHA256: pomSumSHA256,
			MD5:    pomSumMD5,
		},
	}
	uploadParams = append(uploadParams, pomParam)
	buildInfo(ctx.BuildPack, fmt.Sprintf("PUT %s to %s with sum sha256:%s and md5:%s",
		pomParam.Source,
		pomParam.Destination,
		pomParam.SHA256,
		pomParam.MD5))

	if pom.Classifier == "jar" || len(strings.TrimSpace(pom.Classifier)) == 0 {
		//copy jar
		jarName := fmt.Sprintf("%s-%s.jar", pom.ArtifactId, version)
		jarSrc := ctx.buildPathOnRoot(rtModule.Path, "target", jarName)
		jarPublished := ctx.buildPathOnRoot(publishDir, jarName)
		err = copyFile(jarSrc, jarPublished)
		if err != nil {
			return err
		}
		jarSumSHA256, err := checksumSHA256(jarPublished)
		if err != nil {
			return err
		}
		jarSumMD5, err := checksumMD5(jarPublished)
		if err != nil {
			return err
		}
		jarParam := JfrogMVNUpload{
			Destination: fmt.Sprintf("%s/%s/%s/%s", artifact, repository, modulePath, jarName),
			Source:      jarPublished,
			Version:     version,
			Username:    ctx.RuntimeParams.Username,
			Password:    ctx.RuntimeParams.Password,
			CheckSum: CheckSum{
				SHA256: jarSumSHA256,
				MD5:    jarSumMD5,
			},
		}
		uploadParams = append(uploadParams, jarParam)
		buildInfo(ctx.BuildPack, fmt.Sprintf("PUT %s to %s with sum sha256:%s and md5:%s",
			jarParam.Source,
			jarParam.Destination,
			jarParam.SHA256,
			jarParam.MD5))
	}
	p.uploads = uploadParams
	return nil
}

func (p *PublisherJfrogMVN) Publish(ctx PublishContext) error {
	for _, upload := range p.uploads {
		err := uploadFile(upload)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PublisherJfrogMVN) Clean(ctx PublishContext) error {
	return nil
}

func uploadFile(param JfrogMVNUpload) error {
	data, err := os.Open(param.Source)
	if err != nil {
		return err
	}
	defer func() {
		_ = data.Close()
	}()
	req, err := http.NewRequest("PUT", param.Destination, data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-CheckSum-MD5", param.MD5)
	req.SetBasicAuth(param.Username, param.Password)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusCreated {
		return errors.New(res.Status)
	}
	return nil
}

func checksumSHA256(file string) (string, error) {
	hasher := sha256.New()
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func checksumMD5(file string) (string, error) {
	hasher := md5.New()
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
