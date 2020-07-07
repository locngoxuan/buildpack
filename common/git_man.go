package common

import (
	"errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"strings"
	"time"
)

var gitCli GitClient

func GetGitClient() GitClient {
	return gitCli
}

func SetGitClient(cli GitClient) {
	gitCli = cli
}

type GitClient struct {
	WorkDir       string
	AccessToken   string
	Name          string
	Email         string
	Repo          *git.Repository
	CurrentBranch *plumbing.Reference
	Remote        *git.Remote
}

func (cli *GitClient) Close() {
	/**
	close git client if it is necessary
	 */
}

func (cli *GitClient) OpenCurrentRepo() (err error) {
	cli.Repo, err = git.PlainOpen(cli.WorkDir)
	if err != nil {
		return
	}

	if cli.Repo == nil {
		err = errors.New("can not open repo")
		return
	}

	branchRefs, err := cli.Repo.Branches()
	if err != nil {
		return
	}

	headRef, err := cli.Repo.Head()
	if err != nil {
		return
	}

	err = branchRefs.ForEach(func(branchRef *plumbing.Reference) error {
		if branchRef.Hash() == headRef.Hash() {
			cli.CurrentBranch = branchRef
			return nil
		}

		return nil
	})
	if err != nil {
		return
	}

	if cli.CurrentBranch == nil {
		err = errors.New("can not found current branch")
	}

	remotes, err := cli.Repo.Remotes()
	if err != nil {
		return
	}

	if len(remotes) == 0 {
		err = errors.New("not found remote from git config")
		return
	}

	for _, remote := range remotes {
		if _, yes := useHttp(remote); yes {
			cli.Remote = remote
			return
		}
	}

	if cli.Remote == nil {
		err = errors.New("not found URLs started with https from any remote")
		return
	}
	err = cli.Validate()
	return
}

func gitAuth(remote *git.Remote, accessToken string) (transport.AuthMethod, error) {
	if _, yes := useHttp(remote); yes {
		return &http.BasicAuth{
			Username: "token",
			Password: accessToken,
		}, nil
	} else {
		return nil, errors.New("not found http remote")
	}
}

func useHttp(r *git.Remote) (string, bool) {
	for _, url := range r.Config().URLs {
		if strings.HasPrefix(url, "https") {
			return url, true
		}
	}
	return "", false
}

func (c *GitClient) Validate() error {
	url, ok := useHttp(c.Remote)
	if !ok {
		return errors.New("not found URLs started with https in any remote")
	}
	auth, err := gitAuth(c.Remote, c.AccessToken)
	if err != nil {
		return err
	}
	_, err = git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:        url,
		Auth:       auth,
		NoCheckout: true,
	})
	if err != nil {
		return err
	}
	return err
}

func (c *GitClient) signature() *object.Signature {
	return &object.Signature{
		Name:  c.Name,
		Email: c.Email,
		When:  time.Now(),
	}
}

func (c *GitClient) Tag(version string) error {
	reference, err := c.Repo.Storer.Reference(c.CurrentBranch.Name())
	if err != nil {
		return err
	}
	tag := object.Tag{
		Name:       version,
		Message:    "Release of " + version,
		Tagger:     *c.signature(),
		Target:     reference.Hash(),
		TargetType: plumbing.CommitObject,
	}

	e := c.Repo.Storer.NewEncodedObject()
	err = tag.Encode(e)
	if err != nil {
		return err
	}
	hash, err := c.Repo.Storer.SetEncodedObject(e)
	if err != nil {
		return err
	}

	tagReferenceName := plumbing.NewTagReferenceName(version)
	//printOutput(fmt.Sprintf("create tag %s", tagReferenceName))

	tagRefer := plumbing.NewHashReference(tagReferenceName, hash)
	err = c.Repo.Storer.SetReference(tagRefer)
	if err != nil {
		return err
	}

	auth, err := gitAuth(c.Remote, c.AccessToken)
	if err != nil {
		return err
	}
	//printOutput(fmt.Sprintf("push tag %s", tagReferenceName))
	err = c.Repo.Push(&git.PushOptions{
		RemoteName: c.Remote.Config().Name,
		RefSpecs: []config.RefSpec{
			config.RefSpec(tagReferenceName + ":" + tagReferenceName),
		},
		Progress: logOutput,
		Auth:     auth,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *GitClient) CreateNewBranch(branchName string) error {
	newBranchName := plumbing.NewBranchReferenceName(branchName)

	wt, err := c.Repo.Worktree()
	if err != nil {
		return err
	}
	//printOutput(fmt.Sprintf("switch branch %s -> %s", currentBranch, newBranchName))
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: newBranchName,
		Force:  true,
		Create: true,
		Keep:   true,
	})
	if err != nil {
		return err
	}
	auth, err := gitAuth(c.Remote, c.AccessToken)
	if err != nil {
		return err
	}
	//printOutput(fmt.Sprintf("push branch %s ", newBranchName))
	err = c.Repo.Push(&git.PushOptions{
		RemoteName: c.Remote.Config().Name,
		RefSpecs: []config.RefSpec{
			config.RefSpec(newBranchName + ":" + newBranchName),
		},
		Progress: logOutput,
		Auth:     auth,
	})
	if err != nil {
		return err
	}

	//printOutput(fmt.Sprintf("switch back %s -> %s", newBranchName, currentBranch))
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: c.CurrentBranch.Name(),
		Force:  true,
		Create: false,
		Keep:   false,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *GitClient) Add(path string) error {
	wt, err := c.Repo.Worktree()
	if err != nil {
		return err
	}

	_, err = wt.Add(path)
	if err != nil {
		return err
	}

	return nil
}

func (c *GitClient) Commit(msg string) error {
	wt, err := c.Repo.Worktree()
	if err != nil {
		return err
	}
	sign := c.signature()
	_, err = wt.Commit(msg, &git.CommitOptions{
		All:       true,
		Author:    sign,
		Committer: sign,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *GitClient) Push() error {
	auth, err := gitAuth(c.Remote, c.AccessToken)
	if err != nil {
		return err
	}

	err = c.Repo.Push(&git.PushOptions{
		RemoteName: c.Remote.Config().Name,
		RefSpecs: []config.RefSpec{
			config.RefSpec(c.CurrentBranch.Name() + ":" + c.CurrentBranch.Name()),
		},
		Progress: logOutput,
		Auth:     auth,
	})
	if err != nil {
		return err
	}
	return nil
}
