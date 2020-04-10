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
	version       string
	modules       string
	label         string
	devVersion    string
	relVersion    string
	verbose       bool
	skipContainer bool
	skipUnitTest  bool
	skipPublish   bool
	skipBranching bool
	skipTag       bool
	skipClean     bool
	local         bool
	patch         bool
	help          bool
	debug         bool
	release       bool
	devMode       bool
}

func ReadArgument(f *flag.FlagSet) (args RuntimeConfig, err error) {
	args.flag = f
	err = args.readVersion().
		readDevMode().
		readDebug().
		readHelp().
		readShareData().
		readConfigFile().
		readVerbose().
		readModules().
		readPatch().
		readSkipTest().
		readSkipPublish().
		readSkipTag().
		readSkipBranching().
		readSkipContainer().
		readSkipClean().
		readDevVersion().
		readRelVersion().
		readLabel().
		readRelease().
		readLocal().
		parse()

	if err != nil {
		return
	}

	if args.local {
		args.release = false
		args.skipTag = true
		args.skipBranching = true
		args.skipPublish = true
	}

	if args.devMode {
		args.release = false
		args.skipTag = true
		args.skipBranching = true
		args.skipPublish = true
		args.skipClean = true
		args.skipContainer = true
	}

	return
}

func ReadForUsage(f *flag.FlagSet) (args RuntimeConfig, err error) {
	args.flag = f
	args.readVersion().
		readDevMode().
		readDebug().
		readHelp().
		readShareData().
		readConfigFile().
		readVerbose().
		readModules().
		readPatch().
		readSkipTest().
		readSkipPublish().
		readSkipTag().
		readLocal().
		readSkipBranching().
		readSkipContainer().
		readSkipClean().
		readDevVersion().
		readRelVersion().
		readLabel().
		readRelease()
	err = args.flag.Parse(os.Args[1:])
	if err != nil {
		return
	}
	if args.local {
		args.release = false
		args.skipTag = true
		args.skipBranching = true
		args.skipPublish = true
	}
	if args.devMode {
		args.release = false
		args.skipTag = true
		args.skipBranching = true
		args.skipPublish = true
		args.skipClean = true
		args.skipContainer = true
	}
	return
}

func (a *RuntimeConfig) readDevMode() *RuntimeConfig {
	a.flag.BoolVar(&a.devMode, "dev-mode", false, "Build at local machine")
	return a
}

func (a *RuntimeConfig) readDevVersion() *RuntimeConfig {
	a.flag.StringVar(&a.devVersion, "version-dev", "", "Version of development")
	return a
}

func (a *RuntimeConfig) readRelVersion() *RuntimeConfig {
	a.flag.StringVar(&a.relVersion, "version-rel", "", "Latest stable version")
	return a
}

func (a *RuntimeConfig) readDebug() *RuntimeConfig {
	a.flag.BoolVar(&a.debug, "debug", false, "Enable debug mode to keep .buildpack folder after build complete")
	return a
}

func (a *RuntimeConfig) readSkipTag() *RuntimeConfig {
	a.flag.BoolVar(&a.skipTag, "skip-tag", false, "Skip cleaning after build and publish")
	return a
}

func (a *RuntimeConfig) readSkipClean() *RuntimeConfig {
	a.flag.BoolVar(&a.skipClean, "skip-clean", false, "Skip cleaning after build and publish")
	return a
}

func (a *RuntimeConfig) readLocal() *RuntimeConfig {
	a.flag.BoolVar(&a.local, "local", false, "Local mode")
	return a
}

func (a *RuntimeConfig) readRelease() *RuntimeConfig {
	a.flag.BoolVar(&a.release, "release", false, "Enable release mode, name of build does not include label and it will be published to stable channel of repository")
	return a
}

func (a *RuntimeConfig) readHelp() *RuntimeConfig {
	a.flag.BoolVar(&a.help, "h", false, "Print help")
	return a
}

func (a *RuntimeConfig) readShareData() *RuntimeConfig {
	a.flag.StringVar(&a.shareData, "share-data", "", "Share data directory of buildpack")
	return a
}

func (a *RuntimeConfig) readConfigFile() *RuntimeConfig {
	a.flag.StringVar(&a.configFile, "config", "", "Path to specific configuration file")
	return a
}

func (a *RuntimeConfig) readLabel() *RuntimeConfig {
	a.flag.StringVar(&a.label, "label", "", "Label of build. It may be ignored by builder")
	return a
}

func (a *RuntimeConfig) readVerbose() *RuntimeConfig {
	a.flag.BoolVar(&a.verbose, "verbose", false, "Enable verbose mode, more log is going to printed")
	return a
}

func (a *RuntimeConfig) readPatch() *RuntimeConfig {
	a.flag.BoolVar(&a.patch, "patch", false, "Mark this release as a patch (Only apply on release mode)")
	return a
}

func (a *RuntimeConfig) readSkipTest() *RuntimeConfig {
	a.flag.BoolVar(&a.skipUnitTest, "skip-test", false, "Skip run unit test")
	return a
}

func (a *RuntimeConfig) readSkipPublish() *RuntimeConfig {
	a.flag.BoolVar(&a.skipPublish, "skip-publish", false, "Skip deploy build to repository")
	return a
}

func (a *RuntimeConfig) readSkipBranching() *RuntimeConfig {
	a.flag.BoolVar(&a.skipBranching, "skip-branch", false, "Skip create a new branch in GIT")
	return a
}

func (a *RuntimeConfig) readVersion() *RuntimeConfig {
	a.flag.StringVar(&a.version, "version", "", "Print version number")
	return a
}

func (a *RuntimeConfig) readModules() *RuntimeConfig {
	a.flag.StringVar(&a.modules, "m", "", "A specific set of modules to apply your action")
	return a
}

func (a *RuntimeConfig) readSkipContainer() *RuntimeConfig {
	a.flag.BoolVar(&a.skipContainer, "skip-container", false, "Skip using container-build and change to use local command")
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

func (a *RuntimeConfig) V2PDev() string {
	return a.devVersion
}

func (a *RuntimeConfig) V2PRel() string {
	return a.relVersion
}

func (a *RuntimeConfig) SkipContainer() bool {
	return a.skipContainer
}

func (a *RuntimeConfig) SkipPublish() bool {
	return a.skipPublish
}

func (a *RuntimeConfig) SkipClean() bool {
	return a.skipClean
}

func (a *RuntimeConfig) SkipTag() bool {
	return a.skipTag
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

func (a *RuntimeConfig) IsRelease() bool {
	return a.release
}

func (a *RuntimeConfig) IsLocal() bool {
	return a.local
}

func (a *RuntimeConfig) IsHelp() bool {
	return a.help
}

func (a *RuntimeConfig) IsDebug() bool {
	return a.debug
}

func (a *RuntimeConfig) IsDevMode() bool {
	return a.devMode
}