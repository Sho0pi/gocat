package completion

import (
	"bytes"
	"github.com/spf13/cobra"
	"os/exec"
	"strings"
)

func AdbDevices() cobra.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return getADBDevices(), cobra.ShellCompDirectiveNoFileComp
	}
}

func RunningProcesses() cobra.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		device, _ := cmd.Flags().GetString("device")
		return getRunningProcesses(device), cobra.ShellCompDirectiveNoFileComp
	}
}

func LogLevels() cobra.CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []string{"debug", "info", "warn", "error"}, cobra.ShellCompDirectiveNoFileComp
	}
}

func getRunningProcesses(deviceID string) []cobra.Completion {
	cmd := exec.Command("adb", "-s", deviceID, "shell", "ps", "-A")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil
	}

	// TODO: fix some processes have the same name
	lines := strings.Split(out.String(), "\n")
	var processes []cobra.Completion
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 8 {
			process := fields[len(fields)-1]
			if strings.HasPrefix(process, "[") {
				// skip kernel threads
				continue
			}
			pid := fields[1]
			processes = append(processes, cobra.CompletionWithDesc(process, pid))
		}
	}

	return processes
}

func getADBDevices() []cobra.Completion {
	cmd := exec.Command("adb", "devices")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil
	}

	lines := strings.Split(out.String(), "\n")
	var devices []cobra.Completion

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 1 && fields[1] == "device" {
			devices = append(devices, fields[0])
		}
	}

	return devices
}
