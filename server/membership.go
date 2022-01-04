package server

import (
	"unsafe"

	"distributed-cache.io/common"
	"github.com/clockworksoul/smudge"
	log "github.com/sirupsen/logrus"
)

const (
	HEARBEAT_MS        = 500
	c1_32       uint32 = 0xcc9e2d51
	c2_32       uint32 = 0x1b873593
)

type MembershipConfig struct {
	listenPort  int
	healthyNode string
}

type StatusChangeListener struct {
	smudge.StatusListener
}

type BroadcastListener struct {
	smudge.BroadcastListener
}

type LogrusLogger struct {
}

type TokenRing struct {
	nodes *common.SortedCircularLinkedList
}

var tokenRing = TokenRing{
	nodes: new(common.SortedCircularLinkedList),
}

func newRingNode(node *smudge.Node) *common.RingNode {
	hash, tokens := getTokens(node.Address())
	ringNode := &common.RingNode{
		Host:   node,
		Hash:   hash,
		Tokens: tokens,
	}
	return ringNode
}

func (ring TokenRing) isNodePresentInRing(node *common.RingNode) (exist bool) {
	return ring.nodes.IsValueExist(node)
}

func (ring TokenRing) addNode(node *common.RingNode) {
	ring.nodes.Add(node)
}

func (ring TokenRing) getAssignedNode(hash uint32) *common.RingNode {
	return ring.nodes.GetNodeWithGreaterHash(hash)
}

func (l StatusChangeListener) OnChange(node *smudge.Node, status smudge.NodeStatus) {
	log.Infof("Node %s is now status %s", node.Address(), status)
	ringNode := newRingNode(node)
	if status == smudge.StatusAlive && !tokenRing.isNodePresentInRing(ringNode) {
		log.Infof("Adding node %s to the ring", node.Address())
		tokenRing.addNode(ringNode)
	}
	tokenRing.nodes.PrintNodes()
}

func getTokens(host string) (hash uint32, tokens []uint32) {
	hash = murmur3(host)
	tokens = []uint32{hash}
	return
}

func (m BroadcastListener) OnBroadcast(b *smudge.Broadcast) {
	log.Infof("Received broadcast from %s: %s", b.Origin().Address(), string(b.Bytes()))
}

func InitMembershipServer(config MembershipConfig) {

	smudge.SetLogger(LogrusLogger{})
	smudge.SetListenPort(config.listenPort)
	smudge.SetHeartbeatMillis(HEARBEAT_MS)
	smudge.SetListenIP(common.GetOutboundIP())

	smudge.AddStatusListener(StatusChangeListener{})
	smudge.AddBroadcastListener(BroadcastListener{})

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

func murmur3(key string) uint32 {

	data := []byte(key)
	var h1 uint32 = 0

	nblocks := len(data) / 4
	var p uintptr
	if len(data) > 0 {
		p = uintptr(unsafe.Pointer(&data[0]))
	}
	p1 := p + uintptr(4*nblocks)
	for ; p < p1; p += 4 {
		k1 := *(*uint32)(unsafe.Pointer(p))

		k1 *= c1_32
		k1 = (k1 << 15) | (k1 >> 17) // rotl32(k1, 15)
		k1 *= c2_32

		h1 ^= k1
		h1 = (h1 << 13) | (h1 >> 19) // rotl32(h1, 13)
		h1 = h1*5 + 0xe6546b64
	}

	tail := data[nblocks*4:]

	var k1 uint32
	switch len(tail) & 3 {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1_32
		k1 = (k1 << 15) | (k1 >> 17) // rotl32(k1, 15)
		k1 *= c2_32
		h1 ^= k1
	}

	h1 ^= uint32(len(data))

	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return h1
}
