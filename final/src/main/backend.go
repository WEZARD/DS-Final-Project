package main

import (
    "fmt"
    "log"
    "sort"
    "github.com/bradfitz/gomemcache/memcache"
    "time"
    "encoding/gob"
    "bytes"
    "net/rpc"
    "net"
    "errors"
)



type User struct { // capitalize the first letter of each field name as "exported" for gob
    Username string
    Password string
    Following map[string]bool
    Follower map[string]bool
    Messages map[time.Time]string
}

type Message struct {
    Username string
    Timestamp time.Time
    Text string
    DisplayTime string
}

type MessageBox struct {
    Username string
    Following map[string]bool
    Follower map[string]bool
    Messages []Message
    UserMessages []Message
    Status bool
}

type UserSearch struct {
    Followname string
    Userinfo *User
}

type UserSearchReply struct {
    Messages MessageBox
    Reply bool
}

type UserAdd struct {
    Followname string
    Username string
}

type UserPost struct {
    Userinfo *User
    Content string
}

type RegisterReply bool
type LoginReply bool
type UserAddReply bool
type UserPostReply bool
type UserCancel bool
type UserCancelReply bool

type Listener int

func (l *Listener) UserRegister(args *User, rep *RegisterReply) error {
    fmt.Println("This is UserRegister Listener...")
    mc := memcache.New(pbServer.memcacheAddr) // brew services restart memcached, echo 'flush_all' | nc localhost 11211
        
    // ----------------------------------------Test Case---------------------------------------------
    testTime1 := time.Date(2018, 4, 4, 20, 00, 00, 651387237, time.UTC)
    testTime2 := time.Date(2018, 4, 5, 20, 00, 00, 651387237, time.UTC)
    messages1 := map[time.Time]string {testTime1: "Service number: 9292888888!", testTime2: "Welcome to our web service!"}
    messages2 := map[time.Time]string {testTime1: "Post what you like!", testTime2: "Follow the person you're interested!"}
    testUser1 := User {
        Username: "CustomerService",
        Messages: messages1,
    }

    testUser2 := User {
        Username: "ApplicationCenter",
        Messages: messages2,
    }

    _, memErr := mc.Get("CustomerService")
        
    if memErr == nil {

    } else {
        fmt.Println(memErr)
        // Encode struct to byte array
        encBuf := new(bytes.Buffer)
        encErr := gob.NewEncoder(encBuf).Encode(testUser1)
        if encErr != nil {
            log.Fatal(encErr)
        }
        value := encBuf.Bytes()
        mc.Set(&memcache.Item{Key: "CustomerService", Value: []byte(value)}) // store in memcache
    }

    _, memErr = mc.Get("ApplicationCenter")
        
    if memErr == nil {

    } else {
        fmt.Println(memErr)
        // Encode struct to byte array
        encBuf := new(bytes.Buffer)
        encErr := gob.NewEncoder(encBuf).Encode(testUser2)
        if encErr != nil {
            log.Fatal(encErr)
        }
        value := encBuf.Bytes()
        mc.Set(&memcache.Item{Key: "ApplicationCenter", Value: []byte(value)}) // store in memcache
    }

    // ----------------------------------------------------------------------------------------------
        
    username := args.Username
    password := args.Password
        
    following := map[string]bool {testUser1.Username: true, testUser2.Username: true}
    follower := map[string]bool {testUser1.Username: true, testUser2.Username: true}
    //following := make(map[string]bool)
    //follower := make(map[string]bool)
    newMessages := make(map[time.Time]string)

    user := User {
        Username: username,
        Password: password,
        Following: following,
        Follower: follower,
        Messages: newMessages,
    } // initialize all the elements
    _, memErr = mc.Get(username)
        
    if memErr == nil {
        *rep = false
    } else {
        fmt.Println(memErr)
        
        val, memErr := mc.Get("ApplicationCenter")
        if memErr == nil {
            // Decode byte array to struct
            decBuf := bytes.NewBuffer(val.Value)
            userOut := User{}
            decErr := gob.NewDecoder(decBuf).Decode(&userOut)
            if decErr != nil {
                log.Fatal(decErr)
            }
            now := time.Now()
            userOut.Messages[now] = fmt.Sprintf("Welcome new user: %s! Search to see the posts!", username)
            mc := memcache.New(pbServer.memcacheAddr)
            mc.Delete("ApplicationCenter")
            encBuf := new(bytes.Buffer)
            encErr := gob.NewEncoder(encBuf).Encode(userOut)
            if encErr != nil {
                log.Fatal(encErr)
            }
            value := encBuf.Bytes()
            mc.Set(&memcache.Item{Key: "ApplicationCenter", Value: []byte(value)})
        }
        
        // Encode struct to byte array
        encBuf := new(bytes.Buffer)
        encErr := gob.NewEncoder(encBuf).Encode(user)
        if encErr != nil {
            log.Fatal(encErr)
        }
        value := encBuf.Bytes()
        mc.Set(&memcache.Item{Key: username, Value: []byte(value)}) // store in memcache  
        
        *rep = true
    }
    return nil
}

func (l *Listener) UserLogin(args *User, rep *LoginReply) error {
    fmt.Println("This is UserLogin Listener...")
    mc := memcache.New(pbServer.memcacheAddr)
    
    username := args.Username
    password := args.Password
        
    val, memErr := mc.Get(username)

    if memErr == nil {
        // Decode byte array to struct
        decBuf := bytes.NewBuffer(val.Value)
        userOut := User{}
        decErr := gob.NewDecoder(decBuf).Decode(&userOut)
        if decErr != nil {
            log.Fatal(decErr)
        }
        if password == userOut.Password {
            *rep = true
        } else {
            *rep = false
        }
    } else {
        *rep = false
    }
    return nil
}

func (l *Listener) UserInfo(args *User, rep *User) error {
    fmt.Println("This is UserInfo Listener...")
    username := args.Username
    mc := memcache.New(pbServer.memcacheAddr)
    val, _ := mc.Get(username)
    // Decode byte array to struct
    decBuf := bytes.NewBuffer(val.Value)
    decErr := gob.NewDecoder(decBuf).Decode(rep)
    if decErr != nil {
        log.Fatal(decErr)
    }
    return nil
}

func (l *Listener) UserHome(userInfo *User, homeInfoRep *MessageBox) error {
    fmt.Println("This is UserHome Listener...")
    mc := memcache.New(pbServer.memcacheAddr)
    messageBox := MessageBox{}
    messages := []Message{}
    singleMessage := Message{}
        
    for k := range userInfo.Following {
        val, memErr := mc.Get(k)
        if memErr == nil {
            decBuf := bytes.NewBuffer(val.Value)
            userOut := User{}
            decErr := gob.NewDecoder(decBuf).Decode(&userOut)
            if decErr != nil {
                log.Fatal(decErr)
            }
            for k1, v1 := range userOut.Messages {
                singleMessage.Username = k
                singleMessage.Timestamp = k1
                singleMessage.Text = v1
                messages = append(messages, singleMessage)
            }
        } else {
            log.Fatal(memErr)
        }
    }

    sort.Slice(messages, func(i, j int) bool { 
        return messages[i].Timestamp.After(messages[j].Timestamp)
    })
    for i := 0; i < len(messages); i++ {
        messages[i].DisplayTime = messages[i].Timestamp.Format("Mon Jan _2 15:04:05 2006")
    }

    userMessages := []Message{}
    for k,v := range userInfo.Messages {
        singleMessage.Username = userInfo.Username
        singleMessage.Timestamp = k
        singleMessage.Text = v
        userMessages = append(userMessages, singleMessage)
    }
    sort.Slice(userMessages, func(i, j int) bool { 
        return userMessages[i].Timestamp.After(userMessages[j].Timestamp)
    })
    for i := 0; i < len(userMessages); i++ {
        userMessages[i].DisplayTime = userMessages[i].Timestamp.Format("Mon Jan _2 15:04:05 2006")
    }

    messageBox.Username = userInfo.Username
    messageBox.Following = userInfo.Following
    messageBox.Follower = userInfo.Follower
    messageBox.Messages = messages
    messageBox.UserMessages = userMessages

    *homeInfoRep = messageBox
    
    return nil
}

func (l *Listener) UserFollow(followArgs *UserSearch, followInfoRep *UserSearchReply) error {
    fmt.Println("This is UserFollow Listener...")
    followname := followArgs.Followname
    mc := memcache.New(pbServer.memcacheAddr)
    val, memErr := mc.Get(followname)
        
    if memErr == nil {
        messageBox := MessageBox{}
        messages := []Message{}
        singleMessage := Message{}
        decBuf := bytes.NewBuffer(val.Value)
        userOut := User{}
        decErr := gob.NewDecoder(decBuf).Decode(&userOut)
        if decErr != nil {
            log.Fatal(decErr)
        }
        for k1, v1 := range userOut.Messages {
            singleMessage.Username = followname
            singleMessage.Timestamp = k1
            singleMessage.Text = v1
            messages = append(messages, singleMessage)
        }

        for i := 0; i < len(messages); i++ {
        messages[i].DisplayTime = messages[i].Timestamp.Format("Mon Jan _2 15:04:05 2006")
        }

        userInfo := followArgs.Userinfo
        status := true
        if _ , ok := userInfo.Following[followname]; ok {
            status = false
        }
            
        messageBox.Username = followname
        messageBox.Messages = messages
        messageBox.Status = status

        *followInfoRep = UserSearchReply {
            Messages: messageBox,
            Reply: true,
        }
    } else {
        *followInfoRep = UserSearchReply {
            Reply: false,
        }
    }
    return nil
}

func (l *Listener) UserAdd(addArgs *UserAdd, userAddRep *UserAddReply) error {
    fmt.Println("This is UserAdd Listener...")
    mc := memcache.New(pbServer.memcacheAddr)

    username := addArgs.Username
    followname := addArgs.Followname

    val, _ := mc.Get(username)
    decBuf := bytes.NewBuffer(val.Value)
    userSrc := User{}
    decErr := gob.NewDecoder(decBuf).Decode(&userSrc)
    if decErr != nil {
        log.Fatal(decErr)
    }

    val, _ = mc.Get(followname)
    decBuf = bytes.NewBuffer(val.Value)
    userDest := User{}
    decErr = gob.NewDecoder(decBuf).Decode(&userDest)
    if decErr != nil {
        log.Fatal(decErr)
    }

    userSrc.Following[followname] = true
    userDest.Follower[username] = true

    mc.Delete(username)
    encBuf := new(bytes.Buffer)
    encErr := gob.NewEncoder(encBuf).Encode(userSrc)
    if encErr != nil {
        log.Fatal(encErr)
    }
    value := encBuf.Bytes()
    mc.Set(&memcache.Item{Key: username, Value: []byte(value)})

    mc.Delete(followname)
    encBuf = new(bytes.Buffer)
    encErr = gob.NewEncoder(encBuf).Encode(userDest)
    if encErr != nil {
        log.Fatal(encErr)
    }
    value = encBuf.Bytes()
    mc.Set(&memcache.Item{Key: followname, Value: []byte(value)})

    *userAddRep = true

    return nil
}

func (l *Listener) UserPost(postArgs *UserPost, userPostRep *UserPostReply) error {
    fmt.Println("This is UserPost Listener...")
    now := time.Now()
    userInfo := postArgs.Userinfo
    content := postArgs.Content
    username := userInfo.Username
    userInfo.Messages[now] = content
    mc := memcache.New(pbServer.memcacheAddr)
    mc.Delete(username)
    encBuf := new(bytes.Buffer)
    encErr := gob.NewEncoder(encBuf).Encode(userInfo)
    if encErr != nil {
        log.Fatal(encErr)
    }
    value := encBuf.Bytes()
    mc.Set(&memcache.Item{Key: username, Value: []byte(value)})

    *userPostRep = true

    return nil
}

func (l *Listener) UserCancel(cancelArgs *User, userCancelRep *UserCancelReply) error {
    fmt.Println("This is UserCancel Listener...")
    mc := memcache.New(pbServer.memcacheAddr)
    username := cancelArgs.Username
    delErr := mc.Delete(username)
    if delErr != nil {
        *userCancelRep = false
    } else {
        *userCancelRep = true
    }
    return nil
}


//----------------------------------------   PART III   --------------------------------------------
//type Backend struct{
//  //BackendRpcAddr string
//  //Clientend ClientEnd
//  PbServer *PBServer
//}

var pbServer *PBServer

type StartBackendArgs struct {
    ClientEnds []ClientEnd
    Pos int
    MemAddr string
}

type StartBackendReply bool

func (l *Listener) StartBackend(args *StartBackendArgs, reply *StartBackendReply) error {//clientEnds []ClientEnd, pos int, memAddr string) *Backend{
    fmt.Println("This is StartBackend listener...")
    
    pbServer = Make(args.ClientEnds, args.Pos, 0, args.MemAddr)   
   
    return nil
}

type Comm struct {
    BEMeth string
    Args interface{}
    Reply interface{}
}

var resp interface{}

func (l *Listener) ForwardCommand(command *Comm, response *interface{}) error {
    fmt.Println("This is ForwardCommand listener...")
    pbServer.Start(command)
    
    encBuf := new(bytes.Buffer)
    encErr := gob.NewEncoder(encBuf).Encode(resp)
    if encErr != nil {
        log.Fatal(encErr)
    }
    value := encBuf.Bytes()
    
    *response = value
    return nil
}

func (l *Listener) UnpackagePBServerCall(args *CallPBServerArgs, replys *CallPBServerReplys) error {
    fmt.Println("This is UnpackagePBServerCall listener...")
    svcMeth := args.SvcMeth
    //svcArgs := args.SvcArgs
    //svcReply := args.SvcReplys
    gob.Register(Comm{})
    if svcMeth == "PBServer.Prepare" {
        decBuf := bytes.NewBuffer((args.SvcArgs).([]byte))
        prepareArgs := PrepareArgs{}
        decErr := gob.NewDecoder(decBuf).Decode(&prepareArgs)
        if decErr != nil {
            log.Fatal(decErr)
        }

        decBuf = bytes.NewBuffer((args.SvcReplys).([]byte))
        prepareReply := PrepareReply{}
        decErr = gob.NewDecoder(decBuf).Decode(&prepareReply)
        if decErr != nil {
            log.Fatal(decErr)
        }

        pbServer.Prepare(&prepareArgs, &prepareReply, l)
    } else if svcMeth == "PBServer.Recovery" {
        decBuf := bytes.NewBuffer((args.SvcArgs).([]byte))
        recoveryArgs := RecoveryArgs{}
        decErr := gob.NewDecoder(decBuf).Decode(&recoveryArgs)
        if decErr != nil {
            log.Fatal(decErr)
        }

        decBuf = bytes.NewBuffer((args.SvcReplys).([]byte))
        recoveryReply := RecoveryReply{}
        decErr = gob.NewDecoder(decBuf).Decode(&recoveryReply)
        if decErr != nil {
            log.Fatal(decErr)
        }

        pbServer.Recovery(&recoveryArgs, &recoveryReply)
    } else if svcMeth == "PBServer.ViewChange" {
        decBuf := bytes.NewBuffer((args.SvcArgs).([]byte))
        viewChangeArgs := ViewChangeArgs{}
        decErr := gob.NewDecoder(decBuf).Decode(&viewChangeArgs)
        if decErr != nil {
            log.Fatal(decErr)
        }

        decBuf = bytes.NewBuffer((args.SvcReplys).([]byte))
        viewChangeReply := ViewChangeReply{}
        decErr = gob.NewDecoder(decBuf).Decode(&viewChangeReply)
        if decErr != nil {
            log.Fatal(decErr)
        }

        pbServer.ViewChange(&viewChangeArgs, &viewChangeReply)
    } else if svcMeth == "PBServer.StartView" {
        decBuf := bytes.NewBuffer((args.SvcArgs).([]byte))
        startViewArgs := StartViewArgs{}
        decErr := gob.NewDecoder(decBuf).Decode(&startViewArgs)
        if decErr != nil {
            log.Fatal(decErr)
        }

        decBuf = bytes.NewBuffer((args.SvcReplys).([]byte))
        startViewReply := StartViewReply{}
        decErr = gob.NewDecoder(decBuf).Decode(&startViewReply)
        if decErr != nil {
            log.Fatal(decErr)
        }

        pbServer.StartView(&startViewArgs, &startViewReply, l)
    } else {
        return errors.New("Called Function not exist")
    }
    
    *replys = true

    return nil
}

var memcacheAddrPrefix = "127.0.0.1:11211"

func main() {
    addy, err := net.ResolveTCPAddr("tcp", "0.0.0.0:42586")
    if err != nil {
        log.Fatal(err)
    }
    inbound, err := net.ListenTCP("tcp", addy)
    if err != nil {
        log.Fatal(err)
    }
    listener := new(Listener)
    rpc.Register(listener)
    rpc.Accept(inbound)

}