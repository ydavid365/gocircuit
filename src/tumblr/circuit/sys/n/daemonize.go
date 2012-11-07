package n

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"tumblr/circuit/load/config"
)

// pie (Panic-If-Error) panics if err is non-nil
func pie(err interface{}) {
	if err != nil {
		panic(err)
	}
}

// pie2 panics of err is non-nil
func pie2(underscore interface{}, err interface{}) {
	pie(err)
}

// piefwd panics of err is non-nil, in which case it prints the entire 
// contents of stdout and stderr to this process' standard error, followed
// by the panic stack trace
func piefwd(stdout, stderr *os.File, err interface{}) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "PANIC\n")
	os.Stderr.WriteString("Standard output:\n")
	stdout.Seek(0, 0)
	io.Copy(os.Stderr, stdout)
	os.Stderr.WriteString("Standard error:\n")
	stderr.Seek(0, 0)
	io.Copy(os.Stderr, stderr)
	os.Stderr.WriteString("Daemonizer error:\n")
	panic(err)
}

// dbg is like a printf for debugging the interactions between
// daemonizer and runtime where stdandard out and error are not
// available to us to play with.
func dbg(n, s string) {
	cmd := exec.Command("sh")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic("huh")
	}
	cmd.Start()
	defer cmd.Wait()
	fmt.Fprintf(stdin, "echo '%s' >> /Users/petar/tmp/%s\n", s, n)
	stdin.Close()
}

func Daemonize(wc *config.WorkerConfig) {

	// Make jail directory
	jail := path.Join(wc.Install.JailDir(), wc.Spark.ID.String())
	pie(os.MkdirAll(jail, 0700))

	// Prepare exec
	cmd := exec.Command(os.Args[0])
	cmd.Dir = jail

	// Out-of-band pipe for reading child PID and port
	bpr, bpw, err := os.Pipe()
	pie(err)
	cmd.ExtraFiles = []*os.File{bpw}

	// stdin 
	// Relay stdin of daemonizer to stdin of child runtime process
	var w bytes.Buffer
	pie(json.NewEncoder(&w).Encode(wc))
	cmd.Stdin = &w

	// Also save the config as a file for debugging purposes
	u, err := os.Create(path.Join(jail, "config"))
	if err != nil {
		panic(err)
	}
	pie(json.NewEncoder(u).Encode(wc))
	pie(u.Close())

	// Create stdout file
	stdout, err := os.Create(path.Join(jail, "out"))
	if err != nil {
		panic(err)
	}
	defer stdout.Close()
	cmd.Stdout = stdout

	// Create stderr file
	stderr, err := os.Create(path.Join(jail, "err"))
	if err != nil {
		panic(err)
	}
	defer stderr.Close()
	cmd.Stderr = stderr

	// start
	pie(cmd.Start())
	go func() {
		cmd.Wait()
		piefwd(stdout, stderr, bpw.Close())
	}()
	
	// Read the first two lines of stdout. They should hold the Port and PID of the runtime process.
	back := bufio.NewReader(bpr)

	// Read PID
	line, err := back.ReadString('\n')
	piefwd(stdout, stderr, err)

	pid, err := strconv.Atoi(strings.TrimSpace(line))
	piefwd(stdout, stderr, err)

	// Read port
	line, err = back.ReadString('\n')
	piefwd(stdout, stderr, err)
	
	port, err := strconv.Atoi(strings.TrimSpace(line))
	piefwd(stdout, stderr, err)

	// Close the pipe
	piefwd(stdout, stderr, bpr.Close())

	if cmd.Process.Pid != pid {
		piefwd(stdout, stderr, "pid mismatch")
	}

	fmt.Printf("%d\n%d\n", pid, port)
	// Sync is not supported on os.Stdout, at least on OSX
	// os.Stdout.Sync()

	// dbg("d", "daemonize succeeded!")
	os.Exit(0)
}