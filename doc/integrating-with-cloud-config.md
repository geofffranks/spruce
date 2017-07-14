## How can I use Spruce with BOSH's Cloud Config?

With the introduction of Cloud Config in BOSH, a lot of data
moved out of `spruce` templates used to generate BOSH manifests.
For the most part, everything just behaves as normal. The only
exception is the `(( static_ips ))` operator. It relies on
the `networks` key to be defined, to help it determine what
IP ranges are available to pick static IPs from.

If your project still requires static IPs due to external constraints
like load balancers, DNS, or other BOSH deployments that do not
support links, you can work around this by downloading your Cloud Config,
and merging it in with spruce, using a few `--prune` operators to clean
up when complete:

```
# generate base manifest
$ cat <<EOF base.yml
instance_groups:
- name: test_vm
  networks:
  - name: my_network
    static_ips: (( static_ips(0) ))

# clean up the cloud-config related attributes when done merginge,
# so as not to confuse BOSH
azs:           (( prune ))
compilation:   (( prune ))
disk_types:    (( prune ))
networks:      (( prune ))
vm_extensions: (( prune ))
vm_types:      (( prune ))

EOF

# Downloads a cloud-config.yml that includes a network definition
# with static IPs for `my_network`
$ bosh cloud-config > cloud-config.yml
Acting as user 'admin' on 'Bosh Lite Director'

# make sure to merge cloud-config first
$ spruce merge base.yml cloud-config.yml
instance_groups:
- name: test_vm
  networks:
  - name: my_network
    static_ips: [ 10.0.1.2 ]
```

AZ support behaves as it always has, `spruce` will cross-reference the
AZs defined for an `instance_group` with the `az` or `azs` defined for
the `network`, and pull an IP out of the list of available IPs for those
zones.
