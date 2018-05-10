package main

//import "sync"



//
// This is a outline of primary-backup replication based on a simplifed version of Viewstamp replication.
//
//
//

import (
"sync"

//"labrpc"
	"reflect"
	"log"
	"net/rpc"
	"fmt"
	"encoding/gob"
	"bytes"
)

// the 3 possible server status
const (
	NORMAL = iota
	VIEWCHANGE
	RECOVERING
)

type reqMsg struct {
	endname  interface{} // name of sending ClientEnd
	svcMeth  string      // e.g. "Raft.AppendEntries"
	argsType reflect.Type
	args     []byte
	replyCh  chan replyMsg
}

type replyMsg struct {
	ok    bool
	reply []byte
}

type ClientEnd struct {
	//endname interface{} // this end-point's name
	//ch      chan reqMsg // copy of Network.endCh
	TcpAddress string
}

type CallPBServerArgs struct{
	SvcMeth string
	SvcArgs interface{}
	SvcReplys interface{}
}

type CallPBServerReplys bool

// send an RPC, wait for the reply.
// the return value indicates success; false means that
// no reply was received from the server.

func (e *ClientEnd) Call(svcMeth string, args interface{}, reply interface{}) bool {
	fmt.Println("Enter PBServer.go Call function...")
	fmt.Println(reflect.TypeOf(args))
	fmt.Println(reflect.TypeOf(reply))
	
	encBuf := new(bytes.Buffer)
    encErr := gob.NewEncoder(encBuf).Encode(args)
    if encErr != nil {
        log.Fatal(encErr)
    }
    argsEnc := []byte(encBuf.Bytes())

    fmt.Println(reflect.TypeOf(argsEnc))
    //fmt.Println(argsEnc)
            
    encErr = gob.NewEncoder(encBuf).Encode(reply)
    if encErr != nil {
        log.Fatal(encErr)
    }
    replyEnc := encBuf.Bytes()


	callPBServerArgs := CallPBServerArgs {
		SvcMeth: svcMeth,
		SvcArgs: argsEnc,
		SvcReplys: replyEnc,
	}

	client, err := rpc.Dial("tcp", e.TcpAddress);
	if err != nil {
		log.Fatal(err)
		return false
	}

	var callPBServerReplys CallPBServerReplys

	err = client.Call("Listener.UnpackagePBServerCall", &callPBServerArgs, &callPBServerReplys)
	if err != nil {
		fmt.Printf("client Call error... %s\n", err)
		return false
	}
	return true
	
}

// PBServer defines the state of a replica server (either primary or backup)
type PBServer struct {
	mu             sync.Mutex          // Lock to protect shared access to this peer's state
	peers          []ClientEnd // RPC end points of all peers
	me             int                 // this peer's index into peers[]
	currentView    int                 // what this peer believes to be the current active view
	status         int                 // the server's current status (NORMAL, VIEWCHANGE or RECOVERING)
	lastNormalView int                 // the latest view which had a NORMAL status

	log         []interface{} // the log of "commands"
	commitIndex int           // all log entries <= commitIndex are considered to have been committed.
	memcacheAddr string

	// ... other state that you might need ...
}

// Prepare defines the arguments for the Prepare RPC
// Note that all field names must start with a capital letter for an RPC args struct
type PrepareArgs struct {
	View          int         // the primary's current view
	PrimaryCommit int         // the primary's commitIndex
	Index         int         // the index position at which the log entry is to be replicated on backups
	Entry         interface{} // the log entry to be replicated
}

// PrepareReply defines the reply for the Prepare RPC
// Note that all field names must start with a capital letter for an RPC reply struct
type PrepareReply struct {
	View    int  // the backup's current view
	Success bool // whether the Prepare request has been accepted or rejected
}

// RecoverArgs defined the arguments for the Recovery RPC
type RecoveryArgs struct {
	View   int // the view that the backup would like to synchronize with
	Server int // the server sending the Recovery RPC (for debugging)
}

type RecoveryReply struct {
	View          int           // the view of the primary
	Entries       []interface{} // the primary's log including entries replicated up to and including the view.
	PrimaryCommit int           // the primary's commitIndex
	Success       bool          // whether the Recovery request has been accepted or rejected
}

type ViewChangeArgs struct {
	View int // the new view to be changed into
}

type ViewChangeReply struct {
	LastNormalView int           // the latest view which had a NORMAL status at the server
	Log            []interface{} // the log at the server
	Success        bool          // whether the ViewChange request has been accepted/rejected
}

type StartViewArgs struct {
	View int           // the new view which has completed view-change
	Log  []interface{} // the log associated with the new new
}

type StartViewReply struct {
}

// GetPrimary is an auxilary function that returns the server index of the
// primary server given the view number (and the total number of replica servers)
func GetPrimary(view int, nservers int) int {
	return view % nservers
}

//// IsCommitted is called by tester to check whether an index position
//// has been considered committed by this server
//func (srv *PBServer) IsCommitted(index int) (committed bool) {
//	srv.mu.Lock()
//	defer srv.mu.Unlock()
//	if srv.commitIndex >= index {
//		return true
//	}
//	return false
//}

//// ViewStatus is called by tester to find out the current view of this server
//// and whether this view has a status of NORMAL.
//func (srv *PBServer) ViewStatus() (currentView int, statusIsNormal bool) {
//	srv.mu.Lock()
//	defer srv.mu.Unlock()
//	return srv.currentView, srv.status == NORMAL
//}

//// GetEntryAtIndex is called by tester to return the command replicated at
//// a specific log index. If the server's log is shorter than "index", then
//// ok = false, otherwise, ok = true
//func (srv *PBServer) GetEntryAtIndex(index int) (ok bool, command interface{}) {
//	srv.mu.Lock()
//	defer srv.mu.Unlock()
//	if len(srv.log) > index {
//		return true, srv.log[index]
//	}
//	return false, command
//}

//// Kill is called by tester to clean up (e.g. stop the current server)
//// before moving on to the next test
//func (srv *PBServer) Kill() {
//	// Your code here, if necessary
//}

// Make is called by tester to create and initalize a PBServer
// peers is the list of RPC endpoints to every server (including self)
// me is this server's index into peers.
// startingView is the initial view (set to be zero) that all servers start in
func Make(peers []ClientEnd, me int, startingView int, memAddr string) *PBServer {
	srv := &PBServer{
		peers:          peers,
		me:             me,
		currentView:    startingView,
		lastNormalView: startingView,
		status:         NORMAL,
		memcacheAddr:	memAddr,
	}
	// all servers' log are initialized with a dummy command at index 0
	var v interface{}
	srv.log = append(srv.log, v)

	// Your other initialization code here, if there's any
	return srv
}

// Start() is invoked by tester on some replica server to replicate a
// command.  Only the primary should process this request by appending
// the command to its log and then return *immediately* (while the log is being replicated to backup servers).
// if this server isn't the primary, returns false.
// Note that since the function returns immediately, there is no guarantee that this command
// will ever be committed upon return, since the primary
// may subsequently fail before replicating the command to all servers
//
// The first return value is the index that the command will appear at
// *if it's eventually committed*. The second return value is the current
// view. The third return value is true if this server believes it is
// the primary.
func (srv *PBServer) Start(command interface{}) (
	index int, view int, ok bool) {
	fmt.Println("Enter PBServer.go Start function")
	fmt.Println(reflect.TypeOf(command))
	srv.mu.Lock()
	defer srv.mu.Unlock()
	// do not process command if status is not NORMAL
	// and if i am not the primary in the current view
	if srv.status != NORMAL {
		return -1, srv.currentView, false
	} else if GetPrimary(srv.currentView, len(srv.peers)) != srv.me {
		return -1, srv.currentView, false
	}

	// Your code here
	index = len(srv.log)
	prepareReplys := []*PrepareReply{}
	for peer := range srv.peers {
		
		encBuf := new(bytes.Buffer)
        encErr := gob.NewEncoder(encBuf).Encode(command)
        if encErr != nil {
            log.Fatal(encErr)
        }
        commandEnc := []byte(encBuf.Bytes())

		//gob.Register(Comm{})
		prepareArgs := &PrepareArgs{
			View:         srv.currentView,
			PrimaryCommit: srv.commitIndex,
			Index:         index,
			Entry:         commandEnc,
		}
		var prepareReply PrepareReply
		prepareReplys = append(prepareReplys, &prepareReply)
		srv.sendPrepare(peer, prepareArgs, &prepareReply)
	}
	replyNum := 0
	for _,prepareReply := range prepareReplys{
		if (prepareReply.Success){
			replyNum++
		}
	}
	if (replyNum >= len(srv.peers)/2+1){
		srv.commitIndex = index
	}

	view = srv.currentView
	ok = true
	
	
	/*go func(srv *PBServer, command interface{}, logLen int) {
		encBuf := new(bytes.Buffer)
        encErr := gob.NewEncoder(encBuf).Encode(command)
        if encErr != nil {
            log.Fatal(encErr)
        }
        commandEnc := []byte(encBuf.Bytes())
		prepareArgs := &PrepareArgs {View: srv.currentView, PrimaryCommit: srv.commitIndex, Index: logLen - 1, Entry: commandEnc}
		prepareReply := &PrepareReply {}
		var numResponse int
		for i := 0; i < len(srv.peers); i++ {
			fmt.Printf("Send prepare to %d\n", i)
			isSend := srv.sendPrepare(i, prepareArgs, prepareReply)
			if isSend && prepareReply.Success == true {
				numResponse++
				if numResponse > len(srv.peers) / 2 && srv.commitIndex < logLen - 1 {
					srv.commitIndex = logLen - 1
				}
			}
		}
	} (srv, command, len(srv.log))
	
	fmt.Println("Time to return value")
	index = len(srv.log) - 1
	view = srv.currentView
	ok = true*/
	

	return index, view, ok
}

// exmple code to send an AppendEntries RPC to a server.
// server is the index of the target server in srv.peers[].
// expects RPC arguments in args.
// The RPC library fills in *reply with RPC reply, so caller should pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
func (srv *PBServer) sendPrepare(server int, args *PrepareArgs, reply *PrepareReply) bool {
	fmt.Println("Enter PBServer.go sendPrepare function")
	ok := srv.peers[server].Call("PBServer.Prepare", args, reply)
	return ok
}

// Prepare is the RPC handler for the Prepare RPC
func (srv *PBServer) Prepare(args *PrepareArgs, reply *PrepareReply, l *Listener) {
	fmt.Println("Enter PBServer.go Prepare function")
	// Your code here
	if args.Index > len(srv.log) {
		recoveryArgs := &RecoveryArgs{
			View:	args.View,
			Server:	srv.me,
		}
		var recoveryReply RecoveryReply
		primary := GetPrimary(args.View, len(srv.peers))
		ok := srv.peers[primary].Call("PBServer.Recovery", recoveryArgs, &recoveryReply)
		if !ok {
			return//time.Sleep(time.Second/6)
		}
		if recoveryReply.Success{
			srv.currentView = recoveryReply.View
			srv.log = recoveryReply.Entries
			srv.commitIndex = recoveryReply.PrimaryCommit
			if args.Index == len(srv.log) {
				//srv.log = append(srv.log,args.Entry)
				srv.LogAppend(args.Entry, l)
			}
			reply.View = srv.currentView
			reply.Success = true
		}
		return
	}
	if args.View != srv.currentView || args.Index != len(srv.log){
		reply.View = srv.currentView
		reply.Success = false
		return
	}
	//srv.log = append(srv.log,args.Entry)
	
	srv.LogAppend(args.Entry, l)
	
	/*fmt.Println("a test 2")

	decBuf := bytes.NewBuffer((args.Entry).([]byte))
    comm := Comm{}
    decErr := gob.NewDecoder(decBuf).Decode(&comm)
    if decErr != nil {
        log.Fatal(decErr)
    }
	
	fmt.Println(comm)
	
	decBuf = bytes.NewBuffer((comm.Reply).([]byte))
    var registerReply RegisterReply
    decErr = gob.NewDecoder(decBuf).Decode(&registerReply)
    if decErr != nil {
    	log.Fatal(decErr)
    }
    fmt.Println(registerReply)
	*/
	
	reply.View = srv.currentView
	reply.Success = true
	return
}

// Recovery is the RPC handler for the Recovery RPC
func (srv *PBServer) Recovery(args *RecoveryArgs, reply *RecoveryReply) {
	fmt.Println("Enter PBServer.go Recovery function")
	// Your code here
	if args.View == srv.currentView {
		reply.View = srv.currentView
		reply.Entries = srv.log
		reply.PrimaryCommit = srv.commitIndex
		reply.Success = true
	}
}

// Some external oracle prompts the primary of the newView to
// switch to the newView.
// PromptViewChange just kicks start the view change protocol to move to the newView
// It does not block waiting for the view change process to complete.
func (srv *PBServer) PromptViewChange(newView int) {
	fmt.Println("Enter PBServer.go promptViewChange function")
	srv.mu.Lock()
	defer srv.mu.Unlock()
	newPrimary := GetPrimary(newView, len(srv.peers))

	if newPrimary != srv.me { //only primary of newView should do view change
		return
	} else if newView <= srv.currentView {
		return
	}
	vcArgs := &ViewChangeArgs{
		View: newView,
	}
	vcReplyChan := make(chan *ViewChangeReply, len(srv.peers))
	// send ViewChange to all servers including myself
	for i := 0; i < len(srv.peers); i++ {
		go func(server int) {
			var reply ViewChangeReply
			ok := srv.peers[server].Call("PBServer.ViewChange", vcArgs, &reply)
			// fmt.Printf("node-%d (nReplies %d) received reply ok=%v reply=%v\n", srv.me, nReplies, ok, r.reply)
			if ok {
				vcReplyChan <- &reply
			} else {
				vcReplyChan <- nil
			}
		}(i)
	}

	// wait to receive ViewChange replies
	// if view change succeeds, send StartView RPC
	go func() {
		var successReplies []*ViewChangeReply
		var nReplies int
		majority := len(srv.peers)/2 + 1
		for r := range vcReplyChan {
			nReplies++
			if r != nil && r.Success {
				successReplies = append(successReplies, r)
			}
			if nReplies == len(srv.peers) || len(successReplies) == majority {
				break
			}
		}
		ok, log := srv.determineNewViewLog(successReplies)
		if !ok {
			return
		}
		svArgs := &StartViewArgs{
			View: vcArgs.View,
			Log:  log,
		}
		// send StartView to all servers including myself
		for i := 0; i < len(srv.peers); i++ {
			var reply StartViewReply
			go func(server int) {
				// fmt.Printf("node-%d sending StartView v=%d to node-%d\n", srv.me, svArgs.View, server)
				srv.peers[server].Call("PBServer.StartView", svArgs, &reply)
			}(i)
		}
	}()
}

// determineNewViewLog is invoked to determine the log for the newView based on
// the collection of replies for successful ViewChange requests.
// if a quorum of successful replies exist, then ok is set to true.
// otherwise, ok = false.
func (srv *PBServer) determineNewViewLog(successReplies []*ViewChangeReply) (
	ok bool, newViewLog []interface{}) {
	fmt.Println("Enter PBServer.go determineNewViewLog function")
	// Your code here
	if len(successReplies) < len(srv.peers)/2+1 {
		return false, nil
	}
	maxView := 0
	for _,reply := range successReplies{
		if reply.LastNormalView > maxView {
			maxView = reply.LastNormalView
		}
	}
	maxLog := 0
	var maxReply *ViewChangeReply
	for _,reply := range successReplies{
		if reply.LastNormalView == maxView && maxLog < len(reply.Log) {
			maxLog = len(reply.Log)
			maxReply = reply
		}
	}
	return true, maxReply.Log
}

// ViewChange is the RPC handler to process ViewChange RPC.
func (srv *PBServer) ViewChange(args *ViewChangeArgs, reply *ViewChangeReply) {
	fmt.Println("Enter PBServer.go ViewChange function")
	// Your code here
	if args.View <= srv.currentView {
		reply.Success = false
		return
	}
	srv.currentView = args.View
	srv.status = VIEWCHANGE
	reply.Success = true
	reply.LastNormalView = srv.lastNormalView
	reply.Log = srv.log
}

// StartView is the RPC handler to process StartView RPC.
func (srv *PBServer) StartView(args *StartViewArgs, reply *StartViewReply, l *Listener) {
	fmt.Println("Enter PBServer.go StartView function")
	// Your code here
	if args.View < srv.currentView {
		return
	}
	srv.currentView = args.View


	//srv.log = args.Log
	srv.SynchronizeLog(args.Log,l)
	srv.status = NORMAL
}

func (srv *PBServer) SynchronizeLog(newLogs []interface{}, l *Listener){
	fmt.Println("Enter PBServer.go SynchronizeLog function")
	distance := len(newLogs)-len(srv.log)
	for i := 1; i <= distance; i++{
		srv.LogAppend(newLogs[-i], l)
	}
}

func (srv *PBServer) LogAppend(entry interface{}, l *Listener) bool{
	fmt.Println("Enter PBServer.go LogAppend function")
	//srv.mu.Lock()
	//defer srv.mu.Unlock()
	srv.log = append(srv.log, entry)
	decBuf := bytes.NewBuffer((entry).([]byte))
    comm := Comm{}
    decErr := gob.NewDecoder(decBuf).Decode(&comm)
    if decErr != nil {
        log.Fatal(decErr)
    }
	
	fmt.Println(comm)
	if comm.BEMeth == "Listener.UserRegister" {
		fmt.Println("Log Append on UserRegister...")
		decBuf := bytes.NewBuffer((comm.Args).([]byte))
        user := User{}
        decErr := gob.NewDecoder(decBuf).Decode(&user)
        if decErr != nil {
            log.Fatal(decErr)
        }

        decBuf = bytes.NewBuffer((comm.Reply).([]byte))
        var registerReply RegisterReply
        decErr = gob.NewDecoder(decBuf).Decode(&registerReply)
        if decErr != nil {
            log.Fatal(decErr)
        }
		l.UserRegister(&user, &registerReply)
		fmt.Println(registerReply)
		resp = registerReply
		fmt.Printf("User %s Register %v\n", user.Username, registerReply)
		return true
	}
	if comm.BEMeth ==  "Listener.UserLogin" {
		fmt.Println("Log Append on UserLogin...")
		decBuf := bytes.NewBuffer((comm.Args).([]byte))
        user := User{}
        decErr := gob.NewDecoder(decBuf).Decode(&user)
        if decErr != nil {
            log.Fatal(decErr)
        }

        decBuf = bytes.NewBuffer((comm.Reply).([]byte))
        var loginReply LoginReply
        decErr = gob.NewDecoder(decBuf).Decode(&loginReply)
        if decErr != nil {
            log.Fatal(decErr)
        }
		l.UserLogin(&user, &loginReply)
		resp = loginReply
		fmt.Printf("User %s Login %v\n", user.Username, loginReply)
		//l.UserLogin((comm.Args).(*User), (comm.Reply).(*LoginReply))
		return true
	}
	if comm.BEMeth ==  "Listener.UserInfo" {
		fmt.Println("Log Append on UserInfo...")
		decBuf := bytes.NewBuffer((comm.Args).([]byte))
        user := User{}
        decErr := gob.NewDecoder(decBuf).Decode(&user)
        if decErr != nil {
            log.Fatal(decErr)
        }

        decBuf = bytes.NewBuffer((comm.Reply).([]byte))
        userRep := User{}
        decErr = gob.NewDecoder(decBuf).Decode(&userRep)
        if decErr != nil {
            log.Fatal(decErr)
        }
		l.UserInfo(&user, &userRep)
		resp = userRep
		fmt.Printf("User %s getUserInfo\n", user.Username)
		//l.UserInfo((comm.Args).(*User), (comm.Reply).(*User))
		return true
	}
	if comm.BEMeth ==  "Listener.UserHome" {
		fmt.Println("Log Append on UserHome...")
		decBuf := bytes.NewBuffer((comm.Args).([]byte))
        user := User{}
        decErr := gob.NewDecoder(decBuf).Decode(&user)
        if decErr != nil {
            log.Fatal(decErr)
        }

        decBuf = bytes.NewBuffer((comm.Reply).([]byte))
        messageBox := MessageBox{}
        decErr = gob.NewDecoder(decBuf).Decode(&messageBox)
        if decErr != nil {
            log.Fatal(decErr)
        }
		l.UserHome(&user, &messageBox)
		resp = messageBox
		fmt.Printf("User %s Home Page\n", user.Username)
		return true
	}
	if comm.BEMeth ==  "Listener.UserFollow" {
		fmt.Println("Log Append on UserFollow...")
		decBuf := bytes.NewBuffer((comm.Args).([]byte))
        userSearch := UserSearch{}
        decErr := gob.NewDecoder(decBuf).Decode(&userSearch)
        if decErr != nil {
            log.Fatal(decErr)
        }

        decBuf = bytes.NewBuffer((comm.Reply).([]byte))
        userSearchReply := UserSearchReply{}
        decErr = gob.NewDecoder(decBuf).Decode(&userSearchReply)
        if decErr != nil {
            log.Fatal(decErr)
        }
		l.UserFollow(&userSearch, &userSearchReply)
		resp = userSearchReply
		fmt.Printf("User %s Search User %s\n", userSearch.Userinfo.Username, userSearch.Followname)
		//l.UserFollow((comm.Args).(*UserSearch), (comm.Reply).(*UserSearchReply))
		return true
	}
	if comm.BEMeth ==  "Listener.UserAdd" {
		fmt.Println("Log Append on UserAdd...")
		decBuf := bytes.NewBuffer((comm.Args).([]byte))
        userAdd := UserAdd{}
        decErr := gob.NewDecoder(decBuf).Decode(&userAdd)
        if decErr != nil {
            log.Fatal(decErr)
        }

        decBuf = bytes.NewBuffer((comm.Reply).([]byte))
        var userAddReply UserAddReply
        decErr = gob.NewDecoder(decBuf).Decode(&userAddReply)
        if decErr != nil {
            log.Fatal(decErr)
        }
		l.UserAdd(&userAdd, &userAddReply)
		resp = userAddReply
		fmt.Printf("User %s Add User %s\n", userAdd.Username, userAdd.Followname)
		//l.UserFollow((comm.Args).(*UserSearch), (comm.Reply).(*UserSearchReply))
		return true
	}
	if comm.BEMeth ==  "Listener.UserPost" {
		fmt.Println("Log Append on UserPost...")
		decBuf := bytes.NewBuffer((comm.Args).([]byte))
        userPost := UserPost{}
        decErr := gob.NewDecoder(decBuf).Decode(&userPost)
        if decErr != nil {
            log.Fatal(decErr)
        }

        decBuf = bytes.NewBuffer((comm.Reply).([]byte))
        var userPostReply UserPostReply
        decErr = gob.NewDecoder(decBuf).Decode(&userPostReply)
        if decErr != nil {
            log.Fatal(decErr)
        }
		l.UserPost(&userPost, &userPostReply)
		resp = userPostReply
		fmt.Printf("User %s Post Content %s\n", userPost.Userinfo.Username, userPost.Content)
		return true
	}
	if comm.BEMeth ==  "Listener.UserCancel" {
		fmt.Println("Log Append on UserCancel...")
		decBuf := bytes.NewBuffer((comm.Args).([]byte))
        user := User{}
        decErr := gob.NewDecoder(decBuf).Decode(&user)
        if decErr != nil {
            log.Fatal(decErr)
        }

        decBuf = bytes.NewBuffer((comm.Reply).([]byte))
        var userCancelReply UserCancelReply
        decErr = gob.NewDecoder(decBuf).Decode(&userCancelReply)
        if decErr != nil {
            log.Fatal(decErr)
        }
		l.UserCancel(&user, &userCancelReply)
		resp = userCancelReply
		fmt.Printf("User %s Cancel Account %v\n", user.Username, userCancelReply)
		return true
	}

	return false
}