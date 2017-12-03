package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/rs/cors"
	"strings"
	"github.com/philhug/bitwarden-client-go/bitwarden"
)

// The data we get from the client. Only used to parse data
type newCipher struct {
	Type           int       `json:"type"`
	FolderId       string    `json:"folderId"`
	OrganizationId string    `json:"organizationId"`
	Name           string    `json:"name"`
	Notes          string    `json:"notes"`
	Favorite       bool      `json:"favorite"`
	Login          loginData `json:"login"`
}

type loginData struct {
	URI      string `json:"uri"`
	Username string `json:"username"`
	Password string `json:"password"`
	ToTp     string `json:"totp"`
}

func handleCollections(w http.ResponseWriter, req *http.Request) {

	collections := bitwarden.List{Object: "list", Data: []string{}}
	data, err := json.Marshal(collections)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handleCipher(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)

	log.Println(email + " is trying to add data")

	acc, err := db.getAccount(email)
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	var data []byte

	if req.Method == "POST" {
		rCiph, err := unmarshalCipherRequest(req)
		if err != nil {
			log.Fatal("Cipher decode error" + err.Error())
		}

		// Store the new cipher object in db
		newCiph, err := db.newCipher(rCiph, acc.Id)
		if err != nil {
			log.Fatal("newCipher error" + err.Error())
		}
		respCiph := bitwarden.NewCipherResponse(newCiph)
		data, err = json.Marshal(&respCiph)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		ciphers, err := db.getCiphers(acc.Id)
		if err != nil {
			log.Println(err)
		}
		ciphs := make([]bitwarden.CipherResponse, len(ciphers))
		for i, c := range ciphers {
			ciphs[i] = bitwarden.NewCipherResponse(c)
		}
		list := bitwarden.List{Object: "list", Data: ciphs}
		data, err = json.Marshal(&list)
		if err != nil {
			log.Fatal("UAAH")
			log.Fatal(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

type deleteIdRequest struct {
	Ids	[]string
}

// This function handles deleteing / needed by web-UI
func handleCipherDelete(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)
	log.Println(email + " is trying to delete his data")

	// Get the cipher id
	id := strings.TrimPrefix(req.URL.Path, "/api/ciphers/")

	acc, err := db.getAccount(email)
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()

	var dreq deleteIdRequest
	err = decoder.Decode(&dreq)
	if err != nil {
		log.Fatal("Cannot decode request " + err.Error())
	}

	for _, id := range dreq.Ids {
		err = db.deleteCipher(acc.Id, id)
		if err != nil {
			w.Write([]byte("0"))
			log.Println(err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(""))
	log.Println("Cipher " + id + " deleted")
	return
}

// This function handles updates and deleteing
func handleCipherUpdate(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)
	log.Println(email + " is trying to edit his data")

	// Get the cipher id
	id := strings.TrimPrefix(req.URL.Path, "/api/ciphers/")

	acc, err := db.getAccount(email)
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	switch req.Method {
	case "GET":
		log.Println("GET Ciphers for " + acc.Id)
		var data []byte
		ciph, err := db.getCipher(acc.Id, id)
		if err != nil {
			log.Println(err)
		}
		data, err = json.Marshal(&ciph)
		if err != nil {
			log.Fatal("UAAH")
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	case "POST":
		log.Println("POST Ciphers")
		var data []byte

		rCiph, err := unmarshalCipherRequest(req)
		if err != nil {
			log.Fatal("Cipher decode error" + err.Error())
		}

		// Store the new cipher object in db
		newCiph, err := db.newCipher(rCiph, acc.Id)
		if err != nil {
			log.Fatal("newCipher error" + err.Error())
		}

		respCiph := bitwarden.NewCipherResponse(newCiph)
		data, err = json.Marshal(&respCiph)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	case "PUT":
		rCiph, err := unmarshalCipherRequest(req)
		if err != nil {
			log.Fatal("Cipher decode error" + err.Error())
		}

		// Set correct ID
		rCiph.Id = id

		err = db.updateCipher(rCiph, acc.Id, id)
		if err != nil {
			w.Write([]byte("0"))
			log.Println(err)
			return
		}

		// Send response
		respCiph := bitwarden.NewCipherResponse(rCiph)
		data, err := json.Marshal(&respCiph)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		log.Println("Cipher " + id + " updated")
		return

	case "DELETE":
		err := db.deleteCipher(acc.Id, id)
		if err != nil {
			w.Write([]byte("0"))
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(""))
		log.Println("Cipher " + id + " deleted")
		return
	default:
		w.Write([]byte("0"))
		return
	}

}

func handleSync(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)

	log.Println(email + " is trying to sync")

	acc, err := db.getAccount(email)

	prof := bitwarden.Profile{
		Id:               acc.Id,
		Email:            acc.Email,
		EmailVerified:    false,
		Premium:          false,
		Culture:          "en-US",
		TwoFactorEnabled: false,
		Key:              acc.Key,
		SecurityStamp:    "123",
		Organizations:    nil,
		Object:           "profile",
	}

	ciphers, err := db.getCiphers(acc.Id)
	if err != nil {
		log.Println(err)
	}

	ciphs := make([]bitwarden.CipherDetailsResponse, len(ciphers))
	for i, c := range ciphers {
		ciphs[i] = bitwarden.NewCipherDetailsResponse(c)
	}

	folders, err := db.getFolders(acc.Id)
	if err != nil {
		log.Println(err)
	}

	Domains := bitwarden.Domains{
		Object:            "domains",
		EquivalentDomains: nil,
		GlobalEquivalentDomains: []bitwarden.GlobalEquivalentDomains{
			bitwarden.GlobalEquivalentDomains{Type: 1, Domains: []string{"youtube.com", "google.com", "gmail.com"}, Excluded: false},
		},
	}

	data := bitwarden.SyncData{
		Profile: prof,
		Folders: folders,
		Domains: Domains,
		Object:  "sync",
		Ciphers: ciphs,
	}

	jdata, err := json.Marshal(&data)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jdata)
}

func handleFolder(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)

	log.Println(email + " is trying to add a new folder")

	acc, err := db.getAccount(email)
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	var data []byte
	if req.Method == "POST" {
		decoder := json.NewDecoder(req.Body)

		var folderData struct {
			Name string `json:"name"`
		}

		err = decoder.Decode(&folderData)
		if err != nil {
			log.Fatal(err)
		}
		defer req.Body.Close()

		folder, err := db.addFolder(folderData.Name, acc.Id)
		if err != nil {
			log.Fatal("newFolder error" + err.Error())
		}

		data, err = json.Marshal(&folder)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		folders, err := db.getFolders(acc.Id)
		if err != nil {
			log.Println(err)
		}
		list := bitwarden.List{Object: "list", Data: folders}
		data, err = json.Marshal(list)
		if err != nil {
			log.Fatal(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func unmarshalCipherRequest(req *http.Request) (bitwarden.Cipher, error){
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()

	var creq bitwarden.CipherRequest
	err := decoder.Decode(&creq)
	if err != nil {
		log.Fatal("Cannot decode request " + err.Error())
	}

	rCiph, err := creq.ToCipher()
	if err != nil {
		log.Fatal("Cipher decode error" + err.Error())
	}
	return rCiph, err
}

// Interface to make testing easier
type database interface {
	init() error
	addAccount(acc bitwarden.Account) error
	getAccount(username string) (bitwarden.Account, error)
	updateAccountInfo(acc bitwarden.Account) error
	getCipher(owner string, ciphID string) (bitwarden.Cipher, error)
	getCiphers(owner string) ([]bitwarden.Cipher, error)
	newCipher(ciph bitwarden.Cipher, owner string) (bitwarden.Cipher, error)
	updateCipher(newData bitwarden.Cipher, owner string, ciphID string) error
	deleteCipher(owner string, ciphID string) error
	open() error
	close()
	addFolder(name string, owner string) (bitwarden.Folder, error)
	getFolders(owner string) ([]bitwarden.Folder, error)
}

func main() {
	initDB := flag.Bool("init", false, "Initialize the database")
	flag.Parse()

	err := db.open()
	if err != nil {
		log.Fatal(err)
	}

	defer db.close()

	// Create a new database
	if *initDB {
		err := db.init()
		if err != nil {
			log.Fatal(err)
		}
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/accounts/register", handleRegister)
	mux.HandleFunc("/identity/connect/token", handleLogin)

	mux.Handle("/api/accounts/keys", jwtMiddleware(http.HandlerFunc(handleKeysUpdate)))
	mux.Handle("/api/accounts/profile", jwtMiddleware(http.HandlerFunc(handleProfile)))
	mux.Handle("/api/collections", jwtMiddleware(http.HandlerFunc(handleCollections)))
	mux.Handle("/api/folders", jwtMiddleware(http.HandlerFunc(handleFolder)))
	mux.Handle("/apifolders", jwtMiddleware(http.HandlerFunc(handleFolder))) // The android app want's the address like this, will be fixed in the next version. Issue #174
	mux.Handle("/api/sync", jwtMiddleware(http.HandlerFunc(handleSync)))

	mux.Handle("/api/ciphers", jwtMiddleware(http.HandlerFunc(handleCipher)))
	mux.Handle("/api/ciphers/delete", jwtMiddleware(http.HandlerFunc(handleCipherDelete)))
	mux.Handle("/api/ciphers/", jwtMiddleware(http.HandlerFunc(handleCipherUpdate)))

	//mux.Handle("/api/ciphers", jwtMiddleware(http.HandlerFunc(handleCipher)))

	log.Println("Starting server on " + serverAddr)
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
	}).Handler(mux)
	http.ListenAndServe(serverAddr, handler)
}
