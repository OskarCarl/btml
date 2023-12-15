package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const LOGPATH string = "logs/"
const TRACKER_URL string = "127.0.0.1:8923"

func main() {
	var n int
	flag.IntVar(&n, "n", 3, "Number of peers to spawn. Default is 3.")
	flag.Parse()

	log.Default().SetPrefix("[RUNNER] ")

	done := make(chan struct{})
	wgT := &sync.WaitGroup{}
	wgT.Add(1)
	go tracker(done, wgT)

	time.Sleep(time.Second * 1)
	wgP := &sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wgP.Add(1)
		go peer(i, wgP)
	}

	wgP.Wait()
	close(done)
	wgT.Wait()
}

func tracker(done chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	logfile, err := os.Create(LOGPATH + "tracker.log")
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	t := exec.Command("bin/tracker", "-ListenAddress", TRACKER_URL)
	t.Stdout = logfile
	t.Stderr = os.Stderr

	t.Start()
	<-done
	err = t.Process.Signal(syscall.SIGINT)

	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			log.Default().Println("failed executing tracker:", err)
		case *exec.ExitError:
			log.Default().Println("tracker exit rc =", e.ExitCode())
		default:
			panic(err)
		}
	}
}

func peer(i int, wg *sync.WaitGroup) {
	defer wg.Done()

	logfile, err := os.Create(fmt.Sprintf("%speer%d.log", LOGPATH, i))
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	t := exec.Command("bin/peer", "-name", strconv.Itoa(i), "-trackerURL", "http://"+TRACKER_URL)
	t.Stdout = logfile

	if err = t.Run(); err != nil {
		switch e := err.(type) {
		case *exec.Error:
			log.Default().Printf("failed executing peer %d: %v\n", i, err)
		case *exec.ExitError:
			log.Default().Printf("peer %d exit rc = %d\n", i, e.ExitCode())
		default:
			panic(err)
		}
	}
}
