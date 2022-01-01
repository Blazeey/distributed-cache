package server

import (
	"encoding/json"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"distributed-cache.io/common"
	"github.com/panjf2000/gnet"
)

type Callback func(message Message) (respone Response)

type Operation string

type Server struct {
	*gnet.EventServer
	port      int
	multicore bool
	callback  Callback
	cache     *common.Cache
}

type ServerConfig struct {
	serverPort     int
	membershipPort int
	multicore      bool
	healthyNode    string
	callback       Callback
}

const (
	PUT  Operation = "PUT"
	GET  Operation = "GET"
	LIST Operation = "LIST"
)

type Message struct {
	Op    Operation `json:"op"`
	Key   string    `json:"key"`
	Value string    `json:"value"`
}

type Response struct {
	Message string
	Code    int
}

func (m *Message) String() string {
	return fmt.Sprintf("Operation : %s, Key : %s, Value : %s", m.Op, m.Key, m.Value)
}

func (r *Response) String() string {
	return fmt.Sprintf("Code : %d, Message : %s", r.Code, r.Message)
}

func InitServer(config ServerConfig) Server {
	server := new(Server)
	server.port = config.serverPort
	server.multicore = config.multicore
	server.cache = common.InitCache()
	if config.callback == nil {
		server.callback = server.defaultCallback()
	} else {
		server.callback = config.callback
	}

	wg := new(sync.WaitGroup)
	wg.Add(5)

	go startCodecServer(server)
	go InitMembershipServer(MembershipConfig{
		listenPort:  config.membershipPort,
		healthyNode: config.healthyNode,
	})

	wg.Wait()
	return *server
}

func startCodecServer(server *Server) {
	err := gnet.Serve(server, fmt.Sprintf("tcp://:%d", server.port), gnet.WithMulticore(server.multicore))
	if err != nil {
		panic(err)
	}
}

func (s *Server) defaultCallback() Callback {
	return func(message Message) (response Response) {
		m, _ := json.Marshal(message)
		response.Code = 200
		response.Message = string(m)
		return
	}
}

func (s *Server) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	log.Infof("Test codec server is listening on %s (multi-cores: %t, loops: %d)", srv.Addr.String(), srv.Multicore, srv.NumEventLoop)
	return
}

func (s *Server) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Infof("Connection opened. Local Address: %s, Remote Address: %s", c.LocalAddr().String(), c.RemoteAddr().String())
	return
}

func (s *Server) OnShutdown(srv gnet.Server) {
	log.Infoln("Shutting down server")
}

func (s *Server) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	log.Infoln("Closing connection")
	return
}

func (s *Server) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	var message Message
	json.Unmarshal(frame, &message)
	log.Infof("Local Address: %s, Remote Address: %s, Message : %s", c.LocalAddr().String(), c.RemoteAddr().String(), message.String())
	response := s.callback(message)
	log.Infof("Response: %s", response.String())
	out, err := json.Marshal(response)
	if err != nil {
		log.Panicf("ERROR Unmarshalling, %s", response.String())
	}
	return
}

// {"op":"GET","key":"b"}
// {"op":"PUT","key":"LOL","value":"123456"}
// {"op":"PUT","key":"LMAO","value":"abcde"}
// {"op":"GET","key":"LOL"}
// {"op":"GET","key":"LMAO"}
