package spruce

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	. "github.com/geofffranks/spruce/log"
	"github.com/starkandwayne/goutils/ansi"
	"github.com/starkandwayne/goutils/tree"
)

const UNDEFINED_AZ = "__UNDEFINED_AZ__"

// UsedIPs ...
var UsedIPs map[string]string

// StaticIPOperator ...
type StaticIPOperator struct{}

// Setup ...
func (StaticIPOperator) Setup() error {
	UsedIPs = map[string]string{}
	return nil
}

// Phase ...
func (StaticIPOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (StaticIPOperator) Dependencies(ev *Evaluator, _ []*Expr, _ []*tree.Cursor, _ []*tree.Cursor) []*tree.Cursor {
	l := []*tree.Cursor{}

	track := func(path string) {
		c, err := tree.ParseCursor(path)
		if err != nil {
			return
		}
		keys, err := c.Glob(ev.Tree)
		if err != nil {
			return
		}
		for _, alt := range keys {
			l = append(l, alt)
		}
	}

	// top level stuff
	track("networks")
	track("networks.*")
	track("jobs")
	track("jobs.*")
	track("instance_groups")
	track("instance_groups.*")

	// need all the network name decls
	track("networks.*.name")
	track("networks.*.subnets")

	// need all the static range decls
	track("networks.*.subnets.*.static")

	// need all the az decls
	track("networks.*.subnets.*.az")
	track("networks.*.subnets.*.azs")
	track("networks.*.subnets.*.static.*")

	// need all the job instance count decls
	track("jobs.*.instances")
	track("instance_groups.*.instances")

	// need all the job network name decls
	track("jobs.*.networks.*.name")
	track("instance_groups.*.networks.*.name")

	// need all the instance_group azs decls
	track("instance_groups.*.azs")
	track("instance_groups.*.azs.*")

	return l
}

func currentJob(ev *Evaluator) (*tree.Cursor, error) {
	c := ev.Here.Copy()
	for c.Depth() > 0 && c.Parent() != "jobs" && c.Parent() != "instance_groups" {
		c.Pop()
	}

	if c.Depth() == 0 {
		return nil, fmt.Errorf("not currently inside of a job definition block")
	}
	return c, nil
}

func instances(ev *Evaluator, job *tree.Cursor) (int, error) {
	c := job.Copy()
	c.Push("instances")
	inst, err := c.ResolveString(ev.Tree)
	if err != nil {
		return 0, err
	}

	i, err := strconv.ParseInt(inst, 10, 0)
	if err != nil {
		return 0, ansi.Errorf("@R{the `}@c{instances:}@R{` for the current job is not numeric}")
	}
	if i < 0 {
		return 0, ansi.Errorf("@R{negative number found in `}@c{instances:}@R{` for the current job}")
	}
	return int(i), nil
}

func statics(ev *Evaluator) (map[string][]string, []string, error) {
	addrs := map[string][]string{}
	azs := []string{}

	c := ev.Here.Copy()
	c.Pop()
	c.Push("name")
	name, err := c.ResolveString(ev.Tree)
	if err != nil {
		return addrs, azs, err
	}

	c, err = tree.ParseCursor(fmt.Sprintf("networks.%s.subnets.*", name))
	if err != nil {
		return addrs, azs, err
	}
	keys, err := c.Glob(ev.Tree)
	if err != nil {
		return addrs, azs, err
	}

	for _, key := range keys {
		r, err := key.Canonical(ev.Tree)
		if err != nil {
			return addrs, azs, err
		}

		// list of azs associated with this specific subnet
		// do not confuse with `azs`, which is a list of
		// all `azs` for the network.
		subnet_zones := []string{}

		// look for az definition in the `az` key
		c, _ = tree.ParseCursor(fmt.Sprintf("%s.az", r.String()))
		z, err := c.ResolveString(ev.Tree)
		if err == nil && len(z) > 0 {
			azs = append(azs, z) // to preserve subnet ordering
			subnet_zones = append(subnet_zones, z)
		}

		// look for az definitions in the `azs` key
		c, _ = tree.ParseCursor(fmt.Sprintf("%s.azs", r.String()))
		os, err := c.Resolve(ev.Tree)
		if err == nil {
			if zs, ok := os.([]interface{}); ok {
				for _, o := range zs {
					if z, ok := o.(string); ok && len(z) > 0 {
						azs = append(azs, z)
						subnet_zones = append(subnet_zones, z)
					}
				}
			}
		}

		// add a default zone for azs + subnet zones, if
		// this network has no zones specified
		if len(subnet_zones) == 0 {
			azs = append(azs, "z1")
			subnet_zones = append(subnet_zones, "z1")
		}

		c, err = tree.ParseCursor(fmt.Sprintf("%s.static.*", r.String()))
		if err != nil {
			return addrs, azs, err
		}
		keys, err := c.Glob(ev.Tree)
		if err != nil {
			return addrs, azs, err
		}

		for _, key := range keys {
			r, err := key.Resolve(ev.Tree)
			if err != nil {
				return addrs, azs, err
			}

			if _, ok := r.(string); !ok {
				return addrs, azs, ansi.Errorf("@c{%s} @R{is not a well-formed BOSH network}", name)
			}

			segments := strings.Split(r.(string), "-")
			for i, s := range segments {
				segments[i] = strings.TrimSpace(s)
			}

			start := net.ParseIP(segments[0])
			if start == nil {
				return nil, azs, ansi.Errorf("@c{%s}@R{: not a valid IP address}", segments[0])
			}

			for _, az := range subnet_zones {
				addrs[az] = append(addrs[az], start.String())
				if len(segments) == 1 {
					continue
				}
			}

			if len(segments) == 2 {
				end := net.ParseIP(segments[1])
				if end == nil {
					return nil, azs, ansi.Errorf("@c{%s}@R{: not a valid IP address}", segments[1])
				}

				if binary.BigEndian.Uint32(start.To4()) > binary.BigEndian.Uint32(end.To4()) {
					return nil, azs, ansi.Errorf("@R{Static IP pool }@c{[%s - %s]} @R{ends before it starts}", start, end)
				}

				for !start.Equal(end) {
					incrementIP(start, len(start)-1)
					for _, az := range subnet_zones {
						addrs[az] = append(addrs[az], start.String())
					}
				}
			}
		}
	}
	return addrs, azs, nil
}

func allIPs(pools map[string][]string, azs []string) []string {
	var ips []string
	seen := map[string]bool{}

	for _, az := range azs {
		pool, ok := pools[az]
		if !ok {
			continue
		}
		for _, ip := range pool {
			if !seen[ip] {
				ips = append(ips, ip)
				seen[ip] = true
			}
		}
	}
	return ips
}

func incrementIP(ip net.IP, i int) net.IP {
	if ip[i] == 255 {
		ip[i] = 0

		// check next octet
		if ip[i-1] == 255 {
			incrementIP(ip, i-1)
		} else {
			ip[i-1]++
		}
	} else {
		ip[i]++
	}
	return ip
}

// Run ...
func (s StaticIPOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( static_ips ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( static_ips ... )) operation at $%s\n", ev.Here)

	var ips []interface{}

	// detect what job we are in
	DEBUG("  determining what job context (( static_ips ... )) was called in")
	job, err := currentJob(ev)
	if err != nil {
		DEBUG("  failed: %s\n", err)
		return nil, err
	}
	DEBUG("  got it.  $.%s\n", job)

	job.Push("name")
	DEBUG("  extracting job name from $.%s", job)
	jobname, err := job.Resolve(ev.Tree)
	if err != nil {
		DEBUG("  job has no name.  this could be problematic.\n")
		return nil, err
	}
	job.Pop()
	DEBUG("  got it.  job is %s\n", jobname)

	job.Push("azs")
	DEBUG("  extracting azs from $.%s", job)
	var azs []string
	if zs, err := job.Resolve(ev.Tree); err == nil {
		if _, ok := zs.([]interface{}); ok {
			for _, z := range zs.([]interface{}) {
				if _, ok := z.(string); !ok {
					DEBUG("  azs %v: '%v' is not a string literal\n", zs, z)
					return nil, ansi.Errorf("@R{azs} @c{%#v} @R{must be a list of strings}", zs)
				}
				azs = append(azs, z.(string))
			}
		} else {
			DEBUG("  azs must be a list of strings\n")
			return nil, ansi.Errorf("@R{azs} @c{%#v} @R{must be a list of strings}", zs)
		}
	}
	job.Pop()
	DEBUG("  got it.  azs are %v\n", azs)

	// determine if we have any instances
	DEBUG("  determining how many instances of job %s there are", jobname)
	inst, err := instances(ev, job)
	if err != nil {
		DEBUG("  failed: %s\n", err)
		return nil, err
	}
	if inst == 0 {
		DEBUG("  no instances for this job.  skipping static IP address calculations...\n")
		return &Response{
			Type:  Replace,
			Value: ips,
		}, nil
	}
	DEBUG("  got it.  there are %d instances of %s\n", inst, jobname)

	// check to make sure instances matches requested number of static ips
	DEBUG("  checking to see if the caller asked for enough static_ips to provision all job instances (need at least %d)", inst)
	if len(args) < inst {
		DEBUG("  oops.  you asked for %d IPs for a job with %d instances\n", len(args), inst)
		return nil, ansi.Errorf("@R{not enough static IPs requested for} @c{job of %d instances} @R{(only asked for} @c{%d}@R{)}", inst, len(args))
	}
	DEBUG("  looks good.  asking for %d IPs for a job with %d instances\n", len(args), inst)

	// find our network
	DEBUG("  determining the pool of static IPs from which to provision")
	pools, poolAZs, err := statics(ev)
	DEBUG("  static IP pools: %v", pools)
	DEBUG("  static IP pool AZs: %v", poolAZs)
	if err != nil {
		DEBUG("  failed: %s\n", err)
		return nil, err
	}
	count := 0
	for _, pool := range pools {
		count += len(pool)
	}
	DEBUG("  found %d addresses in the pool\n", count)

	// verify that pools contain all specified AZs, just like BOSH
	for _, az := range azs {
		if _, ok := pools[az]; !ok {
			DEBUG("  could not find AZ %s in network AZS: %v\n", az, azs)
			return nil, ansi.Errorf("@R{could not find AZ} @c{%s} (@R{in network AZS} @c{%v})", az, azs)
		}
	}

	// if no AZs are specified on instance_groups, then just use whatever is in networks / pools
	if len(azs) == 0 {
		for _, az := range poolAZs {
			azs = append(azs, az)
		}
	}

	ord := func(n int) string {
		switch {
		case n%100 >= 11 && n%100 <= 13:
			return "th"
		case n%10 == 1:
			return "st"
		case n%10 == 2:
			return "nd"
		case n%10 == 3:
			return "rd"
		}
		return "th"
	}

	// build the list of ips, based on offsets
	for i, arg := range args {
		if i >= inst {
			break
		}

		v, err := arg.Resolve(ev.Tree)
		if err != nil {
			DEBUG("  arg[%d]: failed to resolve expression to a concrete value", i)
			DEBUG("     [%d]: error was: %s", i, err)
			return nil, err
		}

		current := fmt.Sprintf("%s/%d", jobname, i)

		// parse argument, could be in form of <az>:<number>, or just <number>
		var n int64
		az := UNDEFINED_AZ
		a, ok := v.Literal.(string)
		if !ok {
			n, ok = v.Literal.(int64)
			if !ok {
				DEBUG("  arg[%d]: '%v' is not a number literal\n", i, arg)
				return nil, fmt.Errorf("static_ips operator arguments must have format <az>:<number> or <number>")
			}
		} else {
			if strings.Contains(a, ":") {
				// must be of format <az>:<number>
				params := strings.SplitN(a, ":", 2)
				az = params[0]
				a = params[1]
			}
			n, err = strconv.ParseInt(a, 10, 64)
			if err != nil {
				DEBUG("  arg[%d]: '%v' is not a number literal\n", i, arg)
				return nil, fmt.Errorf("static_ips operator arguments must have format <az>:<number> or <number>")
			}
		}

		// get IPs to use
		pool := allIPs(pools, azs)
		if az != UNDEFINED_AZ {
			// check if az is actually in instance_groups azs
			var found bool
			for _, z := range azs {
				if az == z {
					found = true
					break
				}
			}
			if !found {
				DEBUG("  specified az %s is not in instance_groups azs %v\n", az, azs)
				return nil, ansi.Errorf("@R{could not find AZ} @c{%s} @R{in instance_groups AZS} @c{%v}", az, azs)
			}

			pool, ok = pools[az]
			if !ok {
				DEBUG("  could not find pool: %s\n", az)
				return nil, ansi.Errorf("@R{could not find AZ} @c{%s} @R{in IP pool}", az)
			}
		}

		if n < 0 {
			DEBUG("  arg[%d]: '%d' is not a positive number\n", i, n)
			return nil, fmt.Errorf("static_ips operator only accepts literal non-negative numbers for arguments")
		}

		offset := int(n)
		DEBUG("  arg[%d]: asking for the %d%s IP from the static address pool", i, offset, ord(offset))
		if offset >= len(pool) {
			DEBUG("     [%d]: pool only has %d addresses; offset %d is out of bounds\n", i, len(pool), offset)
			return nil, ansi.Errorf("@R{request for} @c{static_ip(%d)} @R{in a pool of only} @c{%d (zero-indexed)} @R{static addresses}", offset, len(pool))
		}

		// check to see if the address is already claimed
		ip := pool[offset]
		DEBUG("     [%d]: checking to see if %s is already claimed", i, ip)
		if thief, taken := UsedIPs[ip]; taken {
			DEBUG("     [%d]: %s is in use by %s\n", i, ip, thief)
			return nil, ansi.Errorf("@R{tried to use IP '}@c{%s}@R{', but that address is already allocated to} @c{%s}", ip, thief)
		}

		// claim this address for ourselves
		DEBUG("     [%d]: claiming %s for job %s", i, ip, current)
		UsedIPs[ip] = current
		ips = append(ips, ip)

		DEBUG("")
	}

	return &Response{
		Type:  Replace,
		Value: ips,
	}, nil
}

func init() {
	RegisterOp("static_ips", StaticIPOperator{})
}
