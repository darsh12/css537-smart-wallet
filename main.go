package main

import (
	"./lib"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var tpl *template.Template

const PrivateKey = "752EF0D8FB4958670DBA40AB1F3C1D0F8FB4958670DBA40AB1F3752EF0DC1D0F"

type token struct {
	SenderWalletID   string
	ReceiverWalletID string
	Amount           string
	Counter          string
	EncryptedToken   string
}

type redisValues struct {
	redisReceiverID      string
	redisReceiverCounter string
}

type receiving struct {
	token     string
	receiveID string
	amount    string
}


type message struct {
	Success bool
	Message []string
}

/*
REDIS keys

amount = {Amount} := The Amount of the sender
sender_id = 1105 :=the senders id
(receiver){id} : {counter} := the receivers id  and its counter

*/
func init() {
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatal("Could not connect to redis server")
	}
	//conn.Do("FLUSHALL")
	conn.Do("SET", "sender_id", "1105")
	conn.Do("SET", "amount", "10")
	defer conn.Close()

	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {

	http.HandleFunc("/", receiveMoney)
	http.HandleFunc("/send", sendMoney)
	http.HandleFunc("/generateSync", sendSync)
	http.HandleFunc("/receiveSync", syncWallet)

	http.ListenAndServe(":8080", nil)

}

func syncWallet(w http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {
		req.ParseForm()

		wallet := new(token)
		wallet.EncryptedToken = strings.TrimSpace(req.PostFormValue("sync_token"))

		if len(wallet.EncryptedToken) != 32 {
			messages := []string{"Please check token"}
			tpl.ExecuteTemplate(w, "receive_sync.gohtml", message{false, messages})
			return
		}

		key := lib.DecodeString(PrivateKey)
		decodeToken := lib.DecodeString(wallet.EncryptedToken)
		plainText, err := lib.Decrypt(key, decodeToken)
		if err != nil {
			messages := []string{"Could not create a block"}
			tpl.ExecuteTemplate(w, "receive_sync.gohtml", message{false, messages})
		}
		if len(plainText) != 16 {
			messages := []string{"Incorrect data from token"}
			tpl.ExecuteTemplate(w, "receive_sync.gohtml", message{false, messages})
		}

		wallet.SenderWalletID = string(plainText[0:4])
		wallet.ReceiverWalletID = string(plainText[4:8])
		wallet.Amount = string(plainText[8:12])
		wallet.Counter = string(plainText[12:16])

		counter, _ := strconv.Atoi(wallet.Counter)
		if counter != 0 {
			messages := []string{"Counter should be 0"}
			tpl.ExecuteTemplate(w, "receive_sync.gohtml", message{false, messages})
		} else {
			counter += 1
		}

		/*
		Start with redis
		*/
		conn, err := redis.Dial("tcp", ":6379")
		defer conn.Close()

		redisVal := redisValues{redisReceiverID: wallet.SenderWalletID, redisReceiverCounter: strconv.Itoa(counter)}
		//Check if the receiver wallet id exists
		redisReceiverCheck, err := redis.Int(conn.Do("EXISTS", redisVal.redisReceiverID))
		if err != nil {
			messages := []string{"Could not connect to redis to check receiver id"}
			tpl.ExecuteTemplate(w, "receive_sync.gohtml", message{false, messages})
			return
		} else if redisReceiverCheck != 0 {
			messages := []string{"Wallet id already synchronised"}
			tpl.ExecuteTemplate(w, "receive_sync.gohtml", message{false, messages})
			return
		}

		/*
		Insert to redis
		*/
		if _, err := conn.Do("SET", redisVal.redisReceiverID, redisVal.redisReceiverCounter); err != nil {
			messages := []string{"Error inserting wallet id in redis"}
			tpl.ExecuteTemplate(w, "generate_sync.gohtml", message{false, messages})
			return
		}
		//
		//r, _ := redis.Strings(conn.Do("GET", "receiver_1105"))
		//fmt.Println(r)
		/*
		End redis
		*/

		messages := []string{wallet.SenderWalletID, redisVal.redisReceiverCounter}
		tpl.ExecuteTemplate(w, "receive_sync.gohtml", message{true, messages})

	} else if req.Method == http.MethodGet {
		err := tpl.ExecuteTemplate(w, "receive_sync.gohtml", nil)
		if err != nil {
			log.Fatal(err)
		}
	}

}

//Method to create a synchronise token to send
func sendSync(w http.ResponseWriter, req *http.Request) {

	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatal(err)
	}

	if req.Method == http.MethodPost {
		req.ParseForm()
		//Get value from the form
		receiverID := strings.TrimSpace(req.PostFormValue("receiver_id"))

		if _, err := strconv.Atoi(receiverID); err != nil {
			messages := []string{"Input intergers only"}
			tpl.ExecuteTemplate(w, "generate_sync.gohtml", message{false, messages})
			return
		} else if len(receiverID) != 4 {
			//tpl.ExecuteTemplate(w, "generate_sync.gohtml", struct {Success bool;string}{false, "Input length of 4 only"})
			messages := []string{"Input length of 4"}
			tpl.ExecuteTemplate(w, "generate_sync.gohtml", message{false, messages})
			return
		}
		//Get the sender id from redis
		senderID, err := redis.String(conn.Do("GET", "sender_id"))
		if err != nil {
			messages := []string{"Could not connect to redis and get the sender_id"}
			tpl.ExecuteTemplate(w, "generate_sync.gohtml", message{false, messages})
			return
		}

		//Initialise with a struct
		sync := token{SenderWalletID: senderID, ReceiverWalletID: receiverID, Amount: "0000", Counter: "0000"}

		//Concat all the fields for encryption
		plainToken := []byte(sync.SenderWalletID + sync.ReceiverWalletID + sync.Amount + sync.Counter)
		//Decode the private key
		key := lib.DecodeString(PrivateKey)
		//Encrypt the token
		sync.EncryptedToken, err = lib.Encrypt(key, plainToken)

		if err != nil {
			messages := []string{"Could not encrypt token. Please try again later"}
			tpl.ExecuteTemplate(w, "generate_sync.gohtml", message{false, messages})
			return
		}

		messages := []string{sync.SenderWalletID, sync.ReceiverWalletID, sync.Amount, sync.Counter, sync.EncryptedToken}
		err = tpl.ExecuteTemplate(w, "generate_sync.gohtml", message{true, messages})
		if err != nil {
			log.Fatal(err)
		}

	}
	if req.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "generate_sync.gohtml", nil)
		return

	}

}
func receiveMoney(w http.ResponseWriter, req *http.Request) {

	err := tpl.ExecuteTemplate(w, "receive.gohtml", nil)

	if err != nil {
		log.Fatal(err)
	}
}
func sendMoney(w http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {
		req.ParseForm()
		receiverID := strings.TrimSpace(req.PostFormValue("receiver_id"))
		amount := strings.TrimSpace(req.PostFormValue("amount"))

		if _, err := strconv.Atoi(receiverID); err != nil {
			messages := []string{"Please input integers only"}
			tpl.ExecuteTemplate(w, "send.gohtml", message{false, messages})
			return
		} else if len(receiverID) != 4 {
			messages := []string{"Input length of 4 only"}
			tpl.ExecuteTemplate(w, "send.gohtml", message{false, messages})
			return
		}

		if _, err := strconv.Atoi(amount); err != nil {
			messages := []string{"Please input integers only"}
			tpl.ExecuteTemplate(w, "send.gohtml", message{false, messages})
			return
		}
		conn, err := redis.Dial("tcp", ":6379")
		if err != nil {
			messages := []string{"Could not connect to redis"}
			tpl.ExecuteTemplate(w, "send.gohtml", message{false, messages})
			return
		}

		receiverIDCheck, err := redis.Int(conn.Do("EXISTS", receiverID))

		if receiverIDCheck != 1 {
			messages := []string{"Wallet ID not found. Please sync first"}
			tpl.ExecuteTemplate(w, "send.gohtml", message{false, messages})
			return
		}

		amountRedis, _ := redis.String(conn.Do("GET", "amount"))

		amountCheck, _ := strconv.Atoi(amountRedis)
		amountInt, _ := strconv.Atoi(amount)

		if amountInt > amountCheck {
			messages := []string{"Insufficient balance", "Current balance: $" + strconv.Itoa(amountCheck)}
			tpl.ExecuteTemplate(w, "send.gohtml", message{false, messages})
			return
		}

		remainingAmount := amountCheck - amountInt
		/*
		Generate Token
		 */
		token := new(token)
		token.SenderWalletID, _ = redis.String(conn.Do("GET", "sender_id"))
		token.ReceiverWalletID = receiverID
		token.Amount, _ = lib.PadStringLeft(amount)
		counterTemp, _ := redis.String(conn.Do("GET", receiverID))
		token.Counter, _ = lib.PadStringLeft(counterTemp)

		//Concat all the fields for encryption
		plainToken := []byte(token.SenderWalletID + token.ReceiverWalletID + token.Amount + token.Counter)
		//Decode the private key
		key := lib.DecodeString(PrivateKey)
		//Encrypt the token
		token.EncryptedToken, err = lib.Encrypt(key, plainToken)

		conn.Do("INCR", receiverID)
		conn.Do("SET", "amount", remainingAmount)

		str, _ := redis.Strings(conn.Do("MGET", receiverID, "amount"))
		fmt.Println(str)

		messages := []string{token.ReceiverWalletID, strconv.Itoa(remainingAmount), token.EncryptedToken}
		tpl.ExecuteTemplate(w, "send.gohtml", message{true,messages})
		return

	} else {
		err := tpl.ExecuteTemplate(w, "send.gohtml", nil)
		if err != nil {
			log.Fatal(err)
		}
	}
}
