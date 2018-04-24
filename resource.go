package godrbdutils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

type host struct {
	id     int // DRBD node-id
	name   string
	ip     string
	volume map[int]volume // key: volume ID
}

// Volume is a DRBD volume
type volume struct {
	id            int // DRBD volume ID
	backingDevice string
	minor         int
}

// Resource is a DRBD resource
type Resource struct {
	name string
	port int
	host map[string]host // key: hostname

	sync.Mutex
}

// NewResource returns a new DRBD resource object
func NewResource(name string, port int) *Resource {
	return &Resource{name: name, port: port, host: make(map[string]host)}
}

func checkVolumes(h host, v volume) error {
	for _, hv := range h.volume {
		if hv.id == v.id {
			return fmt.Errorf("Host '%s' already has a volume with ID: '%d'", h.name, v.id)
		}
		if hv.backingDevice == v.backingDevice {
			return fmt.Errorf("Host '%s' already has a volume with Name: '%s'", h.name, v.backingDevice)
		}
		if hv.minor == v.minor {
			return fmt.Errorf("Host '%s' already has a volume with Minor: '%d'", h.name, v.minor)
		}
	}
	return nil
}

// AddVolume adds DRBD volume information to a resource
func (r *Resource) AddVolume(id, minor int, backingDevice, hostname string) error {
	v := volume{
		id:            id,
		minor:         minor,
		backingDevice: backingDevice,
	}

	r.Lock()
	defer r.Unlock()

	host, ok := r.host[hostname]
	if !ok {
		return fmt.Errorf("Could not find existing host with hostname: %v", hostname)
	}

	if err := checkVolumes(host, v); err != nil {
		return err
	}

	host.volume[id] = v
	r.host[hostname] = host

	return nil
}

func checkHosts(r *Resource, h host) error {
	for _, rh := range r.host {
		if rh.id == h.id {
			return fmt.Errorf("Resource '%s' already contains host with Node-ID: '%d'", r.name, h.id)
		}
		if rh.name == h.name {
			return fmt.Errorf("Resource '%s' already contains host with Name: '%s'", r.name, h.name)
		}
		if rh.ip == h.ip {
			return fmt.Errorf("Resource '%s' already contains host with IP: '%s'", r.name, h.ip)
		}
	}
	return nil
}

// AddHost adds a host information to a resource
func (r *Resource) AddHost(id int, hostname, ip string) error {
	h := host{
		id:     id,
		name:   hostname,
		ip:     ip,
		volume: make(map[int]volume),
	}

	r.Lock()
	defer r.Unlock()

	if err := checkHosts(r, h); err != nil {
		return err
	}
	r.host[hostname] = h

	return nil
}

func indentf(level int, format string, a ...interface{}) string {
	format = strings.Repeat("   ", level) + format
	return fmt.Sprintf(format, a...)
}

// WriteConfig writes the configuration of a DRBD resource to file parsable by drbd-utils
// It is up to the user to check for errors and to check if the file is valid (and to remove it if it isn't).
func (r *Resource) WriteConfig(filename string) error {
	r.Lock()
	defer r.Unlock()

	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("# meta-data-json:{\"updated\": \"%s\"}\n", time.Now().UTC()))
	b.WriteString(fmt.Sprintf("resource %s {\n", r.name))

	var hosts []string
	for _, h := range r.host {
		hosts = append(hosts, h.name)

		b.WriteString(indentf(1, "on %s {\n", h.name))
		b.WriteString(indentf(2, "node-id %d;\n", h.id))
		b.WriteString(indentf(2, "address %s:%d;\n", h.ip, r.port))
		for _, v := range h.volume {
			b.WriteString(indentf(2, "volume %d {\n", v.id))
			b.WriteString(indentf(3, "device minor %d;\n", v.minor))
			b.WriteString(indentf(3, "disk %s;\n", v.backingDevice))
			b.WriteString(indentf(3, "meta-disk internal;\n"))
			b.WriteString(indentf(2, "}\n")) // end volume section
		}
		b.WriteString(indentf(1, "}\n")) // end on section
		b.WriteString("\n")
	}

	b.WriteString(indentf(1, "connection-mesh {\n"))
	b.WriteString(indentf(2, "hosts %s;\n", strings.Join(hosts, " ")))
	b.WriteString(indentf(1, "}\n"))

	b.WriteString("}") // end resource section

	return ioutil.WriteFile(filename, b.Bytes(), 0644)
}
