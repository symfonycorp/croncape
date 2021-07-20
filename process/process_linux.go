package process

import (
	"os/exec"
	"syscall"
)

func Deathsig(sysProcAttr *syscall.SysProcAttr) {
	// the following helps with killing the main process and its children
	// see https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
	sysProcAttr.Setpgid = true
	sysProcAttr.Pdeathsig = syscall.SIGKILL
}

func Kill(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
