package spruce

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	. "github.com/geofffranks/spruce/log"
	"github.com/jhunt/tree"
)

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
func (StaticIPOperator) Dependencies(ev *Evaluator, _ []*Expr, _ []*tree.Cursor) []*tree.Cursor {
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

	// need all the network name decls
	track("networks.*.name")

	// need all the static range decls
	track("networks.*.subnets.*.static")

	// need all the job instance count decls
	track("jobs.*.instances")
	track("instance_groups.*.instances")

	// need all the job network name decls
	track("jobs.*.networks.*.name")
	track("instance_groups.*.networks.*.name")

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
		return 0, fmt.Errorf("the `instances:` for the current job is not numeric")
	}
	if i < 0 {
		return 0, fmt.Errorf("negative number found in `instances:` for the current job")
	}
	return int(i), nil
}

func statics(ev *Evaluator) ([]string, error) {
	addrs := []string{}

	c := ev.Here.Copy()
	c.Pop()
	c.Push("name")
	name, err := c.ResolveString(ev.Tree)
	if err != nil {
		return addrs, err
	}

	c, err = tree.ParseCursor(fmt.Sprintf("networks.%s.subnets.*.static.*", name))
	if err != nil {
		return addrs, err
	}

	keys, err := c.Glob(ev.Tree)
	if err != nil {
		return addrs, err
	}

	for _, key := range keys {
		r, err := key.Resolve(ev.Tree)
		if err != nil {
			return addrs, err
		}

		if _, ok := r.(string); !ok {
			return addrs, fmt.Errorf("%s is not a well-formed BOSH network", name)
		}

		segments := strings.Split(r.(string), "-")
		for i, s := range segments {
			segments[i] = strings.TrimSpace(s)
		}

		start := net.ParseIP(segments[0])
		if start == nil {
			return nil, fmt.Errorf("%s: not a valid IP address", segments[0])
		}

		addrs = append(addrs, start.String())
		if len(segments) == 1 {
			continue
		}

		end := net.ParseIP(segments[1])
		if end == nil {
			return nil, fmt.Errorf("%s: not a valid IP address", segments[1])
		}

		if binary.BigEndian.Uint32(start.To4()) > binary.BigEndian.Uint32(end.To4()) {
			return nil, fmt.Errorf("Static IP pool [%s - %s] ends before it starts", start, end)
		}

		for !start.Equal(end) {
			incrementIP(start, len(start)-1)
			addrs = append(addrs, start.String())
		}
	}
	return addrs, nil
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
		return nil, fmt.Errorf("not enough static IPs requested for job of %d instances (only asked for %d)", inst, len(args))
	}
	DEBUG("  looks good.  asking for %d IPs for a job with %d instances\n", len(args), inst)

	// find our network
	DEBUG("  determining the pool of static IPs from which to provision")
	pool, err := statics(ev)
	DEBUG("  static IP pool: %v", pool)
	if err != nil {
		DEBUG("  failed: %s\n", err)
		return nil, err
	}
	DEBUG("  found %d addresses in the pool\n", len(pool))

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
		n, ok := v.Literal.(int64)
		if !ok {
			DEBUG("  arg[%d]: '%v' is not a string literal\n", i, arg)
			return nil, fmt.Errorf("static_ips operator only accepts literal numbers for arguments")
		}
		if n < 0 {
			DEBUG("  arg[%d]: '%d' is not a positive number\n", i, n)
			return nil, fmt.Errorf("static_ips operator only accepts literal non-negative numbers for arguments")
		}

		offset := int(n)
		DEBUG("  arg[%d]: asking for the %d%s IP from the static address pool", i, offset, ord(offset))
		if offset >= len(pool) {
			DEBUG("     [%d]: pool only has %d addresses; offset %d is out of bounds\n", i, len(pool), offset)
			return nil, fmt.Errorf("request for static_ip(%d) in a pool of only %d (zero-indexed) static addresses", offset, len(pool))
		}

		// check to see if the address is already claimed
		ip := pool[offset]
		DEBUG("     [%d]: checking to see if %s is already claimed", i, ip)
		if thief, taken := UsedIPs[ip]; taken {
			DEBUG("     [%d]: %s is in use by %s\n", i, ip, thief)
			return nil, fmt.Errorf("tried to use IP '%s', but that address is already allocated to %s", ip, thief)
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
