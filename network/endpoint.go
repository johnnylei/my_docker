package network

import (
	"github.com/vishvananda/netlink"
	"net"
)

type Endpoint struct {
	ID string `json:"id"`
	Device netlink.Veth `json:"device"`
	IPAddress net.IP `json:"ip_address"`
	MacAddress net.HardwareAddr `json:"mac_address"`
	PortMapping []string `json:"port_mapping"`
	NW *Network `json:"network"`
}
