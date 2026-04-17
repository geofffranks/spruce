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
	"errors"
)

type ModifiedPair struct {
	From *RenameCandidate
	To   *RenameCandidate
}

type DocumentChanges struct {
	DeletedList []*RenameCandidate
	AddedList   []*RenameCandidate

	modifiedPairs []ModifiedPair
}

var _ Changes = &DocumentChanges{}

func NewDocumentChanges(deleted []*RenameCandidate, added []*RenameCandidate) *DocumentChanges {
	return &DocumentChanges{
		DeletedList: deleted,
		AddedList:   added,
	}
}

func (d *DocumentChanges) Deleted() []file {
	return mapItemsToSlice(d.DeletedList, func(r *RenameCandidate) file { return r })
}

func (d *DocumentChanges) Added() []file {
	return mapItemsToSlice(d.AddedList, func(r *RenameCandidate) file { return r })
}

func (d *DocumentChanges) ModifiedPairs() []ModifiedPair {
	return d.modifiedPairs
}

func (d *DocumentChanges) MarkAsRename(deleted, added file) error {
	var ok bool
	d.DeletedList, ok = reject(d.DeletedList, deleted.(*RenameCandidate))
	if !ok {
		return errors.New("deleted element not found")
	}
	d.AddedList, ok = reject(d.AddedList, added.(*RenameCandidate))
	if !ok {
		return errors.New("added element not found")
	}
	d.modifiedPairs = append(d.modifiedPairs, ModifiedPair{
		From: deleted.(*RenameCandidate),
		To:   added.(*RenameCandidate),
	})
	return nil
}
