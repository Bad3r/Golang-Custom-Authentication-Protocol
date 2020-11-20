package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type authToken struct {
	Auth  string `json:"auth"`
	Token string `json:"token"`
}

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
	// attention: If you do not call ParseForm method, the following data can not be obtained form
	fmt.Println("Form: ", r.Form) // print information on server side.
	fmt.Println("path: ", r.URL.Path)
	fmt.Println("scheme: ", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello Bader!") // write data to response
}

func decodeBase64(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func parseToken(body string) (*authToken, error) {

	var s = new(authToken)
	err := json.Unmarshal([]byte(body), &s)
	if err != nil {
		log.Fatalln(err)
	}
	return s, err
}
func handleLogin(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	requestBody, err := json.Marshal(map[string]string{
		"client_id":     username,
		"client_secret": password,
	})

	if err != nil {
		log.Fatalln(err)
	}

	// initialize http client & set timeout
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest(http.MethodPost, "http://auth_server:3000/", bytes.NewBuffer(requestBody))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	// get the SHA256 of user password
	hasher := sha256.New()
	hasher.Write([]byte(password))

	b64Resp := string(body)
	log.Println("*** client b64 string ***\n" + b64Resp + "\n*** client b64 string ***")
	decodedBody := decodeBase64(b64Resp)
	decryptedBody, err := decrypt(hasher.Sum(nil), decodedBody)
	log.Println("*** decrypted body ***\n" + string(decryptedBody) + "\n*** decrypted body ***")
	tokenStr := string(decryptedBody)

	token, err := parseToken(tokenStr)

	mapD := map[string]string{"auth": token.Auth, "token": token.Token}
	mapB, _ := json.Marshal(mapD)
	log.Println("new json\n" + string(mapB) + "*** new json ***")

	if err != nil {
		log.Fatalln(err)
	}

	// send request to web app with encrypted token

	appReq, err := http.NewRequest(http.MethodPost, "http://web_application:9000/", bytes.NewBuffer(mapB))
	appReq.Header.Set("Content-type", "application/json")

	if err != nil {
		log.Fatalln(err)
	}

	r.Body = ioutil.NopCloser(strings.NewReader(tokenStr))
	r.ContentLength = int64(len(tokenStr))

	http.Redirect(w, r, "http://web_application:9000/", 307)

}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t := template.Must(template.ParseFiles("templates/login.gtpl"))
		t.Execute(w, nil)
	} else {
		handleLogin(w, r)

	}
}

//Go application entrypoint
func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	//Our HTML comes with CSS that go needs to provide when we run the app. Here we tell go to create
	// a handle that looks in the static directory, go then uses the "/static/" as a url that our
	//html can refer to when looking for our css and other files.

	http.Handle("/static/", //final url can be anything
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))
	//Go looks in the relative "static" directory first using http.FileServer(), then matches it to a
	//url of our choice as shown in http.Handle("/static/"). This url is what we need when referencing our css files
	//once the server begins. Our html code would therefore be <link rel="stylesheet" href="/static/stylesheet/...">
	//It is important to note the url in http.Handle can be whatever we like, so long as we are consistent.

	http.HandleFunc("/", sayhelloName) // setting router rule
	http.HandleFunc("/login", login)

	fmt.Println("Listening")
	//Start the web server, set the port to listen to 8080. Without a path it assumes localhost
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
