package server

import (
	"distributed-cache.io/common"
	"distributed-cache.io/swim"
	"github.com/clockworksoul/smudge"
	log "github.com/sirupsen/logrus"
)

const (
	HEARBEAT_MS = 500
)

type MembershipConfig struct {
	listenPort  int
	healthyNode string
}

type StatusChangeListener struct {
	swim.MembershipStatusListener
}

type BroadcastListener struct {
	smudge.BroadcastListener
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
	hash, tokens := getTokens(node.Address())
	ringNode := &swim.RingNode{
		Host:          node,
		Hash:          hash,
		Tokens:        tokens,
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

func (ring TokenRing) getAssignedNode(hash uint32) *swim.RingNode {
	return ring.nodes.GetNodeWithGreaterOrEqualHash(hash)
}

func (l StatusChangeListener) OnChange(node *swim.Node, status swim.Status) {
	log.Infof("Node %s is now status %s", node.Address(), status)
	ringNode := newRingNode(node)
	if status == swim.ALIVE && !tokenRing.isNodePresentInRing(ringNode) {
		log.Infof("Adding node %s to the ring", node.Address())
		tokenRing.addNode(ringNode)
	}
	tokenRing.nodes.PrintNodes()
}

func getTokens(host string) (hash uint32, tokens []uint32) {
	hash = common.Murmur3(host)
	tokens = []uint32{hash}
	return
}

func InitMembershipServer(config MembershipConfig) {

	smudge.SetLogger(LogrusLogger{})
	smudge.SetListenPort(config.listenPort)
	smudge.SetHeartbeatMillis(HEARBEAT_MS)
	smudge.SetListenIP(common.CurrentIP)

	// smudge.AddStatusListener(StatusChangeListener{})
	// smudge.AddBroadcastListener(BroadcastListener{})

	if config.healthyNode != "" {
		node, err := smudge.CreateNodeByAddress(config.healthyNode)
		if err == nil {
			smudge.AddNode(node)
		}
	}

	smudge.Begin()
}

func (l LogrusLogger) Log(level smudge.LogLevel, a ...interface{}) (n int, err error) {

	logFn := log.Infoln
	if level == smudge.LogDebug {
		logFn = log.Debugln
	} else if level == smudge.LogError {
		logFn = log.Errorln
	} else if level == smudge.LogWarn {
		logFn = log.Warnln
	}
	logFn(a...)
	return 0, nil
}

func (l LogrusLogger) Logf(level smudge.LogLevel, format string, a ...interface{}) (n int, err error) {
	logFn := log.Infof
	if level == smudge.LogDebug {
		logFn = log.Debugf
	} else if level == smudge.LogError {
		logFn = log.Errorf
	} else if level == smudge.LogWarn {
		logFn = log.Warnf
	}
	logFn(format, a...)
	return 0, nil
}
