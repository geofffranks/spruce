package main

import (
	"fmt"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// StaticIPGenerator is an implementation of PostProcessor to calcualte static_ips for BOSH jobs
// using the (( static_ips(x, y, x) )) syntax
type StaticIPGenerator struct {
	root map[interface{}]interface{}
}

var seenIP = map[string]string{}

// PostProcess - resolves (( static_ips() )) calls to static IPs for BOSH jobs
func (s StaticIPGenerator) PostProcess(o interface{}, node string) (interface{}, string, error) {
	if o != nil && reflect.TypeOf(o).Kind() == reflect.String {
		staticIPs := []interface{}{}
		re := regexp.MustCompile("^\\Q((\\E\\s*static_ips\\Q(\\E(.*?)\\Q)\\E\\s*\\Q))\\E$")
		if re.MatchString(o.(string)) {
			matches := re.FindStringSubmatch(o.(string))
			if matches[1] != "" {
				DEBUG("%s: Resolving static IPs", node)
				instances, err := instancesForNode(node, s.root)
				if err != nil {
					return nil, "error", err
				}
				if instances == 0 {
					DEBUG("%s: No instances of this job in play, skipping entirely", node)
					return staticIPs, "replace", nil
				}
				DEBUG("%s: Have %d instances", node, instances)
				offsets, err := offsetsForNode(matches[1], node)
				if err != nil {
					return nil, "error", err
				}
				DEBUG("%s: have %d offsets specified", node, len(offsets))
				if len(offsets) < instances {
					return nil, "error", fmt.Errorf("%s: Not enough static IP offsets are defined (need at least %d for the proposed manifest)", node, instances)
				}
				ipRanges, err := ipRangesForNode(node, s.root)
				if err != nil {
					return nil, "error", err
				}

				pool, err := staticIPPool(ipRanges)
				if err != nil {
					return nil, "error", fmt.Errorf("%s: %s", node, err.Error())
				}
				DEBUG("%s: have %d IPs in the static_ip pool", node, len(pool))

				if len(offsets) > len(pool) {
					return nil, "error", fmt.Errorf("%s: Not enough static IP are defined in the IP ranges (need at least %d for the proposed manifest)", node, len(offsets))
				}
				for _, offset := range offsets {
					if offset >= len(pool) {
						return nil, "error", fmt.Errorf("%s: You tried to use static_ip offset '%d' but only have %d static IPs available (offsets are used in a zero-based array)", node, offset, len(pool))
					}
					ip := pool[offset]
					if thief, taken := seenIP[ip]; taken {
						return nil, "error", fmt.Errorf("%s: Tried to use IP '%s', but that is already in use by %s", node, ip, thief)
					}
					seenIP[ip] = node
					staticIPs = append(staticIPs, ip)
				}

				DEBUG("%s: Static IP Pool is %#v", node, staticIPs)
				desiredIPs := staticIPs[0:instances]
				DEBUG("%s: Only have %d instances, returning IPs: %#v", node, instances, desiredIPs)
				return desiredIPs, "replace", nil
			}
			return nil, "error", fmt.Errorf("%s: Could not parse out any offsets to use for resolving static IPs from '%s'", node, o.(string))
		}
	}
	return nil, "ignore", nil
}

func offsetsForNode(offsetsString string, node string) ([]int, error) {
	offsets := []int{}
	offsetStrings := strings.Split(offsetsString, ",")
	for _, offsetString := range offsetStrings {
		offset, err := strconv.Atoi(strings.TrimSpace(offsetString))
		if err != nil {
			return nil, fmt.Errorf("%s: Found '%s' as an IP offset for static_ips(), but that's not an integer", node, offsetString)
		}
		offsets = append(offsets, offset)
	}
	return offsets, nil
}

func ipRangesForNode(node string, root map[interface{}]interface{}) ([]string, error) {
	re := regexp.MustCompile("^(?:\\Q$.\\E)?jobs\\Q.\\E.*?\\Q.\\Enetworks\\Q.\\E(.*?)\\Q.\\Estatic_ips$")
	var ipRanges []string
	if re.MatchString(node) {
		matches := re.FindStringSubmatch(node)
		if matches[1] == "" {
			return nil, fmt.Errorf("%s: Could not detect network name to resolve static_ips()", node)
		}
		findPath := fmt.Sprintf("networks.%s.subnets", matches[1])
		subnetsObj, err := resolveNode(findPath, root)
		if err != nil {
			return nil, fmt.Errorf("%s: `$.%s", node, err)
		}
		if subnets, ok := subnetsObj.([]interface{}); ok {
			for i, subnetObj := range subnets {
				findPath = fmt.Sprintf("%s.[%d]", findPath, i)
				if subnet, ok := subnetObj.(map[interface{}]interface{}); ok {
					findPath = fmt.Sprintf("%s.static", findPath)
					if rangesObj, ok := subnet["static"]; ok {
						if ranges, ok := rangesObj.([]interface{}); ok {
							for j, rangeObj := range ranges {
								findPath = fmt.Sprintf("%s.[%d]", findPath, j)
								if ipRange, ok := rangeObj.(string); ok {
									ipRanges = append(ipRanges, ipRange)
								} else {
									return nil, fmt.Errorf("%s: `$.%s` was not a string", node, findPath)
								}
							}
						} else {
							return nil, fmt.Errorf("%s: `$.%s` is not an array", node, findPath)
						}
					} else {
						return nil, fmt.Errorf("%s: `$.%s` could not be found", node, findPath)
					}
				} else {
					return nil, fmt.Errorf("%s: `$.%s` was not a subnet map", node, findPath)
				}
			}
		} else {
			return nil, fmt.Errorf("%s: `$.%s` was not an array", node, findPath)
		}
	} else {
		return nil, fmt.Errorf("%s: Does not appear to be a valid key for resolving static_ips()", node)
	}
	return ipRanges, nil
}

func instancesForNode(node string, root map[interface{}]interface{}) (int, error) {
	re := regexp.MustCompile("^(?:\\Q$.\\E)?jobs\\Q.\\E(.*?)\\Q.\\E")
	if re.MatchString(node) {
		matches := re.FindStringSubmatch(node)
		if matches[1] == "" {
			return 0, fmt.Errorf("%s: Could not detect job name to resolve static_ips()", node)
		}
		findPath := fmt.Sprintf("jobs.%s.instances", matches[1])
		instancesObj, err := resolveNode(findPath, root)
		if err != nil {
			return 0, fmt.Errorf("%s: `$.%s", node, err)
		}
		if instances, ok := instancesObj.(int); ok {
			return instances, nil
		} else if instancesStr, ok := instancesObj.(string); ok {
			instances, err := strconv.Atoi(strings.TrimSpace(instancesStr))
			if err != nil {
				return 0, fmt.Errorf("%s: `$.%s` did not resolve to an integer", node, findPath)
			}
			return instances, nil
		} else {
			return 0, fmt.Errorf("%s: `$.%s` did not resolve to an integer", node, findPath)
		}
	}
	return 0, fmt.Errorf("%s: Does not appear to be a valid key for resolving static_ips()", node)
}

func staticIPPool(ranges []string) ([]string, error) {
	var ipPool []string

	for _, r := range ranges {
		segments := strings.Split(r, "-")

		var start, end net.IP

		startString := strings.TrimSpace(segments[0])
		start = net.ParseIP(startString)
		if start == nil {
			return nil, fmt.Errorf("Could not parse an IP out of '%s'", startString)
		}

		if len(segments) == 1 {
			end = start
		} else {
			endString := strings.TrimSpace(segments[1])
			end = net.ParseIP(endString)
			if end == nil {
				return nil, fmt.Errorf("Could not parse an IP out of '%s'", endString)
			}
		}

		ipPool = append(ipPool, ipRange(start, end)...)
	}

	return ipPool, nil
}

func ipRange(a, b net.IP) []string {
	prev := a
	ips := []string{a.String()}

	for !prev.Equal(b) {
		next := net.ParseIP(prev.String())
		inc(next)
		ips = append(ips, next.String())
		prev = next
	}

	return ips
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
