package builder

import (
	"fmt"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/common"
	"strings"
)

type MvnLib struct {
	Mvn
}

func (b MvnLib) PostBuild(ctx BuildContext) error {
	pomFile := filepath.Join(ctx.WorkDir, "target", pomXml)
	pom, err := common.ReadPOM(pomFile)
	if err != nil {
		return err
	}

	//copy pom
	pomSrc := filepath.Join(ctx.WorkDir, "target", pomXml)
	pomName := fmt.Sprintf("%s-%s.pom", pom.ArtifactId, ctx.Version)
	pomPublished := filepath.Join(ctx.OutputDir, pomName)
	err = common.CopyFile(pomSrc, pomPublished)
	if err != nil {
		return err
	}

	if pom.Classifier == "jar" || len(strings.TrimSpace(pom.Classifier)) == 0 {
		//copy jar
		jarName := fmt.Sprintf("%s-%s.jar", pom.ArtifactId, ctx.Version)
		jarSrc := filepath.Join(ctx.WorkDir, "target", jarName)
		jarPublished := filepath.Join(ctx.OutputDir, jarName)
		err := common.CopyFile(jarSrc, jarPublished)
		if err != nil {
			return err
		}

		javaDocName := fmt.Sprintf("%s-%s-javadoc.jar", pom.ArtifactId, ctx.Version)
		javaDocSrc := filepath.Join(ctx.WorkDir, "target", javaDocName)
		if common.Exists(javaDocSrc) {
			javaDocPublished := filepath.Join(ctx.OutputDir, javaDocName)
			err := common.CopyFile(javaDocSrc, javaDocPublished)
			if err != nil {
				return err
			}
		}

		javaSourceName := fmt.Sprintf("%s-%s-sources.jar", pom.ArtifactId, ctx.Version)
		javaSourceSrc := filepath.Join(ctx.WorkDir, "target", javaSourceName)
		if common.Exists(javaDocSrc) {
			javaSourcePublished := filepath.Join(ctx.OutputDir, javaSourceName)
			err := common.CopyFile(javaSourceSrc, javaSourcePublished)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func init() {
	registries["mvn_lib"] = &MvnLib{}
}
