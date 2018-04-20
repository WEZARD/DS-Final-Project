package main

import (
	"time"
	"fmt"
	"net/http"
	"log"
	"net/rpc"
	"github.com/bradfitz/gomemcache/memcache"
	"bytes"
	"encoding/gob"
)

type User struct { // capitalize the first letter of each field name as "exported" for gob
	Username string
	Password string
	Following map[string]bool
	Follower map[string]bool
	Messages map[time.Time]string
}

type RegisterArgs struct {
	Username string
	Password string
}

type RegisterRepl bool

type LoginRepl struct {
	UserObj User
	PassCheck bool
}


type Client struct {
	Address string
}

func (c *Client) GetUserByName(username string, reply *User) error {
	mc := memcache.New("127.0.0.1:11211")
	fmt.Print("\nCalled get user by name: ", username)
	user, memErr := mc.Get(username)
	if memErr != nil{
		(*reply).Username = ""
	}else{
		decBuf := bytes.NewBuffer(user.Value)
		userOut := User{}
		decErr := gob.NewDecoder(decBuf).Decode(&userOut)
		if decErr != nil {
			log.Fatal(decErr)
		}
		*reply = userOut
	}
	return nil
}

func (c *Client) SetUserByName(newUser *User, reply *bool) error {
	mc := memcache.New("127.0.0.1:11211")
	fmt.Print("\nCalled set user: ", newUser.Username)

	encBuf := new(bytes.Buffer)
	encErr := gob.NewEncoder(encBuf).Encode(newUser)
	if encErr != nil {
		log.Fatal(encErr)
	}
	value := encBuf.Bytes()
	err := mc.Set(&memcache.Item{Key: newUser.Username, Value: []byte(value)})

	if err != nil{
		*reply= false
	}else{
		*reply = true
	}
	return nil
}

func (c *Client) Register(args *RegisterArgs, repl *RegisterRepl) error{
	mc := memcache.New("127.0.0.1:11211")
	fmt.Print("\nCalled register ", args.Username)
	_, memErr := mc.Get(args.Username)
	if memErr == nil {
		*repl = false
	} else {
		fmt.Println(memErr)
		// Encode struct to byte array
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
		}

		fmt.Println(memErr)
		// Encode struct to byte array
		encBuf := new(bytes.Buffer)
		encErr := gob.NewEncoder(encBuf).Encode(user)
		if encErr != nil {
			log.Fatal(encErr)
		}
		value := encBuf.Bytes()
		mc.Set(&memcache.Item{Key: username, Value: []byte(value)}) // store in memcache
		*repl = true
	}
	return nil
}

func (c *Client) Login(args *RegisterArgs, repl *LoginRepl) error{
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
			(*repl).PassCheck = true
			(*repl).UserObj = userOut
			fmt.Print("login ", (*repl).UserObj.Username)
		} else {
			(*repl).PassCheck = false
		}
	} else {
		(*repl).PassCheck = false
	}
	return nil
}

func (c *Client) DeleteUserByName(username string, repl *bool) error{
	mc := memcache.New("127.0.0.1:11211")
	delErr := mc.Delete(username)
	if delErr != nil {
		*repl = false
	} else {
		*repl = true
	}
	return nil
}

func main() {
	fmt.Println("Backend server started!")

	client := new(Client)
	rpc.Register(client)
	rpc.HandleHTTP()
	log.Fatal(http.ListenAndServe(":1234", nil))
}