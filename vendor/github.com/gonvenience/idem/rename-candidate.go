// Copyright 2025 The Homeport Team
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

package idem

import (
	"bytes"
	"io"

	"github.com/gonvenience/ytbx"
	yamlv3 "go.yaml.in/yaml/v3"
)

var _ file = &RenameCandidate{}

type RenameCandidate struct {
	name string
	Path *ytbx.Path
	Doc  *yamlv3.Node

	content []byte
}

func NewRenameCandidate(name string, path *ytbx.Path, doc *yamlv3.Node) *RenameCandidate {
	return &RenameCandidate{
		name: name,
		Path: path,
		Doc:  doc,
	}
}

func (r *RenameCandidate) Name() string {
	return r.name
}

func (r *RenameCandidate) Reader() (io.ReadCloser, error) {
	if r.content == nil {
		if err := r.marshal(); err != nil {
			return nil, err
		}
	}
	return io.NopCloser(bytes.NewReader(r.content)), nil
}

func (r *RenameCandidate) Size() (int64, error) {
	if r.content == nil {
		if err := r.marshal(); err != nil {
			return 0, err
		}
	}
	return int64(len(r.content)), nil
}

func (r *RenameCandidate) marshal() error {
	var err error
	r.content, err = yamlv3.Marshal(r.Doc)
	return err
}
