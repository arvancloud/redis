package record

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

const SerialTimeFormat = "20060102"

var errCounterExceed = errors.New("counter > 99")
var errSerialInFuture = errors.New("date of serial is in the future")
var errWrongFormat = errors.New("the provided serial is in the wrong format (!= 10 digits)")

// DefaultSerial returns a 10 digit serial based on the current day and the minute of the day
func DefaultSerial() uint32 {
	n := time.Now().UTC()
	// calculate two digit number (0-99) based on the minute of the day, 1440 / 14.4545 = 99,0003
	c := int(math.Floor(((float64(n.Hour() + 1)) * float64(n.Minute()+1)) / 14.5454))
	ser, err := strconv.ParseUint(fmt.Sprintf("%s%02d", n.Format(SerialTimeFormat), c), 10, 32)
	if err != nil {
		return uint32(time.Now().Unix())
	}
	return uint32(ser)
}

// NewSerial returns a new Serial where the date is set to `Now().UTC()` and the
// counter to 0
func NewSerial() uint32 {
	ser, err := strconv.ParseUint(fmt.Sprintf("%s00", time.Now().UTC().Format(SerialTimeFormat)), 10, 32)
	if err != nil {
		panic(err)
	}
	return uint32(ser)
}

// IncrementSerial increments the given 10 digit serial by 1 if it's on the same date,
// otherwise the date is set to `Now().UTC()` and the counter to 0
// If the counter exceeds 99, or the date part is invalid or in the future, an error is returned
func IncrementSerial(serial uint32) (uint32, error) {
	s := strconv.FormatUint(uint64(serial), 10)

	var (
		t = time.Now().UTC().Truncate(24 * time.Hour)
		c = 0
	)
	if len(s) == 10 {
		if ts, err := time.Parse(SerialTimeFormat, s[:8]); err == nil {
			c, err = strconv.Atoi(s[8:])
			if err != nil {
				return 0, err
			}

			if ts.Sub(t) < 0 {

				c = 0
			} else if ts.Sub(t) > 0 {
				return 0, errSerialInFuture
			} else {
				c += 1
			}
		} else {
			return 0, err
		}
	} else if serial > 0 {
		return 0, errWrongFormat
	}

	if c > 99 {
		return 0, errCounterExceed
	}
	ser, err := strconv.ParseUint(fmt.Sprintf("%s%02d", t.Format(SerialTimeFormat), c), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(ser), nil
}
