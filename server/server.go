package server

import (
	"encoding/json"
	"fmt"
	"log"

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

func InitServer(port int, multicore bool, callback Callback) Server {
	server := new(Server)
	server.port = port
	server.multicore = multicore
	server.cache = common.InitCache()
	if callback == nil {
		server.callback = server.defaultCallback()
	} else {
		server.callback = callback
	}
	err := gnet.Serve(server, fmt.Sprintf("tcp://:%d", port), gnet.WithMulticore(multicore))
	if err != nil {
		panic(err)
	}
	return *server
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
	log.Printf("Test codec server is listening on %s (multi-cores: %t, loops: %d)\n", srv.Addr.String(), srv.Multicore, srv.NumEventLoop)
	return
}

func (s *Server) OnShutdown(srv gnet.Server) {
	log.Println("Shutting down server")
}

func (s *Server) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	log.Println("Closing connection")
	return
}

func (s *Server) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	var message Message
	json.Unmarshal(frame, &message)
	log.Printf("Message : %s\n", message.String())
	response := s.callback(message)
	out, err := json.Marshal(response)
	if err != nil {
		log.Panicf("ERROR Unmarshalling, %s", response.String())
	}
	return
}

// {"op":"GET","key":"b"}
// {"op":"PUT","key":"LOL","value":"12345"}
// {"op":"PUT","key":"LMAO","value":"abcde"}
// {"op":"GET","key":"LOL"}
// {"op":"GET","key":"LMAO"}
