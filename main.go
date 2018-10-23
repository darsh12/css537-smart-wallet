package main

import (
	"./lib"
	"github.com/gomodule/redigo/redis"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

var tpl *template.Template

const PrivateKey = "752EF0D8FB4958670DBA40AB1F3C1D0F8FB4958670DBA40AB1F3752EF0DC1D0F"

type token struct {
	SenderWalletID   string
	ReceiverWalletID string
	Amount           string
	Counter          string
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

type syncToken struct {
	Token          token
	encryptedToken string
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
	conn.Do("FLUSHALL")
	conn.Do("SET", "sender_id", "1105")
	conn.Do("SET", "amount", "0")
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

		wallet := new(syncToken)
		wallet.encryptedToken = req.PostFormValue("sync_token")

		if len(wallet.encryptedToken) != 32 {
			messages := []string{"Please check token"}
			tpl.ExecuteTemplate(w, "receive_sync.gohtml", message{false, messages})
			return
		}

		key := lib.DecodeString(PrivateKey)
		decodeToken := lib.DecodeString(wallet.encryptedToken)
		plainText, err := lib.Decrypt(key, decodeToken)
		if err != nil {
			messages := []string{"Could not create a block"}
			tpl.ExecuteTemplate(w, "receive_sync.gohtml", message{false, messages})
		}
		if len(plainText) != 16 {
			messages := []string{"Incorrect data from token"}
			tpl.ExecuteTemplate(w, "receive_sync.gohtml", message{false, messages})
		}

		wallet.Token.SenderWalletID = string(plainText[0:4])
		wallet.Token.ReceiverWalletID = string(plainText[4:8])
		wallet.Token.Amount = string(plainText[8:12])
		wallet.Token.Counter = string(plainText[12:16])

		counter, _ := strconv.Atoi(wallet.Token.Counter)
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

		redisVal := redisValues{redisReceiverID: wallet.Token.SenderWalletID, redisReceiverCounter: strconv.Itoa(counter)}
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

		messages := []string{wallet.Token.SenderWalletID, redisVal.redisReceiverCounter}
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
		receiverID := req.PostFormValue("receiver_id")

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
		sync := syncToken{Token: token{SenderWalletID: senderID, ReceiverWalletID: receiverID, Amount: "0000", Counter: "0000"}}

		//Concat all the fields for encryption
		plainToken := []byte(sync.Token.SenderWalletID + sync.Token.ReceiverWalletID + sync.Token.Amount + sync.Token.Counter)
		//Decode the private key
		key := lib.DecodeString(PrivateKey)
		//Encrypt the token
		sync.encryptedToken, err = lib.Encrypt(key, plainToken)

		if err != nil {
			messages := []string{"Could not encrypt token. Please try again later"}
			tpl.ExecuteTemplate(w, "generate_sync.gohtml", message{false, messages})
			return
		}

		/*
		Start with redis
		*/
		//conn, err := redis.Dial("tcp", ":6379")
		//defer conn.Close()
		//
		//redisVal := redisValues{redisReceiverID: "receiver_" + receiverID, redisReceiverCounterName: "receiver_counter_" + receiverID, redisReceiverCounter: "0000"}
		////Check if the receiver wallet id exists
		//redisReceiverCheck, err := redis.Int(conn.Do("EXISTS", redisVal.redisReceiverID))
		//if err != nil {
		//	messages := []string{"Could not connect to redis to check receiver id"}
		//	tpl.ExecuteTemplate(w, "generate_sync.gohtml", message{false, messages})
		//	return
		//} else if redisReceiverCheck != 0 {
		//	messages := []string{"Wallet id already synchronised"}
		//	tpl.ExecuteTemplate(w, "generate_sync.gohtml", message{false, messages})
		//	return
		//}
		//
		///*
		//Insert to redis
		// */
		//if _, err := conn.Do("SET", redisVal.redisReceiverID, receiverID); err != nil {
		//	messages := []string{"Error inserting wallet id in redis"}
		//	tpl.ExecuteTemplate(w, "generate_sync.gohtml", message{false, messages})
		//	return
		//}
		//if _, err := conn.Do("SET", redisVal.redisReceiverCounterName, redisVal.redisReceiverCounter); err != nil {
		//	messages := []string{"Error inserting wallet id in redis"}
		//	tpl.ExecuteTemplate(w, "generate_sync.gohtml", message{false, messages})
		//	return
		//}
		//
		//r, _ := redis.Strings(conn.Do("KEYS", "*"))
		//fmt.Println(r)
		/*
		End redis
		 */

		messages := []string{sync.Token.SenderWalletID, sync.Token.ReceiverWalletID, sync.Token.Amount, sync.Token.Counter, sync.encryptedToken}
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
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	//if _, err = conn.Do("SET", "Amount","1105"); err != nil {
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

		amount, err := redis.Int(conn.Do("GET", "Amount"))
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
