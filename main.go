package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Worker struct {
	Jobs chan string
}

func main() {
	var interval string
	interval = os.Getenv("INTERVAL")
	intv, err := strconv.Atoi(interval)
	if err != nil {
		panic(fmt.Sprintf("some error"))
	}
	// Lets create the node with a Jobs channel that can handle 10 jobs
	node := &Worker{Jobs: make(chan string, 10)}

	// Start the workers
	go node.StartWorkers()

	// Scan from transmission and write to the channel
	ticker := time.NewTicker(time.Duration(intv) * time.Minute)
	stopTasker := make(chan int, 1)
	go func() {
		for {
			select {
			case <- ticker.C:
				fmt.Println("Scanning ...")
				Scan(node.Jobs)
			case <- stopTasker:
				fmt.Println("Stopping task scheduler...")
				ticker.Stop()
				return
			}
		}
	}()

	// Let the app run until a termination has been signaled
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
	stopTasker<-1
	sleep(1000)
	close(node.Jobs)
	close(quitChannel)
	close(stopTasker)
	fmt.Println("Stopping processes")
}

// worker Handles the worker logic which listens to the node.Jobs channel
// and executes those jobs
func (node Worker) worker(id int, jobs <-chan string, wg *sync.WaitGroup) {
	// Mark worker as done once it is finished
	defer wg.Done()

	// Worker has started
	fmt.Println("Started worker ", id)

	// Listen to the channel for incoming jobs
	for j := range jobs {
		fmt.Println("Worker {", id, "} started  job", j)
		ExecuteJob(j)
		fmt.Println("Worker {", id, "} finished job", j)
	}
}

// StartWorkers launched 3 concurrent workers that listen to node.Jobs channel
func (node Worker) StartWorkers() {

	var wg sync.WaitGroup

	for w := 1; w <= 3; w++ {
		wg.Add(1)
		go node.worker(w, node.Jobs, &wg)
	}
	wg.Wait()
}

// ExecuteJob executes the actual job for unrar-ing the files
func ExecuteJob(dirname string) {
	var (
		out string
		err error
	)

	// Open the directory so that we can look at the files
	f, err := os.Open(dirname)
	if err != nil {
		fmt.Println(err)
		return
	}

	// List all the files
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	norar := true
	for _, file := range files {
		if strings.Contains(file.Name(), ".rar") {
			norar = false
			fmt.Println("Unrar-ing file....")
			out, err = unrar(dirname, file.Name())
			if err != nil {
				fmt.Println("Error:", err)
				norar = true
				os.Remove(fmt.Sprintf("%s/%s", dirname, strings.Replace(file.Name(), ".rar", ".mkv", 1)))
				return
			}
			fmt.Println("Output: ", out)
			return
		}
	}
	if norar == true {
		lock(dirname)
	}
}

// unrar Unrars a rar file into the current directory
func unrar(dir string, name string) (string, error) {
	var (
		out []byte
		err error
	)
	// Change current directory to the directory where the rar file is
	err = os.Chdir(dir)
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%s/%s", dir, name)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	out, err = exec.CommandContext(ctx, "unrar", "e", fileName, dir).Output()
	if err != nil {
		return "", err
	}
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Command timed out")
		return "", ctx.Err()
	}
	// Lock so that it does not try to unnecessarily unrar the same file
	err = lock(dir)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// lock Locks the directory of the file that has been unrar-ed once
func lock(dirname string) error {
	lockfile := fmt.Sprintf("%s/norar", dirname)
	fmt.Println(lockfile)
	file, err := os.Create(lockfile)
	if err != nil {
		return err
	}

	file.Close()
	return nil
}

func sleep(i int) {
	time.Sleep(time.Millisecond * time.Duration(i))
}