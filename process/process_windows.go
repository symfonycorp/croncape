package process

import (
	"os/exec"
	"strconv"
	"syscall"
)

func Deathsig(sysProcAttr *syscall.SysProcAttr) {
}

func Kill(cmd *exec.Cmd) error {
	c := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid))
	if err := c.Run(); err == nil {
		return nil
	}
	return cmd.Process.Signal(syscall.SIGKILL)
}
