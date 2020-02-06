package publisher

//
//import (
//	"crypto/md5"
//	"crypto/sha256"
//	"encoding/xml"
//	"errors"
//	"fmt"
//	"io"
//	"io/ioutil"
//	"net/http"
//	"os"
//	"path/filepath"
//	. "scm.wcs.fortna.com/lngo/buildpack"
//	"strings"
//)
//
//const (
//	pomFile       = "pom.xml"
//	labelSnapshot = "SNAPSHOT"
//)
//
//type ArtifactoryMvn struct {
//	uploads []MVNUploadParam
//}
//
//type MVNUploadParam struct {
//	Destination string
//	Source      string
//	Version     string
//	Username    string
//	Password    string
//	CheckSum
//}
//
//type CheckSum struct {
//	SHA256 string
//	MD5    string
//}
//
//type POM struct {
//	XMLName    xml.Name  `xml:"project"`
//	Parent     ParentPOM `xml:"parent"`
//	GroupId    string    `xml:"groupId"`
//	ArtifactId string    `xml:"artifactId"`
//	Classifier string    `xml:"packaging"`
//}
//
//type ParentPOM struct {
//	GroupId    string `xml:"groupId"`
//	ArtifactId string `xml:"artifactId"`
//}
//
//func (p *ArtifactoryMvn) WriteConfig(bp BuildPack, opt ModuleConfig) error {
//	return nil
//}
//
//func (p *ArtifactoryMvn) CreateContext(bp *BuildPack, rtOpt ModuleRuntime) (PublishContext, error) {
//	ctx := NewPublishContext(rtOpt.Name, rtOpt.Path)
//	ctx.BuildPack = bp
//	ctx.ModuleRuntime = rtOpt
//
//	pwd, err := filepath.Abs(bp.GetModuleWorkingDir(rtOpt.Path))
//	if err != nil {
//		return ctx, err
//	}
//	configFile := filepath.Join(pwd, pomFile)
//	_, err = os.Stat(configFile)
//	if os.IsNotExist(err) {
//		err = errors.New("configuration file not found")
//		return ctx, err
//	}
//
//	yamlFile, err := ioutil.ReadFile(configFile)
//	if err != nil {
//		err = errors.New(fmt.Sprintf("read application config file get error %v", err))
//		return ctx, err
//	}
//
//	var pomProject POM
//	err = xml.Unmarshal(yamlFile, &pomProject)
//	if err != nil {
//		err = errors.New(fmt.Sprintf("unmarshal application config file get error %v", err))
//		return ctx, err
//	}
//
//	if len(strings.TrimSpace(pomProject.GroupId)) == 0 {
//		pomProject.GroupId = pomProject.Parent.GroupId
//	}
//	addPOMToContext(ctx, pomProject)
//	ctx.RepositoryConfig, err = bp.Runtime.GetRepo(rtOpt.RepoId)
//	if err != nil {
//		return ctx, err
//	}
//	return ctx, nil
//}
//
//func getPOMFromContext(ctx PublishContext) POM {
//	v, err := ctx.Get("pom")
//	if err != nil {
//		LogFatal(*ctx.BuildPack.Error("", err))
//	}
//	return v.(POM)
//}
//
//func addPOMToContext(ctx PublishContext, pom POM) {
//	ctx.Add("pom", pom)
//}
//
//func (p *ArtifactoryMvn) Verify(ctx PublishContext) error {
//	return nil
//}
//
//func (p *ArtifactoryMvn) Pre(ctx PublishContext) error {
//	rtModule := ctx.ModuleRuntime
//	pom := getPOMFromContext(ctx)
//	version := ctx.Runtime.VersionRuntime.GetVersion(labelSnapshot, 0)
//
//	uploadParams := make([]MVNUploadParam, 0)
//	//generic information for upload
//	artifact := ctx.RepositoryConfig.URL
//	repository := ctx.RepositoryConfig.ChannelConfig.Stable
//	if !ctx.Runtime.Release {
//		repository = ctx.RepositoryConfig.ChannelConfig.Unstable
//	}
//
//	args := strings.Split(pom.GroupId, ".")
//	args = append(args, pom.ArtifactId)
//	args = append(args, version)
//	modulePath := strings.Join(args, "/")
//	// end generic information
//
//	//copy pom
//	pomSrc := ctx.BuildPathOnRoot(rtModule.Path, "target", pomFile)
//	pomName := fmt.Sprintf("%s-%s.pom", pom.ArtifactId, version)
//	pomPublished := ctx.BuildPathOnRoot(PublishDirectory, pomName)
//	err := CopyFile(pomSrc, pomPublished)
//	if err != nil {
//		return err
//	}
//
//	pomSumSHA256, err := checksumSHA256(pomPublished)
//	if err != nil {
//		return err
//	}
//	pomSumMD5, err := checksumMD5(pomPublished)
//	if err != nil {
//		return err
//	}
//	pomParam := MVNUploadParam{
//		Destination: fmt.Sprintf("%s/%s/%s/%s", artifact, repository, modulePath, pomName),
//		Source:      pomPublished,
//		Version:     version,
//		Username:    ctx.RepositoryConfig.Username,
//		Password:    ctx.RepositoryConfig.Password,
//		CheckSum: CheckSum{
//			SHA256: pomSumSHA256,
//			MD5:    pomSumMD5,
//		},
//	}
//	uploadParams = append(uploadParams, pomParam)
//	LogInfo(*ctx.BuildPack, fmt.Sprintf("Prepare for pushting %s to %s", pomParam.Source, pomParam.Destination))
//
//	if pom.Classifier == "jar" || len(strings.TrimSpace(pom.Classifier)) == 0 {
//		//copy jar
//		jarName := fmt.Sprintf("%s-%s.jar", pom.ArtifactId, version)
//		jarSrc := ctx.BuildPathOnRoot(rtModule.Path, "target", jarName)
//		jarPublished := ctx.BuildPathOnRoot(PublishDirectory, jarName)
//		err := CopyFile(jarSrc, jarPublished)
//		if err != nil {
//			return err
//		}
//		jarSumSHA256, err := checksumSHA256(jarPublished)
//		if err != nil {
//			return err
//		}
//		jarSumMD5, err := checksumMD5(jarPublished)
//		if err != nil {
//			return err
//		}
//		jarParam := MVNUploadParam{
//			Destination: fmt.Sprintf("%s/%s/%s/%s", artifact, repository, modulePath, jarName),
//			Source:      jarPublished,
//			Version:     version,
//			Username:    ctx.RepositoryConfig.Username,
//			Password:    ctx.RepositoryConfig.Password,
//			CheckSum: CheckSum{
//				SHA256: jarSumSHA256,
//				MD5:    jarSumMD5,
//			},
//		}
//		uploadParams = append(uploadParams, jarParam)
//		LogInfo(*ctx.BuildPack, fmt.Sprintf("Prepare for pushting %s to %s", jarParam.Source, jarParam.Destination))
//
//	}
//	p.uploads = uploadParams
//	return nil
//}
//
//func (p *ArtifactoryMvn) Publish(ctx PublishContext) error {
//	for _, upload := range p.uploads {
//		err := uploadFile(upload)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//func (p *ArtifactoryMvn) Clean(ctx PublishContext) error {
//	return nil
//}
//
//func uploadFile(param MVNUploadParam) error {
//	data, err := os.Open(param.Source)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		_ = data.Close()
//	}()
//	req, err := http.NewRequest("PUT", param.Destination, data)
//	if err != nil {
//		return err
//	}
//	req.Header.Set("Content-Type", "text/plain")
//	req.Header.Set("X-CheckSum-MD5", param.MD5)
//	req.SetBasicAuth(param.Username, param.Password)
//
//	client := &http.Client{}
//	res, err := client.Do(req)
//	if err != nil {
//		return err
//	}
//	defer func() {
//		_ = res.Body.Close()
//	}()
//	if res.StatusCode != http.StatusCreated {
//		return errors.New(res.Status)
//	}
//	return nil
//}
//
//func checksumSHA256(file string) (string, error) {
//	hasher := sha256.New()
//	f, err := os.Open(file)
//	if err != nil {
//		return "", err
//	}
//	defer func() {
//		_ = f.Close()
//	}()
//	if _, err := io.Copy(hasher, f); err != nil {
//		return "", err
//	}
//	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
//}
//
//func checksumMD5(file string) (string, error) {
//	hasher := md5.New()
//	f, err := os.Open(file)
//	if err != nil {
//		return "", err
//	}
//	defer func() {
//		_ = f.Close()
//	}()
//	if _, err := io.Copy(hasher, f); err != nil {
//		return "", err
//	}
//	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
//}
