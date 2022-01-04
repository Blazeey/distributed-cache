package common

import (
	"fmt"

	"github.com/clockworksoul/smudge"
	log "github.com/sirupsen/logrus"
)

type SortedCircularLinkedList struct {
	head *node
}

type RingNode struct {
	Host          *smudge.Node
	Hash          uint32
	Tokens        []uint32
	IsCurrentNode bool
}

type node struct {
	value *RingNode
	next  *node
}

func (sll *SortedCircularLinkedList) Add(value *RingNode) {
	newNode := &node{value: value}
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
	for n := sll.head; n != nil; n = n.next {
		if n.value.Hash == value.Hash {
			return true
		}
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
