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
	Host   *smudge.Node
	Hash   uint32
	Tokens []uint32
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
	} else if head.next == nil {
		current := head.value.Hash
		if newValue <= current {
			sll.head = newNode
			newNode.next = head
		} else {
			head.next = newNode
		}
	} else {
		n := head
		for ; n.next != nil; n = n.next {
			if newValue < n.next.value.Hash {
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
	for n := sll.head; n != nil; n = n.next {
		s = fmt.Sprintf("%s %d ", s, n.value.Hash)
	}
	log.Infof("[ %s ]", s)
}

func (sll *SortedCircularLinkedList) GetNodeWithGreaterHash(hash uint32) *RingNode {
	for n := sll.head; n != nil; n = n.next {
		if n.value.Hash > hash {
			return n.value
		}
	}
	return sll.head.value
}
