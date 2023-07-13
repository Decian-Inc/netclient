package wireguard

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gravitl/netclient/ncutils"
	"github.com/gravitl/netmaker/logger"
)

// NCIface.Create - makes a new Wireguard interface for darwin users (userspace)
func (nc *NCIface) Create() error {
	return nc.createUserSpaceWG()
}

// NCIface.ApplyAddrs - applies address for darwin userspace
func (nc *NCIface) ApplyAddrs() error {

	for _, address := range nc.Addresses {
		if address.IP != nil {
			if address.IP.To4() != nil {

				cmd := exec.Command("ifconfig", nc.Name, "inet", "add", address.IP.String(), address.IP.String())
				if out, err := cmd.CombinedOutput(); err != nil {
					logger.Log(0, fmt.Sprintf("adding address command \"%v\" failed with output %s and error: ", cmd.String(), string(out)))
					continue
				}
			} else {

				cmd := exec.Command("ifconfig", nc.Name, "inet6", "add", address.IP.String(), address.IP.String())
				if out, err := cmd.CombinedOutput(); err != nil {
					logger.Log(0, fmt.Sprintf("adding address command \"%v\" failed with output %s and error: ", cmd.String(), string(out)))
					continue
				}
			}

		}
		if address.Network.IP.To4() != nil {
			cmd := exec.Command("route", "add", "-net", "-inet", address.Network.String(), address.IP.String())
			if out, err := cmd.CombinedOutput(); err != nil {
				logger.Log(0, fmt.Sprintf("failed to add route with command %s - %v", cmd.String(), string(out)))
				continue
			}
		} else {
			cmd := exec.Command("route", "add", "-net", "-inet6", address.Network.String(), address.IP.String())
			if out, err := cmd.CombinedOutput(); err != nil {
				logger.Log(0, fmt.Sprintf("failed to add route with command %s - %v", cmd.String(), out))
				continue
			}
		}

	}

	return nil
}

func SetRoutes(addrs []ifaceAddress) {
	for _, addr := range addrs {
		if addr.IP == nil || addr.Network.IP == nil || addr.Network.String() == "0.0.0.0/0" ||
			addr.Network.String() == "::/0" {
			continue
		}

		if addr.Network.IP.To4() != nil {
			cmd := exec.Command("route", "add", "-net", "-inet", addr.Network.String(), addr.IP.String())
			if out, err := cmd.CombinedOutput(); err != nil {
				logger.Log(0, fmt.Sprintf("failed to add route with command %s - %v", cmd.String(), string(out)))
				continue
			}
		} else {
			cmd := exec.Command("route", "add", "-net", "-inet6", addr.Network.String(), addr.IP.String())
			if out, err := cmd.CombinedOutput(); err != nil {
				logger.Log(0, fmt.Sprintf("failed to add route with command %s - %v", cmd.String(), out))
				continue
			}
		}

	}
}

func (nc *NCIface) SetMTU() error {
	// set MTU for the interface
	cmd := exec.Command("ifconfig", nc.Name, "mtu", fmt.Sprint(nc.MTU), "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		logger.Log(0, fmt.Sprintf("failed to set mtu with command %s - %v", cmd.String(), out))
		return err
	}
	return nil
}

func (nc *NCIface) Close() {
	err := nc.Iface.Close()
	if err == nil {
		sockPath := "/var/run/wireguard/" + nc.Name + ".sock"
		if _, statErr := os.Stat(sockPath); statErr == nil {
			os.Remove(sockPath)
		}
	}

}

// DeleteOldInterface - removes named interface
func DeleteOldInterface(iface string) {
	logger.Log(3, "deleting interface", iface)
	ifconfig, err := exec.LookPath("ifconfig")
	if err != nil {
		logger.Log(0, "failed to locate ifconfig", err.Error())
	}
	if _, err := ncutils.RunCmd(ifconfig+" "+iface+" destroy", true); err != nil {
		logger.Log(0, "error removing interface", iface, err.Error())
	}
}
