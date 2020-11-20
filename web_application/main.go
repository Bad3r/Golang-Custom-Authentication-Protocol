package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)

type authToken struct {
	Auth  string `json:"auth"`
	Token string `json:"token"`
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

func formHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println(err)
	}
	newST := "Respose Header:\n" + string(requestDump) + string(body) + "\n***Respose Header***"
	log.Println(newST)

	if r.Method != http.MethodPost {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	// token, err := parseToken(string(body))
	// tokenStr := "Token: " + token.Token
	// authStr := "Authentication Status: " + token.Auth
	// fmt.Fprintln(w, tokenStr)
	fmt.Fprintln(w, "welcome!")

}

func main() {
	http.HandleFunc("/", formHandler)

	fmt.Printf("Starting server at port 9000\n")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}
