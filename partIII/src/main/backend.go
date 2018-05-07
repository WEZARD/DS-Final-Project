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
    mc := memcache.New("127.0.0.1:11211") // brew services restart memcached
        
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
        // Encode struct to byte array
        encBuf := new(bytes.Buffer)
        encErr := gob.NewEncoder(encBuf).Encode(user)
        if encErr != nil {
            log.Fatal(encErr)
        }
        value := encBuf.Bytes()
        mc.Set(&memcache.Item{Key: username, Value: []byte(value)}) // store in memcache  
        

        description := "Register"
        key := username
        val := []byte(value)

        command := Command {
            Description: description,
            Key: key,
            Value: val,
        }

        servers := 3                        //3 servers
        primaryID := GetPrimary(0, servers) //primary ID is determined by view=0
        cfg := make_config(servers, false)
        defer cfg.cleanup()
        cfg.replicateOne(primaryID, command, servers) // replicate command 1000+index, expected successful replication to all servers
        fmt.Printf(" ... Passed\n")
    

        *rep = true

    }

    return nil
}

func (l *Listener) UserLogin(args *User, rep *LoginReply) error {
    mc := memcache.New("127.0.0.1:11211")
    
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
    username := args.Username
    mc := memcache.New("127.0.0.1:11211")
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
    mc := memcache.New("127.0.0.1:11211")
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
    followname := followArgs.Followname
    mc := memcache.New("127.0.0.1:11211")
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
    mc := memcache.New("127.0.0.1:11211")

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
    now := time.Now()
    userInfo := postArgs.Userinfo
    content := postArgs.Content
    username := userInfo.Username
    userInfo.Messages[now] = content
    mc := memcache.New("127.0.0.1:11211") 
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

func (l *Listener) UserCancel(cancelArgs *User, userPostRep *UserPostReply) error {
    mc := memcache.New("127.0.0.1:11211")
    username := cancelArgs.Username
    delErr := mc.Delete(username)
    if delErr != nil {
        *userPostRep = false
    } else {
        *userPostRep = true
    }
    return nil
}

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