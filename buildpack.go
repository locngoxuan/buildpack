package main

type BuildPhase func() error

type BuildPack struct {
}

type Builder interface {
	Clean() error
	Build() error
	Publish() error
}

type Publisher interface {
	Pre() error
	Publish() error
	Post() error
}

func (b *BuildPack) Init() error {
	return nil
}

func (b *BuildPack) Snapshot() error {
	return nil
}

func (b *BuildPack) Release() error {
	return nil
}
