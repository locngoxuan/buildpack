package buildpack

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%+v", *i)
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type SkipOption struct {
	SkipContainer bool
	SkipUnitTest  bool
	SkipPublish   bool
	SkipClean     bool
	SkipBranching bool
}

type RepoArgument struct {
	Id       string
	Username string
	Password string
	Token    string
}

type ActionArguments struct {
	Flag *flag.FlagSet

	shareData          string
	configFile         string
	gitToken           string
	debug              bool
	patch              bool
	backwardCompatible bool
	repoIds            arrayFlags
	repoUsers          arrayFlags
	repoPwds           arrayFlags
	repoTokens         arrayFlags
	skipContainer      bool
	skipUnitTest       bool
	skipPublish        bool
	skipClean          bool
	skipBranching      bool
	version            string
	modules            string
}

func NewActionArguments(f *flag.FlagSet) (*ActionArguments, error) {
	args := &ActionArguments{
		Flag: f,
	}
	err := args.readVersion().
		readShareData().
		readConfigFile().
		readDebug().
		readModules().
		readPatch().
		readBackwardsCompatible().
		readRepoIds().
		readRepoUserName().
		readRepoPassword().
		readRepoAccessToken().
		readGitAccessToken().
		readSkipContainer().
		readSkipClean().
		readSkipPublish().
		readSkipTest().
		parse()
	if err != nil {
		return nil, err
	}
	return args, nil
}

func (a *ActionArguments) readShareData() *ActionArguments {
	a.Flag.StringVar(&a.shareData, "share-data", "", "path to share data folder")
	return a
}

func (a *ActionArguments) readConfigFile() *ActionArguments {
	a.Flag.StringVar(&a.configFile, "config", "", "path to specific configuration file")
	return a
}

func (a *ActionArguments) readGitAccessToken() *ActionArguments {
	a.Flag.StringVar(&a.gitToken, "git-token", "", "access-token of git")
	return a
}

func (a *ActionArguments) readDebug() *ActionArguments {
	a.Flag.BoolVar(&a.debug, "debug", false, "display more log")
	return a
}

func (a *ActionArguments) readPatch() *ActionArguments {
	a.Flag.BoolVar(&a.patch, "patch", false, "true if this release is only apply patch")
	return a
}

func (a *ActionArguments) readBackwardsCompatible() *ActionArguments {
	a.Flag.BoolVar(&a.backwardCompatible, "backwards-compatible", true, "set its to false if there are any backwards incompatible is released")
	return a
}

func (a *ActionArguments) readRepoIds() *ActionArguments {
	a.Flag.Var(&a.repoIds, "repo-id", "list of repository id")
	return a
}

func (a *ActionArguments) readRepoUserName() *ActionArguments {
	a.Flag.Var(&a.repoUsers, "repo-user", "list username follow order of ids")
	return a
}

func (a *ActionArguments) readRepoPassword() *ActionArguments {
	a.Flag.Var(&a.repoPwds, "repo-pass", "list password follow order of ids")
	return a
}

func (a *ActionArguments) readRepoAccessToken() *ActionArguments {
	a.Flag.Var(&a.repoTokens, "repo-token", "list access token follow order of ids")
	return a
}

func (a *ActionArguments) readSkipTest() *ActionArguments {
	a.Flag.BoolVar(&a.skipUnitTest, "skip-ut", false, "skip unit test while running build")
	return a
}

func (a *ActionArguments) readSkipPublish() *ActionArguments {
	a.Flag.BoolVar(&a.skipPublish, "skip-publish", false, "skip publish to artifactory")
	return a
}

func (a *ActionArguments) readSkipClean() *ActionArguments {
	a.Flag.BoolVar(&a.skipClean, "skip-clean", false, "skip cleaning after build and publish")
	return a
}

func (a *ActionArguments) readSkipBranching() *ActionArguments {
	a.Flag.BoolVar(&a.skipBranching, "skip-branch", false, "skip branching after build and publish")
	return a
}

func (a *ActionArguments) readVersion() *ActionArguments {
	a.Flag.StringVar(&a.version, "v", "", "version number")
	return a
}

func (a *ActionArguments) readModules() *ActionArguments {
	a.Flag.StringVar(&a.modules, "m", "", "modules")
	return a
}

func (a *ActionArguments) readSkipContainer() *ActionArguments {
	a.Flag.BoolVar(&a.skipContainer, "skip-container", false, "using docker environment rather than host environment")
	return a
}

func (a *ActionArguments) parse() error {
	return a.Flag.Parse(os.Args[2:])
}

func (a *ActionArguments) ConfigFile() string {
	return strings.TrimSpace(a.configFile)
}

func (a *ActionArguments) ShareData() string {
	return strings.TrimSpace(a.shareData)
}

func (a *ActionArguments) Version() string {
	return strings.TrimSpace(a.version)
}

func (a *ActionArguments) GitAccessToken() string {
	return strings.TrimSpace(a.gitToken)
}

func (a *ActionArguments) Modules() []string {
	v := strings.TrimSpace(a.modules)
	if len(v) == 0 {
		return []string{}
	}
	return strings.Split(v, ",")
}

func (a *ActionArguments) IsDebug() bool {
	return a.debug
}

func (a *ActionArguments) SkipContainer() bool {
	return a.skipContainer
}

func (a *ActionArguments) SkipClean() bool {
	return a.skipClean
}

func (a *ActionArguments) SkipPublish() bool {
	return a.skipPublish
}

func (a *ActionArguments) SkipUnitTest() bool {
	return a.skipUnitTest
}

func (a *ActionArguments) SkipBranching() bool {
	return a.skipBranching
}

func (a *ActionArguments) IsPatch() bool {
	return a.patch
}

func (a *ActionArguments) IsBackwardsCompatible() bool {
	return a.backwardCompatible
}

func (a *ActionArguments) RepoArguments() map[string]RepoArgument {
	rs := make(map[string]RepoArgument)
	for i, id := range a.repoIds {
		r := RepoArgument{
			Id: id,
		}

		if i < len(a.repoUsers) {
			r.Username = a.repoUsers[i]
		}

		if i < len(a.repoPwds) {
			r.Password = a.repoPwds[i]
		}

		if i < len(a.repoTokens) {
			r.Token = a.repoTokens[i]
		}
		rs[id] = r
	}
	return rs
}
