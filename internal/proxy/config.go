package proxy

import (
	"errors"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func SetProxy(port int) error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	return setProxyState(port, "on")
}

func UnsetProxy() error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	return setProxyState(0, "off")
}

// Configures proxy state (set or unset)
func setProxyState(port int, state string) error {
	var err error
	network, err := getNetworkInterface()
	if err != nil {
		return err
	}

	switch state {
	case "on":
		host := "127.0.0.1"
		port := strconv.FormatUint(uint64(port), 10)
		if _, err = runCommand("-setwebproxy", network, host, port); err != nil {
			return err
		}
		if _, err = runCommand("-setsecurewebproxy", network, host, port); err != nil {
			return err
		}
	case "off":
		if _, err = runCommand("-setwebproxystate", network, state); err != nil {
			return err
		}
		if _, err = runCommand("-setsecurewebproxystate", network, state); err != nil {
			return err
		}
	}

	return nil
}

// Retrieves the default network interface
func getNetworkInterface() (string, error) {
	if runtime.GOOS != "darwin" {
		return "", nil
	}

	cmd := "networksetup -listnetworkserviceorder | grep" +
		" `(route -n get default | grep 'interface' || route -n get -inet6 default | grep 'interface') | cut -d ':' -f2`" +
		" -B 1 | head -n 1 | cut -d ' ' -f 2-"
	network, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "", err
	} else if len(network) == 0 {
		return "", errors.New("no available networks")
	}
	return strings.TrimSpace(string(network)), nil
}

func runCommand(args ...string) (string, error) {
	cmd := exec.Command("networksetup", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
