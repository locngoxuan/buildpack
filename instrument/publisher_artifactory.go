package instrument

import (
	"context"
	"fmt"
	"log"
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

func uploadFile(ctx context.Context, param ArtifactoryPackage) error {
	log.Printf("publish package to %s", param.Endpoint)
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
		return fmt.Errorf(res.Status)
	}
	return nil
}
