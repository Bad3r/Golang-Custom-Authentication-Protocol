package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// var (
// 	oauthConfig *oauth2.Config
// )

// func init() {
// 	oauthConfig = &oauth2.Config{
// 		ClientID:     "db",
// 		ClientSecret: "db",
// 		Scopes:       []string{"all"},
// 		RedirectURL: "http://client/",
// 	}
// }

type clientOauthRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type oauthTokenResp struct {
	AccessToken string `json:"access_token"`
	expiresIn   int    `json:"client_secret"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func dumpRequest(st string, r *http.Request, body string) {

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println(err)
	}

	newST := st + "Request Header:\n" + string(requestDump) + body + "\n***Request Header***"
	log.Println(newST)

}

func dumpResponse(st string, r *http.Response, body string) {

	requestDump, err := httputil.DumpResponse(r, true)
	if err != nil {
		log.Println(err)
	}
	newST := st + "Respose Header:\n" + string(requestDump) + body + "\n***Respose Header***"
	log.Println(newST)

}

func getCreds(body []byte) (*clientOauthRequest, error) {

	var s = new(clientOauthRequest)
	err := json.Unmarshal(body, &s)
	if err != nil {
		log.Fatalln(err)
	}
	return s, err
}

func parseOauthToken(body []byte) (*oauthTokenResp, error) {

	var s = new(oauthTokenResp)
	err := json.Unmarshal(body, &s)
	if err != nil {
		log.Fatalln(err)
	}
	return s, err
}

func encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}
func createHash(key string, data []byte) string {
	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(key))
	// Write Data to it
	h.Write([]byte(data))

	// Get result and encode as hexadecimal string
	sha := hex.EncodeToString(h.Sum(nil))

	return sha

}

func getAccessToken(w http.ResponseWriter, r *http.Request) {

	endpoint := "http://oauth_provider/token.php"

	// extract clientID and clientSecret from the client request
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err.Error())
	}

	s, err := getCreds([]byte(reqBody))
	if err != nil {
		log.Fatalln(err)
	}
	//  print the request
	st := "[1] Client -> auth_server"
	dumpRequest(st, r, string(reqBody))
	r.Body.Close()

	// craft the new request data/body
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", s.ClientID)
	data.Set("client_secret", s.ClientSecret)

	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.Header.Set("Content-Length", strconv.Itoa(len(data.Encode())))
	request.SetBasicAuth(s.ClientID, s.ClientSecret)

	if err != nil {
		log.Fatalln(err)
	}

	// create a client and set ransport and timeout
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}

	timeout := time.Duration(5 * time.Second)

	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	//  handle response from Oauth_provider
	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode == 400 {
		mapD := map[string]string{"auth": "fail", "token": ""}
		mapB, _ := json.Marshal(mapD)
		w.Header().Set("Content-Type", "application/json")
		w.Write(mapB)
		return
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	oauthResp, err := parseOauthToken([]byte(respBody))
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(oauthResp.AccessToken)

	// TODO: check if token is empty
	stResp := "[3] Oauth_provider -> auth_server"
	dumpResponse(stResp, resp, string(respBody))

	// key known only to auth_server and web_application
	key := []byte("a very very very very secret key") // 32 bytes
	encToken, err := encrypt(key, respBody)
	tkn := encodeBase64(encToken)
	log.Println("*** encrypted ***\n" + tkn + "\n*** encrypted ***")
	// craft client success response
	mapD := map[string]string{"auth": "success", "token": tkn}
	mapB, _ := json.Marshal(mapD)
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")

	// get the SHA256 of user password
	hasher := sha256.New()
	hasher.Write([]byte(s.ClientSecret))
	// encrypt the response with the hash of the user password
	encResp, err := encrypt(hasher.Sum(nil), mapB)
	b64Resp := encodeBase64(encResp)

	log.Println("*** b64 string ***\n" + b64Resp + "\n" + reflect.TypeOf(b64Resp).String() + "\n*** b64 string ***")
	w.Write([]byte(b64Resp))

}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", getAccessToken)
	http.ListenAndServe(":"+port, mux)

}
