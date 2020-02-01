package main

import (
	"errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	tagName  = "build-pack"
	tagEmail = "buildpack@fortna.com"
)

type GitClient struct {
	Repo          *git.Repository
	CurrentBranch *plumbing.Reference
	Remote        *git.Remote
}

func initGitClient(root string) (cli GitClient, err error) {
	cli.Repo, err = git.PlainOpen(root)
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
		if useHttp(remote) {
			cli.Remote = remote
			return
		}
	}

	if cli.Remote == nil {
		for _, remote := range remotes {
			if remote.Config().Name == "origin" {
				cli.Remote = remote
				return
			}
		}
	}

	if cli.Remote == nil {
		cli.Remote = remotes[0]
	}

	if cli.Remote == nil {
		err = errors.New("not found remote from git config")
		return
	}
	return
}

func auth(remote *git.Remote, gitConfig GitRuntimeParams) (transport.AuthMethod, error) {
	if useHttp(remote) {
		return &http.BasicAuth{
			Username: "token",
			Password: gitConfig.AccessToken,
		}, nil
	} else {
		sshKey, _ := ioutil.ReadFile(gitConfig.SSHPath)
		publicKey, err := ssh.NewPublicKeys("git", []byte(sshKey), gitConfig.SSHPass)
		if err != nil {
			return nil, err
		}
		return publicKey, nil
	}
}

func useHttp(r *git.Remote) bool {
	for _, url := range r.Config().URLs {
		if strings.HasPrefix(url, "https") {
			return true
		}
	}
	return false
}

func (c *GitClient) tag(gitConfig GitRuntimeParams, version string) error {
	reference, err := c.Repo.Storer.Reference(c.CurrentBranch.Name())
	if err != nil {
		return err
	}
	tag := object.Tag{
		Name:    version,
		Message: "Release of " + version,
		Tagger: object.Signature{
			Name:  tagName,
			Email: tagEmail,
			When:  time.Now(),
		},
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

	auth, err := auth(c.Remote, gitConfig)
	if err != nil {
		return err
	}
	//printOutput(fmt.Sprintf("push tag %s", tagReferenceName))
	err = c.Repo.Push(&git.PushOptions{
		RemoteName: c.Remote.Config().Name,
		RefSpecs: []config.RefSpec{
			config.RefSpec(tagReferenceName + ":" + tagReferenceName),
		},
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *GitClient) branch(gitConfig GitRuntimeParams, branchName string) error {
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
	auth, err := auth(c.Remote, gitConfig)
	if err != nil {
		return err
	}
	//printOutput(fmt.Sprintf("push branch %s ", newBranchName))
	err = c.Repo.Push(&git.PushOptions{
		RemoteName: c.Remote.Config().Name,
		RefSpecs: []config.RefSpec{
			config.RefSpec(newBranchName + ":" + newBranchName),
		},
		Progress: os.Stdout,
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
