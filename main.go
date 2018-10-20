package main

import (
	"github.com/gomodule/redigo/redis"
	"html/template"
	"log"
	"net/http"
	"./lib"
	"strconv"
)

var tpl *template.Template

const PrivateKey = "752EF0D8FB4958670DBA40AB1F3C1D0F8FB4958670DBA40AB1F3752EF0DC1D0F"

type sending struct {
	senderWalletID   string
	receiverWalletID string
	amount           string
	counter          string
}

type receiving struct {
	token     string
	receiveID string
	amount    string
}

type syncSend struct {
	receiverID string
	senderID   string
	amount     string
	counter    string
	token      string
}

/*
REDIS keys

amount = {amount} := The amount of the sender
sender_id = 1105 :=the senders id
reciever_{id} : {id} := the receivers id if exists
receiver_counter_{id} := {counter} := the counter of the receiver id

*/
func init() {
	//todo: insert redis amont and sender id
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	http.HandleFunc("/", receiveMoney)
	http.HandleFunc("/send", sendMoney)
	http.HandleFunc("/sendSync", sendSync)

	http.ListenAndServe(":8080", nil)
}

func sendSync(w http.ResponseWriter, req *http.Request) {

	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatal(err)
	}

	if req.Method == http.MethodPost {
		req.ParseForm()
		//Get value from the form
		receiver_id := req.PostFormValue("receiver_id")

		if _,err:= strconv.Atoi(receiver_id);err!=nil {
			tpl.ExecuteTemplate(w,"sync.gohtml",struct{ Success bool }{true})
			return
		}else if len(receiver_id)!=4 {
			tpl.ExecuteTemplate(w,"sync.gohtml",struct{ Success bool }{true})
			return
		}
		//Get the sender id from redis
		sender_id, err := redis.String(conn.Do("GET", "sender_id"))
		if err != nil {
			log.Fatal(err)
		}

		//Initialise with a struct
		sync := new(syncSend)
		sync.receiverID = receiver_id
		sync.senderID = sender_id
		sync.amount = "0000"
		sync.counter = "0000"

		//Concat all the fields for encryption
		plainToken := []byte(sync.senderID + sync.receiverID + sync.amount + sync.counter)
		key := lib.DecodeString(PrivateKey)
		encryptedToken:=lib.Encrypt(key, plainToken)

		err = tpl.ExecuteTemplate(w, "sync.gohtml", encryptedToken)
		if err != nil {
			log.Fatal(err)
		}

	}
	if req.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "sync.gohtml", nil)

	}

}
func receiveMoney(w http.ResponseWriter, req *http.Request) {

	err := tpl.ExecuteTemplate(w, "receive.gohtml", nil)

	if err != nil {
		log.Fatal(err)
	}
}
func sendMoney(w http.ResponseWriter, req *http.Request) {
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	//if _, err = conn.Do("SET", "amount","1105"); err != nil {
	//	log.Fatal(err)
	//}

	// get many keys in a single MGET, ask redigo for []string result
	if req.Method == http.MethodPost {

		receiversID, err := redis.Int(conn.Do("EXISTS", "1106"))
		if err != nil {
			log.Fatal(err)
		} else if receiversID == 0 {
			log.Fatal("Receiver ID not found")
		}

		amount, err := redis.Int(conn.Do("GET", "amount"))
		if err != nil {
			log.Fatal(err)
		} else if amount == 0 {
			log.Fatal("Amount is nil")
		}
		req.ParseForm()
	}
	err = tpl.ExecuteTemplate(w, "send.gohtml", nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Fatalln(err)
	}
}
