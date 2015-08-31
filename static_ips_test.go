package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestStaticIPPostProcess(t *testing.T) {
	Convey("staticIPGenerator.PostProcess()", t, func() {
		goodRoot := map[interface{}]interface{}{
			"jobs": []interface{}{
				map[interface{}]interface{}{
					"name":      "staticIPs_z1",
					"instances": 1,
				},
			},
			"networks": []interface{}{
				map[interface{}]interface{}{
					"name": "net_z1",
					"subnets": []interface{}{
						map[interface{}]interface{}{
							"type": "manual",
							"static": []interface{}{
								"10.0.0.5",
							},
						},
					},
				},
			},
		}
		s := StaticIPGenerator{root: goodRoot}
		Convey("Ignores non-string objects", func() {
			val, action, err := s.PostProcess(1234, "nodepath")
			So(action, ShouldEqual, "ignore")
			So(val, ShouldBeNil)
			So(err, ShouldBeNil)
		})
		Convey("Ignores string objects not matching the static_ips() call", func() {
			val, action, err := s.PostProcess("1234", "nodepath")
			So(action, ShouldEqual, "ignore")
			So(val, ShouldBeNil)
			So(err, ShouldBeNil)
		})
		Convey("Errors if instancesForNode() errored", func() {
			val, action, err := s.PostProcess("(( static_ips(0) ))", "jobs.fakeJob_z1.networks.net_z1.static_ips")
			So(action, ShouldEqual, "error")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.fakeJob_z1.networks.net_z1.static_ips: `$.jobs.fakeJob_z1` could not be found in the YAML datastructure")
		})
		Convey("Errors if offsetsForNode() errored", func() {
			val, action, err := s.PostProcess("(( static_ips(abcdefg) ))", "jobs.staticIPs_z1.networks.net_z1.static_ips")
			So(action, ShouldEqual, "error")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "jobs.staticIPs_z1.networks.net_z1.static_ips: ")
		})
		Convey("Errors if instances exceeds defined offsets", func() {
			badRoot := map[interface{}]interface{}{
				"jobs": []interface{}{
					map[interface{}]interface{}{
						"name":      "staticIPs_z1",
						"instances": 5,
					},
				},
				"networks": []interface{}{
					map[interface{}]interface{}{
						"name": "net_z1",
						"subnets": []interface{}{
							map[interface{}]interface{}{
								"type": "manual",
								"static": []interface{}{
									"10.0.0.5",
								},
							},
						},
					},
				},
			}
			s.root = badRoot
			val, action, err := s.PostProcess("(( static_ips(0) ))", "jobs.staticIPs_z1.networks.net_z1.static_ips")
			s.root = goodRoot
			So(action, ShouldEqual, "error")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIPs_z1.networks.net_z1.static_ips: Not enough static IP offsets are defined (need at least 5 for the proposed manifest)")
		})
		Convey("Errors if ipRangesForNode() errored", func() {
			val, action, err := s.PostProcess("(( static_ips(0) ))", "jobs.staticIPs_z1.networks.fakeNet_z1.static_ips")
			So(action, ShouldEqual, "error")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIPs_z1.networks.fakeNet_z1.static_ips: `$.networks.fakeNet_z1` could not be found in the YAML datastructure")
		})
		Convey("Errors if offsets exceeds IP Pool", func() {
			val, action, err := s.PostProcess("(( static_ips(0, 1) ))", "jobs.staticIPs_z1.networks.net_z1.static_ips")
			So(action, ShouldEqual, "error")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIPs_z1.networks.net_z1.static_ips: Not enough static IP are defined in the IP ranges (need at least 2 for the proposed manifest)")
		})
		Convey("Errors if you specify an offset greater than the availability in the IP Pool", func() {
			val, action, err := s.PostProcess("(( static_ips(1) ))", "jobs.staticIPs_z1.networks.net_z1.static_ips")
			So(action, ShouldEqual, "error")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "jobs.staticIPs_z1.networks.net_z1.static_ips: You tried to use static_ip offset '1' but only have 1 static IPs available")
		})
		Convey("Errors if staticIPPool() errored", func() {
			badRoot := map[interface{}]interface{}{
				"jobs": []interface{}{
					map[interface{}]interface{}{
						"name":      "staticIPs_z1",
						"instances": 1,
					},
				},
				"networks": []interface{}{
					map[interface{}]interface{}{
						"name": "net_z1",
						"subnets": []interface{}{
							map[interface{}]interface{}{
								"type": "manual",
								"static": []interface{}{
									"502.294.312.324",
								},
							},
						},
					},
				},
			}
			s.root = badRoot
			val, action, err := s.PostProcess("(( static_ips(0) ))", "jobs.staticIPs_z1.networks.net_z1.static_ips")
			s.root = goodRoot

			So(action, ShouldEqual, "error")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIPs_z1.networks.net_z1.static_ips: Could not parse an IP out of '502.294.312.324'")
		})
		Convey("Errors if IPs get double-defined", func() {
			val, action, err := s.PostProcess("(( static_ips(0) ))", "jobs.staticIPs_z1.networks.net_z1.static_ips")
			So(action, ShouldEqual, "replace")
			So(val, ShouldResemble, []interface{}{"10.0.0.5"})
			So(err, ShouldBeNil)

			val, action, err = s.PostProcess("(( static_ips(0) ))", "jobs.staticIPs_z1.networks.net_z1.static_ips")
			So(action, ShouldEqual, "error")
			So(val, ShouldResemble, nil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIPs_z1.networks.net_z1.static_ips: Tried to use IP '10.0.0.5', but that is already in use by jobs.staticIPs_z1.networks.net_z1.static_ips")
		})
		Convey("Returns staticIP list, replace, no error for succssful static_ips()", func() {
			seenIP = map[string]string{}
			val, action, err := s.PostProcess("(( static_ips(0) ))", "jobs.staticIPs_z1.networks.net_z1.static_ips")
			So(action, ShouldEqual, "replace")
			So(val, ShouldResemble, []interface{}{"10.0.0.5"})
			So(err, ShouldBeNil)
		})
		Convey("Returns staticIP list, replace, no error for succssful static_ips() (whitespace optional)", func() {
			seenIP = map[string]string{}
			val, action, err := s.PostProcess("((static_ips(0)))", "jobs.staticIPs_z1.networks.net_z1.static_ips")
			So(action, ShouldEqual, "replace")
			So(val, ShouldResemble, []interface{}{"10.0.0.5"})
			So(err, ShouldBeNil)
		})
	})
}

func TestOffsetsforNode(t *testing.T) {
	Convey("offsetsforNode()", t, func() {
		Convey("Errors when passed a non-comma-separated string of non-ints", func() {
			val, err := offsetsForNode("the quick brown fox", "nodepath")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "nodepath: Found 'the quick brown fox' as an IP offset for static_ips(), but that's not an integer")
		})
		Convey("Errors when passed a space-separated string of ints", func() {
			val, err := offsetsForNode("5 4 3 2 1", "nodepath")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "nodepath: Found '5 4 3 2 1' as an IP offset for static_ips(), but that's not an integer")
		})
		Convey("Errors when passed a comma separated string of non-ints", func() {
			val, err := offsetsForNode("jumped, over, the, lazy, dog", "nodepath")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "nodepath: Found 'jumped' as an IP offset for static_ips(), but that's not an integer")
		})
		Convey("Errors when passed an empty string", func() {
			val, err := offsetsForNode("", "nodepath")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "nodepath: Found '' as an IP offset for static_ips(), but that's not an integer")
		})
		Convey("Returns []int when passed a comma separated string of ints", func() {
			val, err := offsetsForNode("1,2,3,4,5", "nodepath")
			So(err, ShouldBeNil)
			So(val, ShouldResemble, []int{1, 2, 3, 4, 5})
		})
		Convey("Returns []int when passed a comma + space separated string of ints", func() {
			val, err := offsetsForNode("1, 2, 3, 4, 5", "nodepath")
			So(err, ShouldBeNil)
			So(val, ShouldResemble, []int{1, 2, 3, 4, 5})
		})
		Convey("Returns 1 element slice when passed a single int-string", func() {
			val, err := offsetsForNode("1", "nodepath")
			So(err, ShouldBeNil)
			So(val, ShouldResemble, []int{1})
		})
	})
}

func TestIpRangesForNode(t *testing.T) {
	Convey("ipRangesForNode()", t, func() {
		Convey("Errors if node is not a job node's network's static_ips key", func() {
			root := map[interface{}]interface{}{
				"networks": []interface{}{
					map[interface{}]interface{}{
						"name": "net_z1",
						"static": []interface{}{
							"10.0.0.2 - 10.0.0.50",
							"10.0.0.64",
						},
					},
				},
			}
			val, err := ipRangesForNode("mynode", root)
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "mynode: Does not appear to be a valid key for resolving static_ips()")
		})
		Convey("Errors if the job's network's subnets could not be found", func() {
			root := map[interface{}]interface{}{
				"networks": []interface{}{
					map[interface{}]interface{}{
						"name": "net_z1",
						"static": []interface{}{
							"10.0.0.2 - 10.0.0.50",
							"10.0.0.64",
						},
					},
				},
			}
			val, err := ipRangesForNode("jobs.staticIP_z1.networks.undefined_z1.static_ips", root)
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIP_z1.networks.undefined_z1.static_ips: `$.networks.undefined_z1` could not be found in the YAML datastructure")
		})
		Convey("Errors if the job's network's subnets was not an array", func() {
			root := map[interface{}]interface{}{
				"networks": []interface{}{
					map[interface{}]interface{}{
						"name":    "net_z1",
						"subnets": 1234,
					},
				},
			}
			val, err := ipRangesForNode("jobs.staticIP_z1.networks.net_z1.static_ips", root)
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIP_z1.networks.net_z1.static_ips: `$.networks.net_z1.subnets` was not an array")
		})
		Convey("Errors if the subnet is not a map", func() {
			root := map[interface{}]interface{}{
				"networks": []interface{}{
					map[interface{}]interface{}{
						"name": "net_z1",
						"subnets": []interface{}{
							1234,
						},
					},
				},
			}
			val, err := ipRangesForNode("jobs.staticIP_z1.networks.net_z1.static_ips", root)
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIP_z1.networks.net_z1.static_ips: `$.networks.net_z1.subnets.[0]` was not a subnet map")
		})
		Convey("Errors if the static IP ranges for the subnet could not be found", func() {
			root := map[interface{}]interface{}{
				"networks": []interface{}{
					map[interface{}]interface{}{
						"name": "net_z1",
						"subnets": []interface{}{
							map[interface{}]interface{}{
								"type": "manual",
							},
						},
					},
				},
			}
			val, err := ipRangesForNode("jobs.staticIP_z1.networks.net_z1.static_ips", root)
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIP_z1.networks.net_z1.static_ips: `$.networks.net_z1.subnets.[0].static` could not be found")
		})
		Convey("Errors if the static IP ranges for the subnet is not an array", func() {
			root := map[interface{}]interface{}{
				"networks": []interface{}{
					map[interface{}]interface{}{
						"name": "net_z1",
						"subnets": []interface{}{
							map[interface{}]interface{}{
								"type":   "manual",
								"static": 1234,
							},
						},
					},
				},
			}
			val, err := ipRangesForNode("jobs.staticIP_z1.networks.net_z1.static_ips", root)
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIP_z1.networks.net_z1.static_ips: `$.networks.net_z1.subnets.[0].static` is not an array")
		})
		Convey("Errors if a static IP range is not a string", func() {
			root := map[interface{}]interface{}{
				"networks": []interface{}{
					map[interface{}]interface{}{
						"name": "net_z1",
						"subnets": []interface{}{
							map[interface{}]interface{}{
								"type": "manual",
								"static": []interface{}{
									1234,
									"1.2.3.4",
								},
							},
						},
					},
				},
			}
			val, err := ipRangesForNode("jobs.staticIP_z1.networks.net_z1.static_ips", root)
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIP_z1.networks.net_z1.static_ips: `$.networks.net_z1.subnets.[0].static.[0]` was not a string")
		})
		Convey("Returns the list of IP ranges from the job's network's subnet's static definition", func() {
			root := map[interface{}]interface{}{
				"networks": []interface{}{
					map[interface{}]interface{}{
						"name": "net_z1",
						"subnets": []interface{}{
							map[interface{}]interface{}{
								"type": "manual",
								"static": []interface{}{
									"10.0.0.5",
									"10.0.0.10 - 10.0.0.32",
								},
							},
						},
					},
				},
			}
			val, err := ipRangesForNode("jobs.staticIP_z1.networks.net_z1.static_ips", root)
			So(val, ShouldResemble, []string{"10.0.0.5", "10.0.0.10 - 10.0.0.32"})
			So(err, ShouldBeNil)
		})
	})
}

func TestInstancesForNode(t *testing.T) {
	Convey("instancesForNode()", t, func() {
		Convey("Errors when passed a non-job node", func() {
			root := map[interface{}]interface{}{
				"jobs": []interface{}{
					map[interface{}]interface{}{
						"name":      "staticIPs_z1",
						"instances": 5,
					},
				},
			}
			val, err := instancesForNode("mynode", root)
			So(val, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "mynode: Does not appear to be a valid key for resolving static_ips()")
		})
		Convey("Errors when the job instances cannot be found", func() {
			root := map[interface{}]interface{}{
				"jobs": []interface{}{
					map[interface{}]interface{}{
						"name":      "staticIPs_z1",
						"instances": 5,
					},
				},
			}
			val, err := instancesForNode("jobs.undefined_z1.networks.net1.static_ips", root)
			So(val, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.undefined_z1.networks.net1.static_ips: `$.jobs.undefined_z1` could not be found in the YAML datastructure")
		})
		Convey("Errors when the instances are not a number", func() {
			root := map[interface{}]interface{}{
				"jobs": []interface{}{
					map[interface{}]interface{}{
						"name":      "staticIPs_z1",
						"instances": map[string]string{},
					},
				},
			}
			val, err := instancesForNode("jobs.staticIPs_z1.networks.net1.static_ips", root)
			So(val, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIPs_z1.networks.net1.static_ips: `$.jobs.staticIPs_z1.instances` did not resolve to an integer")
		})
		Convey("Errors when the instances string could not be converted to a number", func() {
			root := map[interface{}]interface{}{
				"jobs": []interface{}{
					map[interface{}]interface{}{
						"name":      "staticIPs_z1",
						"instances": "five",
					},
				},
			}
			val, err := instancesForNode("jobs.staticIPs_z1.networks.net1.static_ips", root)
			So(val, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.staticIPs_z1.networks.net1.static_ips: `$.jobs.staticIPs_z1.instances` did not resolve to an integer")
		})
		Convey("Returns int when instances was found to be an integer", func() {
			root := map[interface{}]interface{}{
				"jobs": []interface{}{
					map[interface{}]interface{}{
						"name":      "staticIPs_z1",
						"instances": 5,
					},
				},
			}
			val, err := instancesForNode("jobs.staticIPs_z1.networks.net1.static_ips", root)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 5)
		})
		Convey("Returns an int when instances was found to be an integer-string", func() {
			root := map[interface{}]interface{}{
				"jobs": []interface{}{
					map[interface{}]interface{}{
						"name":      "staticIPs_z1",
						"instances": "5",
					},
				},
			}
			val, err := instancesForNode("jobs.staticIPs_z1.networks.net1.static_ips", root)
			So(err, ShouldBeNil)
			So(val, ShouldEqual, 5)
		})
	})
}

func TestStaticIPPool(t *testing.T) {
	Convey("staticIPPool()", t, func() {
		Convey("Errors if given an invalid IP", func() {
			val, err := staticIPPool([]string{"300.432.231.432"})
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Could not parse an IP out of '300.432.231.432'")
		})
		Convey("Errors if given a range with an invalid IP", func() {
			val, err := staticIPPool([]string{"127.0.0.1 - 300.432.231.432"})
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Could not parse an IP out of '300.432.231.432'")
		})
		Convey("Errors if given an empty string", func() {
			val, err := staticIPPool([]string{""})
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Could not parse an IP out of ''")
		})
		Convey("Returns an empty list of IPs if no ranges were provided", func() {
			expect := []string(nil)
			val, err := staticIPPool([]string{})
			So(val, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
		Convey("Returns a list of IPs when provided IPs + ranges", func() {
			expect := []string{
				"192.168.0.1",
				"192.168.0.50",
				"192.168.0.51",
				"192.168.0.52",
				"192.168.0.53",
				"192.168.0.54",
				"192.168.0.55",
				"192.168.0.56",
				"192.168.0.57",
				"192.168.0.58",
				"192.168.0.59",
				"192.168.0.60",
				"127.0.0.1",
				"127.0.0.2",
				"10.10.10.10",
			}
			val, err := staticIPPool([]string{"192.168.0.1", "192.168.0.50 - 192.168.0.60", "127.0.0.1 - 127.0.0.2", "10.10.10.10"})
			So(val, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
	})
}
