package builder

import (
	"fmt"
	"github.com/jhoonb/archivex"
	"os"
	"path/filepath"
)

const (
	mvnAutoBuildTool = "mvn-auto"
)

type MVNAutoBuildTool struct {
	MVNBuildTool
}

func (c *MVNAutoBuildTool) Name() string {
	return mvnAppBuildTool
}

func (c *MVNAutoBuildTool) Build(ctx BuildContext) error {
	return c.MVNBuildTool.Build(ctx)
}

func (c *MVNAutoBuildTool) PostBuild(ctx BuildContext) error {
	moduleInCommon := filepath.Join(ctx.GetCommonDirectory(), ctx.Name)
	err := os.MkdirAll(moduleInCommon, 0777)
	if err != nil {
		return err
	}
	dockerPreDir := filepath.Join(moduleInCommon, "pre-docker")
	// copy docker folder to pre-docker dir
	src := filepath.Join(ctx.WorkingDir, "target", "docker")
	err = copyDirectory(ctx.BuildPack, src, dockerPreDir)
	if err != nil {
		return err
	}

	// tar info
	tarName := fmt.Sprintf("%s-%s.tar", ctx.Name, ctx.Version)
	tarFile := filepath.Join(moduleInCommon, tarName)
	//create tar at common directory
	tar := new(archivex.TarFile)
	err = tar.Create(tarFile)
	if err != nil {
		return err
	}
	err = tar.AddAll(dockerPreDir, false)
	if err != nil {
		return err
	}
	err = tar.Close()
	if err != nil {
		return err
	}
	return nil
}
