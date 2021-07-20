package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/symfonycorp/croncape/process"
)

var version = "dev"

type request struct {
	command   []string
	emails    string
	from      string
	timeout   time.Duration
	transport string
	verbose   bool
}

type result struct {
	request request
	stdout  bytes.Buffer
	stderr  bytes.Buffer
	started time.Time
	stopped time.Time
	killed  bool
	code    int
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	req := request{
		emails: os.Getenv("MAILTO"),
		from:   os.Getenv("MAILFROM"),
	}

	var versionf bool
	flag.DurationVar(&req.timeout, "t", 0, `Timeout for the command, like "-t 2h", "-t 2m", or "-t 30s". After the timeout, the command is killed, disabled by default`)
	flag.StringVar(&req.transport, "p", "auto", `Transport to use, like "-p auto", "-p mail", "-p sendmail"`)
	flag.BoolVar(&req.verbose, "v", false, "Enable sending emails even if command is successful")
	flag.BoolVar(&versionf, "version", false, "Output the version")
	flag.Parse()

	if versionf {
		fmt.Println(version)
		os.Exit(0)
	}

	req.command = flag.Args()
	if len(req.command) == 0 {
		fmt.Println("You must pass a command to execute")
		os.Exit(1)
	}

	r := execCmd(wd, req)

	if r.killed || r.code != 0 || r.request.verbose {
		if r.request.emails == "" {
			fmt.Println(r.render().String())
		} else {
			r.sendEmail()
		}
	}
}

func execCmd(path string, req request) result {
	r := result{
		started: time.Now(),
		request: req,
	}
	cmd := exec.Command(req.command[0], req.command[1:]...)
	cmd.Dir = path
	cmd.Stdout = &r.stdout
	cmd.Stderr = &r.stderr
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	process.Deathsig(cmd.SysProcAttr)
	if err := cmd.Start(); err != nil {
		r.stderr.WriteString("\n" + err.Error() + "\n")
		r.code = 127
	} else {
		var timer *time.Timer
		if req.timeout > 0 {
			timer = time.NewTimer(req.timeout)
			defer timer.Stop()
			go func(timer *time.Timer, cmd *exec.Cmd) {
				for range timer.C {
					r.killed = true
					if err := process.Kill(cmd); err != nil {
						r.stderr.WriteString(fmt.Sprintf("\nUnabled to kill the process: %s\n", err))
					}
				}
			}(timer, cmd)
		}

		if err := cmd.Wait(); err != nil {
			// unsuccessful exit code?
			r.code = -1
			if exitError, ok := err.(*exec.ExitError); ok {
				r.code = exitError.Sys().(syscall.WaitStatus).ExitStatus()
			}
		}
	}

	r.stopped = time.Now()

	return r
}

func (r *result) sendEmail() {
	emails := strings.Split(r.request.emails, ",")
	paths := make(map[string]string)

	if r.request.transport == "auto" {
		paths = map[string]string{"sendmail": "sendmail", "/usr/sbin/sendmail": "sendmail", "mail": "mail", "/usr/bin/mail": "mail"}
	} else if r.request.transport == "sendmail" {
		paths = map[string]string{"sendmail": "sendmail", "/usr/sbin/sendmail": "sendmail"}
	} else if r.request.transport == "mail" {
		paths = map[string]string{"mail": "mail", "/usr/bin/mail": "mail"}
	} else {
		fmt.Printf("Unsupported transport %s\n", r.request.transport)
		os.Exit(1)
	}

	var err error
	var transportType string
	var transportPath string
	for p, t := range paths {
		p, err = exec.LookPath(p)
		if err == nil {
			transportType = t
			transportPath = p
			break
		}
	}

	switch transportType {
	default:
		fmt.Printf("Unable to find a path for %s\n", r.request.transport)
		os.Exit(1)

	case "mail":
		for _, email := range emails {
			args := []string{"-s", r.subject()}
			if from := r.request.from; from != "" {
				args = append(args, "-a", from)
			}
			args = append(args, strings.TrimSpace(email))
			cmd := exec.Command(transportPath, args...)
			cmd.Stdin = r.render()
			cmd.Env = os.Environ()
			if err := cmd.Run(); err != nil {
				fmt.Printf("Could not send email to %s: %s\n", email, err)
				os.Exit(1)
			}
		}

	case "sendmail":
		var message string
		if len(emails) > 1 {
			message = fmt.Sprintf("To: %s\r\nCc: %s\r\nSubject: %s\r\n\r\n%s", emails[0], strings.Join(emails[1:], ","), r.subject(), r.render().String())
		} else {
			message = fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", emails[0], r.subject(), r.render().String())
		}
		if from := r.request.from; from != "" {
			message = fmt.Sprintf("From: %s\r\n%s", from, message)
		}
		cmd := exec.Command(transportPath, "-t")
		cmd.Stdin = strings.NewReader(message)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		if err := cmd.Run(); err != nil {
			fmt.Printf("Could not send email to %s: %s\n", emails, err)
			os.Exit(1)
		}
	}
}

func (r *result) subject() string {
	hostname := "undefined"
	var err error
	if env := os.Getenv("MAILHOST"); env != "" {
		hostname = env
	} else if hostname, err = os.Hostname(); err != nil {
		hostname = "undefined"
	}

	if r.killed {
		return fmt.Sprintf("Cron on host %s: Timeout", hostname)
	}

	if r.code == 0 {
		return fmt.Sprintf("Cron on host %s: Command Successful", hostname)
	}

	return fmt.Sprintf("Cron on host %s: Failure", hostname)
}

func (r *result) title() string {
	var msg string

	if r.killed {
		msg = "Cron timeout detected"
	} else if r.code == 0 {
		msg = "Cron success"
	} else {
		msg = "Cron failure detected"
	}

	return msg + "\n" + strings.Repeat("=", len(msg))
}

func (r *result) duration() time.Duration {
	return r.stopped.Sub(r.started)
}

func (r *result) render() *bytes.Buffer {
	tpl := template.Must(template.New("email").Parse(`{{.Title}}

{{.Command}}

METADATA
--------

Exit Code: {{.Code}}
Start:     {{.Started}}
Stop:      {{.Stopped}}
Duration:  {{.Duration}}

ERROR OUTPUT
------------

{{.Stderr}}

STANDARD OUTPUT
---------------

{{.Stdout}}
`))

	data := struct {
		Title    string
		Command  string
		Started  time.Time
		Stopped  time.Time
		Duration time.Duration
		Code     int
		Stderr   string
		Stdout   string
	}{
		Title:    r.title(),
		Command:  strings.Join(r.request.command, " "),
		Started:  r.started,
		Stopped:  r.stopped,
		Duration: r.duration(),
		Code:     r.code,
		Stderr:   r.stderr.String(),
		Stdout:   r.stdout.String(),
	}

	contents := bytes.Buffer{}
	if err := tpl.Execute(&contents, data); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return &contents
}
