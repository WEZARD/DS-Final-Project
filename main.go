package main

import (
    "fmt"
    "log"
    "sort"
    "net/http"
    //"container/list"
    "html/template"
    "github.com/bradfitz/gomemcache/memcache"
    "github.com/gorilla/mux"
    "time"
    "encoding/gob"
    "bytes"
    "github.com/gorilla/securecookie"
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
        mc := memcache.New("127.0.0.1:11211") // brew services restart memcached
        
        // ----------------------------------------Test Case---------------------------------------------
        testTime1 := time.Date(2018, 4, 4, 20, 00, 00, 651387237, time.UTC)
        testTime2 := time.Date(2018, 4, 5, 20, 00, 00, 651387237, time.UTC)
        messages := map[time.Time]string {testTime1: "Hello World!", testTime2: "Welcome to our web service!"}

        testUser1 := User {
            Username: "Thierry",
            Messages: messages,
        }

        testUser2 := User {
            Username: "Ramsey",
            Messages: messages,
        }

        _, memErr := mc.Get("Thierry")
        
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
            mc.Set(&memcache.Item{Key: "Thierry", Value: []byte(value)}) // store in memcache
        }

        _, memErr = mc.Get("Ramsey")
        
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
            mc.Set(&memcache.Item{Key: "Ramsey", Value: []byte(value)}) // store in memcache
        }

        // ----------------------------------------------------------------------------------------------
        
        username := request.FormValue("username")
        password := request.FormValue("password")
        following := map[string]bool {testUser1.Username: true, testUser2.Username: true}
        follower := map[string]bool {testUser1.Username: true, testUser2.Username: true}
        //following := make(map[string]bool)
        //follower := make(map[string]bool)
        user := User {
            Username: username,
            Password: password,
            Following: following,
            Follower: follower,
        }
        _, memErr = mc.Get(username)
        
        if memErr == nil {
            fmt.Fprintf(response, "<script>alert('this username has been registered')</script>")
            t, _ := template.ParseFiles("register.gtpl")
            t.Execute(response, nil)
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
        mc := memcache.New("127.0.0.1:11211")
        username := request.FormValue("username")
        password := request.FormValue("password")

        val, memErr := mc.Get(username)
        // Decode byte array to struct
        decBuf := bytes.NewBuffer(val.Value)
        userOut := User{}
        decErr := gob.NewDecoder(decBuf).Decode(&userOut)
        if decErr != nil {
            log.Fatal(decErr)
        }

        if memErr == nil {
            if password == userOut.Password {
                setSession(username, response)
                fmt.Println(username, getUserName(request))
                http.Redirect(response, request, "/home", http.StatusSeeOther)
            } else {
                fmt.Fprintf(response, "<script>alert('wrong username or password')</script>")
                t, _ := template.ParseFiles("login.gtpl")
                t.Execute(response, nil)
            }
        }


    }
}

// sessionHandler
var cookieHandler = securecookie.New(
    securecookie.GenerateRandomKey(64),
    securecookie.GenerateRandomKey(32))

func setSession(userName string, response http.ResponseWriter) {
    fmt.Println("setSession start")
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
            mc := memcache.New("127.0.0.1:11211")
            val, _ := mc.Get(username)
            // Decode byte array to struct
            decBuf := bytes.NewBuffer(val.Value)
            userOut := User{}
            decErr := gob.NewDecoder(decBuf).Decode(&userOut)
            if decErr != nil {
                log.Fatal(decErr)
            }
            return &userOut
        }
    }
    return nil   
}

func clearSession(response http.ResponseWriter) {
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
        
        mc := memcache.New("127.0.0.1:11211")

        userInfo := getUserInfo(request)
        fmt.Printf("This is a test for %s\n", userInfo.Username)
        
        messageBox := MessageBox{}
        messages := []Message{}
        singleMessage := Message{}
        
        for k := range userInfo.Following {
            fmt.Printf("%s\n", k)
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

// addfollowHandler
func addfollowHandler(response http.ResponseWriter, request *http.Request) {
    //fmt.Fprintf(response, "This is a test for the broadcasting system %s!", r.URL.Path[1:])
    if request.Method == "GET" {
        t, _ := template.ParseFiles("addfollow.gtpl")
        t.Execute(response, nil)

    }
}

// cancelHandler
func cancelHandler(response http.ResponseWriter, request *http.Request) {
    //fmt.Fprintf(response, "This is a test for the broadcasting system %s!", r.URL.Path[1:])
    if request.Method == "GET" {
        t, _ := template.ParseFiles("cancel.gtpl")
        t.Execute(response, nil)

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
    router.HandleFunc("/addfollow", addfollowHandler)
    router.HandleFunc("/cancel", cancelHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}