package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	daemon "github.com/sevlyar/go-daemon"
)

var (
	quitChan     = make(chan struct{}, 0)
	quitDoneChan = make(chan struct{}, 0)
	signal       = flag.String("nns", "", `stop - fast shutdown`)
)

func main() {
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGQUIT, quitHandler)
	binhome := filepath.Dir(filepath.Dir(getRealPath(os.Args[0])))
	os.Mkdir(binhome+"/var", 0775)
	os.Mkdir(binhome+"/data", 0775)

	cntxt := &daemon.Context{
		PidFileName: binhome + "/var/pid",
		PidFilePerm: 0644,
		LogFileName: binhome + "/var/stdout",
		LogFilePerm: 0644,
		WorkDir:     binhome,
		Umask:       027,
		Args:        os.Args,
	}

	if len(daemon.ActiveFlags()) > 0 {
		if d, err := cntxt.Search(); err != nil {
			fmt.Println("Unable send signal to the daemon:", err)
		} else {
			daemon.SendCommands(d)
		}
		return
	}

	if d, err := cntxt.Reborn(); err != nil {
		log.Fatalln("Reborn:", err)
	} else if d != nil {
		return
	}
	defer cntxt.Release()

	go worker()

	if err := daemon.ServeSignals(); err != nil {
		log.Fatalln("err", err)
	}
	log.Fatalln("daemon terminated")
}

func worker() {
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	//依次加载定时任务
	task := NewTask()

	//阻塞等待关闭
	select {
	case <-quitChan:
		//依次调用task.Shutdown()，并阻塞等待返回
		//todo 考虑并发关闭，需要额外同步机制
		log.Print("got quit msg")
		<-task.Shutdown()
		log.Print("all task shutdown")
		break
	}

	//task全部成功关闭
	close(quitDoneChan)
	log.Print("send quit done msg")
}

func quitHandler(sig os.Signal) error {
	//触发task关闭
	close(quitChan)
	log.Print("send quit msg")

	if sig == syscall.SIGQUIT {
		//等待task全部成功关闭
		<-quitDoneChan
		log.Print("got quit done msg")
	}
	return daemon.ErrStop
}

func getRealPath(path string) string {
	if path[0] == '~' {
		home := os.Getenv("HOME")
		path = home + path[1:]
	}
	rpath, err := filepath.Abs(path)
	if err == nil {
		path = rpath
	}
	return strings.TrimSpace(path)
}
