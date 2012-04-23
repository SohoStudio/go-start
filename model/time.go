package model

import "time"

const unixDateSec int64 = (1969*365 + 1969/4 - 1969/100 + 1969/400) * (24 * 60 * 60)

// Time in milliseconds since January 1, year 1 00:00:00 UTC.
// Time values are always in UTC.
type Time int64

func (self *Time) String() string {
	return self.Time().Format(time.RFC3339Nano)
}

func (self *Time) Parse(s string) error {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err == nil {
		self.SetTime(t)
	}
	return err
}

func (self *Time) Time() time.Time {
	unixsec := int64(*self)/1000 - unixDateSec
	unixmsec := int64(*self) % 1000
	return time.Unix(unixsec, unixmsec*1e6)
}

func (self *Time) SetTime(t time.Time) {
	unixsec := t.Unix()
	unixmsec := int64(t.Nanosecond()) / 1e6
	*self = Time((unixsec+unixDateSec)*1000 + unixmsec)
}

func (self *Time) IsEmpty() string {
	return *self == 0
}
