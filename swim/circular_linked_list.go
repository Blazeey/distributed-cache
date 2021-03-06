package swim

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type listNode struct {
	value *RingNode
	next  *listNode
}

type SortedCircularLinkedList struct {
	head *listNode
}

type RingNode struct {
	Host          *Node
	Hash          uint32
	Tokens        []uint32
	IsCurrentNode bool
}

func (sll *SortedCircularLinkedList) Add(value *RingNode) {
	newNode := &listNode{value: value}
	head := sll.head
	newValue := value.Hash
	if head == nil {
		sll.head = newNode
		newNode.next = newNode
	} else if head.next == head {
		current := head.value.Hash
		if newValue <= current {
			sll.head = newNode
			newNode.next = head
			head.next = newNode
		} else {
			head.next = newNode
			newNode.next = head
		}
	} else {
		n := head
		for ; ; n = n.next {
			if n.value.Hash < newValue && newValue < n.next.value.Hash {
				break
			} else if n.next.value.Hash < n.value.Hash {
				if newValue < n.value.Hash {
					sll.head = newNode
				}
				break
			}
		}
		newNode.next = n.next
		n.next = newNode
	}
}

func (sll *SortedCircularLinkedList) IsValueExist(value *RingNode) (exists bool) {
	exists = false
	n := sll.head
	for ok := true; ok && n != nil; ok = !(n == sll.head) {
		if n.value.Hash == value.Hash {
			return true
		}
		n = n.next
	}
	return
}

func (sll *SortedCircularLinkedList) PrintNodes() {
	s := ""
	n := sll.head
	for ok := true; ok; ok = !(n == sll.head) {
		s = fmt.Sprintf("%s %d ", s, n.value.Hash)
		n = n.next
	}
	log.Infof("[ %s ]", s)
}

func (sll *SortedCircularLinkedList) GetNodeWithGreaterOrEqualHash(hash uint32) *RingNode {
	n := sll.head
	for ok := true; ok; ok = !(n == sll.head) {
		if n.value.Hash >= hash {
			return n.value
		}
		n = n.next
	}
	return sll.head.value
}
