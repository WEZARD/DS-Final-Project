package main

import (
    "fmt"
    "log"
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
    Following map[string]struct{}
    Follower map[string]struct{}
    Messages map[time.Time]string
}

type Message struct {
    timestamp time.Time
    text string
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
        username := request.FormValue("username")
        password := request.FormValue("password")
        following := make(map[string]struct{})
        follower := make(map[string]struct{})
        user := User {
            Username: username,
            Password: password,
            Following: following,
            Follower: follower,
        }
        _, memErr := mc.Get(username)
        
        if memErr == nil {
            fmt.Fprintf(response, "<script>alert('this username has been registered')</script>")
            t, _ := template.ParseFiles("register.gtpl")
            t.Execute(response, nil)
        } else {
            fmt.Println(memErr)
            //TODO  Encode struct to byte array
            encBuf := new(bytes.Buffer)
            encErr := gob.NewEncoder(encBuf).Encode(user)
            if encErr != nil {
                log.Fatal(encErr)
            }
            value := encBuf.Bytes()
            mc.Set(&memcache.Item{Key: username, Value: []byte(value)})
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
        //TODO  Decode byte array to struct
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

func getUserName(request *http.Request) (userName string) {
    if cookie, err := request.Cookie("session"); err == nil {
        cookieValue := make(map[string]string)
        if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
            userName = cookieValue["name"]
        }
    }
    return userName
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
        t.Execute(response, nil)
        username := getUserName(request)
        fmt.Fprintf(response, "This is a test for the broadcasting system %s!", username)

    }
    if request.Method == "POST" {
    }
}

// test
func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "This is a test for the broadcasting system %s!", r.URL.Path[1:])
}

var router = mux.NewRouter()

func main() {
    mc := memcache.New("127.0.0.1:11211")
    mc.Set(&memcache.Item{Key: "key_one", Value: []byte("michael")})
    mc.Set(&memcache.Item{Key: "key_two", Value: []byte("programming")})
    val, err := mc.Get("key_one")

    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Printf("-- %s", val.Value)
    http.Handle("/", router)
    router.HandleFunc("/", handler)
    router.HandleFunc("/register", registerHandler)
    router.HandleFunc("/login", loginHandler)
    router.HandleFunc("/home", homeHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}