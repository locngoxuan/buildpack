package publisher

import (
	"errors"
	"fmt"
)

type RepoManager struct {
	Repos map[string]Repository
}

type Repository struct {
	Name     string
	Stable   *RepoChannel
	Unstable *RepoChannel
}

type RepoChannel struct {
	Address  string
	Username string
	Password string
}

var repoMan RepoManager

func SetRepoManager(rm RepoManager) {
	repoMan = rm
}

func (rm RepoManager) pickOne(name string) (r Repository, err error) {
	r, ok := rm.Repos[name]
	if !ok {
		err = errors.New(fmt.Sprintf("repo %s may be not registered", r))
	}
	return
}

func (rm RepoManager) pickChannel(name string, stable bool) (rc RepoChannel, err error) {
	r, err := rm.pickOne(name)
	if err != nil {
		return
	}

	if stable {
		if r.Stable == nil {
			err = fmt.Errorf("stable channel of repo %s is not configured", name)
			return
		}
		rc = *r.Stable
	} else {
		if r.Unstable == nil {
			err = fmt.Errorf("unstable channel of repo %s is not configured", name)
			return
		}
		rc = *r.Unstable
	}
	return
}
