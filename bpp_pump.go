package main

import (
	"context"
	"fmt"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/core"
	"github.com/locngoxuan/buildpack/utils"
	"gopkg.in/yaml.v2"
	"log"
)

func pump(ctx context.Context) error {
	branchName := cfg.GitConfig.Branch
	if !utils.IsStringEmpty(arg.GitBranch) {
		branchName = arg.GitBranch
	}
	branchName = utils.Trim(branchName)

	if utils.IsStringEmpty(branchName) {
		branchName = "master"
	}

	gitClient := &core.GitClient{
		GitOption: core.GitOption{
			WorkDir:       workDir,
			Branch:        branchName,
			RemoteAddress: cfg.GitConfig.RemoteAddress,
			GitCredential: cfg.GitConfig.GitCredential,
		},
	}
	err := gitClient.CloneIntoMemory(ctx)
	if err != nil {
		return fmt.Errorf("clone error %v", err)
	}

	v, err := core.Parse(buildVersion)
	if err != nil {
		return fmt.Errorf("can not recognize version %v", err)
	}

	log.Printf("create tag of version %s", buildVersion)
	err = gitClient.Tag(ctx, v.String())
	if err != nil {
		return fmt.Errorf("tagging before pumping version error: %v", err)
	}

	v.NextPatch()
	if arg.BuildPath {
		//if pump for patching then push new version then terminate
		return updateVersion(ctx, v.String(), gitClient)
	}

	//it it is pump of releasing, then an branch of 1.0.x must be created
	if arg.BuildRelease {
		branch := v.MinorBranch()
		err = gitClient.CreateNewBranch(branch)
		if err != nil {
			return err
		}
	}

	if arg.SkipBackward {
		//if new change breaks the concept
		v.NextMajor()
	} else {
		v.NextMinor()
	}

	return updateVersion(ctx, v.String(), gitClient)
}

func updateVersion(ctx context.Context, nextVer string, gitClient *core.GitClient) error {
	log.Printf("next version is %s", nextVer)
	cfg.Version = nextVer
	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal data error %v", err)
	}
	err = gitClient.WriteSingleFile(bytes, config.ConfigProject, fmt.Sprintf("increasing version to %s", nextVer))
	if err != nil {
		return fmt.Errorf("write file error %v", err)
	}
	err = gitClient.Push(ctx)
	if err != nil {
		return fmt.Errorf("push error %v", err)
	}
	return nil
}
