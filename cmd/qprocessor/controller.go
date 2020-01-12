package main

import (
	"project/api"
	"project/systemstop"
	"sync"
	"time"
)

type task struct {
	gradermessage *api.GraderMessage
	started       time.Time
	id            string
}

type controller struct {
	processingTasks []*task
	mutex           *sync.Mutex
	overtime        chan *api.MQmessage
}

func createTasksController() *controller {
	c := &controller{}
	c.processingTasks = []*task{}
	c.mutex = &sync.Mutex{}
	c.overtime = make(chan *api.MQmessage, 1)
	stop := systemstop.Subscribe()
	go func() {
		defer stop.Done()
		for {
			c.mutex.Lock()
			if len(c.processingTasks) == 0 {
				c.mutex.Unlock()
				t := time.NewTimer(time.Second)
				select {
				case <-stop.Signal():
					return
				case <-t.C:
					break
				}
				continue
			}

			//t := time.NewTimer()

			select {
			case <-stop.Signal():
				break
			default:
			}

		}
	}()
	return c
}

func (c *controller) addTask(id string, m *api.GraderMessage) {
	c.mutex.Lock()
	t := &task{gradermessage: m, started: time.Now(), id: id}
	c.processingTasks = append(c.processingTasks, t)
	c.mutex.Unlock()

	/*go func() {
		tmr := time.NewTimer(time.Minute)
		<-tmr.C
	}()*/

}

func (c *controller) deleteTask(id string) {
	c.mutex.Lock()
	for i, t := range c.processingTasks {
		if t.id == id {
			c.processingTasks = append(c.processingTasks[:i], c.processingTasks[i+1:]...)
			break
		}
	}
	c.mutex.Unlock()
}

func (c *controller) overdue() <-chan *api.MQmessage {
	return c.overtime
}
