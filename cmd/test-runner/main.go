package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

const LOGPATH string = "logs/"

func main() {
	var n int
	flag.IntVar(&n, "n", 3, "Number of peers to spawn. Default is 3.")

	done := make(chan bool, 1)
	wgT := &sync.WaitGroup{}
	wgT.Add(1)
	go tracker(done, wgT)

	wgP := &sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wgP.Add(1)
		go peer(i, wgP)
	}

	wgP.Wait()
	done <- true
	wgT.Wait()
}

func tracker(done chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	logfile, err := os.Create(LOGPATH + "tracker.log")
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	t := exec.Command("bin/tracker")
	t.Stdout = logfile
	t.Stderr = os.Stderr

	t.Start()
	<-done

	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			fmt.Println("failed executing tracker:", err)
		case *exec.ExitError:
			fmt.Println("tracker exit rc =", e.ExitCode())
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

	t := exec.Command("bin/peer")
	t.Stdout = logfile

	if err = t.Run(); err != nil {
		switch e := err.(type) {
		case *exec.Error:
			fmt.Printf("failed executing peer %d: %v\n", i, err)
		case *exec.ExitError:
			fmt.Printf("peer %d exit rc = %d\n", i, e.ExitCode())
		default:
			panic(err)
		}
	}
}
