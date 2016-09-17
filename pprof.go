package main

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
	"time"

	_ "net/http/pprof"
)

/*
func getInitialize(w http.ResponseWriter, r *http.Request) {
	noprofile := r.URL.Query().Get("noprofile")
	if noprofile == "" {
		StartProfile(time.Minute)
	}
	...
}
*/

var (
	enableProfile     = true
	isProfiling       = false
	cpuProfileFile    = "/tmp/cpu.pprof"
	memProfileFile    = "/tmp/mem.pprof"
	blockProfileFile  = "/tmp/block.pprof"
	onStartProfileCmd = "/home/isucon/bench-tools/on-start-profile"
	onEndProfileCmd   = "/home/isucon/bench-tools/on-end-profile"
)

func callOnStartProfile() {
	if _, err := os.Stat(onStartProfileCmd); os.IsNotExist(err) {
		log.Println("OnStartProfile command not found:", err)
		return
	}
	cmd := exec.Command(onStartProfileCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println("OnStartProfile command error:", err)
	}
	log.Printf("OnStartProfile Output: %s\n", out.String())
}

func callOnEndProfile() {
	if _, err := os.Stat(onEndProfileCmd); os.IsNotExist(err) {
		log.Println("OnEndProfile command not found:", err)
		return
	}
	cmd := exec.Command(onEndProfileCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println("OnEndProfile command error:", err)
	}
	log.Printf("OnEndProfile Output: %s\n", out.String())
}

func StartProfile(duration time.Duration) error {
	f, err := os.Create(cpuProfileFile)
	if err != nil {
		return err
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		return err
	}
	runtime.SetBlockProfileRate(1)
	isProfiling = true
	if 0 < duration.Seconds() {
		go func() {
			time.Sleep(duration)
			err := EndProfile()
			if err != nil {
				log.Println(err)
			}
		}()
	}
	log.Println("Profile start")
	go callOnStartProfile()
	return nil
}

func EndProfile() error {
	if !isProfiling {
		return nil
	}
	isProfiling = false
	pprof.StopCPUProfile()
	runtime.SetBlockProfileRate(0)
	log.Println("Profile end")
	defer func() {
		go callOnEndProfile()
	}()

	mf, err := os.Create(memProfileFile)
	if err != nil {
		return err
	}
	pprof.WriteHeapProfile(mf)

	bf, err := os.Create(blockProfileFile)
	if err != nil {
		return err
	}
	pprof.Lookup("block").WriteTo(bf, 0)
	return nil
}

func init() {
	log.Println("add handler /startprof /endprof")
	http.HandleFunc("/startprof", func(w http.ResponseWriter, r *http.Request) {
		err := StartProfile(time.Minute)
		if err != nil {
			w.Write([]byte(err.Error()))
		} else {
			w.Write([]byte("profile started\n"))
		}
	})

	http.HandleFunc("/endprof", func(w http.ResponseWriter, r *http.Request) {
		err := EndProfile()
		w.Write([]byte("profile ended\n"))
		if err != nil {
			w.Write([]byte(err.Error() + "\n"))
		}
	})
}
