package buildpack

import (
	"flag"
	"os"
	"strings"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type RepoArgument struct {
	Id       string
	Username string
	Password string
	Token    string
}

type ActionArguments struct {
	Flag   *flag.FlagSet
	Values map[string]interface{}
}

func NewActionArguments(f *flag.FlagSet) *ActionArguments {
	return &ActionArguments{
		Flag:   f,
		Values: make(map[string]interface{}),
	}
}

func InitCommanActionArguments(f *flag.FlagSet) *ActionArguments {
	args := NewActionArguments(f)
	return args.ReadVersion().
		ReadModules().
		ReadContainer().
		ReadRepoIds().
		ReadRepoUserName().
		ReadRepoPassword().
		ReadRepoAccessToken().
		ReadGitAccessToken().
		ReadSkipClean().
		ReadSkipPublish().
		ReadSkipTest()
}

func (a *ActionArguments) ReadGitAccessToken() *ActionArguments {
	s := a.Flag.String("git-token", "", "access-token of git")
	a.Values["git-token"] = s
	return a
}

func (a *ActionArguments) ReadRepoIds() *ActionArguments {
	var arrVals arrayFlags
	a.Flag.Var(&arrVals, "repo-id", "list of repository id")
	a.Values["repo-id"] = arrVals
	return a
}

func (a *ActionArguments) ReadRepoUserName() *ActionArguments {
	var arrVals arrayFlags
	a.Flag.Var(&arrVals, "repo-user", "list username follow order of ids")
	a.Values["repo-user"] = arrVals
	return a
}

func (a *ActionArguments) ReadRepoPassword() *ActionArguments {
	var arrVals arrayFlags
	a.Flag.Var(&arrVals, "repo-pass", "list password follow order of ids")
	a.Values["repo-pass"] = arrVals
	return a
}

func (a *ActionArguments) ReadRepoAccessToken() *ActionArguments {
	var arrVals arrayFlags
	a.Flag.Var(&arrVals, "repo-token", "list access token follow order of ids")
	a.Values["repo-token"] = arrVals
	return a
}

func (a *ActionArguments) ReadSkipTest() *ActionArguments {
	s := a.Flag.Bool("skip-ut", false, "skip unit test while running build")
	a.Values["skip-ut"] = s
	return a
}

func (a *ActionArguments) ReadSkipPublish() *ActionArguments {
	s := a.Flag.Bool("skip-publish", false, "skip publish to artifactory")
	a.Values["skip-publish"] = s
	return a
}

func (a *ActionArguments) ReadSkipClean() *ActionArguments {
	s := a.Flag.Bool("skip-clean", false, "skip cleaning after build and publish")
	a.Values["skip-clean"] = s
	return a
}

func (a *ActionArguments) ReadVersion() *ActionArguments {
	s := a.Flag.String("v", "", "version number")
	a.Values["v"] = s
	return a
}

func (a *ActionArguments) ReadModules() *ActionArguments {
	s := a.Flag.String("m", "", "modules")
	a.Values["m"] = s
	return a
}

func (a *ActionArguments) ReadContainer() *ActionArguments {
	s := a.Flag.Bool("container", false, "using docker environment rather than host environment")
	a.Values["container"] = s
	return a
}

func (a *ActionArguments) Parse() error {
	return a.Flag.Parse(os.Args[2:])
}

func (a *ActionArguments) Version() string {
	s, ok := a.Values["v"]
	if !ok {
		return ""
	}
	return strings.TrimSpace(*(s.(*string)))
}

func (a *ActionArguments) Modules() []string {
	s, ok := a.Values["m"]
	if !ok {
		return []string{}
	}
	v := strings.TrimSpace(*(s.(*string)))
	if len(v) == 0 {
		return []string{}
	}
	return strings.Split(v, ",")
}

func (a *ActionArguments) Container() bool {
	s, ok := a.Values["container"]
	if !ok {
		return false
	}
	return *(s.(*bool))
}

func (a *ActionArguments) SkipClean() bool {
	s, ok := a.Values["skip-clean"]
	if !ok {
		return false
	}
	return *(s.(*bool))
}

func (a *ActionArguments) SkipPublish() bool {
	s, ok := a.Values["skip-publish"]
	if !ok {
		return false
	}
	return *(s.(*bool))
}

func (a *ActionArguments) SkipUnitTest() bool {
	s, ok := a.Values["skip-ut"]
	if !ok {
		return false
	}
	return *(s.(*bool))
}

func (a *ActionArguments) RepoArguments() map[string]RepoArgument {
	rs := make(map[string]RepoArgument)

	repoIds, _ := a.Values["repo-id"].([]string)
	repoUsers, _ := a.Values["repo-user"].([]string)
	repoPwds, _ := a.Values["repo-pass"].([]string)
	repoTokens, _ := a.Values["repo-token"].([]string)

	for i, id := range repoIds {
		r := RepoArgument{
			Id: id,
		}

		if i < len(repoUsers) {
			r.Username = repoUsers[i]
		}

		if i < len(repoPwds) {
			r.Password = repoPwds[i]
		}

		if i < len(repoTokens) {
			r.Token = repoTokens[i]
		}
		rs[id] = r
	}
	return rs
}
