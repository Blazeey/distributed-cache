package swim

import (
	"context"
	"net"
	"sync"
	"time"

	"distributed-cache.io/common"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SwimService struct {
	UnimplementedSwimServer
	membershipList *MembershipList
	client         *SwimClientConnectionPool
}

func (s SwimService) Ping(ctx context.Context, in *PingRequest) (*PingResponse, error) {
	log.Infof("Received Ping Request from %s:%d", in.Source.Ip, in.Source.Port)
	return &PingResponse{
		Code: ResponseCode_SUCCESS,
	}, nil
}

func (s SwimService) SecondaryPing(ctx context.Context, in *SecondaryPingRequest) (*SecondaryPingResponse, error) {
	log.Infof("Received SecondaryPing Request from %s:%d", in.Source.Ip, in.Source.Port)
	pingTarget := in.PingTarget
	client := s.client.getRemoteConnection(pingTarget.Ip, uint16(pingTarget.Port))
	secondaryCtx, cancel := context.WithTimeout(context.Background(), time.Second) // TODO: Check the timeout
	defer cancel()
	pingResponse, err := client.Ping(secondaryCtx, &PingRequest{
		Source: &Host{
			Ip:   in.Source.Ip,
			Port: in.Source.Port,
		},
	})
	response := &SecondaryPingResponse{}
	if err != nil {
		response.Code = ResponseCode_ERROR
	} else {
		response.Code = pingResponse.Code
	}
	return response, err
}

func (s SwimService) AddNode(ctx context.Context, in *NodeAdditionRequest) (*NodeAdditionResponse, error) {
	log.Infof("Received AddNode Request from %s:%d", in.Source.Ip, in.Source.Port)
	addedNode := in.AddedNode
	source := in.Source
	s.membershipList.addNode(addedNode.Ip, uint16(addedNode.Port), source.Ip, uint16(source.Port))
	s.printMembership()
	return &NodeAdditionResponse{
		Code: ResponseCode_SUCCESS,
	}, nil
}

func (s SwimService) RemoveNode(ctx context.Context, in *NodeRemovalRequest) (*NodeRemovalResponse, error) {
	log.Infof("Received RemoveNode Request from %s:%d", in.Source.Ip, in.Source.Port)
	s.membershipList.removeNode(in.RemovedNode.Ip, uint16(in.RemovedNode.Port))
	s.printMembership()
	return &NodeRemovalResponse{
		Code: ResponseCode_SUCCESS,
	}, nil
}

func (s SwimService) Join(ctx context.Context, in *JoinRequest) (*JoinResponse, error) {
	log.Infof("Received Join Request from %s:%d", in.Source.Ip, in.Source.Port)
	nodes := s.membershipList.getGroupList()
	s.membershipList.addNode(in.Source.Ip, uint16(in.Source.Port), CURRENT_IP, LISTEN_PORT) // TODO: Check the timeout

	var wg sync.WaitGroup
	for _, node := range nodes {
		if !node.isCurrentNode {
			wg.Add(1)
			client := s.client.getRemoteConnection(node.ip.String(), node.port)
			go sendAddNodeRequests(in.Source.Ip, uint32(in.Source.Port), &wg, client)
		}
	}
	wg.Wait()

	s.printMembership()
	return &JoinResponse{
		Code:                ResponseCode_SUCCESS,
		GroupMembershipList: mapNodesToNodeDetails(nodes),
	}, nil
}

func (s SwimService) sendPingRequest(ip string, port uint16) bool {
	log.Infof("Sending Ping Reguest to %s:%d", ip, port)
	client := s.client.getRemoteConnection(ip, port)
	ctx, cancel := context.WithTimeout(context.Background(), PING_TIME_PERIOD)
	defer cancel()
	_, err := client.Ping(ctx, &PingRequest{
		Source: &CURRENT_HOST,
	})
	if err != nil {
		log.Errorf("ERROR sending Ping request to %s:%d", ip, port, err)
	}
	return err == nil
}

func (s SwimService) sendSecondaryPingRequest(ip string, port uint16, pingTargetIp string, pingTargetPort uint16) bool {
	log.Infof("Sending Secondary Ping request to %s:%d", ip, port)
	client := s.client.getRemoteConnection(ip, port)
	ctx, cancel := context.WithTimeout(context.Background(), PROTOCOL_PERIOD-PING_TIME_PERIOD)
	defer cancel()
	response, err := client.SecondaryPing(ctx, &SecondaryPingRequest{
		Source:     &CURRENT_HOST,
		PingTarget: getHost(pingTargetIp, pingTargetPort),
	})
	if err != nil {
		log.Errorf("ERROR sending Secondary Ping request to %s:%d", ip, port, err)
	}
	return err == nil && response.Code == ResponseCode_SUCCESS
}

func (s SwimService) sendJoinRequest(ip string, port uint16) {
	client := s.client.getRemoteConnection(ip, port)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	log.Infof("Sending Join Request to %s:%d", ip, port)
	response, err := client.Join(ctx, &JoinRequest{
		Source: &CURRENT_HOST,
	})
	if err != nil {
		log.Errorf("ERROR sending join request to %s:%d", ip, port, err)
	}
	s.updateGroupMembership(response.GroupMembershipList)
	s.printMembership()
}

func (s SwimService) updateGroupMembership(list []*NodeDetails) {
	for _, n := range list {
		memberNode := mapNodeDetails(n)
		s.membershipList.addMemberNode(memberNode)
	}
	for _, n := range list {
		_, _, nodeHash := getHostIdentifiers(n.Host)
		_, _, statusSourceHash := getHostIdentifiers(n.StatusSource)

		if nodeHash == statusSourceHash {
			statusSourceHash = CURRENT_NODE_HASH
		}
		s.membershipList.updateStatusSource(nodeHash, statusSourceHash)
	}
	s.printMembership()
}

func (s SwimService) printMembership() {
	s.membershipList.printMembership()
}

func mapNodeDetails(n *NodeDetails) *node {
	ip, port, hash := getHostIdentifiers(n.Host)
	status := Status(n.Status)
	isCurrentNode := CURRENT_IP == ip.String() && port == LISTEN_PORT
	if isCurrentNode {
		status = ALIVE
	}
	return &node{
		ip:            ip,
		port:          port,
		hash:          hash,
		status:        status,
		latestPing:    n.LatestPing,
		isCurrentNode: isCurrentNode,
	}
}

func mapNodesToNodeDetails(nodes []*node) []*NodeDetails {
	mappedNodes := make([]*NodeDetails, 0, len(nodes))
	for _, n := range nodes {
		nodeDetails := mapNodeToNodeDetails(n)
		mappedNodes = append(mappedNodes, nodeDetails)
	}

	return mappedNodes
}

func mapNodeToNodeDetails(n *node) *NodeDetails {
	return &NodeDetails{
		Host:         getHost(n.ip.String(), n.port),
		Status:       NodeStatus(n.status),
		LatestPing:   n.latestPing,
		StatusSource: getHost(n.statusSource.ip.String(), n.statusSource.port),
	}
}

func getHostIdentifiers(h *Host) (ip net.IP, port uint16, hash uint32) {
	ip = net.ParseIP(h.Ip)
	port = uint16(h.Port)
	address := common.GetAddress(ip.String(), port)
	hash = common.Murmur3(address)
	return
}

func getHost(ip string, port uint16) *Host {
	return &Host{
		Ip:   ip,
		Port: uint32(port),
	}
}

func sendAddNodeRequests(addedNodeIp string, addedNodePort uint32, wg *sync.WaitGroup, client SwimClient) {
	secondaryCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client.AddNode(secondaryCtx, &NodeAdditionRequest{
		Source: &Host{
			Ip:   CURRENT_IP,
			Port: uint32(LISTEN_PORT),
		},
		AddedNode: &Host{
			Ip:   addedNodeIp,
			Port: addedNodePort,
		},
	})
	wg.Done()
}

func handleError(err error) {

	errStatus, _ := status.FromError(err)
	log.Infoln(errStatus.Message())
	log.Infoln(errStatus.Code())
	if codes.InvalidArgument == errStatus.Code() {
		// do your stuff here
		log.Fatal()
	}
}
