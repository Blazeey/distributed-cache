package common

import (
	log "github.com/sirupsen/logrus"
)

type SortedLinkedList struct {
	head *node
}

type node struct {
	value SortValue
	next  *node
}

type SortValue interface {
	Value() uint32
}

type Comparator func(newNode interface{}, existingNode interface{}) bool

func (cll *SortedLinkedList) Add(value SortValue) {
	newNode := &node{value: value}
	head := cll.head
	newValue := value.Value()
	if head == nil {
		cll.head = newNode
	} else if head.next == nil {
		current := head.value.Value()
		if newValue <= current {
			cll.head = newNode
			newNode.next = head
		} else {
			head.next = newNode
		}
	} else {
		n := head
		for ; n.next != nil; n = n.next {
			if newValue < n.next.value.Value() {
				break
			}
		}
		newNode.next = n.next
		n.next = newNode
	}
}

func (cll *SortedLinkedList) IsValueExist(value SortValue) (exists bool) {
	exists = false
	for n := cll.head; n != nil; n = n.next {
		if n.value.Value() == value.Value() {
			return true
		}
	}
	return
}

type Val struct {
	V uint32
}

func (x *Val) Value() uint32 {
	return x.V
}

func (cll *SortedLinkedList) PrintNodes() {
	for n := cll.head; n != nil; n = n.next {
		log.Infof("VAL : %d", n.value.Value())
	}
}
