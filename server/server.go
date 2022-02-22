package server

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"distributed-cache.io/cache"
	"distributed-cache.io/common"
	"distributed-cache.io/swim"
	"github.com/panjf2000/gnet"
)

type Operation string

type Server struct {
	*gnet.EventServer
	port           int
	grpcPort       int
	multicore      bool
	grpcServer     *grpc.Server
	requestHandler *RequestHandler
}

type ServerConfig struct {
	serverPort     int
	membershipPort int
	grpcPort       int
	multicore      bool
	healthyNode    string
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
	Code    int32
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
	server.grpcPort = config.grpcPort

	localCache := common.InitCache()
	grpcServer := grpc.NewServer()
	cacheService := &cache.CacheService{Cache: localCache}

	swimService := swim.InitMembershipServer(uint16(config.grpcPort), config.healthyNode)
	swimService.AddStatusChangeListener(StatusChangeListener{})

	server.grpcServer = grpcServer
	server.requestHandler = &RequestHandler{
		connections: make(map[uint32]cache.CacheClient),
		cache:       cacheService,
	}
	cache.RegisterCacheServer(grpcServer, cacheService)
	swim.RegisterSwimServer(grpcServer, swimService)

	wg := new(sync.WaitGroup)
	wg.Add(5)

	go startCodecServer(server)
	go startGrpcServer(server)
	go swimService.Begin()
	// go InitMembershipServer(MembershipConfig{
	// 	listenPort:  config.membershipPort,
	// 	healthyNode: config.healthyNode,
	// })

	wg.Wait()
	return *server
}

func startGrpcServer(server *Server) {
	ip := common.CurrentIP
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip.String(), server.grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}
	log.Infof("Starting gRPC server at %v", listener.Addr())
	if err := server.grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

func startCodecServer(server *Server) {
	err := gnet.Serve(server, fmt.Sprintf("tcp://:%d", server.port), gnet.WithMulticore(server.multicore))
	if err != nil {
		panic(err)
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
	response := s.requestHandler.processMessage(message)
	log.Infof("Response: %s", response.String())
	out, err := json.Marshal(response)
	if err != nil {
		log.Panicf("ERROR Unmarshalling, %s", response.String())
	}
	return
}

// {"op":"GET","key":"b"}
// {"op":"GET","key":"a"}
// {"op":"GET","key":"c"}
// {"op":"GET","key":"d"}
// {"op":"GET","key":"e"}
// {"op":"GET","key":"f"}
// {"op":"GET","key":"fffffffffffff"}
// {"op":"PUT","key":"LOL","value":"123456"}
// {"op":"PUT","key":"LMAO","value":"abcde"}
// {"op":"GET","key":"LOL"}
// {"op":"GET","key":"LMAO"}
// {"op":"GET","key":"TEMP"}
// {"op":"PUT","key":"TEMP","value":"HELLO WHASSSSSSUPPP - DISTRIBUTED CACHE"}
