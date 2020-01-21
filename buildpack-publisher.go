package main

import (
	"github.com/pkg/errors"
	"strings"
)

var publishers map[string]Publisher

func init() {
	publishers = make(map[string]Publisher)
	publishers["jfrog"] = &PublisherJfrog{}
	publishers["docker"] = nil
}

func publisherOptions() string {
	names := make([]string, 0)
	for name, _ := range publishers {
		names = append(names, name)
	}
	return strings.Join(names, "/")
}

func getPublisher(builderName string) (Publisher, error) {
	publisher, ok := publishers[builderName]
	if !ok {
		return nil, errors.New("can not find builder by name " + builderName)
	}
	return publisher, nil
}
