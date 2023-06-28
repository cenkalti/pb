package main

import (
	"bufio"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Wing924/shellwords"
	"github.com/cheggaaa/pb"
)

var epoch = time.Now()

type Line struct {
	// Durations since epoch for each occurance of the line.
	Delta []time.Duration
}

type State struct {
	Lines map[string]*Line
	// Duration passed after epoch
	Duration time.Duration
}

func readLines(r io.Reader, previousState *State, logFile io.Writer, done chan map[string]*Line, exitCodeCh chan int) {
	var bar *pb.ProgressBar
	if previousState != nil {
		// Show progress bar only if the program run successfully before.
		bar = pb.New64(int64(previousState.Duration))
		bar.SetUnits(pb.U_DURATION)
		bar.ShowTimeLeft = false
		bar.ShowElapsedTime = true
		bar.ShowFinalTime = false
		bar.Start()
	}
	h := md5.New()
	lines := make(map[string]*Line)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() { // Scanner stops when the stream is closed.
		s := scanner.Text()
		b := scanner.Bytes()
		h.Reset()
		h.Write(b)
		sum := string(h.Sum(nil))
		delta := time.Since(epoch)
		line, ok := lines[sum]
		if !ok {
			line = &Line{Delta: []time.Duration{delta}}
			lines[sum] = line
		} else {
			line.Delta = append(line.Delta, delta)
		}
		_, _ = logFile.Write([]byte(s))
		_, _ = logFile.Write([]byte("\n"))
		if previousState == nil {
			fmt.Println(s)
			continue
		}
		previousLine, ok := previousState.Lines[sum]
		if !ok {
			continue
		}
		if len(line.Delta) > len(previousLine.Delta) {
			continue
		}
		idx := len(line.Delta) - 1
		previousDelta := previousLine.Delta[idx]
		bar.Set64(int64(previousDelta))
	}
	if bar != nil {
		// The process must be stopped. Check exit code.
		exitCode := <-exitCodeCh
		if exitCode == 0 {
			bar.Set64(int64(previousState.Duration))
		}
		// Stops printing the progress bar to console.
		bar.Finish()
	}
	done <- lines
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("argument required")
	}

	baseFilename := getFilename(args)
	stateFilename := baseFilename + ".state"
	logFilename := baseFilename + ".log"

	previousState, err := readState(stateFilename)
	if err != nil {
		log.Fatal(err)
	}

	logFile, err := os.Create(logFilename)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Log:", logFile.Name())

	var name string
	shell := os.Getenv("SHELL")
	if shell != "" {
		name = shell
		args = append([]string{"-ic", shellwords.Join(args)})
	} else {
		name = args[0]
		args = args[1:]
	}

	cmd := exec.Command(name, args...)
	newState := runCmd(cmd, previousState, logFile)

	err = writeState(stateFilename, newState)
	if err != nil {
		log.Fatal(err)
	}

	if previousState != nil {
		pager := os.Getenv("PAGER")
		if pager != "" {
			cmd = exec.Command(pager, logFile.Name())
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Foreground: true,
			}
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func runCmd(cmd *exec.Cmd, previousState *State, logFile *os.File) *State {
	// Create a Pipe to redirect both stdout and stderr to the same stream.
	pr, pw, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stdout = pw
	cmd.Stderr = pw

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan map[string]*Line)
	exitCodeCh := make(chan int, 1)
	go readLines(pr, previousState, logFile, done, exitCodeCh)

	// Order of the following 4 lines is important here to avoid deadlock.
	waitErr := cmd.Wait()
	exitCodeCh <- cmd.ProcessState.ExitCode()
	pr.Close()
	lines := <-done

	duration := time.Since(epoch)
	if previousState == nil {
		fmt.Printf("Duration: %s\n", duration.Truncate(time.Second).String())
	}

	os.Stdout.Write([]byte{7}) // Ring the bell

	err = logFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	if waitErr != nil {
		fmt.Println("Log:", logFile.Name())
		log.Fatal(waitErr)
	}
	return &State{
		Lines:    lines,
		Duration: duration,
	}
}

func readState(filename string) (*State, error) {
	f, err := os.Open(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// previous state does not exist
		return nil, nil
	}
	defer f.Close()

	var state State
	err = gob.NewDecoder(f).Decode(&state)
	return &state, err
}

func writeState(filename string, state *State) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	err = gob.NewEncoder(f).Encode(state)
	if err != nil {
		return err
	}
	return f.Close()
}

func getFilename(args []string) string {
	wd, _ := os.Getwd()
	hasher := md5.New()
	hasher.Write([]byte(wd))
	for _, arg := range args {
		hasher.Write([]byte(arg))
	}
	sum := hasher.Sum(nil)
	id := hex.EncodeToString(sum[:])[:7]
	return filepath.Join(os.TempDir(), "pb."+id)
}
