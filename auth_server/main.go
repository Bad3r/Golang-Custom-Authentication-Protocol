package main

import (
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

func dumpRequest(r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("######## the request")
	fmt.Println(string(requestDump))
	fmt.Println("######## the body")
	log.Println(string(body))
	fmt.Println("######## end of body")

	r.Body.Close()
}

func dumpResponse(r *http.Response) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}

	requestDump, err := httputil.DumpResponse(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("######## the request")
	fmt.Println(string(requestDump))
	fmt.Println("######## the body")
	log.Println(string(body))

	r.Body.Close()

}

func getAccessToken(w http.ResponseWriter, r *http.Request) {

	// fmt.Println("#### dumping client request:")
	// dumpRequest(r)

	endpoint := "http://oauth_provider/token.php"
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", "db")
	data.Set("client_secret", "db")

	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	request.Header.Set("Content-type", "application/x-www-form-urlencoded")
	request.Header.Set("Content-Length", strconv.Itoa(len(data.Encode())))

	if err != nil {
		log.Fatalln(err)
	}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}

	timeout := time.Duration(500 * time.Second)
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	// fmt.Println("#### dumping crafted request:")
	// dumpRequest(request)

	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

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
