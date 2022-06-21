package server

import (
	"distributed-cache.io/swim"
	log "github.com/sirupsen/logrus"
)

const (
	HEARBEAT_MS = 500
)

type StatusChangeListener struct {
	swim.MembershipStatusListener
}

type LogrusLogger struct {
}

type TokenRing struct {
	nodes *swim.SortedCircularLinkedList
}

var tokenRing = TokenRing{
	nodes: new(swim.SortedCircularLinkedList),
}

func newRingNode(node *swim.Node) *swim.RingNode {
	ringNode := &swim.RingNode{
		Host:          node,
		Hash:          node.Hash(),
		Tokens:        node.Tokens(),
		IsCurrentNode: node.IsCurrentNode(),
	}
	return ringNode
}

func (ring TokenRing) isNodePresentInRing(node *swim.RingNode) (exist bool) {
	return ring.nodes.IsValueExist(node)
}

func (ring TokenRing) addNode(node *swim.RingNode) {
	ring.nodes.Add(node)
}

func (ring TokenRing) removeNode(node *swim.RingNode) {
	ring.nodes.Remove(node)
}

func (ring TokenRing) getAssignedNode(hash uint32) *swim.RingNode {
	return ring.nodes.GetNodeWithGreaterOrEqualHash(hash)
}

func (l StatusChangeListener) OnChange(node *swim.Node, status swim.Status) {
	log.Infof("Node %s is now status %s", node.Address(), status)
	ringNode := newRingNode(node)
	if status == swim.ALIVE && !tokenRing.isNodePresentInRing(ringNode) {
		log.Infof("Adding node %s to the ring with %d tokens", node.Address(), len(ringNode.Tokens))
		tokenRing.addNode(ringNode)
	}
	if status == swim.DEAD {
		log.Infof("Removing node %s from the ring", node.Address())
		tokenRing.removeNode(ringNode)
	}
	tokenRing.nodes.PrintNodes()
}
