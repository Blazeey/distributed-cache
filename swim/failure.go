package swim

import (
	"time"

	log "github.com/sirupsen/logrus"
)

const PROTOCOL_PERIOD = time.Millisecond * 30000 // Determine based on average roundtrip
const PING_TIME_PERIOD = time.Millisecond * 200
const K_VALUE = 3

func (s SwimService) begin() {

	log.Infoln("Started failure detection")
	ticker := time.NewTicker(PROTOCOL_PERIOD)

	for {
		<-ticker.C
		s.pingNode()
	}
}

func (s SwimService) pingNode() {
	if len(s.membershipList.nodes) == 1 {
		return
	}
	s.membershipList.printMembership()
	probeTarget := s.membershipList.getProbeTarget()
	isSuccess := s.sendPingRequest(probeTarget.ip.String(), probeTarget.port)
	if isSuccess {
		log.Infof("Successfully pinged %s:%d", probeTarget.ip, probeTarget.port)
		s.membershipList.updateLatestPing(probeTarget.hash)
	} else {
		nodes := s.membershipList.getFailureDetectionSubgroups(probeTarget.hash)
		responseChannel := make(chan bool)
		for _, node := range nodes {
			go s.secondaryPingNode(probeTarget, node, responseChannel)
		}
		nodeCount := len(nodes)
		var isNodeReachable bool
		for ; nodeCount > 0; nodeCount-- {
			response := <-responseChannel
			if response {
				isNodeReachable = true
				log.Infof("Successfully secondary pinged %s:%d", probeTarget.ip, probeTarget.port)
				s.membershipList.updateLatestPing(probeTarget.hash)
			}
		}
		if !isNodeReachable {
			s.membershipList.suspectNode(probeTarget.hash)
		}
		close(responseChannel)
	}
	s.membershipList.updateNextProbeTarget()
}

func (s SwimService) secondaryPingNode(probeTarget *node, pingNode *node, responseChannel chan<- bool) {
	responseChannel <- s.sendSecondaryPingRequest(pingNode.ip.String(), pingNode.port, probeTarget.ip.String(), probeTarget.port)
}
