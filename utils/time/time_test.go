/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package time

import (
	"encoding/json"
	"fmt"
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
	log.Println(Time(time.Now()).String())
}

func TestTimeUnmarshalJSON(t *testing.T) {
	timeJson := `{"Time":"2023-04-25 16:15:00.133846"}`
	var myTime TimeData
	if err := json.Unmarshal([]byte(timeJson), &myTime); err != nil {
		t.Fatal(err)
	}
	log.Println(myTime)
	log.Println(Time{}.IsZero())
	log.Println(myTime.Time.IsZero())
}

func TestTimeParse(t *testing.T) {
	t0 := &Time{}
	if err := t0.Parse("2023-04-25 16:15:00.133846"); err != nil {
		fmt.Println(err)
	}
	fmt.Println(t0)
	if err := t0.Parse("2023-04-25 16:15:00"); err != nil {
		fmt.Println(err)
	}
	fmt.Println(t0)
	if err := t0.Parse("2023-04-25"); err != nil {
		fmt.Println(err)
	}
	fmt.Println(t0)
}
