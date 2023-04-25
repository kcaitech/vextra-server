package time

import (
	"encoding/json"
	"log"
	"testing"
	"time"
)

type TimeData struct {
	Time Time
}

func TestTimeMarshal(t *testing.T) {
	data := TimeData{
		Time: Time(time.Now()),
	}
	dataJson, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(string(dataJson))
}

func TestTimeUnmarshalJSON(t *testing.T) {
	timeJson := `{"Time":"2023-04-25 16:15:00.133846"}`
	var myTime TimeData
	if err := json.Unmarshal([]byte(timeJson), &myTime); err != nil {
		t.Fatal(err)
	}
	log.Println(myTime)
}
