package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	UserKey       = "user"
	ImprintKey    = "imprint"
	AuthCookieKey = "wakapi_auth"
)

type MigrationFunc func(db *gorm.DB) error

type KeyStringValue struct {
	Key   string `gorm:"primary_key"`
	Value string `gorm:"type:text"`
}

type Interval struct {
	Start time.Time
	End   time.Time
}

type CustomTime time.Time

func (j *CustomTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.String())
}

func (j *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Replace(strings.Trim(string(b), "\""), ".", "", 1)
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	t := time.Unix(0, i*int64(math.Pow10(19-len(s))))
	*j = CustomTime(t)
	return nil
}

// heartbeat timestamps arrive as strings for sqlite and as time.Time for postgres
func (j *CustomTime) Scan(value interface{}) error {
	var (
		t   time.Time
		err error
	)

	switch value.(type) {
	case string:
		t, err = time.Parse("2006-01-02 15:04:05-07:00", value.(string))
		if err != nil {
			return errors.New(fmt.Sprintf("unsupported date time format: %s", value))
		}
	case int64:
		t = time.Unix(0, value.(int64))
		break
	case time.Time:
		t = value.(time.Time)
		break
	default:
		return errors.New(fmt.Sprintf("unsupported type: %T", value))
	}

	t = time.Unix(0, (t.UnixNano()/int64(time.Millisecond))*int64(time.Millisecond)) // round to millisecond precision
	*j = CustomTime(t)

	return nil
}

func (j *CustomTime) Hash() (uint64, error) {
	return uint64((j.T().UnixNano() / 1000) / 1000), nil
}

func (j CustomTime) Value() (driver.Value, error) {
	t := time.Unix(0, j.T().UnixNano()/int64(time.Millisecond)*int64(time.Millisecond)) // round to millisecond precision
	return t, nil
}

func (j CustomTime) String() string {
	t := time.Time(j)
	return t.Format("2006-01-02 15:04:05.000")
}

func (j CustomTime) T() time.Time {
	return time.Time(j)
}

func (j CustomTime) Valid() bool {
	return j.T().Unix() >= 0
}
