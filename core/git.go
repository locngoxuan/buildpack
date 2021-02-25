package core

import (
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/utils"
	"io/ioutil"
	"os"
	"time"
)

func signature() *object.Signature {
	return &object.Signature{
		Name: "system",
		When: time.Now(),
	}
}

type GitOption struct {
	WorkDir       string
	Branch        string
	RemoteAddress string
	config.GitCredential
}

type GitClient struct {
	GitOption
	ReferenceName plumbing.ReferenceName
	Repo          *git.Repository
}

func (c *GitClient) auth() (transport.AuthMethod, error) {
	switch c.GitCredential.Type {
	case config.CredentialToken:
		return &http.BasicAuth{
			Username: "token",
			Password: utils.ReadEnvVariableIfHas(c.GitCredential.AccessToken),
		}, nil
	case config.CredentialAccount:
		return &http.BasicAuth{
			Username: utils.ReadEnvVariableIfHas(c.GitCredential.Username),
			Password: utils.ReadEnvVariableIfHas(c.GitCredential.Password),
		}, nil
	}
	return nil, fmt.Errorf("can not recognize credential type")
}

func (c *GitClient) CloneIntoMemory() error {
	auth, err := c.auth()
	if err != nil {
		return err
	}
	c.ReferenceName = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", c.Branch))
	c.Repo, err = git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		ReferenceName: c.ReferenceName,
		URL:           c.GitOption.RemoteAddress,
		Auth:          auth,
		SingleBranch:  true,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *GitClient) WriteSingleFile(data []byte, file, commitMsg string) error {
	wt, err := c.Repo.Worktree()
	if err != nil {
		return err
	}
	f, err := wt.Filesystem.OpenFile(file, os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	_, err = wt.Add(file)
	if err != nil {
		return err
	}
	_, err = wt.Commit(commitMsg, &git.CommitOptions{
		All:       true,
		Author:    signature(),
		Committer: signature(),
	})
	if err != nil {
		return err
	}
	obj, err := c.Repo.CommitObjects()
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("commit object is null")
	}
	return nil
}

func (c *GitClient) Push() error {
	auth, err := c.auth()
	if err != nil {
		return err
	}
	err = c.Repo.Push(&git.PushOptions{
		RefSpecs: []gitconfig.RefSpec{
			gitconfig.RefSpec(c.ReferenceName + ":" + c.ReferenceName),
		},
		Force:    true,
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *GitClient) Tag(version string) error {
	reference, err := c.Repo.Storer.Reference(c.ReferenceName)
	if err != nil {
		return err
	}
	tag := object.Tag{
		Name:       version,
		Message:    "v" + version,
		Tagger:     *signature(),
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

	tagRefer := plumbing.NewHashReference(tagReferenceName, hash)
	err = c.Repo.Storer.SetReference(tagRefer)
	if err != nil {
		return err
	}

	auth, err := c.auth()
	if err != nil {
		return err
	}
	err = c.Repo.Push(&git.PushOptions{
		RefSpecs: []gitconfig.RefSpec{
			gitconfig.RefSpec(tagReferenceName + ":" + tagReferenceName),
		},
		Progress: ioutil.Discard,
		Auth:     auth,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *GitClient) CreateNewBranch(branchName string) error {
	newBranchName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName))

	wt, err := c.Repo.Worktree()
	if err != nil {
		return err
	}
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: newBranchName,
		Force:  true,
		Create: true,
		Keep:   true,
	})
	if err != nil {
		return err
	}
	auth, err := c.auth()
	if err != nil {
		return err
	}
	err = c.Repo.Push(&git.PushOptions{
		RefSpecs: []gitconfig.RefSpec{
			gitconfig.RefSpec(newBranchName + ":" + newBranchName),
		},
		Progress: ioutil.Discard,
		Auth:     auth,
		Force:    true,
	})
	if err != nil {
		return err
	}

	oldBranchName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", c.Branch))
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: oldBranchName,
		Force:  true,
		Create: false,
		Keep:   false,
	})
	if err != nil {
		return err
	}

	return nil
}
