package main

type PublisherJfrog struct {
	BuildPack
}

func (p *PublisherJfrog) SetBuildPack(bp BuildPack) {
	p.BuildPack = bp
}
func (p *PublisherJfrog) LoadConfig(rtOpt BuildPackModuleRuntimeParams, bp BuildPack) error {
	return nil
}
func (p *PublisherJfrog) Pre() error {
	return nil
}
func (p *PublisherJfrog) Publish() error {
	return nil
}
func (p *PublisherJfrog) Clean() error {
	return nil
}
