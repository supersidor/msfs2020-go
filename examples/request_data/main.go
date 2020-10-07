package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/supersidor/msfs2020-go/simconnect"
	"net/http"
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

func main() {

	req := &Request{Title: "Frank", Altitude: 1, Latitude: 0.1, Longitude: 0.1, Heading: 360}
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

	s, err := simconnect.New("Request Data")
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to Flight Simulator!")

	report := &Report{}
	s.RegisterDataDefinition(report)
	report.RequestData(s)
	for {
		ppData, r1, err := s.GetNextDispatch()

		if r1 < 0 {
			if uint32(r1) == simconnect.E_FAIL {
				// skip error, means no new messages?
				continue
			} else {
				panic(fmt.Errorf("GetNextDispatch error: %d %s", r1, err))
			}
		}

		recvInfo := *(*simconnect.Recv)(ppData)
		//fmt.Println(ppData, pcbData, recvInfo)

		switch recvInfo.ID {
		case simconnect.RECV_ID_EXCEPTION:
			recvErr := *(*simconnect.RecvException)(ppData)
			fmt.Printf("SIMCONNECT_RECV_ID_EXCEPTION %#v\n", recvErr)

		case simconnect.RECV_ID_OPEN:
			recvOpen := *(*simconnect.RecvOpen)(ppData)
			fmt.Println("SIMCONNECT_RECV_ID_OPEN", fmt.Sprintf("%s", recvOpen.ApplicationName))
			//spew.Dump(recvOpen)
		case simconnect.RECV_ID_EVENT:
			recvEvent := *(*simconnect.RecvEvent)(ppData)
			fmt.Println("SIMCONNECT_RECV_ID_EVENT")
			//spew.Dump(recvEvent)

			switch recvEvent.EventID {
			//case eventSimStartID:
			//	fmt.Println("SimStart Event")
			default:
				fmt.Println("unknown SIMCONNECT_RECV_ID_EVENT", recvEvent.EventID)
			}

		case simconnect.RECV_ID_SIMOBJECT_DATA_BYTYPE:
			recvData := *(*simconnect.RecvSimobjectDataByType)(ppData)
			fmt.Println("SIMCONNECT_RECV_SIMOBJECT_DATA_BYTYPE")

			switch recvData.RequestID {
			case s.DefineMap["Report"]:
				report := (*Report)(ppData)
				fmt.Printf("REPORT: %s: GPS: %.6f,%.6f Altitude: %.0f\n", report.Title, report.Latitude, report.Longitude, report.Altitude)
				if report.Longitude > 0.1 || report.Latitude > 0.1 {
					req := &Request{
						Title:     string(report.Title[:bytes.IndexByte(report.Title[:], 0)]),
						Altitude:  report.Altitude,
						Latitude:  report.Latitude,
						Longitude: report.Longitude,
						Heading:   report.Heading,
						Timestamp: makeTimestamp(),
					}
					sendData(req)
				}
				report.RequestData(s)
			}

		default:
			fmt.Println("recvInfo.dwID unknown", recvInfo.ID)
		}

		time.Sleep(1000 * time.Millisecond)
	}

	fmt.Println("close")

	if err = s.Close(); err != nil {
		panic(err)
	}
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
