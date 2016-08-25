package master

import (
	"testing"
	"github.com/coldog/sked/tools"
	"time"
)

func TestQueue_Enqueue(t *testing.T) {
	q := NewSchedulerQueue()

	q.Push("hello1")
	q.Push("hello2")
	q.Push("hello3")
	q.Push("hello1")
	q.Push("hello1")

	time.Sleep(100 * time.Millisecond)
	tools.Assert(t, len(q.queue) == 3, "queue is not 3 wide")
}

func TestQueue_Blocking(t *testing.T) {
	q := NewSchedulerQueue()

	go func() {
		time.Sleep(100 * time.Millisecond)
		q.Push("hello2")
	}()

	l := make(chan string)
	q.Pop(l)
	val := <- l
	tools.Assert(t, val == "hello2", "blocking didn't work")
}
