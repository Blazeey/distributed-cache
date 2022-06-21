package swim

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type listNode struct {
	value *vNode
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

type vNode struct {
	ringNode *RingNode
	token    uint32
}

func (node *vNode) String() string {
	return fmt.Sprintf("%11d %s", node.token, node.ringNode.Host)
}

func (node *RingNode) equals(checkNode *RingNode) bool {
	return node.Hash == checkNode.Hash
}

func (sll *SortedCircularLinkedList) Add(ringNode *RingNode) {
	for _, token := range ringNode.Tokens {
		newVNode := &vNode{ringNode, token}
		newNode := &listNode{value: newVNode}
		head := sll.head
		newValue := newVNode.token
		if head == nil {
			sll.head = newNode
			newNode.next = newNode
		} else if head.next == head {
			current := head.value.token
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
				if n.value.token < newValue && newValue < n.next.value.token {
					break
				} else if n.next.value.token < n.value.token {
					if newValue < n.value.token {
						sll.head = newNode
					}
					break
				}
			}
			newNode.next = n.next
			n.next = newNode
		}
	}
}

func (sll *SortedCircularLinkedList) Remove(ringNode *RingNode) {
	n := sll.head
	currentHead := sll.head
	newHead := sll.head
	for ; ; newHead = newHead.next {
		if !newHead.value.ringNode.equals(ringNode) {
			break
		}
	}
	end := false
	for ; !end; n = n.next {
		for n.next.value.ringNode.equals(ringNode) {
			n.next = n.next.next
			if n.next == currentHead {
				end = true
			}
		}
	}
	sll.head = newHead
}

func (sll *SortedCircularLinkedList) IsValueExist(value *RingNode) (exists bool) {
	exists = false
	n := sll.head
	for ok := true; ok && n != nil; ok = !(n == sll.head) {
		if n.value.token == value.Hash {
			return true
		}
		n = n.next
	}
	return
}

func (sll *SortedCircularLinkedList) PrintNodes() {
	n := sll.head
	for ok := true; ok; ok = !(n == sll.head) {
		log.Warnf("%s", n.value)
		n = n.next
	}
}

func (sll *SortedCircularLinkedList) GetNodeWithGreaterOrEqualHash(hash uint32) *RingNode {
	n := sll.head
	for ok := true; ok; ok = !(n == sll.head) {
		if n.value.token >= hash {
			return n.value.ringNode
		}
		n = n.next
	}
	return sll.head.value.ringNode
}
