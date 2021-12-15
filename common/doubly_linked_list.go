package common

type Node struct {
	value *CacheValue
	prev  *Node
	next  *Node
}

type DoublyLinkedList struct {
	size int
	head *Node
	tail *Node
}

func (dll *DoublyLinkedList) AddToFront(value *CacheValue) *Node {
	newNode := &Node{value: value}
	if dll.head == nil {
		dll.head = newNode
		dll.tail = newNode
	} else {
		dll.head.prev = newNode
		newNode.next = dll.head
		dll.head = newNode
	}
	dll.size++
	return newNode
}

func (dll *DoublyLinkedList) AddToRear(value *CacheValue) *Node {
	newNode := &Node{value: value}
	if dll.tail == nil {
		dll.head = newNode
		dll.tail = newNode
	} else {
		newNode.prev = dll.tail
		dll.tail.next = newNode
		dll.tail = newNode
	}
	dll.size++
	return newNode
}

func (dll *DoublyLinkedList) MoveToFront(node *Node) {
	if node.prev != nil {
		//its already in the front
	}
	if node.next != nil {

	}
}
