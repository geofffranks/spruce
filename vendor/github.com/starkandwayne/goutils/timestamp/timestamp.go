package timestamp

import (
    "fmt"
    "time"
)

const Format = "2006-01-02 15:04:05"

type Timestamp struct {
    when time.Time
}

func Now() Timestamp {
    return Timestamp{when: time.Now()}
}

func NewTimestamp(t time.Time) Timestamp {
    return Timestamp{when: t}
}

func (t Timestamp) Add(d time.Duration) Timestamp {
    return Timestamp{when: t.when.Add(d)}
}

func (t Timestamp) After(other Timestamp) bool {
    return t.when.After(other.when)
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
    if t.when.IsZero() {
        return []byte("\"\""), nil
    }
    return []byte(fmt.Sprintf("\"%s\"", t.when.Format(Format))), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) (err error) {
    if string(b) == "\"\"" {
        return
    }
    t.when, err = time.Parse(Format, string(b[1:len(b)-1]))
    return
}

func (t Timestamp) Format(layout string) string {
    return t.when.Format(layout)
}

func (t Timestamp) IsZero() bool {
    return t.when.IsZero()
}

func (t Timestamp) Time() time.Time {
    return t.when
}
