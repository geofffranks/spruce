package spruce

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/geofffranks/simpleyaml"
	"github.com/starkandwayne/goutils/tree"
)

var _ = Describe("We should key-based merge arrays of hashes", func() {
	It("If the element is a string with the right key-merge token", func() {
		yes, key := shouldKeyMergeArray([]interface{}{"(( merge ))", "stuff"})
		Expect(yes).To(BeTrue())
		Expect(key).To(Equal("name"))
	})
	It("If the element is a string with the right key-merge token and custom key specified", func() {
		yes, key := shouldKeyMergeArray([]interface{}{"(( merge on id ))", "stuff"})
		Expect(yes).To(BeTrue())
		Expect(key).To(Equal("id"))
	})
	Context("when DEFAULT_ARRAY_MERGE_KEY is set", func() {
		BeforeEach(func() {
			os.Setenv("DEFAULT_ARRAY_MERGE_KEY", "id")
		})
		AfterEach(func() {
			os.Setenv("DEFAULT_ARRAY_MERGE_KEY", "")
		})
		It("shouldKeyMergeArray picks up on it", func() {
			yes, key := shouldKeyMergeArray([]interface{}{"(( merge ))", "stuff"})
			Expect(yes).To(BeTrue())
			Expect(key).To(Equal("id"))
		})
	})
	Context("Is whitespace agnostic", func() {
		It("No surrounding whitespace", func() {
			yes, key := shouldKeyMergeArray([]interface{}{"((merge))"})
			Expect(yes).To(BeTrue())
			Expect(key).To(Equal("name"))
		})
		It("Surrounding tabs", func() {
			yes, key := shouldKeyMergeArray([]interface{}{"((	merge	))"})
			Expect(yes).To(BeTrue())
			Expect(key).To(Equal("name"))
		})
		It("Multiple surrounding whitespaces", func() {
			yes, key := shouldKeyMergeArray([]interface{}{"((  merge  ))"})
			Expect(yes).To(BeTrue())
			Expect(key).To(Equal("name"))
		})
		It("Multiple whitespace in key spec", func() {
			yes, key := shouldKeyMergeArray([]interface{}{"((  merge  on  id  ))"})
			Expect(yes).To(BeTrue())
			Expect(key).To(Equal("id"))
		})
		It("Tabs in key spec", func() {
			yes, key := shouldKeyMergeArray([]interface{}{"((  merge	on	id  ))"})
			Expect(yes).To(BeTrue())
			Expect(key).To(Equal("id"))
		})
		It("No trailing whitespace in key spec", func() {
			yes, key := shouldKeyMergeArray([]interface{}{"((  merge	on	id))"})
			Expect(yes).To(BeTrue())
			Expect(key).To(Equal("id"))
		})
		It("No whitespace in key spec", func() {
			yes, key := shouldKeyMergeArray([]interface{}{"((  mergeonid ))"})
			Expect(yes).To(BeFalse())
			Expect(key).To(Equal(""))
		})
		It("Only leading whitespace in key spec", func() {
			yes, key := shouldKeyMergeArray([]interface{}{"((  merge onid ))"})
			Expect(yes).To(BeFalse())
			Expect(key).To(Equal(""))
		})
	})
	It("But not if the element is a string with the wrong token", func() {
		yes, key := shouldKeyMergeArray([]interface{}{"not a magic token"})
		Expect(yes).To(BeFalse())
		Expect(key).To(Equal(""))
	})
	It("But not if the element is not a string", func() {
		yes, key := shouldKeyMergeArray([]interface{}{42})
		Expect(yes).To(BeFalse())
		Expect(key).To(Equal(""))
	})
	It("But not if the slice has no elements", func() {
		yes, key := shouldKeyMergeArray([]interface{}{})
		Expect(yes).To(BeFalse())
		Expect(key).To(Equal(""))
	})
})

var _ = Describe("Should recognize string patterns for", func() {
	// Helper: returns true if result has index equal to expIndex
	shouldInsertAt := func(result ModificationDefinition, expIndex int) bool {
		return result.index == expIndex
	}

	shouldBeAppend := func(result ModificationDefinition) bool {
		return shouldInsertAt(result, -1)
	}

	shouldBePrepend := func(result ModificationDefinition) bool {
		return shouldInsertAt(result, 0)
	}

	shouldBeDelete := func(result ModificationDefinition) bool {
		return result.listOp == listOpDelete
	}

	shouldBeDefault := func(result ModificationDefinition) bool {
		return result.listOp == listOpMergeDefault
	}

	shouldBeMergeOnKey := func(result ModificationDefinition) bool {
		return result.listOp == listOpMergeOnKey
	}

	shouldBeReplace := func(result ModificationDefinition) bool {
		return result.listOp == listOpReplace
	}

	Context("(( merge ))", func() {
		cases := map[string]bool{
			"(( merge ))":          true,
			"((merge))":            true,
			"((	merge	))":          true,
			"((  merge  ))":        true,
			"((     merge))":       true,
			"(( notmerge ))":       false,
			"(( mergenot ))":       false,
			"(( not even merge ))": false,
			"(( somethingelse ))":  false,
		}
		for input, shouldMatch := range cases {
			input, shouldMatch := input, shouldMatch
			Context(fmt.Sprintf("with case %s", input), func() {
				It("matches correctly", func() {
					results := getArrayModifications([]interface{}{input}, false)
					if shouldMatch {
						Expect(results).To(HaveLen(2))
						Expect(shouldBeDefault(results[0])).To(BeTrue())
						Expect(shouldBeMergeOnKey(results[1])).To(BeTrue())
					} else {
						Expect(results).To(HaveLen(1))
						Expect(shouldBeDefault(results[0])).To(BeTrue())
					}
				})
			})
		}
	})

	Context("(( replace ))", func() {
		cases := map[string]bool{
			"(( replace ))":          true,
			"((replace))":            true,
			"((	replace	))":          true,
			"((  replace  ))":        true,
			"((     replace))":       true,
			"(( notreplace ))":       false,
			"(( replacenot ))":       false,
			"(( not even replace ))": false,
			"(( somethingelse ))":    false,
		}
		for input, shouldMatch := range cases {
			input, shouldMatch := input, shouldMatch
			Context(fmt.Sprintf("with case %s", input), func() {
				It("matches correctly", func() {
					results := getArrayModifications([]interface{}{input}, false)
					if shouldMatch {
						Expect(results).To(HaveLen(2))
						Expect(shouldBeDefault(results[0])).To(BeTrue())
						Expect(shouldBeReplace(results[1])).To(BeTrue())
					} else {
						Expect(results).To(HaveLen(1))
						Expect(shouldBeDefault(results[0])).To(BeTrue())
					}
				})
			})
		}
	})

	Context("(( append ))", func() {
		cases := map[string]bool{
			"(( append ))":          true,
			"((append))":            true,
			"((	append	))":          true,
			"((  append  ))":        true,
			"((     append))":       true,
			"(( notappend ))":       false,
			"(( appendnot ))":       false,
			"(( not even append ))": false,
			"(( somethingelse ))":   false,
		}
		for input, shouldMatch := range cases {
			input, shouldMatch := input, shouldMatch
			Context(fmt.Sprintf("with case %s", input), func() {
				It("matches correctly", func() {
					results := getArrayModifications([]interface{}{input}, false)
					if shouldMatch {
						Expect(results).To(HaveLen(2))
						Expect(shouldBeDefault(results[0])).To(BeTrue())
						Expect(shouldBeAppend(results[1])).To(BeTrue())
					} else {
						Expect(results).To(HaveLen(1))
						Expect(shouldBeDefault(results[0])).To(BeTrue())
					}
				})
			})
		}
	})

	Context("(( prepend ))", func() {
		cases := map[string]bool{
			"(( prepend ))":          true,
			"((prepend))":            true,
			"((	prepend	))":          true,
			"((  prepend  ))":        true,
			"((     prepend))":       true,
			"(( notprepend ))":       false,
			"(( prependnot ))":       false,
			"(( not even prepend ))": false,
			"(( somethingelse ))":    false,
		}
		for input, shouldMatch := range cases {
			input, shouldMatch := input, shouldMatch
			Context(fmt.Sprintf("with case %s", input), func() {
				It("matches correctly", func() {
					results := getArrayModifications([]interface{}{input}, false)
					if shouldMatch {
						Expect(results).To(HaveLen(2))
						Expect(shouldBeDefault(results[0])).To(BeTrue())
						Expect(shouldBePrepend(results[1])).To(BeTrue())
					} else {
						Expect(results).To(HaveLen(1))
						Expect(shouldBeDefault(results[0])).To(BeTrue())
					}
				})
			})
		}
	})

	Context("(( insert ... ))", func() {
		for _, rel := range []string{"before", "after"} {
			rel := rel
			for _, index := range []int{0, 10, 100} {
				index := index
				// index based insert cases
				indexCases := map[string]bool{
					fmt.Sprintf("(( insert %s %d ))", rel, index):          true,
					fmt.Sprintf("(( insert %s	 %d ))", rel, index):         true,
					fmt.Sprintf("(( insert	 %s %d ))", rel, index):         true,
					fmt.Sprintf("(( insert   %s   %d ))", rel, index):      true,
					fmt.Sprintf("((   insert %s %d ))", rel, index):        true,
					fmt.Sprintf("(( insert %s %d   ))", rel, index):        true,
					fmt.Sprintf("((    insert %s %d ))", rel, index):       true,
					fmt.Sprintf("((   insert   %s   %d	 ))", rel, index):   true,
					fmt.Sprintf("(( insert%s%d ))", rel, index):            false,
					fmt.Sprintf("(( insert%s %d ))", rel, index):           false,
					fmt.Sprintf("(( notinsert %s %d ))", rel, index):       false,
					fmt.Sprintf("(( insertnot %s %d ))", rel, index):       false,
					fmt.Sprintf("(( not even insert %s %d ))", rel, index): false,
					fmt.Sprintf("(( somethingelse %s %d ))", rel, index):   false,
				}
				for input, shouldMatch := range indexCases {
					input, shouldMatch := input, shouldMatch
					Context(fmt.Sprintf("with case %s", input), func() {
						It("matches correctly", func() {
							results := getArrayModifications([]interface{}{input}, false)
							if shouldMatch {
								Expect(results).To(HaveLen(2))
								Expect(shouldBeDefault(results[0])).To(BeTrue())
								Expect(shouldInsertAt(results[1], index)).To(BeTrue())
								Expect(results[1].relative).To(Equal(rel))
							} else {
								Expect(results).To(HaveLen(1))
								Expect(shouldBeDefault(results[0])).To(BeTrue())
							}
						})
					})
				}
			}
			for _, key := range []string{"", "foo"} {
				key := key
				// name based insert cases
				nameCases := map[string]bool{
					fmt.Sprintf(`(( insert %s %s "spruce" ))`, rel, key):                true,
					fmt.Sprintf(`(( insert %s %s	 "spruce" ))`, rel, key):               true,
					fmt.Sprintf(`(( insert	 %s %s "spruce" ))`, rel, key):               true,
					fmt.Sprintf(`(( insert   %s %s   "spruce" ))`, rel, key):            true,
					fmt.Sprintf(`((   insert %s %s "spruce" ))`, rel, key):              true,
					fmt.Sprintf(`(( insert %s %s "spruce"   ))`, rel, key):              true,
					fmt.Sprintf(`((    insert %s %s "spruce" ))`, rel, key):             true,
					fmt.Sprintf(`(( insert %s    %s "spruce" ))`, rel, key):             true,
					fmt.Sprintf(`((    insert    %s    %s     "spruce"   ))`, rel, key): true,
					fmt.Sprintf(`((   insert   %s %s   "spruce"	 ))`, rel, key):         true,
					fmt.Sprintf("(( insert%s%sspruce ))", rel, key):                     false,
					fmt.Sprintf("(( insert%s%s spruce ))", rel, key):                    false,
					fmt.Sprintf("(( notinsert %s %s ))", rel, key):                      false,
					fmt.Sprintf("(( insertnot %s %s ))", rel, key):                      false,
					fmt.Sprintf("(( not even insert %s %s ))", rel, key):                false,
					fmt.Sprintf("(( somethingelse %s %s ))", rel, key):                  false,
				}
				for input, shouldMatch := range nameCases {
					input, shouldMatch, capturedKey := input, shouldMatch, key
					Context(fmt.Sprintf("with case %s", input), func() {
						It("matches correctly", func() {
							results := getArrayModifications([]interface{}{input}, false)
							if shouldMatch {
								Expect(results).To(HaveLen(2))
								Expect(shouldBeDefault(results[0])).To(BeTrue())
								Expect(results[1].name).To(Equal("spruce"))
								testKey := capturedKey
								if testKey == "" {
									testKey = "name"
								}
								Expect(results[1].key).To(Equal(testKey))
								Expect(results[1].relative).To(Equal(rel))
							} else {
								Expect(results).To(HaveLen(1))
								Expect(shouldBeDefault(results[0])).To(BeTrue())
							}
						})
					})
				}
			}
		}
	})

	Context("(( delete ... ))", func() {
		for _, key := range []string{"", "foo"} {
			key := key
			// name based delete cases
			nameCases := map[string]bool{
				fmt.Sprintf(`(( delete %s "spruce" ))`, key):        true,
				fmt.Sprintf(`(( delete %s	 "spruce" ))`, key):       true,
				fmt.Sprintf(`(( delete	 %s "spruce" ))`, key):       true,
				fmt.Sprintf(`(( delete   %s   "spruce" ))`, key):    true,
				fmt.Sprintf(`((   delete %s "spruce" ))`, key):      true,
				fmt.Sprintf(`(( delete %s "spruce"   ))`, key):      true,
				fmt.Sprintf(`(( delete %s "spruce"    ))`, key):     true,
				fmt.Sprintf(`((    delete %s "spruce" ))`, key):     true,
				fmt.Sprintf(`((   delete   %s   "spruce"	 ))`, key): true,
				fmt.Sprintf("(( delete%sspruce ))", key):            false,
				fmt.Sprintf("(( notdelete %s ))", key):              false,
				fmt.Sprintf("(( deletenot %s ))", key):              false,
				fmt.Sprintf("(( not even delete %s ))", key):        false,
				fmt.Sprintf("(( somethingelse %s ))", key):          false,
			}
			for input, shouldMatch := range nameCases {
				input, shouldMatch, capturedKey := input, shouldMatch, key
				Context(fmt.Sprintf("with case %s", input), func() {
					It("matches correctly", func() {
						results := getArrayModifications([]interface{}{input}, false)
						if shouldMatch {
							Expect(results).To(HaveLen(2))
							Expect(shouldBeDefault(results[0])).To(BeTrue())
							Expect(results[1].name).To(Equal("spruce"))
							testKey := capturedKey
							if testKey == "" {
								testKey = "name"
							}
							Expect(results[1].key).To(Equal(testKey))
						} else {
							Expect(results).To(HaveLen(1))
							Expect(shouldBeDefault(results[0])).To(BeTrue())
						}
					})
				})
			}
		}

		simpleCases := map[string]bool{
			`(( delete  "MrSpiff" ))`:    true,
			`(( delete 	 "MrSpiff" ))`:   true,
			`((   delete  "MrSpiff" ))`:  true,
			`(( delete  "MrSpiff"   ))`:  true,
			`(( delete  "MrSpiff"    ))`: true,
			`(( delete  MrSpiff ))`:      true,
			`(( deletespruce ))`:         false,
			`(( delete  Mr Spiff ))`:     false,
			`(( delete  "" "MrSpiff" ))`: false,
		}
		for input, shouldMatch := range simpleCases {
			input, shouldMatch := input, shouldMatch
			Context(fmt.Sprintf("with case %s on simple list", input), func() {
				It("matches correctly", func() {
					results := getArrayModifications([]interface{}{input}, true)
					if shouldMatch {
						Expect(results).To(HaveLen(2))
						Expect(shouldBeDefault(results[0])).To(BeTrue())
						Expect(results[1].key).To(BeEmpty())
						Expect(results[1].name).To(Equal("MrSpiff"))
					} else {
						Expect(results).To(HaveLen(1))
						Expect(shouldBeDefault(results[0])).To(BeTrue())
					}
				})
			})
		}

		for _, index := range []int{0, 10, 100} {
			index := index
			// index based delete cases
			indexDeleteCases := map[string]bool{
				fmt.Sprintf(`(( delete %d ))`, index):         true,
				fmt.Sprintf(`(( delete     %d ))`, index):     true,
				fmt.Sprintf(`((   delete %d ))`, index):       true,
				fmt.Sprintf(`(( delete %d   ))`, index):       true,
				fmt.Sprintf(`((   delete     %d	 ))`, index):  true,
				fmt.Sprintf(`(( delete%d ))`, index):          false,
				fmt.Sprintf(`(( notdelete %d))`, index):       false,
				fmt.Sprintf(`(( deletenot %d))`, index):       false,
				fmt.Sprintf(`(( not even delete %d))`, index): false,
				fmt.Sprintf(`(( something else %d))`, index):  false,
			}
			for input, shouldDelete := range indexDeleteCases {
				input, shouldDelete := input, shouldDelete
				Context(fmt.Sprintf("with case %s", input), func() {
					It("matches correctly", func() {
						results := getArrayModifications([]interface{}{input}, false)
						if shouldDelete {
							Expect(results).To(HaveLen(2))
							Expect(shouldBeDefault(results[0])).To(BeTrue())
							Expect(results[1].index).To(Equal(index))
							Expect(shouldBeDelete(results[1])).To(BeTrue())
						} else {
							Expect(results).To(HaveLen(1))
							Expect(shouldBeDefault(results[0])).To(BeTrue())
						}
					})
				})
			}
		}
	})

	It("Don't return an insert if index is obviously out of bounds", func() {
		results := getArrayModifications([]interface{}{"(( insert before -1 ))", "stuff"}, false)
		Expect(results).To(HaveLen(1)) //Just the default merge
		Expect(results[0].listOp).To(Equal(listOpMergeDefault))
	})

	It("If there are multiple insert token with after/before, different key names, and names (only technical usecase)", func() {
		results := getArrayModifications([]interface{}{
			"(( insert after name \"nats\" ))",
			"stuff1",
			"stuff2",
			"stuff3",
			"(( insert before id \"consul\" ))",
			"stuffX1",
			"stuffX2",
		}, false)
		Expect(results).To(HaveLen(3))
		Expect(shouldBeDefault(results[0])).To(BeTrue())
		Expect(results[1].relative).To(Equal("after"))
		Expect(results[1].key).To(Equal("name"))
		Expect(results[1].name).To(Equal("nats"))
		Expect(results[1].list).To(Equal([]interface{}{"stuff1", "stuff2", "stuff3"}))
		Expect(results[2].relative).To(Equal("before"))
		Expect(results[2].key).To(Equal("id"))
		Expect(results[2].name).To(Equal("consul"))
		Expect(results[2].list).To(Equal([]interface{}{"stuffX1", "stuffX2"}))
	})

	It("Only default merge if no operators given", func() {
		results := getArrayModifications([]interface{}{"not a magic token", "stuff"}, false)
		Expect(results).To(HaveLen(1))
		Expect(shouldBeDefault(results[0])).To(BeTrue())
	})

	It("Can specify operators without one at the 0th index", func() {
		results := getArrayModifications([]interface{}{"foo", "(( append ))", "stuff"}, false)
		Expect(results).To(HaveLen(2))
		Expect(results[0].listOp).To(Equal(listOpMergeDefault))
		Expect(shouldBeAppend(results[1])).To(BeTrue())
		Expect(results[0].list).To(HaveLen(1))
		Expect(results[1].list).To(HaveLen(1))
		Expect(results[0].list[0]).To(Equal("foo"))
		Expect(results[1].list[0]).To(Equal("stuff"))
	})
})

var _ = Describe("Passing a map to m.mergeObj merges as a map", func() {
	It("merges as a map under normal conditions", func() {
		orig := map[interface{}]interface{}{"first": 1, "second": 2}
		n := map[interface{}]interface{}{"second": 4, "third": 3}
		expect := map[interface{}]interface{}{"first": 1, "second": 4, "third": 3}

		m := &Merger{}
		o := m.mergeObj(orig, n, "node-path")
		err := m.Error()
		Expect(o).To(Equal(expect))
		Expect(err).NotTo(HaveOccurred())
	})
	It("returns an error when m.mergeMap throws an error", func() {
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
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("node-path.array.0: new object is a string, not a map - cannot merge by key"))
	})
	It("returns an error for any (( merge ... )) operators found in non-list context", func() {
		orig := map[interface{}]interface{}{}
		n := map[interface{}]interface{}{"map": "(( merge || nil ))"}
		m := &Merger{}
		m.mergeObj(orig, n, "node-path")
		err := m.Error()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("node-path.map: inappropriate use of (( merge )) operator outside of a list (this is spruce, after all)"))
	})
})

var _ = Describe("Passing a slice to m.mergeObj", func() {
	It("without magical merge token replaces entire array", func() {
		orig := []interface{}{"first", "second"}
		array := []interface{}{"my", "new", "array"}
		expect := []interface{}{"my", "new", "array"}

		m := &Merger{}
		o := m.mergeObj(orig, array, "node-path")
		err := m.Error()
		Expect(o).To(Equal(expect))
		Expect(err).NotTo(HaveOccurred())
	})
	It("When passed a slice, but original item is nil", func() {
		val := []interface{}{"(( append ))", "two"}
		expect := []interface{}{"two"}

		m := &Merger{}
		o := m.mergeObj(nil, val, "node-path")
		err := m.Error()
		Expect(o).To(Equal(expect))
		Expect(err).NotTo(HaveOccurred())
	})
	It("Returns an error when m.mergeArray throws an error, when merging array to array", func() {
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
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("node-path.0: new object is a string, not a map - cannot merge by key"))
	})
	It("Returns an error when m.mergeArray throws an error, when merging nil to array", func() {
		array := []interface{}{
			"(( merge on name ))",
			"second",
		}

		m := &Merger{}
		m.mergeObj(nil, array, "node-path")
		err := m.Error()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("node-path.0: new object is a string, not a map - cannot merge by key"))
	})
})

var _ = Describe("m.mergeObj merges in place", func() {
	It("When passed a string", func() {
		orig := 42
		val := "asdf"

		m := &Merger{}
		o := m.mergeObj(orig, val, "node-path")
		err := m.Error()
		Expect(o).To(Equal("asdf"))
		Expect(err).NotTo(HaveOccurred())
	})
	It("When passed an int", func() {
		orig := "fdsa"
		val := 10

		m := &Merger{}
		o := m.mergeObj(orig, val, "node-path")
		err := m.Error()
		Expect(o).To(Equal(10))
		Expect(err).NotTo(HaveOccurred())
	})
	It("When passed an float64", func() {
		orig := "fdsa"
		val := 10.4

		m := &Merger{}
		o := m.mergeObj(orig, val, "node-path")
		err := m.Error()
		Expect(o).To(Equal(10.4))
		Expect(err).NotTo(HaveOccurred())
	})
	It("When passed nil", func() {
		orig := "fdsa"
		val := interface{}(nil)

		m := &Merger{}
		o := m.mergeObj(orig, val, "node-path")
		err := m.Error()
		Expect(o).To(BeNil())
		Expect(err).NotTo(HaveOccurred())
	})
	It("When passed a map, but original item is a scalar", func() {
		orig := "value"
		val := map[interface{}]interface{}{"key": "value"}
		expect := map[interface{}]interface{}{"key": "value"}

		m := &Merger{}
		o := m.mergeObj(orig, val, "node-path")
		err := m.Error()
		Expect(o).To(Equal(expect))
		Expect(err).NotTo(HaveOccurred())
	})
	It("When passed a map, but original item is nil", func() {
		val := map[interface{}]interface{}{"key": "value"}
		expect := map[interface{}]interface{}{"key": "value"}

		m := &Merger{}
		o := m.mergeObj(nil, val, "node-path")
		err := m.Error()
		Expect(o).To(Equal(expect))
		Expect(err).NotTo(HaveOccurred())
	})
	It("When passed a slice, but original item is a scalar", func() {
		orig := "value"
		val := []interface{}{"one", "two"}
		expect := []interface{}{"one", "two"}

		m := &Merger{}
		o := m.mergeObj(orig, val, "node-path")
		err := m.Error()
		Expect(o).To(Equal(expect))
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("TestMergeMap", func() {
	It("with map elements updates original map", func() {
		origMap := map[interface{}]interface{}{"k1": "v1", "k2": "v2"}
		newMap := map[interface{}]interface{}{"k3": "v3", "k2": "v2.new"}
		expectMap := map[interface{}]interface{}{"k2": "v2.new", "k3": "v3", "k1": "v1"}

		m := &Merger{}
		m.mergeMap(origMap, newMap, "node-path")
		Expect(origMap).To(Equal(expectMap))
		Expect(m.Error()).NotTo(HaveOccurred())
	})
	It("m.mergeMap re-throws an error if it finds one while merging data", func() {
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
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("node-path.array.0: new object is a string, not a map - cannot merge by key"))
	})
})

var _ = Describe("m.mergeArray", func() {
	It("with initial element '(( prepend ))' prepends new data", func() {
		orig := []interface{}{"first", "second"}
		array := []interface{}{"(( prepend ))", "zeroth"}
		expect := []interface{}{"zeroth", "first", "second"}

		m := &Merger{}
		a := m.mergeArray(orig, array, "node-path")
		err := m.Error()
		Expect(a).To(Equal(expect))
		Expect(err).NotTo(HaveOccurred())
	})
	It("with initial element '(( append ))' appends new data", func() {
		orig := []interface{}{"first", "second"}
		array := []interface{}{"(( append ))", "third"}
		expect := []interface{}{"first", "second", "third"}

		m := &Merger{}
		a := m.mergeArray(orig, array, "node-path")
		err := m.Error()
		Expect(a).To(Equal(expect))
		Expect(err).NotTo(HaveOccurred())
	})

	Context("with initial element '(( inline ))'", func() {
		It("and len(orig) == len(new)", func() {
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})
		It("and len(orig) > len(new)", func() {
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})
		It("and len(orig < len(new)", func() {
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})
		It("and empty orig slice", func() {
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns error when mergeObj returns error (merged data)", func() {
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
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("node-path.0.0: new object is a string, not a map - cannot merge by key"))
		})
		It("returns error when mergeObj returns error (appended data)", func() {
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
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("node-path.0.0: new object is a string, not a map - cannot merge by key"))
		})
	})

	Context("with initial element '(( merge ))'", func() {
		It("merges and defaults to 'name' for a key, if not specified", func() {
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})
		It("Default merging falls back to inline if one of the original map's target keys' values are a map", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org1"},
				map[interface{}]interface{}{"name": map[interface{}]interface{}{"beep": "boop"}, "org": "org2"},
			}
			array := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org3"},
				map[interface{}]interface{}{"name": "bar", "org": "org4"},
			}
			expect := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org3"},
				map[interface{}]interface{}{"name": "bar", "org": "org4"},
			}
			m := &Merger{}
			output := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(expect))
		})
		It("Default merging falls back to inline if one of the original map's target keys' values are a sequence", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org1"},
				map[interface{}]interface{}{"name": []interface{}{"beep", "boop"}},
			}
			array := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org3"},
				map[interface{}]interface{}{"name": "bar", "org": "org4"},
			}
			expect := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org3"},
				map[interface{}]interface{}{"name": "bar", "org": "org4"},
			}
			m := &Merger{}
			output := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(expect))
		})
		It("Default merging falls back to inline if one of the new map's target keys' values are a map", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org1"},
				map[interface{}]interface{}{"name": "bar", "org": "org2"},
			}
			array := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org3"},
				map[interface{}]interface{}{"name": map[interface{}]interface{}{"beep": "boop"}, "org": "org3"},
			}
			expect := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org3"},
				map[interface{}]interface{}{"name": map[interface{}]interface{}{"beep": "boop"}, "org": "org3"},
			}
			m := &Merger{}
			output := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(expect))
		})
		It("Default merging falls back to inline if one of the new map's target keys' values are a sequence", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org1"},
				map[interface{}]interface{}{"name": "bar", "org": "org2"},
			}
			array := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org3"},
				map[interface{}]interface{}{"name": []interface{}{"beep", "boop"}},
			}
			expect := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org3"},
				map[interface{}]interface{}{"name": []interface{}{"beep", "boop"}, "org": "org2"},
			}
			m := &Merger{}
			output := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(expect))
		})
		It("Explicit merging fails if one of the original map's target keys' values are a map", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org1"},
				map[interface{}]interface{}{"name": map[interface{}]interface{}{"beep": "boop"}, "org": "org2"},
			}
			array := []interface{}{
				"(( merge ))",
				map[interface{}]interface{}{"name": "foo", "org": "org3"},
				map[interface{}]interface{}{"name": "bar", "org": "org4"},
			}
			m := &Merger{}
			m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(err).To(HaveOccurred())
		})
		It("Explicit merging fails if one of the original map's target keys' values are a sequence", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org1"},
				map[interface{}]interface{}{"name": []interface{}{"beep", "boop"}},
			}
			array := []interface{}{
				"(( merge ))",
				map[interface{}]interface{}{"name": "foo", "org": "org3"},
				map[interface{}]interface{}{"name": "bar", "org": "org4"},
			}
			m := &Merger{}
			m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(err).To(HaveOccurred())
		})
		It("Explicit merging fails if one of the new map's target keys' values are a map", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org1"},
				map[interface{}]interface{}{"name": "bar", "org": "org2"},
			}
			array := []interface{}{
				"(( merge ))",
				map[interface{}]interface{}{"name": "foo", "org": "org3"},
				map[interface{}]interface{}{"name": map[interface{}]interface{}{"beep": "boop"}, "org": "org3"},
			}
			m := &Merger{}
			m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(err).To(HaveOccurred())
		})
		It("Explicit merging fails if one of the new map's target keys' values are a sequence", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "foo", "org": "org1"},
				map[interface{}]interface{}{"name": "bar", "org": "org2"},
			}
			array := []interface{}{
				"(( merge ))",
				map[interface{}]interface{}{"name": "foo", "org": "org3"},
				map[interface{}]interface{}{"name": []interface{}{"beep", "boop"}},
			}
			m := &Merger{}
			m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(err).To(HaveOccurred())
		})
		It("allows custom key merging to be specified", func() {
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})
		It("But not if any of the original array elements are not maps", func() {
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
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("node-path.3: original object is a string, not a map - cannot merge by key"))
		})
		It("But not if any of the new array elements are not maps", func() {
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
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("node-path.3: new object is a string, not a map - cannot merge by key"))
		})
		It("But not if any of the elements of the original array don't have the key requested", func() {
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
			Expect(err).To(HaveOccurred())
		})
		It("But not if any of the elements of the new array don't have the key requested", func() {
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
			Expect(a).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("node-path.2: new object does not contain the key 'id' - cannot merge by key"))
		})
		It("Returns an error if m.mergeObj() returns an error", func() {
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
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("node-path.first.val.0: new object is a string, not a map - cannot merge by key"))
		})
	})

	It("arrays of maps can be merged inline", func() {
		origMapSlice := map[interface{}]interface{}{"k1": "v1", "k2": "v2"}
		newMapSlice := map[interface{}]interface{}{"k3": "v3", "k2": "v2.new"}
		expectMapSlice := map[interface{}]interface{}{"k1": "v1", "k2": "v2.new", "k3": "v3"}
		orig := []interface{}{origMapSlice}
		array := []interface{}{newMapSlice}
		expect := []interface{}{expectMapSlice}

		m := &Merger{}
		o := m.mergeArrayInline(orig, array, "node-path")
		err := m.Error()
		Expect(o).To(Equal(expect))
		Expect(err).NotTo(HaveOccurred())
	})
	It("merges arrays of maps by default", func() {
		origMapSlice := map[interface{}]interface{}{"k1": "v1", "k2": "v2"}
		newMapSlice := map[interface{}]interface{}{"k3": "v3", "k2": "v2.new"}
		expectMapSlice := map[interface{}]interface{}{"k1": "v1", "k2": "v2.new", "k3": "v3"}
		orig := []interface{}{origMapSlice}
		array := []interface{}{newMapSlice}
		expect := []interface{}{expectMapSlice}

		m := &Merger{}
		o := m.mergeObj(orig, array, "node-path")
		err := m.Error()
		Expect(o).To(Equal(expect))
		Expect(err).NotTo(HaveOccurred())
	})
	It("uses key-merge if possible", func() {
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
		Expect(o).To(Equal([]interface{}{
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
		}))
		Expect(err).NotTo(HaveOccurred())
	})

	Context("setting DEFAULT_ARRAY_MERGE_KEY", func() {
		BeforeEach(func() {
			os.Setenv("DEFAULT_ARRAY_MERGE_KEY", "id")
		})
		AfterEach(func() {
			os.Setenv("DEFAULT_ARRAY_MERGE_KEY", "")
		})
		It("can override key-merge by default", func() {
			first := []interface{}{
				map[interface{}]interface{}{
					"name": "first",
					"k1":   "v1",
					"id":   "1",
				},
				map[interface{}]interface{}{
					"name": "second",
					"done": "yes",
					"id":   "2",
				},
			}
			second := []interface{}{
				map[interface{}]interface{}{
					"name": "second",
					"2":    "best",
					"test": "test",
					"id":   "2",
				},
				map[interface{}]interface{}{
					"name": "first",
					"k1":   "1",
					"k2":   "2",
					"id":   "1",
				},
			}

			m := &Merger{}
			o := m.mergeObj(first, second, "an.inlined.merge")
			err := m.Error()
			Expect(o).To(Equal([]interface{}{
				map[interface{}]interface{}{
					"name": "first",
					"k1":   "1",
					"k2":   "2",
					"id":   "1",
				},
				map[interface{}]interface{}{
					"name": "second",
					"2":    "best",
					"done": "yes",
					"test": "test",
					"id":   "2",
				},
			}))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	It("without magical merge token replaces entire array", func() {
		orig := []interface{}{"first", "second"}
		array := []interface{}{"my", "new", "array"}
		expect := []interface{}{"my", "new", "array"}

		m := &Merger{}
		o := m.mergeObj(orig, array, "node-path")
		err := m.Error()
		Expect(o).To(Equal(expect))
		Expect(err).NotTo(HaveOccurred())
	})

	Context("with element '(( insert ))' inserts new data where wanted", func() {
		It("After #0 put the new entry", func() {
			orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
			array := []interface{}{"(( insert after 0 ))", "new-kid-on-the-block"}
			expect := []interface{}{"first", "new-kid-on-the-block", "second", "third", "fourth", "fifth", "sixth"}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Before #0 put the new entry", func() {
			orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
			array := []interface{}{"(( insert before 0 ))", "new-kid-on-the-block"}
			expect := []interface{}{"new-kid-on-the-block", "first", "second", "third", "fourth", "fifth", "sixth"}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("After #4 put the new entry", func() {
			orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
			array := []interface{}{"(( insert after 4 ))", "new-kid-on-the-block"}
			expect := []interface{}{"first", "second", "third", "fourth", "fifth", "new-kid-on-the-block", "sixth"}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("After #5 put the new entry", func() {
			orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
			array := []interface{}{"(( insert after 5 ))", "new-kid-on-the-block"}
			expect := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth", "new-kid-on-the-block"}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("throw an error if insertion point is out of bounds", func() {
			orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
			array := []interface{}{"(( insert after 6 ))", "new-kid-on-the-block"}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to modify the list, because specified index 7 is out of bounds"))
		})

		It("After '<default>: first' put the new entry", func() {
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Before '<default>: first' put the new entry", func() {
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("After 'id: second' put the new entry", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( insert after id \"second\" ))",
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Before 'id: second' put the new entry", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( insert before id \"second\" ))",
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Insert multiple new entries before second and after first", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( insert before id \"second\" ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
				"(( insert after id \"first\" ))",
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Insert multiple new entries in two batches after first", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( insert after id \"first\" ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
				"(( insert after id \"first\" ))",
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Insert multiple new entries in three batches after first", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( insert after id \"first\" ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
				"(( insert after id \"first\" ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-3", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-4", "release": "vNext"},
				"(( insert after id \"first\" ))",
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
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Insert multiple new entries in different batches", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
				map[interface{}]interface{}{"id": "third", "release": "v1"},
				map[interface{}]interface{}{"id": "fourth", "release": "v1"},
				map[interface{}]interface{}{"id": "fifth", "release": "v1"},
				map[interface{}]interface{}{"id": "sixth", "release": "v1"},
				map[interface{}]interface{}{"id": "seventh", "release": "v1"},
				map[interface{}]interface{}{"id": "eighth", "release": "v1"},
			}

			array := []interface{}{
				"(( insert after id \"second\" ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
				"(( insert after id \"fourth\" ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
				"(( insert after id \"sixth\" ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-3", "release": "vNext"},
			}

			expect := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
				map[interface{}]interface{}{"id": "third", "release": "v1"},
				map[interface{}]interface{}{"id": "fourth", "release": "v1"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
				map[interface{}]interface{}{"id": "fifth", "release": "v1"},
				map[interface{}]interface{}{"id": "sixth", "release": "v1"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-3", "release": "vNext"},
				map[interface{}]interface{}{"id": "seventh", "release": "v1"},
				map[interface{}]interface{}{"id": "eighth", "release": "v1"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Insert multiple new entries in different batches with empty lists", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
				map[interface{}]interface{}{"id": "third", "release": "v1"},
				map[interface{}]interface{}{"id": "fourth", "release": "v1"},
				map[interface{}]interface{}{"id": "fifth", "release": "v1"},
				map[interface{}]interface{}{"id": "sixth", "release": "v1"},
				map[interface{}]interface{}{"id": "seventh", "release": "v1"},
				map[interface{}]interface{}{"id": "eighth", "release": "v1"},
			}

			array := []interface{}{
				"(( insert after id \"second\" ))",
				"(( insert after id \"fourth\" ))",
				"(( insert after id \"sixth\" ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block", "release": "vNext"},
			}

			expect := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
				map[interface{}]interface{}{"id": "third", "release": "v1"},
				map[interface{}]interface{}{"id": "fourth", "release": "v1"},
				map[interface{}]interface{}{"id": "fifth", "release": "v1"},
				map[interface{}]interface{}{"id": "sixth", "release": "v1"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block", "release": "vNext"},
				map[interface{}]interface{}{"id": "seventh", "release": "v1"},
				map[interface{}]interface{}{"id": "eighth", "release": "v1"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Insert multiple new entries in with different insertion styles", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
				map[interface{}]interface{}{"id": "third", "release": "v1"},
				map[interface{}]interface{}{"id": "fourth", "release": "v1"},
				map[interface{}]interface{}{"id": "fifth", "release": "v1"},
				map[interface{}]interface{}{"id": "sixth", "release": "v1"},
				map[interface{}]interface{}{"id": "seventh", "release": "v1"},
				map[interface{}]interface{}{"id": "eighth", "release": "v1"},
			}

			array := []interface{}{
				"(( insert after id \"second\" ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
				"(( prepend ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-3", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-4", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-5", "release": "vNext"},
				"(( append ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-6", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-7", "release": "vNext"},
				"(( insert before id \"fourth\" ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-8", "release": "vNext"},
				"(( insert after 0 ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-9", "release": "vNext"},
			}

			expect := []interface{}{
				map[interface{}]interface{}{"id": "new-kid-on-the-block-3", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-9", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-4", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-5", "release": "vNext"},
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-1", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-2", "release": "vNext"},
				map[interface{}]interface{}{"id": "third", "release": "v1"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-8", "release": "vNext"},
				map[interface{}]interface{}{"id": "fourth", "release": "v1"},
				map[interface{}]interface{}{"id": "fifth", "release": "v1"},
				map[interface{}]interface{}{"id": "sixth", "release": "v1"},
				map[interface{}]interface{}{"id": "seventh", "release": "v1"},
				map[interface{}]interface{}{"id": "eighth", "release": "v1"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-6", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-7", "release": "vNext"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Insert multiple new entries in with different insertion styles which depend on each other (without name keys)", func() {
			orig := []interface{}{
				"1",
				"2",
				"3",
			}

			array := []interface{}{
				"(( prepend ))",
				"this thing",
				"that thing",
				"(( insert before 1 ))",
				"first insert",
				"(( insert before 2 ))",
				"second insert",
				"(( insert before 6 ))",
				"stuff",
			}

			expect := []interface{}{
				"this thing",
				"first insert",
				"second insert",
				"that thing",
				"1",
				"2",
				"stuff",
				"3",
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Insert multiple new entries in with different insertion styles which depend on each other (with id keys)", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "1"},
				map[interface{}]interface{}{"id": "2"},
				map[interface{}]interface{}{"id": "3"},
			}

			array := []interface{}{
				"(( prepend ))",
				map[interface{}]interface{}{"id": "this thing"},
				map[interface{}]interface{}{"id": "that thing"},
				"(( insert after id \"this thing\" ))",
				map[interface{}]interface{}{"id": "first insert"},
				"(( insert after id \"first insert\" ))",
				map[interface{}]interface{}{"id": "second insert"},
				"(( insert after id \"2\" ))",
				map[interface{}]interface{}{"id": "stuff"},
			}

			expect := []interface{}{
				map[interface{}]interface{}{"id": "this thing"},
				map[interface{}]interface{}{"id": "first insert"},
				map[interface{}]interface{}{"id": "second insert"},
				map[interface{}]interface{}{"id": "that thing"},
				map[interface{}]interface{}{"id": "1"},
				map[interface{}]interface{}{"id": "2"},
				map[interface{}]interface{}{"id": "stuff"},
				map[interface{}]interface{}{"id": "3"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("throw an error when insertion point cannot be found", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "first", "release": "v1"},
				map[interface{}]interface{}{"name": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( insert after name \"not-existing\" ))",
				map[interface{}]interface{}{"name": "new-kid-on-the-block", "release": "vNext"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find specified modification point"))
		})

		It("throw an error when key cannot be found in new list", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( insert after id \"second\" ))",
				map[interface{}]interface{}{"name": "new-kid-on-the-block", "release": "vNext"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("new object does not contain the key"))
		})

		It("throw an error when key cannot be found in original list", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "first", "release": "v1"},
				map[interface{}]interface{}{"name": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( insert after id \"second\" ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block", "release": "vNext"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("original object does not contain the key"))
		})

		It("throw an error when entry is already in target list", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "first", "release": "v1"},
				map[interface{}]interface{}{"name": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( insert after name \"second\" ))",
				map[interface{}]interface{}{"name": "new-kid-on-the-block", "release": "vNext"},
				map[interface{}]interface{}{"name": "second", "release": "vNext"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("because new list entry 'name: second' is detected multiple times"))
		})
	})

	Context("with element '(( delete ))' deletes data where wanted", func() {
		It("Delete entry #0", func() {
			orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
			array := []interface{}{"(( delete 0 ))"}
			expect := []interface{}{"second", "third", "fourth", "fifth", "sixth"}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Delete entry #4", func() {
			orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
			array := []interface{}{"(( delete 4 ))"}
			expect := []interface{}{"first", "second", "third", "fourth", "sixth"}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("throw an error if delete point is out of bounds", func() {
			orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
			array := []interface{}{"(( delete 6 ))"}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to modify the list, because specified index 6 is out of bounds"))
		})

		It("throw an error if delete point is negative", func() {
			orig := []interface{}{"first", "second", "third", "fourth", "fifth", "sixth"}
			array := []interface{}{"(( delete -2 ))"}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to modify the list, because specified index -2 is out of bounds"))
		})

		It("Delete '<default>: first'", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "first", "release": "v1"},
				map[interface{}]interface{}{"name": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( delete \"first\" ))",
			}

			expect := []interface{}{
				map[interface{}]interface{}{"name": "second", "release": "v1"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Delete '<default>: first' (no quotes)", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "first", "release": "v1"},
				map[interface{}]interface{}{"name": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( delete first ))",
			}

			expect := []interface{}{
				map[interface{}]interface{}{"name": "second", "release": "v1"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Delete 'id: second'", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( delete id \"second\" ))",
			}

			expect := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Allow inquoted names in delete", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( delete id second ))",
			}

			expect := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("delete multiple entries, second and first", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( delete id \"second\" ))",
				"(( delete id \"first\" ))",
			}

			expect := []interface{}{}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("delete multiple entries in different batches", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
				map[interface{}]interface{}{"id": "third", "release": "v1"},
				map[interface{}]interface{}{"id": "fourth", "release": "v1"},
				map[interface{}]interface{}{"id": "fifth", "release": "v1"},
				map[interface{}]interface{}{"id": "sixth", "release": "v1"},
				map[interface{}]interface{}{"id": "seventh", "release": "v1"},
				map[interface{}]interface{}{"id": "eighth", "release": "v1"},
			}

			array := []interface{}{
				"(( delete id \"second\" ))",
				"(( delete id \"fourth\" ))",
				"(( delete id \"sixth\" ))",
			}

			expect := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "third", "release": "v1"},
				map[interface{}]interface{}{"id": "fifth", "release": "v1"},
				map[interface{}]interface{}{"id": "seventh", "release": "v1"},
				map[interface{}]interface{}{"id": "eighth", "release": "v1"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("delete multiple entries, together with different modification styles", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "second", "release": "v1"},
				map[interface{}]interface{}{"id": "third", "release": "v1"},
				map[interface{}]interface{}{"id": "fourth", "release": "v1"},
				map[interface{}]interface{}{"id": "fifth", "release": "v1"},
				map[interface{}]interface{}{"id": "sixth", "release": "v1"},
				map[interface{}]interface{}{"id": "seventh", "release": "v1"},
				map[interface{}]interface{}{"id": "eighth", "release": "v1"},
			}

			array := []interface{}{
				"(( delete id \"second\" ))",
				"(( prepend ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-3", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-4", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-5", "release": "vNext"},
				"(( append ))",
				map[interface{}]interface{}{"id": "new-kid-on-the-block-6", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-7", "release": "vNext"},
				"(( delete id \"fourth\" ))",
				"(( delete 0 ))",
			}

			expect := []interface{}{
				map[interface{}]interface{}{"id": "new-kid-on-the-block-4", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-5", "release": "vNext"},
				map[interface{}]interface{}{"id": "first", "release": "v1"},
				map[interface{}]interface{}{"id": "third", "release": "v1"},
				map[interface{}]interface{}{"id": "fifth", "release": "v1"},
				map[interface{}]interface{}{"id": "sixth", "release": "v1"},
				map[interface{}]interface{}{"id": "seventh", "release": "v1"},
				map[interface{}]interface{}{"id": "eighth", "release": "v1"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-6", "release": "vNext"},
				map[interface{}]interface{}{"id": "new-kid-on-the-block-7", "release": "vNext"},
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(Equal(expect))
			Expect(err).NotTo(HaveOccurred())
		})

		It("throw an error when delete point cannot be found", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "first", "release": "v1"},
				map[interface{}]interface{}{"name": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( delete name \"not-existing\" ))",
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find specified modification point with 'name: not-existing'"))
		})

		It("throw an error when key cannot be found in original list", func() {
			orig := []interface{}{
				map[interface{}]interface{}{"name": "first", "release": "v1"},
				map[interface{}]interface{}{"name": "second", "release": "v1"},
			}

			array := []interface{}{
				"(( delete id \"second\" ))",
			}

			m := &Merger{}
			a := m.mergeArray(orig, array, "node-path")
			err := m.Error()
			Expect(a).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("original object does not contain the key"))
		})
	})
})

var _ = Describe("Merge()", func() {
	var (
		YAML = func(s string) map[interface{}]interface{} {
			y, err := simpleyaml.NewYaml([]byte(s))
			Expect(err).NotTo(HaveOccurred())

			data, err := y.Map()
			Expect(err).NotTo(HaveOccurred())

			return data
		}

		valueIs = func(t interface{}, path string, expect string) {
			c, err := tree.ParseCursor(path)
			Expect(err).NotTo(HaveOccurred())

			v, err := c.ResolveString(t)
			Expect(err).NotTo(HaveOccurred())

			Expect(v).To(Equal(expect))
		}

		notPresent = func(t interface{}, path string) {
			c, err := tree.ParseCursor(path)
			Expect(err).NotTo(HaveOccurred())

			_, err = c.ResolveString(t)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not be found"))
		}
	)

	It("leaves original object untouched when merging", func() {
		template := YAML(`props:
  toplevel: TEMPLATE VALUE
  sub:
    key: ANOTHER TEMPLATE VALUE
`)
		other := YAML(`props:
  toplevel: override
`)

		merged, err := Merge(template, other)
		Expect(err).NotTo(HaveOccurred())

		valueIs(template, "props.toplevel", "TEMPLATE VALUE")
		valueIs(template, "props.sub.key", "ANOTHER TEMPLATE VALUE")

		valueIs(other, "props.toplevel", "override")
		notPresent(other, "props.sub.key")

		valueIs(merged, "props.toplevel", "override")
		valueIs(merged, "props.sub.key", "ANOTHER TEMPLATE VALUE")
	})

	It("handles array operators when previous data was nil", func() {
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
		Expect(err).NotTo(HaveOccurred())

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
})
