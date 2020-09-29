package record

import (
	"strconv"
	"testing"
	"time"
)

func TestIncrementSerial(t *testing.T) {

	sn, err := strconv.ParseUint(time.Now().UTC().Add(-24*time.Hour).Format(SerialTimeFormat), 10, 32)
	if err != nil {
		t.Error(err)
	}

	s0, err := strconv.ParseUint(time.Now().UTC().Format(SerialTimeFormat), 10, 32)
	if err != nil {
		t.Error(err)
	}

	s1, err := strconv.ParseUint(time.Now().UTC().Add(48*time.Hour).Format(SerialTimeFormat), 10, 32)
	if err != nil {
		t.Error(err)
	}
	// add two zeros to the end
	sn *= 100
	s0 *= 100
	s1 *= 100

	tests := []struct {
		name    string
		serial  uint32
		want    uint32
		wantErr bool
		expErr  error
	}{
		{name: "no serial", serial: 0, want: uint32(s0), wantErr: false, expErr: nil},
		{name: "today 00", serial: uint32(s0), want: uint32(s0 + 1), wantErr: false, expErr: nil},
		{name: "today 09", serial: uint32(s0 + 9), want: uint32(s0 + 10), wantErr: false, expErr: nil},
		{name: "today 41", serial: uint32(s0 + 41), want: uint32(s0 + 42), wantErr: false, expErr: nil},
		{name: "today 90", serial: uint32(s0 + 90), want: uint32(s0 + 91), wantErr: false, expErr: nil},
		{name: "today 90", serial: uint32(s0 + 90), want: uint32(s0 + 91), wantErr: false, expErr: nil},
		{name: "today 99", serial: uint32(s0 + 99), want: 0, wantErr: true, expErr: errCounterExceed},
		{name: "yesterday 00", serial: uint32(sn), want: uint32(s0), wantErr: false, expErr: nil},
		{name: "yesterday 42", serial: uint32(sn + 42), want: uint32(s0), wantErr: false, expErr: nil},
		{name: "yesterday 99", serial: uint32(sn + 99), want: uint32(s0), wantErr: false, expErr: nil},
		{name: "tomorrow 00", serial: uint32(s1), want: 0, wantErr: true, expErr: errSerialInFuture},
		{name: "< 10 digits", serial: uint32(20200929), want: 0, wantErr: true, expErr: errWrongFormat},
		{name: "max uint32", serial: uint32(4294967295), want: 0, wantErr: true, expErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IncrementSerial(tt.serial)
			if (err != nil) != tt.wantErr {
				t.Errorf("IncrementSerial() error = %v", err)
				return
			} else if err != nil && tt.wantErr && tt.expErr != nil && tt.expErr != err {
				t.Errorf("IncrementSerial() error = %v, wantErr %v", err, tt.expErr)
				return
			}

			if got != tt.want {
				t.Errorf("IncrementSerial() got = %v, want %v", got, tt.want)
			} else {
				println(t.Name(), got)
			}
		})
	}
}
