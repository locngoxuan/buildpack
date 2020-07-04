package publisher

import (
	"scm.wcs.fortna.com/lngo/buildpack/common"
)

type DockerSql struct {
	DockerPublisher
}

func init() {
	docker := &DockerSql{}
	docker.PrepareImage = func(ctx PublishContext, client common.DockerClient) (strings []string, e error) {
		return nil, nil
	}
	registries["docker_sql"] = docker
}
