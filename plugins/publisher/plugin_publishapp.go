package main

import (
	"fmt"
	"scm.wcs.fortna.com/lngo/buildpack/publisher"
)

func GetPublisher() publisher.Interface {
	return &AppPublisher{}
}

type AppPublisher struct {
}

func (b AppPublisher) PrePublish(ctx publisher.PublishContext) error {
	fmt.Println("doing something")
	return nil
}

func (b AppPublisher) Publish(ctx publisher.PublishContext) error {
	fmt.Println("doing something")
	return nil
}

func (b AppPublisher) PostPublish(ctx publisher.PublishContext) error {
	fmt.Println("doing something")
	return nil
}
