package main

import (
    "net/http"
    "os/exec"
    "github.com/gin-gonic/gin"
    "os"
	"net/url"
    "encoding/json"
	"bytes"
	"log"
	"fmt"
	"strconv"
	"strings"
	"io/ioutil"
)

var credData = url.Values{
	"audience":       {"https://api.cloud.armory.io"},
	"grant_type": {"client_credentials"},
	"client_secret": {os.Getenv("client-secret")},
	"client_id": {os.Getenv("client-id")},
}

type callback_data struct {
    success    bool  `json:"success"`
    mdMessage  string  `json:"mdMessage"`
}

type auth_data struct {
    access_token  string
	scope	string
	expires_in int
	token_type	string
}

// getAlbums responds with the list of all albums as JSON.
func runCmd(c *gin.Context) {
    fmt.Println("request recieved: ")
	cmd:=c.Query("cmd")
	arg:=c.Query("arg")
	fmt.Println(cmd)
	fmt.Println(arg)
	callbackURL:=c.Query("callbackURL")
	go invoke(cmd,arg,callbackURL)
	c.IndentedJSON(http.StatusOK,"")
}

func invoke(cmd string,arg string,callbackURL string) {
	//kubectl wait -n=borealis-argo rollout/potato-facts --for=condition=Completed
	//out, err := exec.Command(cmd, string.Fields(arg)).Output()
	out, err := exec.Command("kubectl", "wait","-n=borealis-argo","rollout/potato-facts","--for=condition=Completed").Output()
	message:=""
	success:=true
	if err!=nil {
		fmt.Println("error on command")
		message=err.Error()
		success=false
	} else {
		message=string(out[:])
		success=true
	}
    fmt.Println(message)
	
	token:=auth()
    fmt.Println("Authorized")
	callback(token, callbackURL,success,message)
}

func callback(token string,callbackURL string, success bool, message string){
	data := callback_data{success, message}
	serialized, err :=json.Marshal(data)

	/*var jsonData = []byte(`{"success": `+strconv.FormatBool(success)+`,"mdMessage": "`+message+`"}`)*/
	var dataToPass=`{ "success": `+strconv.FormatBool(success)+`, "mdMessage": "`+strings.Trim(message,"\r\n")+`"}`

    if err != nil {
        log.Fatal(err)
    }

    var bearer = "Bearer " + token
	client := &http.Client{}
	
    fmt.Println("posting: "+bearer)
    fmt.Println("posting data:")
	fmt.Println(serialized)
	fmt.Println(dataToPass)
    fmt.Println("posting to:")
    fmt.Println(callbackURL)
	req,err := http.NewRequest("POST",callbackURL,bytes.NewBuffer([]byte(dataToPass)))
    req.Header.Add("Authorization", bearer)
    req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

    if err != nil {
        log.Fatal(err)
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(body))
}

func auth() string{
    resp, err := http.PostForm("https://auth.cloud.armory.io/oauth/token",credData)

    if err != nil {
        log.Fatal(err)
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(body))
	var access_token map[string]interface{}
	err = json.Unmarshal([]byte(body),&access_token)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("unmarshalled:")
    fmt.Println(access_token["access_token"].(string))
	return access_token["access_token"].(string)
}

func main() {
    router := gin.Default()
    router.GET("/cmd", runCmd)

    fmt.Println("starting")
    //router.Run("localhost:8080")
    router.Run()
}