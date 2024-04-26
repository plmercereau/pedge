package api

import (
    "log"
    "net"
	"sync"

    "github.com/insomniacslk/dhcp/dhcpv4"
    "github.com/insomniacslk/dhcp/dhcpv4/server4"
	"netboot-provider/dhcp"
)

// StartDHCPServer starts a DHCP server that listens on all interfaces
func StartDHCPServer(machines *sync.Map) {
    laddr := net.UDPAddr{
        IP:   net.ParseIP("0.0.0.0"),
        Port: 67,
    }

    server, err := server4.NewServer("", &laddr,
        func(conn net.PacketConn, peer net.Addr, m *dhcpv4.DHCPv4) {
            log.Printf("Received DHCP %s from %s", m.MessageType(), m.ClientHWAddr)
            if m.MessageType() != dhcpv4.MessageTypeDiscover {
                return
            }

            log.Printf("DHCP Discover !!!")

            if archTypeData := m.GetOneOption(dhcpv4.OptionClientSystemArchitectureType); archTypeData != nil {
				archType, err := dhcp.BytesToArchType(archTypeData)
				if err != nil {
					log.Printf("Error decoding architecture type: %v", err)
					return
				}
				log.Printf("OptionClientSystemArchitectureType: %s", archType.String())
				
				machines.Store(m.ClientHWAddr.String(), archType)

				// * Log all machines
				machines.Range(func(key, value interface{}) bool {
					log.Printf("%s: %s\n", key, value)
					return true
				})

            }
        })

    if err != nil {
        log.Fatalf("Failed to start DHCP server: %v", err)
    }

    log.Println("DHCP server is listening on all interfaces")
    if err := server.Serve(); err != nil {
        log.Fatalf("DHCP server failed: %v", err)
    }
}
