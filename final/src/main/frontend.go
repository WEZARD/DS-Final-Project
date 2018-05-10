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
    "reflect"
    "encoding/gob"
    "bytes"
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

        var registerRep RegisterReply
        fmt.Println("Prepare to packageCommand register...")
        
        answer := packageCommand("Listener.UserRegister", &registerArgs, &registerRep)
        
        fmt.Println("This is the UserRegister answer...")
        
        decBuf := bytes.NewBuffer(answer.([]byte))
        decErr := gob.NewDecoder(decBuf).Decode(&registerRep)
        if decErr != nil {
            log.Fatal(decErr)
        }

        /*if dialErr == false {
            fmt.Fprintf(response, "<script>alert('register dial client error')</script>")
            t, _ := template.ParseFiles("register.gtpl")
            t.Execute(response, nil)
        }*/

        /*client, err := rpc.Dial("tcp", rpcAdrr);
        if err != nil {
            log.Fatal(err)
        }
        var registerRep RegisterReply
        err = client.Call("Listener.UserRegister", &registerArgs, &registerRep)
        if err != nil {
            fmt.Print("client error:", err)
        }*/

        if registerRep == false {
            fmt.Fprintf(response, "<script>alert('this username has been registered')</script>")
            t, _ := template.ParseFiles("register.gtpl")
            t.Execute(response, nil)
        }
        if registerRep == true {    
            setSession(username, response)
            fmt.Println("get next...")
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

		var loginRep LoginReply 

        fmt.Println("Prepare to packageCommand login...")
        
        answer := packageCommand("Listener.UserLogin", &loginArgs, &loginRep)
        
        fmt.Println("This is the UserLogin answer...")
        
        decBuf := bytes.NewBuffer(answer.([]byte))
        decErr := gob.NewDecoder(decBuf).Decode(&loginRep)
        if decErr != nil {
            log.Fatal(decErr)
        }
        
        //client, err := rpc.Dial("tcp", rpcAdrr);
        //if err != nil {
        //    log.Fatal(err)
        //}
        //var loginRep LoginReply
        //err = client.Call("Listener.UserLogin", &loginArgs, &loginRep)
        //if err != nil {
        //    fmt.Print("client error:", err)
        //}
		//
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

            var infoRep User
            
            fmt.Println("Prepare to packageCommand getUserInfo...")
        
            answer := packageCommand("Listener.UserInfo", &infoArgs, &infoRep)

            fmt.Println("This is the UserInfo answer...")
        
            decBuf := bytes.NewBuffer(answer.([]byte))
            decErr := gob.NewDecoder(decBuf).Decode(&infoRep)
            if decErr != nil {
                log.Fatal(decErr)
            }

            /*
            client, err := rpc.Dial("tcp", rpcAdrr);
            if err != nil {
                log.Fatal(err)
            }
            err = client.Call("Listener.UserInfo", &infoArgs, &infoRep)
            if err != nil {
                fmt.Print("client error:", err)
            }*/

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
        
        var messageBox MessageBox

        fmt.Println("Prepare to packageCommand home...")
        
        answer := packageCommand("Listener.UserHome", &userInfo, &messageBox)
        
        fmt.Println("This is the UserHome answer...")
        
        decBuf := bytes.NewBuffer(answer.([]byte))
        decErr := gob.NewDecoder(decBuf).Decode(&messageBox)
        if decErr != nil {
            log.Fatal(decErr)
        }
        
        /*client, err := rpc.Dial("tcp", rpcAdrr);
        if err != nil {
            log.Fatal(err)
        }*/
        
        /*err = client.Call("Listener.UserHome", &userInfo, &messageBox)
        if err != nil {
            fmt.Print("client error:", err)
        }*/
        
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
        
        /*client, err := rpc.Dial("tcp", rpcAdrr);
        if err != nil {
            log.Fatal(err)
        }*/
        
        var messageBox UserSearchReply
        /*err = client.Call("Listener.UserFollow", &followArgs, &messageBox)
        if err != nil {
            fmt.Print("client error:", err)
        }*/
        fmt.Println("Prepare to packageCommand follow...")
        
        answer := packageCommand("Listener.UserFollow", &followArgs, &messageBox)
        
        fmt.Println("This is the UserFollow answer...")
        
        decBuf := bytes.NewBuffer(answer.([]byte))
        decErr := gob.NewDecoder(decBuf).Decode(&messageBox)
        if decErr != nil {
            log.Fatal(decErr)
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

        /*client, err := rpc.Dial("tcp", rpcAdrr);
        if err != nil {
            log.Fatal(err)
        }*/
        var userAddReply UserAddReply
        /*err = client.Call("Listener.UserAdd", &addArgs, &userAddReply)
        if err != nil {
            fmt.Print("client error:", err)
        }*/
        fmt.Println("Prepare to packageCommand add...")
        
        answer := packageCommand("Listener.UserAdd", &addArgs, &userAddReply)
        
        fmt.Println("This is the UserAdd answer...")
        
        decBuf := bytes.NewBuffer(answer.([]byte))
        decErr := gob.NewDecoder(decBuf).Decode(&userAddReply)
        if decErr != nil {
            log.Fatal(decErr)
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

        /*client, err := rpc.Dial("tcp", rpcAdrr);
        if err != nil {
            log.Fatal(err)
        }*/
        var userPostReply UserPostReply
        
        fmt.Println("Prepare to packageCommand post...")
        
        answer := packageCommand("Listener.UserPost", &postArgs, &userPostReply)
        
        fmt.Println("This is the UserPost answer...")
        
        decBuf := bytes.NewBuffer(answer.([]byte))
        decErr := gob.NewDecoder(decBuf).Decode(&userPostReply)
        if decErr != nil {
            log.Fatal(decErr)
        }

        /*err = client.Call("Listener.UserPost", &postArgs, &userPostReply)
        if err != nil {
            fmt.Print("client error:", err)
        }*/

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
		cancelArgs := User{
			Username: getUserName(request),
		}

		var userCancelReply UserCancelReply
		
        fmt.Println("Prepare to packageCommand cancel...")
        
        answer := packageCommand("Listener.UserCancel", &cancelArgs, &userCancelReply)
        
        fmt.Println("This is the UserCancel answer...")
        
        decBuf := bytes.NewBuffer(answer.([]byte))
        decErr := gob.NewDecoder(decBuf).Decode(&userCancelReply)
        if decErr != nil {
            log.Fatal(decErr)
        }

        /*client, err := rpc.Dial("tcp", rpcAdrr);
		if err != nil {
			log.Fatal(err)
		}
		err = client.Call("Listener.UserCancel", &cancelArgs, &userCancelReply)
		if err != nil {
			fmt.Print("client error:", err)
		}*/

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

var serverNum = 3
var rpcAddrs  = []string {"52.87.162.252:42586", "54.146.6.116:42586", "54.88.235.80:42586"}

type Comm struct {
	BEMeth string
	Args interface{}
	Reply interface{}
}

func packageCommand(beMeth string, args interface{}, reply interface{}) interface{} {

	var flag = false
	for j := 0; j < serverNum; j++ {
		client, err := rpc.Dial("tcp", rpcAddrs[j]);
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println(reflect.TypeOf(args))
            encBuf := new(bytes.Buffer)
            encErr := gob.NewEncoder(encBuf).Encode(args)
            if encErr != nil {
                log.Fatal(encErr)
            }
            argsEnc := []byte(encBuf.Bytes())

            fmt.Println(reflect.TypeOf(argsEnc))
            fmt.Println(argsEnc)
            
            encBuf1 := new(bytes.Buffer)
            encErr = gob.NewEncoder(encBuf1).Encode(reply)
            if encErr != nil {
                log.Fatal(encErr)
            }
            replyEnc := encBuf1.Bytes()

            comm := Comm {
				BEMeth: beMeth,
				Args: argsEnc,
				Reply: replyEnc,
			}
            fmt.Println(reflect.TypeOf(comm.Args))
            fmt.Println(comm.Args)
            fmt.Println(comm)
			var respond interface{}
			fmt.Println("Calling Listener.ForwardCommand...")
            err = client.Call("Listener.ForwardCommand", &comm, &respond)            
            return respond
			if err != nil {
				fmt.Printf("client Call ForwardCommand error... %s\n", err)
			} else {
				flag = true
			}
		}
	}
    fmt.Println("pacakgeCommmand is done...")
	return flag
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