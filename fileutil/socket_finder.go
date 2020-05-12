package fileutil

import (
	"fmt"
	"os/exec"
	"strings"
)

func ListSockets() (string, error) {
	cmd := exec.Command("bash", "-c", "lsof -U | grep testmanagerd")
	out, err := cmd.Output()
	return string(out), err
}

func FirstSocket() (string, error) {
	list, err := ListSockets()
	if err != nil {
		return "", err
	}
	index1 := strings.Index(list, "/private/tmp/com.apple.launchd.")
	index2 := strings.Index(list, "unix-domain.socket")
	if index1 == -1 || index2 == -1 {
		return "", fmt.Errorf("no socket found in list: %s", list)
	}
	return list[index1 : index2+len("unix-domain.socket")], nil

}
