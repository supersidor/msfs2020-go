package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"github.com/supersidor/msfs2020-go/simconnect"
	"log"
	"net/http"
	"net/url"
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
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	queryParts, _ := url.ParseQuery(r.URL.RawQuery)
	token := queryParts["token"][0]
	log.Printf("token: %s\n", token)

	//
	//// Use the authorization code that is pushed to the redirect
	//// URL.
	//code := queryParts["code"][0]
	//log.Printf("code: %s\n", code)
	//
	//// Exchange will do the handshake to retrieve the initial access token.
	//tok, err := conf.Exchange(ctx, code)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Printf("Token: %s", tok)
	//// The HTTP Client returned by conf.Client will refresh the token as necessary.
	//client := conf.Client(ctx, tok)
	//
	//resp, err := client.Get("http://some-server.example.com/")
	//if err != nil {
	//	log.Fatal(err)
	//} else {
	//	log.Println(color.CyanString("Authentication successful"))
	//}
	//defer resp.Body.Close()
	//
	//// show succes page
	//msg := "<p><strong>Success!</strong></p>"
	//msg = msg + "<p>You are authenticated and can now return to the CLI.</p>"
	//fmt.Fprintf(w, msg)
}
func main() {
	url := "http://localhost:3000/ui/login?redirect_uri=http://localhost:9999/oauth2/callback"

	open.Run(url)

	http.HandleFunc("/oauth2/callback", callbackHandler)
	log.Fatal(http.ListenAndServe(":9999", nil))

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
