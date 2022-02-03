package swim

import (
	"fmt"
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
	IsCurrentNode bool
}

type MembershipList struct {
	nodes       map[uint32]*node
	currentNode *node
}

type MembershipConfig struct {
	ListenPort  int
	HealthyNode string
}

var CURRENT_IP string
var LISTEN_PORT uint16
var CURRENT_NODE_HASH uint32

func InitMembershipServer(listenPort uint16, healthyNode string) *SwimService {
	ip := common.CurrentIP.String()
	port := listenPort
	CURRENT_IP = ip
	LISTEN_PORT = port
	CURRENT_NODE_HASH = common.Murmur3(common.GetAddress(CURRENT_IP, LISTEN_PORT))

	membership := &MembershipList{
		nodes: make(map[uint32]*node),
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

	swimService.printMembership()
	return swimService
}

func (n node) format() {
	log.Infof("%s:%d\t%11d\t%10s\t%d\t%s:%d", n.ip, n.port, n.hash, n.status, n.latestPing, n.statusSource.ip, n.statusSource.port)
	// log.Infof("%s:%d %d %s %d", n.ip, n.port, n.hash, n.status, n.latestPing)
}

func (s Status) String() string {
	return []string{"ALIVE", "DEAD", "SUSPECTED"}[s]
}

func (m MembershipList) updateLatestPing(hash uint32) {
	node := m.nodes[hash]
	node.latestPing = time.Now().Unix()
}

func (m MembershipList) addNode(ip string, port uint16, sourceIp string, sourcePort uint16) *node {
	address := common.GetAddress(ip, port)
	hash := common.Murmur3(address)
	newNode := &node{
		ip:            net.ParseIP(ip),
		port:          port,
		hash:          hash,
		status:        ALIVE,
		latestPing:    time.Now().Unix(),
		IsCurrentNode: ip == sourceIp && port == sourcePort,
	}
	sourceNodeHash := common.Murmur3(fmt.Sprintf("%s:%d", sourceIp, sourcePort))
	newNode.statusSource = m.nodes[sourceNodeHash]
	m.nodes[hash] = newNode
	return newNode
}

func (m MembershipList) addMemberNode(n *node) {
	m.nodes[n.hash] = n
}

func (m MembershipList) removeNode(ip string, port uint16) {
	address := common.GetAddress(ip, port)
	hash := common.Murmur3(address)
	delete(m.nodes, hash)
}

func (m MembershipList) getGroupList() []*node {
	nodes := make([]*node, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (m MembershipList) updateStatusSource(nodeHash uint32, statusSourcehash uint32) {
	node := m.nodes[nodeHash]
	node.statusSource = m.nodes[statusSourcehash]
}
