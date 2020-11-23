package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/skratchdot/open-golang/open"
	"github.com/supersidor/msfs2020-go/simconnect"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// ported from: MSFS-SDK/Samples/SimConnectSamples/RequestData/RequestData.cpp
// build: GOOS=windows GOARCH=amd64 go build github.com/lian/msfs2020-go/examples/request_data

type Report struct {
	simconnect.RecvSimobjectDataByType
	Title         [256]byte `name:"TITLE"`
	Altitude      float64   `name:"INDICATED ALTITUDE" unit:"feet"` // PLANE ALTITUDE or PLANE ALT ABOVE GROUND
	Latitude      float64   `name:"PLANE LATITUDE" unit:"degrees"`
	Longitude     float64   `name:"PLANE LONGITUDE" unit:"degrees"`
	Heading       float64   `name:"PLANE HEADING DEGREES TRUE" unit:"degrees"`
	Airspeed      float64   `name:"AIRSPEED INDICATED" unit:"knot"`
	AirspeedTrue  float64   `name:"AIRSPEED TRUE" unit:"knot"`
	VerticalSpeed float64   `name:"VERTICAL SPEED" unit:"ft/min"`
	Flaps         float64   `name:"TRAILING EDGE FLAPS LEFT ANGLE" unit:"degrees"`
	Trim          float64   `name:"ELEVATOR TRIM PCT" unit:"percent"`
	RudderTrim    float64   `name:"RUDDER TRIM PCT" unit:"percent"`
}

func (r *Report) RequestData(s *simconnect.SimConnect) {
	defineID := s.GetDefineID(r)
	requestID := defineID
	s.RequestDataOnSimObjectType(requestID, defineID, 0, simconnect.SIMOBJECT_TYPE_USER)
}

type Request struct {
	Title     string  `json:"title"`
	Altitude  float64 `json:"altitude"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Heading   float64 `json:"heading"`
	Timestamp int64   `json:"timestamp"`
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

type UserInfo struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

const baseUrl string = "http://localhost:3000/ui/login_console"
const serverPort = 9999

var token string

const tokenFileName = "token.jwt"

func authenticate(baseUrl string, serverPort int) (string, error) {
	targetUrl := baseUrl + "?redirect_uri=http://localhost:" + strconv.Itoa(serverPort) + "/oauth2/callback"
	srv := &http.Server{Addr: ":" + strconv.Itoa(serverPort)}
	var result string
	chQuit := make(chan bool)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		queryParts, _ := url.ParseQuery(r.URL.RawQuery)
		result = queryParts["token"][0]
		log.Printf("token: %s\n", token)

		msg := "<h1><strong>Success!</strong></h1>"
		msg = msg + "<p>You are authenticated.You could close window now and return to the CLI.</p>"
		fmt.Fprintf(w, msg)
		chQuit <- true
	})
	go func() {
		log.Println("server starting")
		if err := srv.ListenAndServe(); err != nil {
			//log.Fatalf("listenAndServe failed: %v", err)
		}
	}()
	fmt.Println("server started")
	open.Run(targetUrl)

	_ = <-chQuit

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	return result, nil
}

func loadToken(fileName string) string {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) || !info.IsDir() {
		return ""
	}
	content, err := ioutil.ReadFile(fileName)

	return string(content)
}
func saveToken(fileName string, token string) {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
	} else {
		file.WriteString(token)
	}
	file.Close()
}
func getUser(token string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", "http://localhost:8080/api/user/me", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		var userInfo UserInfo
		err = json.Unmarshal(bodyBytes, &userInfo)
		fmt.Println(string(bodyBytes))
		if err != nil {
			return nil, err
		}
		return &userInfo, nil
	} else {
		return nil, fmt.Errorf("Got status code %d", resp.StatusCode)
	}

	//resp, err := http.Get("https://localhost:8080/api/user/me")
}
func main() {

	var userInfo *UserInfo
	token := loadToken(tokenFileName)
	passed := false
	if len(token) > 0 {
		userInfo, _ = getUser(token)
		if userInfo != nil {
			fmt.Println(userInfo)
			passed = true
		} else {
			fmt.Println("Failed to get userInfo.Request authentication")
		}
	}
	if !passed {
		log.Println(color.CyanString("You will now be taken to your browser for authentication"))
		time.Sleep(1 * time.Second)
		token, _ = authenticate(baseUrl, serverPort)
		if len(token) > 0 {
			userInfo, _ := getUser(token)
			if userInfo != nil {
				saveToken(tokenFileName, token)
				fmt.Println(userInfo)
			} else {
				log.Fatal("Failed to get userInfo")
			}
		}
	}

	//http.HandleFunc("/oauth2/callback", callbackHandler)
	//log.Fatal(http.ListenAndServe(":9999", nil))

	//heading := 0
	//for{
	//	req := &Request{
	//		Title:"test",
	//		Altitude:10.0,
	//		Latitude:50.4501,
	//		Longitude:30.5234,
	//		Heading:float64(heading),
	//		Timestamp: makeTimestamp(),
	//	}
	//	sendData(req);
	//
	//	time.Sleep(1000 * time.Millisecond)
	//	heading = heading + 10;
	//	if heading > 360 {
	//		heading = heading - 360
	//	}
	//
	//}
}

func sendData(req *Request) {
	b, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))
	_, err = http.Post("http://localhost:8080/sim", "application/json", bytes.NewBuffer(b))
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}
