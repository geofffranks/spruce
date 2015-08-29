# Added support for static_ips() and referencing multiple values at once

You can now do things like this:

```yml
jobs:
- name: api_z1
  instances: 1
  networks:
  - name: net1
    static_ips: (( static_ips(1, 2, 3) ))
- name: api_z2
  instances: 1
  networks:
  - name: net2
    static_ips: (( static_ips(1, 2, 3) ))

networks:
- name: net1
  subnets:
  - cloud_properties: random
    static:
    - 192.168.1.2 - 192.168.1.30
- name: net2
  subnets:
  - cloud_properties: random
    static:
    - 192.168.2.2 - 192.168.2.30

properties:
  api_servers: (( grab jobs.api_z1.networks.net1.static_ips jobs.api_z2.networks.net2.static_ips ))
```

And get back the following YAML:

```yml
jobs:
- instances: 1
  name: api_z1
  networks:
  - name: net1
    static_ips:
    - 192.168.1.2
    - 192.168.1.3
    - 192.168.1.4
- instances: 1
  name: api_z2
  networks:
  - name: net2
    static_ips:
    - 192.168.2.2
    - 192.168.2.3
    - 192.168.2.4
networks:
- name: net1
  subnets:
  - cloud_properties: random
    static:
    - 192.168.1.2 - 192.168.1.30
- name: net2
  subnets:
  - cloud_properties: random
    static:
    - 192.168.2.2 - 192.168.2.30
properties:
  api_servers:
  - 192.168.1.2
  - 192.168.1.3
  - 192.168.1.4
  - 192.168.2.2
  - 192.168.2.3
  - 192.168.2.4
```
