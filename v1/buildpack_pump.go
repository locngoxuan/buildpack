package v1

import (
	"context"
	"github.com/locngoxuan/buildpack/v1/common"
)

func (bp *BuildPack) pump(ctx context.Context) error {
	cli := common.GetGitClient()
	defer cli.Close()

	ver, err := common.Parse(bp.GetVersion())
	if err != nil {
		common.PrintLog("err %v %v", ver, err)
		return err
	}

	common.PrintLog("tagging version %v", ver.String())
	//tagging current version, i.e 1.0.0, 1.0.1, etc.
	err = cli.Tag(ver.String())
	if err != nil {
		common.PrintLog("tagging error %+v", err)
	}

	//increase 1.0.0 -> 1.0.1
	ver.NextPatch()
	if bp.BuildPath {
		//if pump for patching then push new version then terminate
		common.PrintLog("next version is %v", ver.String())
		return gitUpdateConfig(cli, *bp, ver)
	}

	//it it is pump of releasing, then an branch of 1.0.x must be created
	if bp.BuildRelease {
		branch := ver.MinorBranch()
		common.PrintLog("creating new branch (%v) to archive latest published", branch)
		err = cli.CreateNewBranch(branch)
		if err != nil {
			return err
		}
	}

	if bp.SkipBackward {
		//if new change breaks the concept
		ver.NextMajor()
	} else {
		ver.NextMinor()
	}

	common.PrintLog("next version is %v", ver.String())
	err = gitUpdateConfig(cli, *bp, ver)
	if err != nil {
		return err
	}
	return nil
}
