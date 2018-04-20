package main

import (
    "fmt"
    "log"
    "sort"
    "net/http"
    //"container/list"
    "html/template"
    "github.com/gorilla/mux"
    "time"
    "github.com/gorilla/securecookie"
    "net/rpc"
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
    Status bool
}

var rpcAddr = "localhost:1234"

type RegisterArgs struct {
    Username string
    Password string
}

type RegisterRepl bool

type LoginRepl struct {
    UserObj User
    PassCheck bool
}

// registerHandler
func registerHandler(response http.ResponseWriter, request *http.Request) {
    fmt.Println("method:", request.Method) // get request method
    if request.Method == "GET" {
        t, _ := template.ParseFiles("register.gtpl")
        t.Execute(response, nil)
    } else {
        request.ParseForm()
        fmt.Println("username:", request.Form["username"])
        username := request.FormValue("username")
        password := request.FormValue("password")

        //------------------------------rpc test------------------------------
        client, err := rpc.DialHTTP("tcp", rpcAddr)
        if err != nil {
            fmt.Print("dialing:",err)
        }
        registerArgs := RegisterArgs{
            Username:   username,
            Password:   password,
        }
        var registerRepl RegisterRepl

        err = client.Call("Client.Register", &registerArgs, &registerRepl)
        if err != nil {
            fmt.Print("client error:", err)
        }
        if registerRepl == false {
            fmt.Fprintf(response, "<script>alert('this username has been registered')</script>")
            t, _ := template.ParseFiles("register.gtpl")
            t.Execute(response, nil)
        } else {
            setSession(username, response)
            http.Redirect(response, request, "/home", http.StatusSeeOther)
            fmt.Printf("user registered: %s", username)
        }
        //--------------------------------------------------------------------
    }
}

// loginHandler
func loginHandler(response http.ResponseWriter, request *http.Request) {
    if request.Method == "GET" {
        t, _ := template.ParseFiles("login.gtpl")
        t.Execute(response, nil)
    } else {
        request.ParseForm()
        fmt.Println("username:", request.Form["username"])
        username := request.FormValue("username")
        password := request.FormValue("password")

        loginArgs := RegisterArgs{
            Username:   username,
            Password:   password,
        }
        var loginRepl LoginRepl

        client, err := rpc.DialHTTP("tcp", rpcAddr)
        if err != nil {
            fmt.Print("dialing:",err)
        }

        err = client.Call("Client.Login", &loginArgs, &loginRepl)
        if err != nil {
            fmt.Print("client error:", err)
        }

        if loginRepl.PassCheck == false {
            fmt.Fprintf(response, "<script>alert('wrong username or password')</script>")
            t, _ := template.ParseFiles("login.gtpl")
            t.Execute(response, nil)
        } else {
            setSession(username, response)
            fmt.Println(username, getUserName(request))
            http.Redirect(response, request, "/home", http.StatusSeeOther)
        }
    }
}

// sessionHandler
var cookieHandler = securecookie.New(
    securecookie.GenerateRandomKey(64),
    securecookie.GenerateRandomKey(32))

func setSession(userName string, response http.ResponseWriter) {
    fmt.Println("setSession starts")
    value := map[string]string{
        "name": userName,
    }
    if encoded, err := cookieHandler.Encode("session", value); err == nil {
        cookie := &http.Cookie{
            Name:  "session",
            Value: encoded,
            Path:  "/",
        }
        http.SetCookie(response, cookie)
    }
}

func getUserName(request *http.Request) (username string) {
    if cookie, err := request.Cookie("session"); err == nil {
        cookieValue := make(map[string]string)
        if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
            username = cookieValue["name"]
        }
    }
    return username
}

func getUserInfo(request *http.Request) *User {
    if cookie, err := request.Cookie("session"); err == nil {
        cookieValue := make(map[string]string)
        if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
            username := cookieValue["name"]
            userOut := User{}
            client, err := rpc.DialHTTP("tcp", rpcAddr)
            if err != nil {
                fmt.Print("dialing:",err)
            }

            err = client.Call("Client.GetUserByName", username, &userOut)
            if err != nil {
                fmt.Print("client error:", err)
            }
            return &userOut
        }
    }
    return nil   
}

func clearSession(response http.ResponseWriter) {
    fmt.Println("clear session...")
    cookie := &http.Cookie{
        Name:   "session",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
    }
    http.SetCookie(response, cookie)
}

func setSearch(userName string, response http.ResponseWriter) {
    fmt.Println("search result stored...")
    value := map[string]string{
        "name": userName,
    }
    if encoded, err := cookieHandler.Encode("search", value); err == nil {
        cookie := &http.Cookie{
            Name:  "search",
            Value: encoded,
            Path:  "/",
        }
        http.SetCookie(response, cookie)
    }
}

func getSearch(request *http.Request) (username string) {
    if cookie, err := request.Cookie("search"); err == nil {
        cookieValue := make(map[string]string)
        if err = cookieHandler.Decode("search", cookie.Value, &cookieValue); err == nil {
            username = cookieValue["name"]
        }
    }
    return username
}

func clearSearch(response http.ResponseWriter) {
    fmt.Println("clear search...")
    cookie := &http.Cookie{
        Name:   "search",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
    }
    http.SetCookie(response, cookie)
}

// homeHandler
func homeHandler(response http.ResponseWriter, request *http.Request) {
    if request.Method == "GET" {
        t, _ := template.ParseFiles("home.gtpl")

        userInfo := getUserInfo(request)
        fmt.Printf("This is a test for %s\n", userInfo.Username)
        
        messageBox := MessageBox{}
        messages := []Message{}
        singleMessage := Message{}
        
        for k := range userInfo.Following {
            fmt.Printf("%s\n", k)

            //Get following user by rpc function GetUserByName
            userOut := User{}
            client, err := rpc.DialHTTP("tcp", rpcAddr)
            if err != nil {
                fmt.Print("dialing:",err)
            }

            err = client.Call("Client.GetUserByName", k, &userOut)
            if err != nil {
                fmt.Print("client error:", err)
            }

            for k1, v1 := range userOut.Messages {
                singleMessage.Username = k
                singleMessage.Timestamp = k1
                singleMessage.Text = v1
                messages = append(messages, singleMessage)
            }
        }

        sort.Slice(messages, func(i, j int) bool { 
            return messages[i].Timestamp.After(messages[j].Timestamp)
        })

        for i := 0; i < len(messages); i++ {
            messages[i].DisplayTime = messages[i].Timestamp.Format("Mon Jan _2 15:04:05 2006")
        }

        messageBox.Username = userInfo.Username
        messageBox.Following = userInfo.Following
        messageBox.Follower = userInfo.Follower
        messageBox.Messages = messages

        t.Execute(response, messageBox)
    }

    if request.Method == "POST" {
        t, _ := template.ParseFiles("home.gtpl")
        t.Execute(response, nil)
    }
}

// followHandler
func followHandler(response http.ResponseWriter, request *http.Request) {
    if request.Method == "GET" {
        t, _ := template.ParseFiles("follow.gtpl")
        t.Execute(response, nil)
    }
    if request.Method == "POST" {
        followname := request.FormValue("username")

        userOut := User{}
        client, err := rpc.DialHTTP("tcp", rpcAddr)
        if err != nil {
            fmt.Print("dialing:",err)
        }

        err = client.Call("Client.GetUserByName", followname, &userOut)
        if err != nil {
            fmt.Print("client error:", err)
        }
        
        if userOut.Username != "" {
            messageBox := MessageBox{}
            messages := []Message{}
            singleMessage := Message{}
            for k1, v1 := range userOut.Messages {
                singleMessage.Username = followname
                singleMessage.Timestamp = k1
                singleMessage.Text = v1
                messages = append(messages, singleMessage)
            }

            for i := 0; i < len(messages); i++ {
            messages[i].DisplayTime = messages[i].Timestamp.Format("Mon Jan _2 15:04:05 2006")
            }

            userInfo := getUserInfo(request)
            status := true
            if _ , ok := userInfo.Following[followname]; ok {
                status = false
            }
            
            messageBox.Username = followname
            messageBox.Messages = messages
            messageBox.Status = status

            clearSearch(response)
            setSearch(followname, response)

            t, _ := template.ParseFiles("follow.gtpl")
            t.Execute(response, messageBox)
        } else {
            fmt.Fprintf(response, "<script>alert('No user found!')</script>")
            t, _ := template.ParseFiles("follow.gtpl")
            t.Execute(response, nil)
        }
    }
}

// addHandler
func addHandler(response http.ResponseWriter, request *http.Request) {
    if request.Method == "POST" {
        followname := getSearch(request)
        fmt.Println(followname)
        username := getUserName(request)
        fmt.Println(username)

        userSrc := User{}
        userDest := User{}
        client, err := rpc.DialHTTP("tcp", rpcAddr)
        if err != nil {
            fmt.Print("dialing:",err)
        }

        err = client.Call("Client.GetUserByName", username, &userSrc)
        if err != nil {
            fmt.Print("client error:", err)
        }

        err = client.Call("Client.GetUserByName", followname, &userDest)
        if err != nil {
            fmt.Print("client error:", err)
        }

        userSrc.Following[followname] = true
        userDest.Follower[username] = true

        var deleteRepl bool
        err = client.Call("Client.DeleteUserByName", username, &deleteRepl)
        if err != nil {
            fmt.Print("client error:", err)
        }

        var setRepl bool
        err = client.Call("Client.SetUserByName", userSrc, &setRepl)
        if err != nil {
            fmt.Print("client error:", err)
        }

        err = client.Call("Client.DeleteUserByName", followname, &deleteRepl)
        if err != nil {
            fmt.Print("client error:", err)
        }

        err = client.Call("Client.SetUserByName", userDest, &setRepl)
        if err != nil {
            fmt.Print("client error:", err)
        }

        http.Redirect(response, request, "/home", http.StatusSeeOther)
    }
}

// postlHandler
func postHandler(response http.ResponseWriter, request *http.Request) {
    if request.Method == "POST" {
        now := time.Now()
        username := getUserName(request)
        userInfo := getUserInfo(request)
        userInfo.Messages[now] = request.FormValue("postcontent")

        client, err := rpc.DialHTTP("tcp", rpcAddr)
        if err != nil {
            fmt.Print("dialing:",err)
        }
        var deleteRepl bool
        err = client.Call("Client.DeleteUserByName", username, &deleteRepl)
        if err != nil {
            fmt.Print("client error:", err)
        }

        var setRepl bool
        err = client.Call("Client.SetUserByName", userInfo, &setRepl)
        if err != nil {
            fmt.Print("client error:", err)
        }

        http.Redirect(response, request, "/home", http.StatusSeeOther)
    }
}

// cancelHandler
func cancelHandler(response http.ResponseWriter, request *http.Request) {
    if request.Method == "GET" {
        t, _ := template.ParseFiles("cancel.gtpl")
        t.Execute(response, nil)
    } 
    if request.Method == "POST" {
        username := getUserName(request)
        var repl bool

        client, err := rpc.DialHTTP("tcp", rpcAddr)
        if err != nil {
            fmt.Print("dialing:",err)
        }

        err = client.Call("Client.DeleteUserByName", username, &repl)
        if err != nil {
            fmt.Print("client error:", err)
        }

        if repl == false{
            fmt.Fprintf(response, "<script>alert('Failed to cancel account')</script>")
            t, _ := template.ParseFiles("cancel.gtpl")
            t.Execute(response, nil)
        } else {
            clearSession(response)
            http.Redirect(response, request, "/login", http.StatusSeeOther)
        }
     }
}

// test
func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "This is a test for the broadcasting system %s!", r.URL.Path[1:])
}

var router = mux.NewRouter()

func main() {
    //mc := memcache.New("127.0.0.1:11211")
    fmt.Println("Welcome to message web application!")

    http.Handle("/", router)
    router.HandleFunc("/", handler)
    router.HandleFunc("/register", registerHandler)
    router.HandleFunc("/login", loginHandler)
    router.HandleFunc("/home", homeHandler)
    router.HandleFunc("/follow", followHandler)
    router.HandleFunc("/add", addHandler)
    router.HandleFunc("/post", postHandler)
    router.HandleFunc("/cancel", cancelHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}