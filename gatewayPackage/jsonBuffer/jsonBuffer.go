package jsonBuffer

import (
	"fmt"
	"strings"
	"time"
	"sync"  // mutex
)


/************************************************************************/
// JsonBuffer Queue chanel store data
// 
type JsonBuffer struct {
	S          strings.Builder
	Size       int
  MaxLength  int
	Mux        sync.Mutex
	Queue      chan string
}
/************************************************************************/

// AddKeyValue  add "key":"value", to S
func (m *JsonBuffer) AddKeyValue(key string, format string, value interface{}) {
	m.S.WriteString(key)
	m.S.WriteString(`:"`)
	fmt.Fprintf(&m.S, format, value)
	m.S.WriteString(`",`)
}
/************************************************************************/

// ObjectBegin begin object { "ts":"123456", "value":{
func (m *JsonBuffer) ObjectBegin(start string) (*strings.Builder) {
	m.Mux.Lock()

	if m.S.Len() == 0 {
		m.S.WriteString(start)  // start JsonBuffer
	}
	
	m.S.WriteString(`{`)
	m.AddKeyValue(`"ts"`, "%d", time.Now().UnixNano() / int64(time.Millisecond))
	m.S.WriteString(`"values":{`)
	return &m.S
}


/************************************************************************/
func (m *JsonBuffer) GetObject() (string) {
	m.S.WriteString(`"":""}}`)   // ObjectEnd without ","
	
	str := m.S.String()

	if m.Size < m.S.Len() {
		m.Size = m.S.Len()
	}

	m.S.Reset()
	m.S.Grow(m.Size)
	
	m.Mux.Unlock()
	return str
}

/************************************************************************/

// ObjectEnd add null object to S
func (m *JsonBuffer) ObjectEnd() {
	m.S.WriteString(`"":""}},`)
	m.Mux.Unlock()
}
/************************************************************************/

// ObjectEndCheckLength if len s > MaxLength push data to Queue
func (m *JsonBuffer) ObjectEndCheckLength() {
	m.ObjectEnd()
	
	if m.MaxLength < m.S.Len() {
		m.Queue <- m.GetString()
	}
}

/************************************************************************/
func (m *JsonBuffer) AddObject(obj string) {
	m.Mux.Lock()
	
	if m.S.Len() == 0 {
		m.S.WriteString(`[`)  // check start JsonBuffer
	}
	m.S.WriteString(obj)
	m.S.WriteString(",")
	
	m.Mux.Unlock()
	
	
	if m.MaxLength < m.S.Len() {
		m.Queue <- m.GetString()
	}
}

/************************************************************************/
func (m *JsonBuffer) GetString() (string) {
	m.Mux.Lock()
	
	m.S.WriteString(`{}]`)  // end JsonBuffer
	str  := m.S.String()

	if m.Size < m.S.Len() {
		m.Size = m.S.Len()
	}

	m.S.Reset()
	m.S.Grow(m.Size)
	
	m.Mux.Unlock()
	return str
}

/************************************************************************/
// func (m *JsonBuffer) PeakString() (string) {
	// m.Mux.Lock()

	// str := m.S.String() + `{}]`  // end JsonBuffer

	// m.Mux.Unlock()
	// return str
// }

/************************************************************************/
