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
    access_token  string  `json:"access_token"`
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	cmd:=c.Query("cmd")
	arg:=c.Query("arg")
	callbackURL:=c.Query("callbackURL")
	out, err := exec.Command(cmd, arg).Output()
	message:=""
	success:=true
	if err!=nil {
		message=err.Error()
		c.IndentedJSON(http.StatusInternalServerError,err.Error())
		success=false
	} else {
		c.IndentedJSON(http.StatusOK,string(out[:]))
		message=string(out[:])
	}
	
	token:=auth()
	callback(token, callbackURL,success,message)
}

func callback(token string,callbackURL string, success bool, messae string){
	data := &callback_data{true, "message"}
	serialized, err :=json.Marshal(data)

    var bearer = "Bearer " + token
	client := &http.Client{}
	req,err := http.NewRequest("POST",callbackURL,bytes.NewBuffer(serialized))
    req.Header.Add("Authorization", bearer)
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
	var access_token auth_data
	json.Unmarshal([]byte(body),&access_token)
	return access_token.access_token
}

func main() {
    router := gin.Default()
    router.GET("/albums", getAlbums)

    router.Run("localhost:8080")
}