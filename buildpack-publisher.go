package main

import (
	"errors"
	"strings"
)

type Publisher interface {
	WriteConfig(bp BuildPack, opt BuildPackModuleConfig) error
	CreateContext(bp BuildPack, rtOpt BuildPackModuleRuntimeParams) (PublishContext, error)

	Verify(ctx PublishContext) error
	Pre(ctx PublishContext) error
	Publish(ctx PublishContext) error
	Clean(ctx PublishContext) error
}

type PublishContext struct {
	Name     string
	Path     string
	metadata map[string]interface{}
	BuildPack
	BuildPackModuleRuntimeParams
}

func newPublishContext(name, path string) PublishContext {
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

func (p *EmptyPublisher) WriteConfig(bp BuildPack, opt BuildPackModuleConfig) error {
	return nil
}

func (p *EmptyPublisher) CreateContext(bp BuildPack, rtOpt BuildPackModuleRuntimeParams) (PublishContext, error) {
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
	publisherJfrogMvn = "jfrog-mvn"
)

func init() {
	publishers = make(map[string]Publisher)
	publishers[publisherJfrogMvn] = &PublisherJfrogMVN{}
}

func publisherOptions() string {
	names := make([]string, 0)
	for name, _ := range publishers {
		names = append(names, name)
	}
	return strings.Join(names, "/")
}

func doesPublisherExist(publisherName string) bool {
	_, ok := publishers[publisherName]
	return ok
}

func getPublisher(publisherName string) Publisher {
	publisher, ok := publishers[publisherName]
	if !ok {
		return &EmptyPublisher{}
	}
	return publisher
}
