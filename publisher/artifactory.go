package publisher

import (
	"errors"
	"net/http"
	"os"
)

type ArtifactPackage struct {
	Source      string
	Destination string
	MD5         string
	Username    string
	Password    string
}

func uploadFile(param ArtifactPackage) error {
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
