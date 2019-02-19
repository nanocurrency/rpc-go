package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"pkg/nanoipc"
	"strconv"
	"sync/atomic"
)

// RestServer is a REST interface to the Node IPC
type RestServer struct {
	// Session pool
	sessions      []*nanoipc.Session
	nextsession   int32
	conf          *_Conf
	needReconnect bool
}

type _Conf struct {
	Port     int       `json:"port"`
	Hostname string    `json:"hostname"`
	Node     _ConfNode `json:"node"`
}

type _ConfNode struct {
	Connection string `json:"connection"`
	Poolsize   int    `json:"poolsize"`
}

// Start serving requests
func (server *RestServer) listen() {

	server.conf = &_Conf{
		Hostname: "", Port: 8080,
		Node: _ConfNode{Connection: "local:///tmp/nano", Poolsize: 1},
	}

	if configBytes, err := ioutil.ReadFile("config.json"); err != nil {
		log.Println("No config file found, using defaults")
	} else {
		if err := json.Unmarshal(configBytes, server.conf); err != nil {
			log.Fatal(err)
		}
	}

	server.tryConnectNode()

	http.HandleFunc("/", server.handler)
	http.ListenAndServe(server.conf.Hostname+":"+strconv.Itoa(server.conf.Port), nil)
}

// Try connecting to the Nano node
func (server *RestServer) tryConnectNode() *nanoipc.Error {
	var err *nanoipc.Error
	server.sessions = make([]*nanoipc.Session, 0)
	for i := 0; err == nil && i < server.conf.Node.Poolsize; i++ {
		session := &nanoipc.Session{}
		server.sessions = append(server.sessions, session)
		err = session.Connect(server.conf.Node.Connection)
	}
	if len(server.sessions) < server.conf.Node.Poolsize {
		log.Println("Reconnection attempt required")
		server.needReconnect = true
	} else {
		server.needReconnect = false
	}

	if err != nil {
		log.Println(err.Message)
	}

	return err
}

func (server *RestServer) reconnectNode() *nanoipc.Error {
	for _, element := range server.sessions {
		element.Close()
	}
	server.sessions = nil
	server.nextsession = 0
	return server.tryConnectNode()
}

// getSession returns the next available session in a round-robin fashion
func (server *RestServer) getSession() *nanoipc.Session {
	next := atomic.AddInt32(&server.nextsession, 1)
	next = next % int32(server.conf.Node.Poolsize)
	return server.sessions[next]
}

// Request handler. This automatically translates between JSON and protobuf messages.
func (server *RestServer) handler(resp http.ResponseWriter, req *http.Request) {

	if req.Method == "POST" {
		// Reconnect if necessary
		if server.needReconnect {
			if err := server.tryConnectNode(); err != nil {
				if json, jsonErr := json.Marshal(err); jsonErr == nil {
					resp.Write(json)
				}
				return
			}
			log.Println("Reconnected successfully to node")
		}

		sc := &nanoipc.CallChain{}
		var responseBytes []byte
		var bodyString string
		sc.Do(func() {
			if bodyBytes, err := ioutil.ReadAll(req.Body); err != nil {
				sc.Err = &nanoipc.Error{Code: 1, Message: err.Error(), Category: "REST"}
			} else {
				bodyString = string(bodyBytes)
				log.Println(bodyString)
			}
		}).Do(func() {
			if rb, err := server.getSession().Request(bodyString); err != nil {
				sc.Err = err
				log.Println("Request failed:")
				log.Println(err.Error())
			} else {
				responseBytes = rb
			}
		}).Do(func() {
			// log.Println(string(responseBytes))
			resp.Write(responseBytes)
		}).Failure(func() {
			log.Println("Request failed")
			log.Println(sc.Err)

			if sc.Err.Category == "Network" {
				if err := server.reconnectNode(); err != nil {
					log.Println("Unable to reconnect to node")
				}
			}
		})
	} else {
		resp.Write([]byte("Invalid request method. Use POST."))
	}
}

// Start REST server
func main() {
	s := &RestServer{needReconnect: false}
	s.listen()
}
