package buildpack

import (
	"flag"
	"os"
	"strings"
)

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

type RuntimeConfig struct {
	flag *flag.FlagSet

	shareData     string
	configFile    string
	verbose       bool
	skipContainer bool
	skipUnitTest  bool
	skipPublish   bool
	skipBranching bool
	patch         bool
	version       string
	modules       string
	label         string
	help          bool
}

func ReadArgument(f *flag.FlagSet) (args RuntimeConfig, err error) {
	args.flag = f
	err = args.readVersion().
		readHelp().
		readShareData().
		readConfigFile().
		readVerbose().
		readModules().
		readPatch().
		readSkipTest().
		readSkipPublish().
		readSkipBranching().
		readSkipContainer().
		readLabel().
		parse()
	return
}

func (a *RuntimeConfig) readHelp() *RuntimeConfig {
	a.flag.BoolVar(&a.help, "h", false, "print help")
	return a
}

func (a *RuntimeConfig) readShareData() *RuntimeConfig {
	a.flag.StringVar(&a.shareData, "share-data", "", "path to share data folder")
	return a
}

func (a *RuntimeConfig) readConfigFile() *RuntimeConfig {
	a.flag.StringVar(&a.configFile, "config", "", "path to specific configuration file")
	return a
}

func (a *RuntimeConfig) readLabel() *RuntimeConfig {
	a.flag.StringVar(&a.label, "label", "", "label of build")
	return a
}

func (a *RuntimeConfig) readVerbose() *RuntimeConfig {
	a.flag.BoolVar(&a.verbose, "v", false, "print verbose while running")
	return a
}

func (a *RuntimeConfig) readPatch() *RuntimeConfig {
	a.flag.BoolVar(&a.patch, "patch", false, "build and publish patch")
	return a
}

func (a *RuntimeConfig) readSkipTest() *RuntimeConfig {
	a.flag.BoolVar(&a.skipUnitTest, "skip-ut", false, "skip unit test while running build")
	return a
}

func (a *RuntimeConfig) readSkipPublish() *RuntimeConfig {
	a.flag.BoolVar(&a.skipPublish, "skip-publish", false, "skip publish to artifactory")
	return a
}

func (a *RuntimeConfig) readSkipBranching() *RuntimeConfig {
	a.flag.BoolVar(&a.skipBranching, "skip-branch", false, "skip branching after build and publish")
	return a
}

func (a *RuntimeConfig) readVersion() *RuntimeConfig {
	a.flag.StringVar(&a.version, "version", "", "version number")
	return a
}

func (a *RuntimeConfig) readModules() *RuntimeConfig {
	a.flag.StringVar(&a.modules, "m", "", "modules")
	return a
}

func (a *RuntimeConfig) readSkipContainer() *RuntimeConfig {
	a.flag.BoolVar(&a.skipContainer, "skip-container", false, "using docker environment rather than host environment")
	return a
}

func (a *RuntimeConfig) parse() error {
	return a.flag.Parse(os.Args[2:])
}

func (a *RuntimeConfig) ConfigFile() string {
	return strings.TrimSpace(a.configFile)
}

func (a *RuntimeConfig) ShareData() string {
	return strings.TrimSpace(a.shareData)
}

func (a *RuntimeConfig) Version() string {
	return strings.TrimSpace(a.version)
}

func (a *RuntimeConfig) Label() string {
	return strings.TrimSpace(a.label)
}

func (a *RuntimeConfig) Modules() []string {
	v := strings.TrimSpace(a.modules)
	if len(v) == 0 {
		return []string{}
	}
	return strings.Split(v, ",")
}

func (a *RuntimeConfig) Verbose() bool {
	return a.verbose
}

func (a *RuntimeConfig) SkipContainer() bool {
	return a.skipContainer
}

func (a *RuntimeConfig) SkipPublish() bool {
	return a.skipPublish
}

func (a *RuntimeConfig) SkipUnitTest() bool {
	return a.skipUnitTest
}

func (a *RuntimeConfig) SkipBranching() bool {
	return a.skipBranching
}

func (a *RuntimeConfig) IsPatch() bool {
	return a.patch
}

func (a *RuntimeConfig) IsHelp() bool {
	return a.help
}
