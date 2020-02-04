package publisher

import (
	"errors"
	. "scm.wcs.fortna.com/lngo/buildpack"
)

type Publisher interface {
	WriteConfig(bp BuildPack, opt ModuleConfig) error
	CreateContext(bp *BuildPack, rtOpt ModuleRuntime) (PublishContext, error)

	Verify(ctx PublishContext) error
	Pre(ctx PublishContext) error
	Publish(ctx PublishContext) error
	Clean(ctx PublishContext) error
}

type PublishContext struct {
	Name     string
	Path     string
	metadata map[string]interface{}
	*BuildPack
	RepositoryConfig
	ModuleRuntime
}

func NewPublishContext(name, path string) PublishContext {
	return PublishContext{
		Name:     name,
		Path:     path,
		metadata: make(map[string]interface{}),
	}
}

func (c *PublishContext) Add(key string, value interface{}) {
	c.metadata[key] = value
}

func (c *PublishContext) Get(key string) (interface{}, error) {
	v, ok := c.metadata[key]
	if !ok {
		return nil, errors.New("not found metadata by key " + key)
	}
	return v, nil
}

type EmptyPublisher struct {
}

func (p *EmptyPublisher) WriteConfig(bp BuildPack, opt ModuleConfig) error {
	return nil
}

func (p *EmptyPublisher) CreateContext(bp *BuildPack, rtOpt ModuleRuntime) (PublishContext, error) {
	return PublishContext{}, nil
}

func (p *EmptyPublisher) Verify(ctx PublishContext) error {
	return nil
}

func (p *EmptyPublisher) Pre(ctx PublishContext) error {
	return nil
}

func (p *EmptyPublisher) Publish(ctx PublishContext) error {
	return nil
}

func (p *EmptyPublisher) Clean(ctx PublishContext) error {
	return nil
}

var publishers map[string]Publisher

const (
	artifactoryMvnPublisher = "artifactory"
)

func init() {
	publishers = make(map[string]Publisher)
	publishers[artifactoryMvnPublisher] = &ArtifactoryMvn{}
}

func VerifyPublisher(publisherName string) bool {
	_, exist := publishers[publisherName]
	return exist
}

func GetPublisher(publisherName string) Publisher {
	publisher, ok := publishers[publisherName]
	if !ok {
		return &EmptyPublisher{}
	}
	return publisher
}
