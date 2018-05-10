package main

import (
	"net/rpc"
	"log"
	"fmt"
)

var memcacheAddrPrefix = "127.0.0.1:11211"

type StartBackendArgs struct {
    ClientEnds []ClientEnd
    Pos int
    MemAddr string
}

type StartBackendReply bool

type ClientEnd struct {
	TcpAddress string
}

func main() {
	serverNum := 3

	var cliendEnds []ClientEnd
	cliendEnds = make([]ClientEnd,serverNum)

	memAddrs := [3]string{}

	//var backends []*Backend
	rpcAddrs := []string{"52.87.162.252:42586", "54.146.6.116:42586", "54.88.235.80:42586"}
	for j := 0; j < serverNum; j++ {
		cliendEnds[j].TcpAddress = rpcAddrs[j]
		memAddrs[j] = memcacheAddrPrefix
	}
	//start backend
	for j := 0; j < serverNum; j++{
		client, err := rpc.Dial("tcp", rpcAddrs[j]);
		if err != nil {
			log.Fatal(err)
			//flag=false
		} else {
			startBackendArg := StartBackendArgs{
				ClientEnds: cliendEnds,
				Pos: j,
				MemAddr: memAddrs[j],
			}
			var startBackendReply bool
			err = client.Call("Listener.StartBackend", &startBackendArg, &startBackendReply)
			if err != nil {
				fmt.Print("client error:", err)
				//flag=false
			}
		}
	}

}
