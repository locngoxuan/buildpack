package publisher

import (
	"errors"
	"github.com/locngoxuan/buildpack/common"
	"golang.org/x/net/context"
	"net/http"
	"os"
)

type ArtifactoryPackage struct {
	Source   string
	Endpoint string
	Md5      string
	Username string
	Password string
}

func upload(ctx PublishContext, packages []ArtifactoryPackage) error {
	for _, upload := range packages {
		common.PrintLogW(ctx.LogWriter, "uploading %s with md5:%s", upload.Endpoint, upload.Md5)
		err := uploadFile(ctx.Ctx, upload)
		if err != nil {
			return err
		}
	}
	return nil
}

func uploadFile(ctx context.Context, param ArtifactoryPackage) error {
	data, err := os.Open(param.Source)
	if err != nil {
		return err
	}
	defer func() {
		_ = data.Close()
	}()
	req, err := http.NewRequestWithContext(ctx, "PUT", param.Endpoint, data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("X-CheckSum-MD5", param.Md5)
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

type PreparePackage func(ctx PublishContext) ([]ArtifactoryPackage, error)

type ArtifactoryPublisher struct {
	PreparePackage
}

func (n ArtifactoryPublisher) PrePublish(ctx PublishContext) error {
	return nil
}

func (n ArtifactoryPublisher) Publish(ctx PublishContext) error {
	packages, err := n.PreparePackage(ctx)
	if err != nil {
		return err
	}
	if packages == nil || len(packages) == 0 {
		return errors.New("not found any package for publishing")
	}
	return upload(ctx, packages)
}

func (n ArtifactoryPublisher) PostPublish(ctx PublishContext) error {
	return nil
}
