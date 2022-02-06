package swim

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"distributed-cache.io/common"
	log "github.com/sirupsen/logrus"
)

type Status byte

const (
	ALIVE Status = iota
	DEAD
	SUSPECTED
)

type node struct {
	ip            net.IP
	port          uint16
	hash          uint32
	status        Status
	latestPing    int64
	statusSource  *node
	isCurrentNode bool
}

type MembershipList struct {
	nodesMap    map[uint32]*node
	nodes       []*node
	currentNode *node
	probeIndex  uint32
}

type MembershipConfig struct {
	ListenPort  int
	HealthyNode string
}

var CURRENT_IP string
var LISTEN_PORT uint16
var CURRENT_NODE_HASH uint32

var CURRENT_HOST Host

func InitMembershipServer(listenPort uint16, healthyNode string) *SwimService {
	ip := common.CurrentIP.String()
	port := listenPort
	CURRENT_IP = ip
	LISTEN_PORT = port
	CURRENT_HOST = *getHost(ip, port)
	CURRENT_NODE_HASH = common.Murmur3(common.GetAddress(CURRENT_IP, LISTEN_PORT))

	membership := &MembershipList{
		nodesMap: make(map[uint32]*node),
	}

	currentNode := membership.addNode(CURRENT_IP, LISTEN_PORT, CURRENT_IP, LISTEN_PORT)
	currentNode.statusSource = currentNode

	membership.currentNode = currentNode
	clientConnections := &SwimClientConnectionPool{
		connections: make(map[string]SwimClient, 0),
	}

	swimService := &SwimService{
		membershipList: membership,
		client:         clientConnections,
	}

	if healthyNode != "" {
		ip, port := common.ParseIP(healthyNode)
		go swimService.sendJoinRequest(ip, port)
	}

	go swimService.begin()
	return swimService
}

func (n *node) format() {
	log.Infof("%s:%d\t%11d\t%10s\t%d\t%s:%d\t%t", n.ip, n.port, n.hash, n.status, n.latestPing, n.statusSource.ip, n.statusSource.port, n.isCurrentNode)
}

func (n *node) hashIn(hashes ...uint32) bool {
	for _, hash := range hashes {
		if n.hash == hash {
			return true
		}
	}
	return false
}

func (s Status) String() string {
	return []string{"ALIVE", "DEAD", "SUSPECTED"}[s]
}

func (m *MembershipList) updateLatestPing(hash uint32) {
	node := m.nodesMap[hash]
	node.latestPing = time.Now().Unix()
	node.statusSource = m.currentNode
	node.status = ALIVE
}

func (m *MembershipList) addNode(ip string, port uint16, sourceIp string, sourcePort uint16) *node {
	address := common.GetAddress(ip, port)
	hash := common.Murmur3(address)
	newNode := &node{
		ip:            net.ParseIP(ip),
		port:          port,
		hash:          hash,
		status:        ALIVE,
		latestPing:    time.Now().Unix(),
		isCurrentNode: ip == sourceIp && port == sourcePort,
	}
	sourceNodeHash := common.Murmur3(fmt.Sprintf("%s:%d", sourceIp, sourcePort))
	newNode.statusSource = m.nodesMap[sourceNodeHash]
	m.nodesMap[hash] = newNode
	m.nodes = append(m.nodes, newNode)
	return newNode
}

func (m *MembershipList) addMemberNode(n *node) {
	m.nodesMap[n.hash] = n
	m.nodes = append(m.nodes, n)
}

func (m *MembershipList) removeNode(ip string, port uint16) {
	address := common.GetAddress(ip, port)
	hash := common.Murmur3(address)
	delete(m.nodesMap, hash)
}

func (m *MembershipList) getGroupList() []*node {
	nodes := make([]*node, 0, len(m.nodesMap))
	for _, node := range m.nodesMap {
		nodes = append(nodes, node)
	}
	return nodes
}

func (m *MembershipList) updateStatusSource(nodeHash uint32, statusSourcehash uint32) {
	node := m.nodesMap[nodeHash]
	node.statusSource = m.nodesMap[statusSourcehash]
}

func (m *MembershipList) getProbeTarget() *node {
	probeIndex := m.probeIndex
	probeNode := m.nodes[probeIndex]
	if probeNode.isCurrentNode {
		m.updateNextProbeTarget()
		probeIndex = m.probeIndex
		probeNode = m.nodes[probeIndex]
	}
	return probeNode
}

func (m *MembershipList) updateNextProbeTarget() {
	groupLength := len(m.nodes)
	var nextProbeIndex = m.probeIndex + 1

	for {
		if nextProbeIndex == uint32(groupLength) {
			// Shuffle the list
			nextProbeIndex = 0
			m.shuffleNodes()
		} else if m.nodes[nextProbeIndex].isCurrentNode {
			nextProbeIndex++
		} else {
			break
		}
	}
	m.probeIndex = nextProbeIndex
}

func (m *MembershipList) shuffleNodes() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(m.nodes), func(i, j int) {
		m.nodes[i], m.nodes[j] = m.nodes[j], m.nodes[i]
	})
}

func (m *MembershipList) getFailureDetectionSubgroups(ignoreNodes ...uint32) []*node {
	nodes := m.getOtherNodes(ignoreNodes...)
	nodeCount := common.MinInt(len(nodes), K_VALUE)
	subGroup := make([]*node, 0, nodeCount)
	rand.Seed(time.Now().UnixNano())
	p := rand.Perm(len(nodes))
	for _, r := range p {
		subGroup = append(subGroup, nodes[r])
	}
	return subGroup
}

func (m *MembershipList) getAllNodes() []*node {
	return m.nodes
}

func (m *MembershipList) getOtherNodes(ignoreNodes ...uint32) []*node {
	nodes := make([]*node, 0, len(m.nodes)-1)
	for _, n := range m.nodes {
		if !n.isCurrentNode && n.hashIn(ignoreNodes...) {
			nodes = append(nodes, n)
		}
	}
	return nodes
}

func (m *MembershipList) suspectNode(hash uint32) {
	node := m.nodesMap[hash]
	node.status = SUSPECTED
	log.Warnf("Suspecting Node %s:%d failed", node.ip, node.port)
	m.printMembership()
}

func (m *MembershipList) printMembership() {
	log.Info("Printing Membership Info")
	for _, n := range m.nodesMap {
		n.format()
	}
}
