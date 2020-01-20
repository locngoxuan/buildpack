package main

import "github.com/pkg/errors"

var publishers map[string]Publisher

func init() {
	publishers = make(map[string]Publisher)
}

func getPublisher(builderName string) (Publisher, error) {
	publisher, ok := publishers[builderName]
	if !ok {
		return nil, errors.New("can not find builder by name " + builderName)
	}
	return publisher, nil
}
