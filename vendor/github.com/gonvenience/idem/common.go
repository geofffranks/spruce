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

func mapItemsToSlice[E any, S ~[]E, T any](slice S, fn func(e E) T) []T {
	ret := make([]T, len(slice))
	for i, e := range slice {
		ret[i] = fn(e)
	}
	return ret
}

func reject[E comparable, S ~[]E](slice S, elt E) (ret S, ok bool) {
	ret = make(S, 0, len(slice))
	for _, e := range slice {
		if elt == e {
			ok = true
		} else {
			ret = append(ret, e)
		}
	}
	return
}
