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
	id   int // DRBD node-id
	name string
	ip   string
}

// Volume is a DRBD volume
type volume struct {
	id            int // DRBD volume ID
	backingDevice string
	minor         int
}

// Resource is a DRBD resource
type Resource struct {
	name   string
	port   int
	volume []volume
	host   []host

	sync.Mutex
}

// NewResource returns a new DRBD resource object
func NewResource(name string, port int) *Resource {
	return &Resource{name: name, port: port}
}

func checkVolumes(r *Resource, v volume) error {
	for _, rv := range r.volume {
		if rv.id == v.id {
			return fmt.Errorf("Resource '%s' already contains volume with ID: '%d'", r.name, v.id)
		}
		if rv.backingDevice == v.backingDevice {
			return fmt.Errorf("Resource '%s' already contains volume with Name: '%s'", r.name, v.backingDevice)
		}
		if rv.minor == v.minor {
			return fmt.Errorf("Resource '%s' already contains volume with Minor: '%d'", r.name, v.minor)
		}
	}
	return nil
}

// AddVolume adds DRBD volume information to a resource
func (r *Resource) AddVolume(id, minor int, backingDevice string) error {
	v := volume{
		id:            id,
		minor:         minor,
		backingDevice: backingDevice,
	}

	r.Lock()
	defer r.Unlock()

	if err := checkVolumes(r, v); err != nil {
		return err
	}
	r.volume = append(r.volume, v)

	return nil
}

func checkHosts(r *Resource, h host) error {
	for _, rh := range r.host {
		if rh.id == h.id {
			return fmt.Errorf("Resource '%s' already contains host with ID: '%d'", r.name, h.id)
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
func (r *Resource) AddHost(id int, name, ip string) error {
	h := host{
		id:   id,
		name: name,
		ip:   ip,
	}

	r.Lock()
	defer r.Unlock()

	if err := checkHosts(r, h); err != nil {
		return err
	}
	r.host = append(r.host, h)

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
		for _, v := range r.volume {
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
