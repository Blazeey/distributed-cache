package server

import (
	"net"

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
	smudge.StatusListener
}

type BroadcastListener struct {
	smudge.BroadcastListener
}

type LogrusLogger struct {
}

func (l StatusChangeListener) OnChange(node *smudge.Node, status smudge.NodeStatus) {
	log.Infof("Node %s is now status %s", node.Address(), status)
	for _, node := range smudge.AllNodes() {
		log.Infof("NODE : %s", node.Address())
	}
}

func (m BroadcastListener) OnBroadcast(b *smudge.Broadcast) {
	log.Infof("Received broadcast from %s: %s", b.Origin().Address(), string(b.Bytes()))
}

func InitMembershipServer(config MembershipConfig) {

	smudge.SetLogger(LogrusLogger{})
	smudge.SetListenPort(config.listenPort)
	smudge.SetHeartbeatMillis(HEARBEAT_MS)
	smudge.SetListenIP(getOutboundIP())

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

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
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
