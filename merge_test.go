package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
			So(shouldInlineMergeArray([]interface{}{}), ShouldBeFalse)
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

func TestMergeObj(t *testing.T) {
	Convey("Passing a map to mergeObj merges as a map", t, func() {
		Convey("merges as a map under normal conditions", func() {
			orig := map[interface{}]interface{}{"first": 1, "second": 2}
			n := map[interface{}]interface{}{"second": 4, "third": 3}
			expect := map[interface{}]interface{}{"first": 1, "second": 4, "third": 3}

			o, err := mergeObj(orig, n, "node-path")
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("returns an error when mergeMap throws an error", func() {
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
			o, err := mergeObj(origMap, newMap, "node-path")
			So(o, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "node-path.array.0: new object is a string, not a map - cannot merge using keys")
		})
	})
	Convey("Passing a slice to mergeObj", t, func() {
		Convey("without magical merge token replaces entire array", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"my", "new", "array"}
			expect := []interface{}{"my", "new", "array"}

			o, err := mergeObj(orig, array, "node-path")
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("When passed a slice, but original item is nil", func() {
			val := []interface{}{"(( append ))", "two"}
			expect := []interface{}{"two"}

			o, err := mergeObj(nil, val, "node-path")
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("Returns an error when mergeArray throws an error, when merging array to array", func() {
			orig := []interface{}{
				"first",
			}
			array := []interface{}{
				"(( merge on name ))",
				"second",
			}

			o, err := mergeObj(orig, array, "node-path")
			So(o, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "node-path.0: new object is a string, not a map - cannot merge using keys")
		})
		Convey("Returns an error when mergeArray throws an error, when merging nil to array", func() {
			array := []interface{}{
				"(( merge on name ))",
				"second",
			}

			o, err := mergeObj(nil, array, "node-path")
			So(o, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "node-path.0: new object is a string, not a map - cannot merge using keys")
		})
	})
	Convey("mergeObj merges in place", t, func() {
		Convey("When passed a string", func() {
			orig := 42
			val := "asdf"

			o, err := mergeObj(orig, val, "node-path")
			So(o, ShouldEqual, "asdf")
			So(err, ShouldBeNil)
		})
		Convey("When passed an int", func() {
			orig := "fdsa"
			val := 10

			o, err := mergeObj(orig, val, "node-path")
			So(o, ShouldEqual, 10)
			So(err, ShouldBeNil)
		})
		Convey("When passed an float64", func() {
			orig := "fdsa"
			val := 10.4

			o, err := mergeObj(orig, val, "node-path")
			So(o, ShouldEqual, 10.4)
			So(err, ShouldBeNil)
		})
		Convey("When passed nil", func() {
			orig := "fdsa"
			val := interface{}(nil)

			o, err := mergeObj(orig, val, "node-path")
			So(o, ShouldBeNil)
			So(err, ShouldBeNil)
		})
		Convey("When passed a map, but original item is a scalar", func() {
			orig := "value"
			val := map[interface{}]interface{}{"key": "value"}
			expect := map[interface{}]interface{}{"key": "value"}

			o, err := mergeObj(orig, val, "node-path")
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("When passed a map, but original item is nil", func() {
			val := map[interface{}]interface{}{"key": "value"}
			expect := map[interface{}]interface{}{"key": "value"}

			o, err := mergeObj(nil, val, "node-path")
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("When passed a slice, but original item is a scalar", func() {
			orig := "value"
			val := []interface{}{"one", "two"}
			expect := []interface{}{"one", "two"}

			o, err := mergeObj(orig, val, "node-path")
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

		err := mergeMap(origMap, newMap, "node-path")
		So(origMap, ShouldResemble, expectMap)
		So(err, ShouldBeNil)
	})
	Convey("mergeMap re-throws an error if it finds one while merging data", t, func() {
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
		err := mergeMap(origMap, newMap, "node-path")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "node-path.array.0: new object is a string, not a map - cannot merge using keys")
	})
}

func TestMergeArray(t *testing.T) {
	Convey("mergeArray", t, func() {
		Convey("with initial element '(( prepend ))' prepends new data", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"(( prepend ))", "zeroth"}
			expect := []interface{}{"zeroth", "first", "second"}

			a, err := mergeArray(orig, array, "node-path")
			So(a, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("with initial element '(( append ))' appends new data", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"(( append ))", "third"}
			expect := []interface{}{"first", "second", "third"}

			a, err := mergeArray(orig, array, "node-path")
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
				a, err := mergeArray(orig, array, "node-path")
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
				a, err := mergeArray(orig, array, "node-path")
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
				a, err := mergeArray(orig, array, "node-path")
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
				a, err := mergeArray(orig, array, "node-path")
				So(a, ShouldResemble, expect)
				So(err, ShouldBeNil)
			})
			Convey("returns error when mergObj returns error (merged data)", func() {
				orig := []interface{}{
					[]interface{}{
						"string",
					},
				}
				array := []interface{}{
					"(( inline ))",
					[]interface{}{
						"(( merge ))",
						"string",
					},
				}
				a, err := mergeArray(orig, array, "node-path")
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "node-path.0.0: new object is a string, not a map - cannot merge using keys")
			})
			Convey("returns error when mergObj returns error (appended data)", func() {
				orig := []interface{}{}
				array := []interface{}{
					"(( inline ))",
					[]interface{}{
						"(( merge ))",
						"string",
					},
				}
				a, err := mergeArray(orig, array, "node-path")
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "node-path.0.0: new object is a string, not a map - cannot merge using keys")
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

				a, err := mergeArray(orig, array, "node-path")
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

				a, err := mergeArray(orig, array, "node-path")
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

				a, err := mergeArray(orig, array, "node-path")
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "node-path.3: original object is a string, not a map - cannot merge using keys")
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

				a, err := mergeArray(orig, array, "node-path")
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "node-path.3: new object is a string, not a map - cannot merge using keys")
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

				a, err := mergeArray(orig, array, "node-path")
				So(a, ShouldBeNil)
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

				a, err := mergeArray(orig, array, "node-path")
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "node-path.2: new object does not contain the key 'id' - cannot merge")
			})
			Convey("Returns an error if mergeObj() returns an error", func() {
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
				a, err := mergeArray(orig, array, "node-path")
				So(a, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "node-path.0.val.0: new object is a string, not a map - cannot merge using keys")
			})
		})
		Convey("arrays of maps can be merged inline", func() {
			origMapSlice := map[interface{}]interface{}{"k1": "v1", "k2": "v2"}
			newMapSlice := map[interface{}]interface{}{"k3": "v3", "k2": "v2.new"}
			expectMapSlice := map[interface{}]interface{}{"k1": "v1", "k2": "v2.new", "k3": "v3"}
			orig := []interface{}{origMapSlice}
			array := []interface{}{newMapSlice}
			expect := []interface{}{expectMapSlice}

			o, err := mergeArrayInline(orig, array, "node-path")
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

			o, err := mergeObj(orig, array, "node-path")
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

			o, err := mergeObj(first, second, "an.inlined.merge")
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

			o, err := mergeObj(orig, array, "node-path")
			So(o, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
	})
}
