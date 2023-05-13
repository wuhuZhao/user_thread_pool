package user_thread_pool

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var _ Queue = (*DefaultTaskQueue)(nil)

type Queue interface {
	Poll() interface{}
	Offer(v interface{})
}

type node struct {
	Next  unsafe.Pointer
	Value interface{}
}

func NewDefaultTaskQueue() *DefaultTaskQueue {
	node := &node{
		Next:  nil,
		Value: nil,
	}
	return &DefaultTaskQueue{
		Head: unsafe.Pointer(node),
		Tail: unsafe.Pointer(node),
	}
}

// DefaultTaskQueue DefaultTaskQueue: lock-free的数据结构
type DefaultTaskQueue struct {
	Head unsafe.Pointer // 头结点
	Tail unsafe.Pointer // 尾节点
}

// load: 加载这个指针的值
func load(p *unsafe.Pointer) *node {
	return (*node)(atomic.LoadPointer(p))
}

// cas: go的cas实现
func cas(p *unsafe.Pointer, old, new *node) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}

func (d *DefaultTaskQueue) Offer(v interface{}) {
	// 创建要插入的Node节点
	node := &node{
		Next:  nil,
		Value: v,
	}
	// cas不能确保一次成功，所以需要不断轮训去获取最新的tail
	for {
		// 拿到tail和head
		tail := load(&d.Tail)
		next := load(&tail.Next)
		// 如果此时的tail和我们获取的tail相等的情况下,即其他线程还没有插入新的节点到tail中
		if tail == load(&d.Tail) {
			// 再次确认tail的next没有任何插入
			if next == nil {
				// 尝试node插入到tail.Next中,cas的定义可以参考上面的wiki的c实现，原理是一样的
				// 尝试插入失败就证明tail.Next已经有别的协程插入了，所以重新循环获取最新值
				if cas(&tail.Next, next, node) {
					// 插入成功后，将队列struct的tail替换成node节点，然后插入成功返回
					cas(&d.Tail, tail, node)
					return
				}
			} else {
				// next != nil 即表示有节点已经入队了，需要更新尾部，不需要确认更新是否成功，
				//因为有其他协程也会进行这一步，下几次for循环后就会更新掉，我们这个插入过程就能获取到最新的尾部了
				cas(&d.Tail, tail, next)
			}
		}
	}
}

func (d *DefaultTaskQueue) Poll() interface{} {
	for {
		// 获取head和tail，以及我们需要出列的节点
		tail := load(&d.Tail)
		head := load(&d.Head)
		next := load(&head.Next)
		// 确保当前head和队列的head还是一样，即还没有其他协程弹出的节点
		if head == load(&d.Head) {
			// 空队列的情况
			if head == tail {
				// 空队列直接返回nil
				if next == nil {
					return nil
				}
				// next != nil 即代表有插入的情况在另一个协程发生了，所以需要更新一下尾部的节点，
				//不需要确保更新成功，然后进行下一次循环，获取最新头尾节点
				cas(&d.Tail, tail, next)
			} else {
				// 不是空队列，将next中的值提出，然后释放掉value的内存空间
				cur := next.Value
				next.Value = nil
				// 将head尝试更新成next,因为出列了一个节点，自然要往后挪，如果失败了，就需要重新循环获取最新值，因为有其他协程已经弹出这个节点
				if cas(&d.Head, head, next) {
					// 替换成功即返回
					return cur
				}
			}
		}
	}
}

type Node struct {
	Next *Node
	Val  interface{}
}

type TwoLockQueue struct {
	HeadLock *sync.Mutex
	TailLock *sync.Mutex
	Head     *Node
	Tail     *Node
}

func (t *TwoLockQueue) Offer(v interface{}) {
	node := &Node{
		Next: nil,
		Val:  v,
	}
	t.TailLock.Lock()
	defer t.TailLock.Unlock()
	t.Tail.Next = node
	t.Tail = t.Tail.Next
}

func (t *TwoLockQueue) Poll() interface{} {
	t.HeadLock.Lock()
	defer t.HeadLock.Unlock()
	if t.Head.Next == nil {
		return nil
	}
	cur := t.Head.Next.Val
	t.Head.Next.Val = nil
	t.Head = t.Head.Next
	return cur
}

type LockQueue struct {
	lock *sync.Mutex
	q    []interface{}
}

func (l *LockQueue) Offer(v interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.q = append(l.q, v)
}

func (l *LockQueue) Poll() interface{} {
	l.lock.Lock()
	l.lock.Unlock()
	if len(l.q) == 0 {
		return nil
	}
	cur := l.q[0]
	l.q = l.q[1:]
	return cur
}
