package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
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

func dumpRequest(st string, r *http.Request, body string) {

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}

	newST := st + "Request Header:\n" + string(requestDump) + body
	fmt.Println(newST)

}

func dumpResponse(st string, r *http.Response, body string) {

	requestDump, err := httputil.DumpResponse(r, true)
	if err != nil {
		fmt.Println(err)
	}
	newST := st + "Respose Header:\n" + string(requestDump) + body
	fmt.Println(newST)

}

func getCreds(body []byte) (*clientOauthRequest, error) {

	var s = new(clientOauthRequest)
	err := json.Unmarshal(body, &s)
	if err != nil {
		log.Fatalln(err)
	}
	return s, err
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

	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(body))

	// fmt.Println("#### dumping Oauth response:")
	// dumpResponse(resp)

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
