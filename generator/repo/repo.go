// Copyright 2016 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package repo deals with a Github repository of generators.
package repo

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/go-github/github"
	"github.com/stratumn/go/generator"
)

const (
	// StatesDir is the name of the states directory.
	StatesDir = "states"

	// StateFile is the name of the state file.
	StateFile = "repo.json"

	// StateDirPerm is the file mode for a state directory.
	StateDirPerm = 0755

	// StateFilePerm is the file mode for a state file.
	StateFilePerm = 0644

	// SrcDir is the name of the directory where sources are stored.
	SrcDir = "src"

	// SrcPerm is the file mode for a state directory.
	SrcPerm = 0755
)

// State stateibes a repository.
type State struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	Ref   string `json:"ref"`
	SHA1  string `json:"sha1"`
}

// Repo manages a Github repository.
type Repo struct {
	path   string
	owner  string
	repo   string
	client *github.Client
}

// New instantiates a repository.
func New(path, owner, repo string) *Repo {
	return &Repo{
		path:   path,
		owner:  owner,
		repo:   repo,
		client: github.NewClient(nil),
	}
}

// Update download the latest release if needed.
// Ref can be branch, a tag, or a commit SHA1.
func (r *Repo) Update(ref string) (*State, bool, error) {
	state, err := r.GetState(ref)
	if err != nil {
		return nil, false, err
	}

	sha1 := ""
	if state != nil {
		sha1 = state.SHA1
	}

	sha1, res, err := r.client.Repositories.GetCommitSHA1(r.owner, r.repo, ref, sha1)
	if res != nil {
		defer res.Body.Close()
		if res.StatusCode == http.StatusNotModified {
			// No update is available.
			return state, false, nil
		}
	}
	if err != nil {
		return nil, false, err
	}

	state, err = r.download(ref, sha1)
	if err != nil {
		return nil, false, err
	}

	path := filepath.Join(r.path, StatesDir, ref, StateFile)
	if err := os.MkdirAll(filepath.Dir(path), StateDirPerm); err != nil {
		return nil, false, err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, StateFilePerm)
	if err != nil {
		return nil, false, err
	}

	enc := json.NewEncoder(f)
	if err := enc.Encode(state); err != nil {
		return nil, false, err
	}

	return state, true, nil
}

// GetState returns the state of the repository.
// Ref can be branch, a tag, or a commit SHA1.
// If the repository does not exist, it returns nil.
func (r *Repo) GetState(ref string) (*State, error) {
	path := filepath.Join(r.path, StatesDir, ref, StateFile)
	var state *State
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	state = &State{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&state); err != nil {
		return nil, err
	}
	return state, err
}

// GetStateOrCreate returns the state of the repository.
// If the repository does not exist, it returns creates it by calling Update().
// Ref can be branch, a tag, or a commit SHA1.
func (r *Repo) GetStateOrCreate(ref string) (*State, error) {
	state, err := r.GetState(ref)
	if err != nil {
		return nil, err
	}
	if state == nil {
		if state, _, err = r.Update(ref); err != nil {
			return nil, err
		}
	}
	return state, nil
}

// List lists the generators of the repository.
// Ref can be branch, a tag, or a commit SHA1.
func (r *Repo) List(ref string) ([]*generator.Definition, error) {
	_, err := r.GetStateOrCreate(ref)
	if err != nil {
		return nil, err
	}

	matches, err := filepath.Glob(filepath.Join(r.path, SrcDir, ref, "*", generator.DefinitionFile))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)

	var defs []*generator.Definition
	for _, p := range matches {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		var def generator.Definition
		if err = dec.Decode(&def); err != nil {
			return nil, err
		}
		defs = append(defs, &def)
	}

	return defs, nil
}

// Generate executes a generator by name.
// Ref can be branch, a tag, or a commit SHA1.
func (r *Repo) Generate(name, dst string, opts *generator.Options, ref string) error {
	_, err := r.GetStateOrCreate(ref)
	if err != nil {
		return err
	}

	matches, err := filepath.Glob(filepath.Join(r.path, SrcDir, ref, "*", generator.DefinitionFile))
	if err != nil {
		return err
	}
	sort.Strings(matches)

	for _, p := range matches {
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		var def generator.Definition
		if err = dec.Decode(&def); err != nil {
			return err
		}

		if def.Name == name {
			gen, err := generator.NewFromDir(filepath.Dir(p), opts)
			if err != nil {
				return err
			}
			return gen.Exec(dst)
		}
	}

	return fmt.Errorf("could not find generator %q", name)
}

func (r *Repo) download(ref, sha1 string) (*State, error) {
	opts := github.RepositoryContentGetOptions{Ref: sha1}
	url, ghres, err := r.client.Repositories.GetArchiveLink(r.owner, r.repo, github.Tarball, &opts)
	if err != nil {
		return nil, err
	}
	defer ghres.Body.Close()

	res, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	gr, err := gzip.NewReader(res.Body)
	if err != nil {
		return nil, err
	}

	if err := os.RemoveAll(filepath.Join(r.path, SrcDir, ref)); err != nil {
		return nil, err
	}

	tr := tar.NewReader(gr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if hdr.Typeflag == tar.TypeReg {
			parts := strings.Split(hdr.Name, "/")
			parts = parts[1:]
			dst := filepath.Join(r.path, SrcDir, ref, filepath.Join(parts...))
			err = os.MkdirAll(filepath.Dir(dst), SrcPerm)
			if err != nil {
				return nil, err
			}
			mode := hdr.FileInfo()
			f, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Mode())
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(f, tr)
			if err != nil {
				return nil, err
			}
		}
	}

	return &State{
		Owner: r.owner,
		Repo:  r.repo,
		Ref:   ref,
		SHA1:  sha1,
	}, nil
}