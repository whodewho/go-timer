package main

import (
	"log"
	"time"
)

//task:runner=1:N, runner:ticker=1:N
//task业务隔离；runner并发(推荐)；ticker序列化(随机\无叠加)
type Task struct {
	taskQuitChan     chan bool
	taskQuitDoneChan chan bool
}

//runner退出超时时间，应大于最大的ticker周期, eg. 20 Seconds
var runnerQuitTimeout = time.Duration(20)

func NewTask() *Task {
	this := new(Task)

	this.taskQuitChan = make(chan bool)
	this.taskQuitDoneChan = make(chan bool)

	go this.Run1()
	go this.Run2()
	return this
}

func (this *Task) Shutdown() chan bool {
	//发信号终止所有runner
	close(this.taskQuitChan)

	c := make(chan bool)
	go func() {

		//收集runner退出成功信号
		for i := 0; i < 2; i++ {
			select {
			case <-this.taskQuitDoneChan:
				log.Print("shutdown runner success ", i)
			case <-time.After(runnerQuitTimeout * time.Second):
				//在超时时间内等待runner退出
				log.Print("shutdown runner timeout ", i)
				break
			}
		}

		//超时机制保证且必会释放信号
		c <- true
	}()

	return c
}

func (this *Task) Run1() {
	log.Print("runner1 start running")

	ticker1 := time.NewTicker(time.Second * 5)
	ticker2 := time.NewTicker(time.Second * 10)

	for i := 1; ; i++ {
		select {
		case <-ticker1.C:
			//业务逻辑日志可用log打印
			log.Print("runner1-ticker1 begin ", i)

			//todo 此处添加业务逻辑
			time.Sleep(time.Second * 15)

			log.Print("runner1-ticker1 finish ", i)
		case <-ticker2.C:
			//业务逻辑日志可用log打印
			log.Print("runner1-ticker2 begin ", i)

			//todo 此处添加业务逻辑
			time.Sleep(time.Second * 2)

			log.Print("runner1-ticker2 finish ", i)
		case <-this.taskQuitChan:
			//接收Shutdown()消息
			log.Print("runner1 quit ", i)
			this.taskQuitDoneChan <- true
			return
		}
	}
}

func (this *Task) Run2() {
	log.Print("runner2 start running")

	ticker1 := time.NewTicker(time.Second * 20)

	for i := 1; ; i++ {
		select {
		case <-ticker1.C:
			log.Print("runner2-ticker1 begin ", i)
			time.Sleep(time.Second * 1)
			log.Print("runner2-ticker1 finish ", i)
		case <-this.taskQuitChan:
			log.Print("runner2 quit ", i)
			this.taskQuitDoneChan <- true
			return
		}
	}
}
