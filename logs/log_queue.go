package logs

import (
	"encoding/json"
	"sync"
)

type Queue struct {
	items []interface{}
	limit int
	lock  sync.Mutex
}

// 初始化队列
func NewQueue(limit int) *Queue {
	q := &Queue{
		limit: limit,
	}

	return q
}

// 入队操作
func (q *Queue) Enqueue(item interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// 如果队列已满，先出队一个元素
	for len(q.items) >= q.limit {
		q.items = q.items[1:]
	}

	q.items = append(q.items, item)
}

// 出队操作
func (q *Queue) Dequeue() interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()

	// 等待队列非空
	for len(q.items) == 0 {
		return nil
	}

	item := q.items[0]
	q.items = q.items[1:]

	return item
}

// 获取队列长度
func (q *Queue) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.items)
}

// 判断队列是否为空
func (q *Queue) IsEmpty() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.items) == 0
}

func (q *Queue) MarshalJSON() ([]byte, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	return json.Marshal(q.items)
}
