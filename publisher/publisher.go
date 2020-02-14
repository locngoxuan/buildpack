package publisher

import (
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack"
)

type Publisher struct {
	PublishTool
	PublishContext
}

type PublishTool interface {
	Name() string
	GenerateConfig(ctx PublishContext) error
	LoadConfig(ctx PublishContext) error
	Clean(ctx PublishContext) error
	PrePublish(ctx PublishContext) error
	Publish(ctx PublishContext) error
	PostPublish(ctx PublishContext) error
}

type PublishContext struct {
	Name string
	Path string
	buildpack.BuildPack
	Release    bool
	WorkingDir string
	Version    string
	Values     map[string]interface{}
}

func (bc *PublishContext) GetFile(args ...string) string {
	parts := []string{
		bc.WorkingDir,
	}
	parts = append(parts, args...)
	p, err := filepath.Abs(filepath.Join(parts...))
	if err != nil {
		buildpack.LogFatal(bc.Error("", err))
	}
	return p
}

func CreatePublisher(bp buildpack.BuildPack, moduleConfig buildpack.ModuleConfig, release bool, version string) (Publisher, error) {
	p := Publisher{
		PublishContext: PublishContext{
			moduleConfig.Name,
			moduleConfig.Path,
			bp,
			release,
			bp.GetModuleWorkingDir(moduleConfig.Path),
			bp.Config.Version,
			make(map[string]interface{}),
		},
	}
	repoConfig, err := bp.GetRepoById(moduleConfig.RepoId)
	if err != nil {
		return p, err
	}

	tool, ok := publishTools[repoConfig.Name]
	if !ok {
		tool = &DoNothingPublishTool{}
	}
	p.PublishTool = tool
	p.PublishContext.Version = version
	return p, p.LoadConfig(p.PublishContext)
}

func (b *Publisher) ToolName() string {
	return b.PublishTool.Name()
}

func (b *Publisher) GenerateConfig() error {
	return b.PublishTool.GenerateConfig(b.PublishContext)
}

func (b *Publisher) Clean() error {
	return b.PublishTool.Clean(b.PublishContext)
}

func (b *Publisher) PrePublish() error {
	return b.PublishTool.PrePublish(b.PublishContext)
}

func (b *Publisher) Publish() error {
	return b.PublishTool.Publish(b.PublishContext)
}

func (b *Publisher) PostPublish() error {
	return b.PublishTool.PostPublish(b.PublishContext)
}

// DUMMY TOOL

type DoNothingPublishTool struct {
}

func (c *DoNothingPublishTool) Name() string {
	return "dump-publish-tool"
}
func (c *DoNothingPublishTool) GenerateConfig(ctx PublishContext) error {
	return nil
}
func (c *DoNothingPublishTool) LoadConfig(ctx PublishContext) error {
	return nil
}
func (c *DoNothingPublishTool) Clean(ctx PublishContext) error {
	return nil
}
func (c *DoNothingPublishTool) PrePublish(ctx PublishContext) error {
	return nil
}
func (c *DoNothingPublishTool) Publish(ctx PublishContext) error {
	return nil
}
func (c *DoNothingPublishTool) PostPublish(ctx PublishContext) error {
	return nil
}