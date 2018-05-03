package main

import (
    "fmt"
    "log"
    "net/http"
    "html/template"
    "github.com/gorilla/mux"
    "time"
    "github.com/gorilla/securecookie"
    "net/rpc"
)

var rpcAdrr = "107.180.40.18:42586"

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
type UserCancelReply bool

// registerHandler
func registerHandler(response http.ResponseWriter, request *http.Request) {
    fmt.Println("method:", request.Method) // get request method
    if request.Method == "GET" {
        t, _ := template.ParseFiles("register.gtpl")
        t.Execute(response, nil)
    } else {
        request.ParseForm()
        username := request.FormValue("username")
        password := request.FormValue("password")
        registerArgs := User {
            Username: username,
            Password: password,
        }

        client, err := rpc.Dial("tcp", rpcAdrr);
        if err != nil {
            log.Fatal(err)
        }
        var registerRep RegisterReply
        err = client.Call("Listener.UserRegister", &registerArgs, &registerRep)
        if err != nil {
            fmt.Print("client error:", err)
        }

        if registerRep == false {
            fmt.Fprintf(response, "<script>alert('this username has been registered')</script>")
            t, _ := template.ParseFiles("register.gtpl")
            t.Execute(response, nil)
        }
        if registerRep == true {    
            setSession(username, response)
            http.Redirect(response, request, "/home", http.StatusSeeOther)
        }            
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
        loginArgs := User {
            Username: username,
            Password: password,
        }

        client, err := rpc.Dial("tcp", rpcAdrr);
        if err != nil {
            log.Fatal(err)
        }
        var loginRep LoginReply
        err = client.Call("Listener.UserLogin", &loginArgs, &loginRep) 
        if err != nil {
            fmt.Print("client error:", err)
        }

        if loginRep == false {
            fmt.Fprintf(response, "<script>alert('wrong username or password')</script>")
            t, _ := template.ParseFiles("login.gtpl")
            t.Execute(response, nil)
        }
        if loginRep == true {
            setSession(username, response)
            http.Redirect(response, request, "/home", http.StatusSeeOther)
        }
    }
}

// sessionHandler
var cookieHandler = securecookie.New(
    securecookie.GenerateRandomKey(64),
    securecookie.GenerateRandomKey(32))

// User info session
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
            infoArgs := User {
                Username: username,
            }

            client, err := rpc.Dial("tcp", rpcAdrr);
            if err != nil {
                log.Fatal(err)
            }
            var infoRep User
            err = client.Call("Listener.UserInfo", &infoArgs, &infoRep)
            if err != nil {
                fmt.Print("client error:", err)
            }
            return &infoRep
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

// homeHandler
func homeHandler(response http.ResponseWriter, request *http.Request) {
    if request.Method == "GET" {
        t, _ := template.ParseFiles("home.gtpl")
        userInfo := getUserInfo(request)
        client, err := rpc.Dial("tcp", rpcAdrr);
        if err != nil {
            log.Fatal(err)
        }
        var messageBox MessageBox
        err = client.Call("Listener.UserHome", &userInfo, &messageBox)
        if err != nil {
            fmt.Print("client error:", err)
        }
        t.Execute(response, messageBox)
    }
    
    if request.Method == "POST" {
        t, _ := template.ParseFiles("home.gtpl")
        t.Execute(response, nil)
    }
}

// User search session
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

// followHandler
func followHandler(response http.ResponseWriter, request *http.Request) {
    if request.Method == "GET" {
        t, _ := template.ParseFiles("follow.gtpl")
        t.Execute(response, nil)
    }
    
    if request.Method == "POST" {
        followname := request.FormValue("username")
        userinfo := getUserInfo(request)
        followArgs := UserSearch {
            Followname: followname,
            Userinfo: userinfo,
        }
        
        client, err := rpc.Dial("tcp", rpcAdrr);
        if err != nil {
            log.Fatal(err)
        }
        var messageBox UserSearchReply
        err = client.Call("Listener.UserFollow", &followArgs, &messageBox)
        if err != nil {
            fmt.Print("client error:", err)
        }

        if messageBox.Reply == true {
            clearSearch(response)
            setSearch(followname, response)

            t, _ := template.ParseFiles("follow.gtpl")
            t.Execute(response, messageBox.Messages)
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
        username := getUserName(request) 
        addArgs := UserAdd {
            Followname: followname,
            Username: username,
        }

        client, err := rpc.Dial("tcp", rpcAdrr);
        if err != nil {
            log.Fatal(err)
        }
        var userAddReply UserAddReply
        err = client.Call("Listener.UserAdd", &addArgs, &userAddReply)
        if err != nil {
            fmt.Print("client error:", err)
        }

        if userAddReply == true {
            http.Redirect(response, request, "/home", http.StatusSeeOther)
        }
    }
}

// postlHandler
func postHandler(response http.ResponseWriter, request *http.Request) {
    if request.Method == "POST" {
        userinfo := getUserInfo(request)
        content := request.FormValue("postcontent")
        postArgs := UserPost {
            Userinfo: userinfo,
            Content:  content,
        }

        client, err := rpc.Dial("tcp", rpcAdrr);
        if err != nil {
            log.Fatal(err)
        }
        var userPostReply UserPostReply
        err = client.Call("Listener.UserPost", &postArgs, &userPostReply)
        if err != nil {
            fmt.Print("client error:", err)
        }

        if userPostReply == true {
            http.Redirect(response, request, "/home", http.StatusSeeOther)
        }
    }
}

// cancelHandler
func cancelHandler(response http.ResponseWriter, request *http.Request) {
    if request.Method == "GET" {
        t, _ := template.ParseFiles("cancel.gtpl")
        t.Execute(response, nil)
    } 
    
    if request.Method == "POST" {
        cancelArgs := User {
            Username: getUserName(request),
        }

        var userCancelReply UserCancelReply
        client, err := rpc.Dial("tcp", rpcAdrr);
        if err != nil {
            log.Fatal(err)
        }
        err = client.Call("Listener.UserCancel", &cancelArgs, &userCancelReply)
        if err != nil {
            fmt.Print("client error:", err)
        }
        
        if userCancelReply == true {
            clearSession(response)
            http.Redirect(response, request, "/login", http.StatusSeeOther)
        } else {
            fmt.Fprintf(response, "<script>alert('Failed to cancel account')</script>")
            t, _ := template.ParseFiles("cancel.gtpl")
            t.Execute(response, nil)
        }
     }
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "This is a test for the broadcasting system %s!", r.URL.Path[1:])
}

var router = mux.NewRouter()

func main() {
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