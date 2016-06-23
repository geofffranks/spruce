package spruce

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/jhunt/tree"
	"github.com/smallfish/simpleyaml"
)

func TestShouldReplaceArray(t *testing.T) {
	Convey("We should replace arrays", t, func() {
		Convey("If the element is a string with the right append token", func() {
			So(shouldReplaceArray([]interface{}{"(( replace ))", "stuff"}), ShouldBeTrue)
		})
		Convey("But not if the element is a string with the wrong token", func() {
			So(shouldReplaceArray([]interface{}{"not a magic token"}), ShouldBeFalse)
		})
		Convey("But not if the element is not a string", func() {
			So(shouldReplaceArray([]interface{}{42}), ShouldBeFalse)
		})
		Convey("But not if the slice has no elements", func() {
			So(shouldReplaceArray([]interface{}{}), ShouldBeFalse)
		})
		Convey("Is whitespace agnostic", func() {
			Convey("No surrounding whitespace", func() {
				yes := shouldReplaceArray([]interface{}{"((replace))"})
				So(yes, ShouldBeTrue)
			})
			Convey("Surrounding tabs", func() {
				yes := shouldReplaceArray([]interface{}{"((	replace	))"})
				So(yes, ShouldBeTrue)
			})
			Convey("Multiple surrounding whitespaces", func() {
				yes := shouldReplaceArray([]interface{}{"((  replace  ))"})
				So(yes, ShouldBeTrue)
			})
		})
	})
}
func TestShouldAppendToArray(t *testing.T) {
	Convey("We should append to arrays", t, func() {
		Convey("If the element is a string with the right append token", func() {
			So(shouldAppendToArray([]interface{}{"(( append ))", "stuff"}), ShouldBeTrue)
		})
		Convey("But not if the element is a string with the wrong token", func() {
			So(shouldAppendToArray([]interface{}{"not a magic token"}), ShouldBeFalse)
		})
		Convey("But not if the element is not a string", func() {
			So(shouldAppendToArray([]interface{}{42}), ShouldBeFalse)
		})
		Convey("But not if the slice has no elements", func() {
			So(shouldAppendToArray([]interface{}{}), ShouldBeFalse)
		})
		Convey("Is whitespace agnostic", func() {
			Convey("No surrounding whitespace", func() {
				yes := shouldAppendToArray([]interface{}{"((append))"})
				So(yes, ShouldBeTrue)
			})
			Convey("Surrounding tabs", func() {
				yes := shouldAppendToArray([]interface{}{"((	append	))"})
				So(yes, ShouldBeTrue)
			})
			Convey("Multiple surrounding whitespaces", func() {
				yes := shouldAppendToArray([]interface{}{"((  append  ))"})
				So(yes, ShouldBeTrue)
			})
		})
	})
}
func TestShouldPrependToArray(t *testing.T) {
	Convey("We should prepend to arrays", t, func() {
		Convey("If the element is a string with the right prepend token", func() {
			So(shouldPrependToArray([]interface{}{"(( prepend ))", "stuff"}), ShouldBeTrue)
		})
		Convey("But not if the element is a string with the wrong token", func() {
			So(shouldPrependToArray([]interface{}{"not a magic token"}), ShouldBeFalse)
		})
		Convey("But not if the element is not a string", func() {
			So(shouldPrependToArray([]interface{}{42}), ShouldBeFalse)
		})
		Convey("But not if the slice has no elements", func() {
			So(shouldInlineMergeArray([]interface{}{}), ShouldBeFalse)
		})
		Convey("Is whitespace agnostic", func() {
			Convey("No surrounding whitespace", func() {
				yes := shouldPrependToArray([]interface{}{"((prepend))"})
				So(yes, ShouldBeTrue)
			})
			Convey("Surrounding tabs", func() {
				yes := shouldPrependToArray([]interface{}{"((	prepend	))"})
				So(yes, ShouldBeTrue)
			})
			Convey("Multiple surrounding whitespaces", func() {
				yes := shouldPrependToArray([]interface{}{"((  prepend  ))"})
				So(yes, ShouldBeTrue)
			})
		})
	})
}
func TestShouldInlineMergeArray(t *testing.T) {
	Convey("We should inline merge arrays", t, func() {
		Convey("If the element is a string with the right inline-merge token", func() {
			So(shouldInlineMergeArray([]interface{}{"(( inline ))", "stuff"}), ShouldBeTrue)
		})
		Convey("But not if the element is a string with the wrong token", func() {
			So(shouldInlineMergeArray([]interface{}{"not a magic token"}), ShouldBeFalse)
		})
		Convey("But not if the element is not a string", func() {
			So(shouldInlineMergeArray([]interface{}{42}), ShouldBeFalse)
		})
		Convey("But not if the slice has no elements", func() {
			So(shouldInlineMergeArray([]interface{}{}), ShouldBeFalse)
		})
		Convey("Is whitespace agnostic", func() {
			Convey("No surrounding whitespace", func() {
				yes := shouldInlineMergeArray([]interface{}{"((inline))"})
				So(yes, ShouldBeTrue)
			})
			Convey("Surrounding tabs", func() {
				yes := shouldInlineMergeArray([]interface{}{"((	inline	))"})
				So(yes, ShouldBeTrue)
			})
			Convey("Multiple surrounding whitespaces", func() {
				yes := shouldInlineMergeArray([]interface{}{"((  inline  ))"})
				So(yes, ShouldBeTrue)
			})
		})
	})
}
func TestShouldKeyMergeArrayOfHashes(t *testing.T) {
	Convey("We should key-based merge arrays of hashes", t, func() {
		Convey("If the element is a string with the right key-merge token", func() {
			yes, key := shouldKeyMergeArray([]interface{}{"(( merge ))", "stuff"})
			So(yes, ShouldBeTrue)
			So(key, ShouldEqual, "name")
		})
		Convey("If the element is a string with the right key-merge token and custom key specified", func() {
			yes, key := shouldKeyMergeArray([]interface{}{"(( merge on id ))", "stuff"})
			So(yes, ShouldBeTrue)
			So(key, ShouldEqual, "id")
		})
		Convey("Is whitespace agnostic", func() {
			Convey("No surrounding whitespace", func() {
				yes, key := shouldKeyMergeArray([]interface{}{"((merge))"})
				So(yes, ShouldBeTrue)
				So(key, ShouldEqual, "name")
			})
			Convey("Surrounding tabs", func() {
				yes, key := shouldKeyMergeArray([]interface{}{"((	merge	))"})
				So(yes, ShouldBeTrue)
				So(key, ShouldEqual, "name")
			})
			Convey("Multiple surrounding whitespaces", func() {
				yes, key := shouldKeyMergeArray([]interface{}{"((  merge  ))"})
				So(yes, ShouldBeTrue)
				So(key, ShouldEqual, "name")
			})
			Convey("Multiple whitespace in key spec", func() {
				yes, key := shouldKeyMergeArray([]interface{}{"((  merge  on  id  ))"})
				So(yes, ShouldBeTrue)
				So(key, ShouldEqual, "id")
			})
			Convey("Tabs in key spec", func() {
				yes, key := shouldKeyMergeArray([]interface{}{"((  merge	on	id  ))"})
				So(yes, ShouldBeTrue)
				So(key, ShouldEqual, "id")
			})
			Convey("No trailing whitespace in key spec", func() {
				yes, key := shouldKeyMergeArray([]interface{}{"((  merge	on	id))"})
				So(yes, ShouldBeTrue)
				So(key, ShouldEqual, "id")
			})
			Convey("No whitespace in key spec", func() {
				yes, key := shouldKeyMergeArray([]interface{}{"((  mergeonid ))"})
				So(yes, ShouldBeFalse)
				So(key, ShouldEqual, "")
			})
			Convey("Only leading whitespace in key spec", func() {
				yes, key := shouldKeyMergeArray([]interface{}{"((  merge onid ))"})
				So(yes, ShouldBeFalse)
				So(key, ShouldEqual, "")
			})
		})
		Convey("But not if the element is a string with the wrong token", func() {
			yes, key := shouldKeyMergeArray([]interface{}{"not a magic token"})
			So(yes, ShouldBeFalse)
			So(key, ShouldEqual, "")
		})
		Convey("But not if the element is not a string", func() {
			yes, key := shouldKeyMergeArray([]interface{}{42})
			So(yes, ShouldBeFalse)
			So(key, ShouldEqual, "")
		})
		Convey("But not if the slice has no elements", func() {
			yes, key := shouldKeyMergeArray([]interface{}{})
			So(yes, ShouldBeFalse)
			So(key, ShouldEqual, "")
		})
	})
}

func TestShouldInsertIntoArrayBasedOnIndex(t *testing.T) {
	Convey("We should insert into arrays based on index", t, func() {
		Convey("If insert token with after and index is found", func() {
			result, idx := shouldInsertIntoArrayBasedOnIndex([]interface{}{"(( insert after 0 ))", "stuff"})
			So(result, ShouldBeTrue)
			So(idx, ShouldEqual, 1)
		})

		Convey("If insert token with before and index is found", func() {
			result, idx := shouldInsertIntoArrayBasedOnIndex([]interface{}{"(( insert before 0 ))", "stuff"})
			So(result, ShouldBeTrue)
			So(idx, ShouldEqual, 0)
		})

		Convey("If insert token with after and index is found independent of missing whitespaces", func() {
			result, idx := shouldInsertIntoArrayBasedOnIndex([]interface{}{"((insert after 0))", "stuff"})
			So(result, ShouldBeTrue)
			So(idx, ShouldEqual, 1)
		})

		Convey("If insert token with after and index is found with a lot of additional whitespaces", func() {
			result, idx := shouldInsertIntoArrayBasedOnIndex([]interface{}{"((  insert   after   0  ))", "stuff"})
			So(result, ShouldBeTrue)
			So(idx, ShouldEqual, 1)
		})

		Convey("But not if index is obviously out of bounds", func() {
			result, idx := shouldInsertIntoArrayBasedOnIndex([]interface{}{"(( insert before -1 ))", "stuff"})
			So(result, ShouldBeFalse)
			So(idx, ShouldEqual, -1)
		})

		Convey("But not if the magic token is not specified", func() {
			result, idx := shouldInsertIntoArrayBasedOnIndex([]interface{}{"not a magic token", "stuff"})
			So(result, ShouldBeFalse)
			So(idx, ShouldEqual, -1)
		})

		Convey("But not if the index is somehow quoted (and therefore not integer)", func() {
			result, idx := shouldInsertIntoArrayBasedOnIndex([]interface{}{"(( insert after \"0\" ))", "stuff"})
			So(result, ShouldBeFalse)
			So(idx, ShouldEqual, -1)
		})

		Convey("But not if there is actually a different type of position style is wanted", func() {
			result, idx := shouldInsertIntoArrayBasedOnIndex([]interface{}{"(( insert after \"foobar\" ))", "stuff"})
			So(result, ShouldBeFalse)
			So(idx, ShouldEqual, -1)
		})
	})
}

func TestShouldInsertIntoArrayBasedOnName(t *testing.T) {
	Convey("We should insert into arrays based on key and name", t, func() {
		Convey("If insert token with after, and insertion-name was found", func() {
			result, insertSlices := shouldInsertIntoArrayBasedOnName([]interface{}{"(( insert after \"nats\" ))", "stuff"})
			So(result, ShouldBeTrue)
			So(insertSlices, ShouldNotBeNil)
			So(len(insertSlices), ShouldEqual, 1)
			So(insertSlices[0].relative, ShouldEqual, "after")
			So(insertSlices[0].key, ShouldEqual, "name")
			So(insertSlices[0].name, ShouldEqual, "nats")
		})

		Convey("If insert token with after, key-name and insertion-name was found", func() {
			result, insertSlices := shouldInsertIntoArrayBasedOnName([]interface{}{"(( insert after on name \"nats\" ))", "stuff"})
			So(result, ShouldBeTrue)
			So(insertSlices, ShouldNotBeNil)
			So(len(insertSlices), ShouldEqual, 1)
			So(insertSlices[0].relative, ShouldEqual, "after")
			So(insertSlices[0].key, ShouldEqual, "name")
			So(insertSlices[0].name, ShouldEqual, "nats")
		})

		Convey("If insert token with before, another custom key-name and insertion-name was found", func() {
			result, insertSlices := shouldInsertIntoArrayBasedOnName([]interface{}{"(( insert before on id \"ccdb\" ))", "stuff"})
			So(result, ShouldBeTrue)
			So(insertSlices, ShouldNotBeNil)
			So(len(insertSlices), ShouldEqual, 1)
			So(insertSlices[0].relative, ShouldEqual, "before")
			So(insertSlices[0].key, ShouldEqual, "id")
			So(insertSlices[0].name, ShouldEqual, "ccdb")
		})

		Convey("If insert token with after, key-name and insertion-name was found without additional whitespaces", func() {
			result, insertSlices := shouldInsertIntoArrayBasedOnName([]interface{}{"((insert after on name \"nats\"))", "stuff"})
			So(result, ShouldBeTrue)
			So(insertSlices, ShouldNotBeNil)
			So(len(insertSlices), ShouldEqual, 1)
			So(insertSlices[0].relative, ShouldEqual, "after")
			So(insertSlices[0].key, ShouldEqual, "name")
			So(insertSlices[0].name, ShouldEqual, "nats")
		})

		Convey("If insert token with after, key-name and insertion-name was found with additional whitespaces", func() {
			result, insertSlices := shouldInsertIntoArrayBasedOnName([]interface{}{"((   insert   after  on    name   \"nats\"   ))", "stuff"})
			So(result, ShouldBeTrue)
			So(insertSlices, ShouldNotBeNil)
			So(len(insertSlices), ShouldEqual, 1)
			So(insertSlices[0].relative, ShouldEqual, "after")
			So(insertSlices[0].key, ShouldEqual, "name")
			So(insertSlices[0].name, ShouldEqual, "nats")
		})

		Convey("If insert token with after, key-name and insertion-name was found, but the list is empty", func() {
			result, insertSlices := shouldInsertIntoArrayBasedOnName([]interface{}{"(( insert after on name \"nats\" ))"})
			So(result, ShouldBeTrue)
			So(insertSlices, ShouldNotBeNil)
			So(len(insertSlices), ShouldEqual, 1)
			So(insertSlices[0].relative, ShouldEqual, "after")
			So(insertSlices[0].key, ShouldEqual, "name")
			So(insertSlices[0].name, ShouldEqual, "nats")
			So(insertSlices[0].list, ShouldResemble, []interface{}{})
		})

		Convey("If there are multiple insert token with after/before, different key names, and names (only technical usecase)", func() {
			result, insertSlices := shouldInsertIntoArrayBasedOnName([]interface{}{
				"(( insert after on name \"nats\" ))",
				"stuff1",
				"stuff2",
				"stuff3",
				"(( insert before on id \"consul\" ))",
				"stuffX1",
				"stuffX2",
			})
			So(result, ShouldBeTrue)
			So(insertSlices, ShouldNotBeNil)
			So(len(insertSlices), ShouldEqual, 2)
			So(insertSlices[0].relative, ShouldEqual, "after")
			So(insertSlices[0].key, ShouldEqual, "name")
			So(insertSlices[0].name, ShouldEqual, "nats")
			So(insertSlices[0].list, ShouldResemble, []interface{}{"stuff1", "stuff2", "stuff3"})
			So(insertSlices[1].relative, ShouldEqual, "before")
			So(insertSlices[1].key, ShouldEqual, "id")
			So(insertSlices[1].name, ShouldEqual, "consul")
			So(insertSlices[1].list, ShouldResemble, []interface{}{"stuffX1", "stuffX2"})
		})

		Convey("But not if the magic token is not specified", func() {
			result, insertSlices := shouldInsertIntoArrayBasedOnName([]interface{}{"not a magic token", "stuff"})
			So(result, ShouldBeFalse)
			So(insertSlices, ShouldBeNil)
		})
	})
}

func TestMergeObj(t *testing.T) {
	Convey("Passing a map to m.mergeObj merges as a map", t, func() {
		Convey("merges as a map under normal conditions", func() {
			orig := map[interface{}]interface{}{"first": 1, "second": 2}
			n := map[interface{}]interface{}{"second": 4, "third": 3}
			expect := map[interface{}]interface{}{"first": 1, "second": 4, "third": 3}

			m := &Merger{}
			o := m.mergeObj(orig, n, "node-path")
			err := m.Error()
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("returns an error when m.mergeMap throws an error", func() {
			origMap := map[interface{}]interface{}{
				"array": []interface{}{
					"string",
				},
			}
			newMap := map[interface{}]interface{}{
				"array": []interface{}{
					"(( merge on name ))",
					"string",
				},
			}
			m := &Merger{}
			m.mergeObj(origMap, newMap, "node-path")
			err := m.Error()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "node-path.array.0: new object is a string, not a map - cannot merge using keys")
		})
		Convey("returns an error for any (( merge ... )) operators found in non-list context", func() {
			orig := map[interface{}]interface{}{}
			n := map[interface{}]interface{}{"map": "(( merge || nil ))"}
			m := &Merger{}
			m.mergeObj(orig, n, "node-path")
			err := m.Error()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "node-path.map: inappropriate use of (( merge )) operator outside of a list (this is spruce, after all)")
		})
	})
	Convey("Passing a slice to m.mergeObj", t, func() {
		Convey("without magical merge token replaces entire array", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"my", "new", "array"}
			expect := []interface{}{"my", "new", "array"}

			m := &Merger{}
			o := m.mergeObj(orig, array, "node-path")
			err := m.Error()
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("When passed a slice, but original item is nil", func() {
			val := []interface{}{"(( append ))", "two"}
			expect := []interface{}{"two"}

			m := &Merger{}
			o := m.mergeObj(nil, val, "node-path")
			err := m.Error()
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("Returns an error when m.mergeArray throws an error, when merging array to array", func() {
			orig := []interface{}{
				"first",
			}
			array := []interface{}{
				"(( merge on name ))",
				"second",
			}

			m := &Merger{}
			m.mergeObj(orig, array, "node-path")
			err := m.Error()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "node-path.0: new object is a string, not a map - cannot merge using keys")
		})
		Convey("Returns an error when m.mergeArray throws an error, when merging nil to array", func() {
			array := []interface{}{
				"(( merge on name ))",
				"second",
			}

			m := &Merger{}
			m.mergeObj(nil, array, "node-path")
			err := m.Error()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "node-path.0: new object is a string, not a map - cannot merge using keys")
		})
	})
	Convey("m.mergeObj merges in place", t, func() {
		Convey("When passed a string", func() {
			orig := 42
			val := "asdf"

			m := &Merger{}
			o := m.mergeObj(orig, val, "node-path")
			err := m.Error()
			So(o, ShouldEqual, "asdf")
			So(err, ShouldBeNil)
		})
		Convey("When passed an int", func() {
			orig := "fdsa"
			val := 10

			m := &Merger{}
			o := m.mergeObj(orig, val, "node-path")
			err := m.Error()
			So(o, ShouldEqual, 10)
			So(err, ShouldBeNil)
		})
		Convey("When passed an float64", func() {
			orig := "fdsa"
			val := 10.4

			m := &Merger{}
			o := m.mergeObj(orig, val, "node-path")
			err := m.Error()
			So(o, ShouldEqual, 10.4)
			So(err, ShouldBeNil)
		})
		Convey("When passed nil", func() {
			orig := "fdsa"
			val := interface{}(nil)

			m := &Merger{}
			o := m.mergeObj(orig, val, "node-path")
			err := m.Error()
			So(o, ShouldBeNil)
			So(err, ShouldBeNil)
		})
		Convey("When passed a map, but original item is a scalar", func() {
			orig := "value"
			val := map[interface{}]interface{}{"key": "value"}
			expect := map[interface{}]interface{}{"key": "value"}

			m := &Merger{}
			o := m.mergeObj(orig, val, "node-path")
			err := m.Error()
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("When passed a map, but original item is nil", func() {
			val := map[interface{}]interface{}{"key": "value"}
			expect := map[interface{}]interface{}{"key": "value"}

			m := &Merger{}
			o := m.mergeObj(nil, val, "node-path")
			err := m.Error()
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("When passed a slice, but original item is a scalar", func() {
			orig := "value"
			val := []interface{}{"one", "two"}
			expect := []interface{}{"one", "two"}

			m := &Merger{}
			o := m.mergeObj(orig, val, "node-path")
			err := m.Error()
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
	})
}

func TestMergeMap(t *testing.T) {
	Convey("with map elements updates original map", t, func() {
		origMap := map[interface{}]interface{}{"k1": "v1", "k2": "v2"}
		newMap := map[interface{}]interface{}{"k3": "v3", "k2": "v2.new"}
		expectMap := map[interface{}]interface{}{"k2": "v2.new", "k3": "v3", "k1": "v1"}

		m := &Merger{}
		m.mergeMap(origMap, newMap, "node-path")
		So(origMap, ShouldResemble, expectMap)
		So(m.Error(), ShouldBeNil)
	})
	Convey("m.mergeMap re-throws an error if it finds one while merging data", t, func() {
		origMap := map[interface{}]interface{}{
			"array": []interface{}{
				"string",
			},
		}
		newMap := map[interface{}]interface{}{
			"array": []interface{}{
				"(( merge on name ))",
				"string",
			},
		}
		m := &Merger{}
		m.mergeMap(origMap, newMap, "node-path")
		err := m.Error()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "node-path.array.0: new object is a string, not a map - cannot merge using keys")
	})
}

func TestMergeArray(t *testing.T) {
	Convey("m.mergeArray", t, func() {
		Convey("with initial element '(( prepend ))' prepends new data", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"(( prepend ))", "zeroth"}
			expect := []interface{}{"zeroth", "first", "second"}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			So(a, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("with initial element '(( append ))' appends new data", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"(( append ))", "third"}
			expect := []interface{}{"first", "second", "third"}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			So(a, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("with initial element '(( inline ))'", func() {
			Convey("and len(orig) == len(new)", func() {
				orig := []interface{}{
					"orig_first",
					[]interface{}{"subfirst", "subsecond"},
					map[interface{}]interface{}{
						"name": "original",
						"val":  "original",
					},
					"orig_last",
				}
				array := []interface{}{
					"(( inline ))",
					"overwritten_first",
					[]interface{}{"(( append ))", "subthird"},
					map[interface{}]interface{}{
						"val": "overwritten",
					},
					"overwritten_last",
				}
				expect := []interface{}{
					"overwritten_first",
					[]interface{}{"subfirst", "subsecond", "subthird"},
					map[interface{}]interface{}{
						"name": "original",
						"val":  "overwritten",
					},
					"overwritten_last",
				}
				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})
			Convey("and len(orig) > len(new)", func() {
				orig := []interface{}{
					"orig_first",
					[]interface{}{"subfirst", "subsecond"},
					map[interface{}]interface{}{
						"name": "original",
						"val":  "original",
					},
					"orig_last",
				}
				array := []interface{}{
					"(( inline ))",
					"overwritten_first",
					[]interface{}{"(( append ))", "subthird"},
					map[interface{}]interface{}{
						"val": "overwritten",
					},
				}
				expect := []interface{}{
					"overwritten_first",
					[]interface{}{"subfirst", "subsecond", "subthird"},
					map[interface{}]interface{}{
						"name": "original",
						"val":  "overwritten",
					},
					"orig_last",
				}
				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})
			Convey("and len(orig < len(new)", func() {
				orig := []interface{}{
					"orig_first",
					[]interface{}{"subfirst", "subsecond"},
				}
				array := []interface{}{
					"(( inline ))",
					"overwritten_first",
					[]interface{}{"(( append ))", "subthird"},
					map[interface{}]interface{}{
						"val": "overwritten",
					},
					"overwritten_last",
				}
				expect := []interface{}{
					"overwritten_first",
					[]interface{}{"subfirst", "subsecond", "subthird"},
					map[interface{}]interface{}{
						"val": "overwritten",
					},
					"overwritten_last",
				}
				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})
			Convey("and empty orig slice", func() {
				orig := []interface{}{}
				array := []interface{}{
					"(( inline ))",
					"overwritten_first",
					[]interface{}{"(( append ))", "subthird"},
					map[interface{}]interface{}{
						"val": "overwritten",
					},
					"overwritten_last",
				}
				expect := []interface{}{
					"overwritten_first",
					[]interface{}{"subthird"},
					map[interface{}]interface{}{
						"val": "overwritten",
					},
					"overwritten_last",
				}
				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})
			Convey("returns error when mergObj returns error (merged data)", func() {
				m := &Merger{}
				m.mergeArray(
					// old
					[]interface{}{
						[]interface{}{
							"string",
						},
					},

					// new
					[]interface{}{
						"(( inline ))",
						[]interface{}{
							"(( merge ))",
							"string",
						},
					}, "node-path")

				err := m.Error()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "node-path.0.0: new object is a string, not a map - cannot merge using keys")
			})
			Convey("returns error when mergObj returns error (appended data)", func() {
				m := &Merger{}
				m.mergeArray(
					// old
					[]interface{}{}, // empty array

					// new
					[]interface{}{
						"(( inline ))",
						[]interface{}{
							"(( merge ))",
							"string", // this is not a map
						},
					}, "node-path")

				err := m.Error()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "node-path.0.0: new object is a string, not a map - cannot merge using keys")
			})
		})
		Convey("with initial element '(( merge ))'", func() {
			Convey("merges and defaults to 'name' for a key, if not specified", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"name": "job1", "id": "1", "org": "org1"},
					map[interface{}]interface{}{"name": "job3", "id": "3", "org": "org3"},
					map[interface{}]interface{}{"name": "job2", "id": "2", "org": "org2"},
				}
				array := []interface{}{
					"(( merge ))",
					map[interface{}]interface{}{"name": "job2", "org": "myorg2"},
					map[interface{}]interface{}{"name": "job4", "org": "myorg4"},
					map[interface{}]interface{}{"name": "job1", "org": "myorg1"},
				}
				expect := []interface{}{
					map[interface{}]interface{}{"name": "job1", "id": "1", "org": "myorg1"},
					map[interface{}]interface{}{"name": "job3", "id": "3", "org": "org3"},
					map[interface{}]interface{}{"name": "job2", "id": "2", "org": "myorg2"},
					map[interface{}]interface{}{"name": "job4", "org": "myorg4"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})
			Convey("allows custom key merging to be specified", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"id": "1", "org": "org1"},
					map[interface{}]interface{}{"id": "3", "org": "org3"},
					map[interface{}]interface{}{"id": "2", "org": "org2"},
				}
				array := []interface{}{
					"(( merge on id ))",
					map[interface{}]interface{}{"id": "2", "org": "myorg2"},
					map[interface{}]interface{}{"id": "4", "org": "myorg4"},
					map[interface{}]interface{}{"id": "1", "org": "myorg1"},
				}
				expect := []interface{}{
					map[interface{}]interface{}{"id": "1", "org": "myorg1"},
					map[interface{}]interface{}{"id": "3", "org": "org3"},
					map[interface{}]interface{}{"id": "2", "org": "myorg2"},
					map[interface{}]interface{}{"id": "4", "org": "myorg4"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})
			Convey("But not if any of the original array elements are not maps", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"id": "1", "org": "org1"},
					map[interface{}]interface{}{"id": "3", "org": "org3"},
					map[interface{}]interface{}{"id": "2", "org": "org2"},
					"this will make it fail",
				}
				array := []interface{}{
					"(( merge on id ))",
					map[interface{}]interface{}{"id": "2", "org": "myorg2"},
					map[interface{}]interface{}{"id": "4", "org": "myorg4"},
					map[interface{}]interface{}{"id": "1", "org": "myorg1"},
				}

				m := &Merger{}
				m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "node-path.3: original object is a string, not a map - cannot merge using keys")
			})
			Convey("But not if any of the new array elements are not maps", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"id": "1", "org": "org1"},
					map[interface{}]interface{}{"id": "3", "org": "org3"},
					map[interface{}]interface{}{"id": "2", "org": "org2"},
				}
				array := []interface{}{
					"(( merge on id ))",
					map[interface{}]interface{}{"id": "2", "org": "myorg2"},
					map[interface{}]interface{}{"id": "4", "org": "myorg4"},
					map[interface{}]interface{}{"id": "1", "org": "myorg1"},
					"this will make it fail",
				}

				m := &Merger{}
				m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "node-path.3: new object is a string, not a map - cannot merge using keys")
			})
			Convey("But not if any of the elements of the original array don't have the key requested", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"id": "1", "org": "org1"},
					map[interface{}]interface{}{"id": "3", "org": "org3"},
					map[interface{}]interface{}{"name": "2", "org": "org2"},
				}
				array := []interface{}{
					"(( merge on id ))",
					map[interface{}]interface{}{"id": "2", "org": "myorg2"},
					map[interface{}]interface{}{"id": "4", "org": "myorg4"},
					map[interface{}]interface{}{"id": "1", "org": "myorg1"},
				}

				m := &Merger{}
				m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(err, ShouldNotBeNil)
			})
			Convey("But not if any of the elements of the new array don't have the key requested", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"id": "1", "org": "org1"},
					map[interface{}]interface{}{"id": "3", "org": "org3"},
					map[interface{}]interface{}{"id": "2", "org": "org2"},
				}
				array := []interface{}{
					"(( merge on id ))",
					map[interface{}]interface{}{"id": "2", "org": "myorg2"},
					map[interface{}]interface{}{"id": "4", "org": "myorg4"},
					map[interface{}]interface{}{"name": "1", "org": "myorg1"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "node-path.2: new object does not contain the key 'id' - cannot merge")
			})
			Convey("Returns an error if m.mergeObj() returns an error", func() {
				orig := []interface{}{
					map[interface{}]interface{}{
						"name": "first",
						"val": []interface{}{
							"string",
						},
					},
				}
				array := []interface{}{
					"(( merge ))",
					map[interface{}]interface{}{
						"name": "first",
						"val": []interface{}{
							"(( merge ))",
							"string",
						},
					},
				}
				m := &Merger{}
				m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "node-path.0.val.0: new object is a string, not a map - cannot merge using keys")
			})
		})
		Convey("arrays of maps can be merged inline", func() {
			origMapSlice := map[interface{}]interface{}{"k1": "v1", "k2": "v2"}
			newMapSlice := map[interface{}]interface{}{"k3": "v3", "k2": "v2.new"}
			expectMapSlice := map[interface{}]interface{}{"k1": "v1", "k2": "v2.new", "k3": "v3"}
			orig := []interface{}{origMapSlice}
			array := []interface{}{newMapSlice}
			expect := []interface{}{expectMapSlice}

			m := &Merger{}
			o := m.mergeArrayInline(orig, array, "node-path")
			err := m.Error()
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("merges arrays of maps by default", func() {
			origMapSlice := map[interface{}]interface{}{"k1": "v1", "k2": "v2"}
			newMapSlice := map[interface{}]interface{}{"k3": "v3", "k2": "v2.new"}
			expectMapSlice := map[interface{}]interface{}{"k1": "v1", "k2": "v2.new", "k3": "v3"}
			orig := []interface{}{origMapSlice}
			array := []interface{}{newMapSlice}
			expect := []interface{}{expectMapSlice}

			m := &Merger{}
			o := m.mergeObj(orig, array, "node-path")
			err := m.Error()
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("uses key-merge if possible", func() {
			first := []interface{}{
				map[interface{}]interface{}{
					"name": "first",
					"k1":   "v1",
				},
				map[interface{}]interface{}{
					"name": "second",
					"done": "yes",
				},
			}
			second := []interface{}{
				map[interface{}]interface{}{
					"name": "second",
					"2":    "best",
					"test": "test",
				},
				map[interface{}]interface{}{
					"name": "first",
					"k1":   "1",
					"k2":   "2",
				},
			}

			m := &Merger{}
			o := m.mergeObj(first, second, "an.inlined.merge")
			err := m.Error()
			So(o, ShouldResemble, []interface{}{
				map[interface{}]interface{}{
					"name": "first",
					"k1":   "1",
					"k2":   "2",
				},
				map[interface{}]interface{}{
					"name": "second",
					"2":    "best",
					"done": "yes",
					"test": "test",
				},
			})
			So(err, ShouldBeNil)
		})
		Convey("without magical merge token replaces entire array", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"my", "new", "array"}
			expect := []interface{}{"my", "new", "array"}

			m := &Merger{}
			o := m.mergeObj(orig, array, "node-path")
			err := m.Error()
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})

		Convey("with element '(( insert ))' inserts new data where wanted", func() {
			Convey("After #0 put the new entry", func() {
				orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
				array := []interface{}{"(( insert after 0 ))", "new-kid-on-the-block"}
				expect := []interface{}{"first", "new-kid-on-the-block", "second", "third", "fourth", "fifth", "sixth"}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})

			Convey("Before #0 put the new entry", func() {
				orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
				array := []interface{}{"(( insert before 0 ))", "new-kid-on-the-block"}
				expect := []interface{}{"new-kid-on-the-block", "first", "second", "third", "fourth", "fifth", "sixth"}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})

			Convey("After #4 put the new entry", func() {
				orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
				array := []interface{}{"(( insert after 4 ))", "new-kid-on-the-block"}
				expect := []interface{}{"first", "second", "third", "fourth", "fifth", "new-kid-on-the-block", "sixth"}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})

			Convey("After #5 put the new entry", func() {
				orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
				array := []interface{}{"(( insert after 5 ))", "new-kid-on-the-block"}
				expect := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth", "new-kid-on-the-block"}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})

			Convey("throw an error if insertion point is out of bounds", func() {
				orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
				array := []interface{}{"(( insert after 6 ))", "new-kid-on-the-block"}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "specified insertion index 7 is out of bounds")
			})

			Convey("After '<default>: first' put the new entry", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"name": "first", "release": "v1"},
					map[interface{}]interface{}{"name": "second", "release": "v1"},
				}

				array := []interface{}{
					"(( insert after \"first\" ))",
					map[interface{}]interface{}{"name": "new-kid-on-the-block", "release": "vNext"},
				}

				expect := []interface{}{
					map[interface{}]interface{}{"name": "first", "release": "v1"},
					map[interface{}]interface{}{"name": "new-kid-on-the-block", "release": "vNext"},
					map[interface{}]interface{}{"name": "second", "release": "v1"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})

			Convey("Before '<default>: first' put the new entry", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"name": "first", "release": "v1"},
					map[interface{}]interface{}{"name": "second", "release": "v1"},
				}

				array := []interface{}{
					"(( insert before \"first\" ))",
					map[interface{}]interface{}{"name": "new-kid-on-the-block", "release": "vNext"},
				}

				expect := []interface{}{
					map[interface{}]interface{}{"name": "new-kid-on-the-block", "release": "vNext"},
					map[interface{}]interface{}{"name": "first", "release": "v1"},
					map[interface{}]interface{}{"name": "second", "release": "v1"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})

			Convey("After 'id: second' put the new entry", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"id": "first", "release": "v1"},
					map[interface{}]interface{}{"id": "second", "release": "v1"},
				}

				array := []interface{}{
					"(( insert after on id \"second\" ))",
					map[interface{}]interface{}{"id": "new-kid-on-the-block", "release": "vNext"},
				}

				expect := []interface{}{
					map[interface{}]interface{}{"id": "first", "release": "v1"},
					map[interface{}]interface{}{"id": "second", "release": "v1"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block", "release": "vNext"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})

			Convey("Before 'id: second' put the new entry", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"id": "first", "release": "v1"},
					map[interface{}]interface{}{"id": "second", "release": "v1"},
				}

				array := []interface{}{
					"(( insert before on id \"second\" ))",
					map[interface{}]interface{}{"id": "new-kid-on-the-block", "release": "vNext"},
				}

				expect := []interface{}{
					map[interface{}]interface{}{"id": "first", "release": "v1"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block", "release": "vNext"},
					map[interface{}]interface{}{"id": "second", "release": "v1"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})

			Convey("Insert multiple new entries before second and after first", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"id": "first", "release": "v1"},
					map[interface{}]interface{}{"id": "second", "release": "v1"},
				}

				array := []interface{}{
					"(( insert before on id \"second\" ))",
					map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
					"(( insert after on id \"first\" ))",
					map[interface{}]interface{}{"id": "new-kid-on-the-block-3", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-4", "release": "vNext"},
				}

				expect := []interface{}{
					map[interface{}]interface{}{"id": "first", "release": "v1"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-3", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-4", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
					map[interface{}]interface{}{"id": "second", "release": "v1"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})

			Convey("Insert multiple new entries in two batches after first", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"id": "first", "release": "v1"},
					map[interface{}]interface{}{"id": "second", "release": "v1"},
				}

				array := []interface{}{
					"(( insert after on id \"first\" ))",
					map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
					"(( insert after on id \"first\" ))",
					map[interface{}]interface{}{"id": "new-kid-on-the-block-3", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-4", "release": "vNext"},
				}

				expect := []interface{}{
					map[interface{}]interface{}{"id": "first", "release": "v1"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-3", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-4", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
					map[interface{}]interface{}{"id": "second", "release": "v1"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})

			Convey("Insert multiple new entries in three batches after first", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"id": "first", "release": "v1"},
					map[interface{}]interface{}{"id": "second", "release": "v1"},
				}

				array := []interface{}{
					"(( insert after on id \"first\" ))",
					map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
					"(( insert after on id \"first\" ))",
					map[interface{}]interface{}{"id": "new-kid-on-the-block-3", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-4", "release": "vNext"},
					"(( insert after on id \"first\" ))",
					map[interface{}]interface{}{"id": "new-kid-on-the-block-5", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-6", "release": "vNext"},
				}

				expect := []interface{}{
					map[interface{}]interface{}{"id": "first", "release": "v1"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-5", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-6", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-3", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-4", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
					map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
					map[interface{}]interface{}{"id": "second", "release": "v1"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})

			Convey("throw an error when insertion point cannot be found", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"name": "first", "release": "v1"},
					map[interface{}]interface{}{"name": "second", "release": "v1"},
				}

				array := []interface{}{
					"(( insert after on name \"not-existing\" ))",
					map[interface{}]interface{}{"name": "new-kid-on-the-block", "release": "vNext"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "unable to find specified insertion point")
			})

			Convey("throw an error when key cannot be found in new list", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"id": "first", "release": "v1"},
					map[interface{}]interface{}{"id": "second", "release": "v1"},
				}

				array := []interface{}{
					"(( insert after on id \"second\" ))",
					map[interface{}]interface{}{"name": "new-kid-on-the-block", "release": "vNext"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "new object does not contain the key")
			})

			Convey("throw an error when key cannot be found in original list", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"name": "first", "release": "v1"},
					map[interface{}]interface{}{"name": "second", "release": "v1"},
				}

				array := []interface{}{
					"(( insert after on id \"second\" ))",
					map[interface{}]interface{}{"id": "new-kid-on-the-block", "release": "vNext"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "original object does not contain the key")
			})

			Convey("throw an error when entry is already in target list", func() {
				orig := []interface{}{
					map[interface{}]interface{}{"name": "first", "release": "v1"},
					map[interface{}]interface{}{"name": "second", "release": "v1"},
				}

				array := []interface{}{
					"(( insert after on name \"second\" ))",
					map[interface{}]interface{}{"name": "new-kid-on-the-block", "release": "vNext"},
					map[interface{}]interface{}{"name": "second", "release": "vNext"},
				}

				m := &Merger{}
				a := m.mergeArray(orig, array, "node-path")
				err := m.Error()
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "because new list entry 'name: second' is detected multiple times")
			})
		})
	})
}

func TestMerge(t *testing.T) {
	YAML := func(s string) map[interface{}]interface{} {
		y, err := simpleyaml.NewYaml([]byte(s))
		So(err, ShouldBeNil)

		data, err := y.Map()
		So(err, ShouldBeNil)

		return data
	}

	valueIs := func(t interface{}, path string, expect string) {
		c, err := tree.ParseCursor(path)
		So(err, ShouldBeNil)

		v, err := c.ResolveString(t)
		So(err, ShouldBeNil)

		So(v, ShouldEqual, expect)
	}
	notPresent := func(t interface{}, path string) {
		c, err := tree.ParseCursor(path)
		So(err, ShouldBeNil)

		_, err = c.ResolveString(t)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "could not be found")
	}

	Convey("Merge()", t, func() {
		Convey("leaves original object untouched when merging", func() {
			template := YAML(`props:
  toplevel: TEMPLATE VALUE
  sub:
    key: ANOTHER TEMPLATE VALUE
`)
			other := YAML(`props:
  toplevel: override
`)

			merged, err := Merge(template, other)
			So(err, ShouldBeNil)

			valueIs(template, "props.toplevel", "TEMPLATE VALUE")
			valueIs(template, "props.sub.key", "ANOTHER TEMPLATE VALUE")

			valueIs(other, "props.toplevel", "override")
			notPresent(other, "props.sub.key")

			valueIs(merged, "props.toplevel", "override")
			valueIs(merged, "props.sub.key", "ANOTHER TEMPLATE VALUE")
		})
	})

	Convey("Merge() handles array operators when previous data was nil", t, func() {
		template := YAML(`
array:
- name: nested
  attrs:
    append:
    - (( append ))
    - two
    - three
    prepend:
    - (( prepend ))
    - two
    - three
    inline:
    - (( inline ))
    - two
    - three
    replace:
    - (( replace ))
    - two
    - three
nested:
  append:
  - (( append ))
  - two
  - three
  prepend:
  - (( prepend ))
  - two
  - three
  inline:
  - (( inline ))
  - two
  - three
  replace:
  - (( replace ))
  - two
  - three

top_append:
- (( append ))
- b
- c
top_prepend:
- (( prepend ))
- b
- c
top_inline:
- (( inline ))
- b
- c
top_replace:
- (( replace ))
- b
- c
`)

		orig := map[interface{}]interface{}{}
		merged, err := Merge(orig, template)
		So(err, ShouldBeNil)

		valueIs(merged, "array.nested.attrs.append.0", "two")
		valueIs(merged, "nested.append.0", "two")
		valueIs(merged, "top_append.0", "b")
		valueIs(merged, "array.nested.attrs.prepend.0", "two")
		valueIs(merged, "nested.prepend.0", "two")
		valueIs(merged, "top_prepend.0", "b")
		valueIs(merged, "array.nested.attrs.inline.0", "two")
		valueIs(merged, "nested.inline.0", "two")
		valueIs(merged, "top_inline.0", "b")
		valueIs(merged, "array.nested.attrs.replace.0", "two")
		valueIs(merged, "nested.replace.0", "two")
		valueIs(merged, "top_replace.0", "b")

	})
}
