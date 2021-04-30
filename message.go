package gelf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

// Message represents the contents of the GELF message.  It is gzipped
// before sending.
type Message struct {
	Version  string                 `json:"version"`
	Host     string                 `json:"host"`
	Short    string                 `json:"short_message"`
	Full     string                 `json:"full_message,omitempty"`
	TimeUnix float64                `json:"timestamp"`
	Level    int32                  `json:"level,omitempty"`
	Facility string                 `json:"facility,omitempty"`
	Extra    map[string]interface{} `json:"-"`
	RawExtra json.RawMessage        `json:"-"`
}

// Syslog severity levels
const (
	LOG_EMERG   = 0
	LOG_ALERT   = 1
	LOG_CRIT    = 2
	LOG_ERR     = 3
	LOG_WARNING = 4
	LOG_NOTICE  = 5
	LOG_INFO    = 6
	LOG_DEBUG   = 7
)

var (
	levelMap = map[string]int32{
		"DEBU": LOG_DEBUG,
		"INFO": LOG_INFO,
		"NOTI": LOG_NOTICE,
		"WARN": LOG_WARNING,
		"ERRO": LOG_ERR,
		"CRIT": LOG_CRIT,
	}
)

func (m *Message) MarshalJSONBuf(buf *bytes.Buffer) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	// write up until the final }
	if _, err = buf.Write(b[:len(b)-1]); err != nil {
		return err
	}
	if len(m.Extra) > 0 {
		eb, err := json.Marshal(m.Extra)
		if err != nil {
			return err
		}
		// merge serialized message + serialized extra map
		if err = buf.WriteByte(','); err != nil {
			return err
		}
		// write serialized extra bytes, without enclosing quotes
		if _, err = buf.Write(eb[1 : len(eb)-1]); err != nil {
			return err
		}
	}

	if len(m.RawExtra) > 0 {
		if err := buf.WriteByte(','); err != nil {
			return err
		}

		// write serialized extra bytes, without enclosing quotes
		if _, err = buf.Write(m.RawExtra[1 : len(m.RawExtra)-1]); err != nil {
			return err
		}
	}

	// write final closing quotes
	return buf.WriteByte('}')
}

func (m *Message) UnmarshalJSON(data []byte) error {
	i := make(map[string]interface{}, 16)
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	for k, v := range i {
		if k[0] == '_' {
			if m.Extra == nil {
				m.Extra = make(map[string]interface{}, 1)
			}
			m.Extra[k] = v
			continue
		}

		ok := true
		switch k {
		case "version":
			m.Version, ok = v.(string)
		case "host":
			m.Host, ok = v.(string)
		case "short_message":
			m.Short, ok = v.(string)
		case "full_message":
			m.Full, ok = v.(string)
		case "timestamp":
			m.TimeUnix, ok = v.(float64)
		case "level":
			var level float64
			level, ok = v.(float64)
			m.Level = int32(level)
		case "facility":
			m.Facility, ok = v.(string)
		}

		if !ok {
			return fmt.Errorf("invalid type for field %s", k)
		}
	}
	return nil
}

func (m *Message) toBytes(buf *bytes.Buffer) (messageBytes []byte, err error) {
	if err = m.MarshalJSONBuf(buf); err != nil {
		return nil, err
	}
	messageBytes = buf.Bytes()
	return messageBytes, nil
}

type LogWrite struct {
	AppName     *string
	Environment *string
	Facility    string
	File        string
	HostName    string
	Line        int
	Payload     []byte
	Version     *string
	// Additional fields to add to each log statement
	Meta map[string]string
}

// %{id:03x}|%{module}|%{shortpkg}|%{shortfunc}|%{level:.4s}|%{message}
const (
	partId     = 0
	partModule = 1
	partPkg    = 2
	partFunc   = 3
	partLevel  = 4
	partMsg    = 5
)

type Parts struct {
	Id     string
	Module string
	Pkg    string
	Func   string
	Level  string
	Msg    []byte
}

func getMessageParts(w LogWrite) Parts {
	result := Parts{}
	parts := strings.Split(string(w.Payload), "|")

	// This is a safety check for transitioning between format changes in event caller backend format isn't in sync
	// with the latest extended format here.
	if len(parts) == 6 {
		result.Id = parts[partId]
		result.Module = parts[partModule]
		result.Pkg = parts[partPkg]
		result.Func = parts[partFunc]
		result.Level = parts[partLevel]
		result.Msg = bytes.TrimSpace([]byte(parts[partMsg]))
	} else if len(parts) == 4 {
		result.Id = parts[0]
		result.Func = parts[1]
		result.Level = parts[2]
		result.Msg = bytes.TrimSpace([]byte(parts[3]))
	}

	return result
}

func constructMessage(w LogWrite) (m *Message) {
	// See gelf.LogFormat and constants above
	parts := getMessageParts(w)

	// If there are newlines in the message, use the first line
	// for the short message and set the full message to the original input.
	// If the input has no newlines, stick the whole thing in Short.
	short := parts.Msg
	full := []byte("")

	if i := bytes.IndexRune(parts.Msg, '\n'); i > 0 {
		short = parts.Msg[:i]
		full = parts.Msg
	}

	appName := w.AppName
	env := w.Environment

	// Allows caller to override app name otherwise default to executable name
	if appName == nil {
		appName = &w.Facility
	}

	if env == nil {
		s := ""
		env = &s
	}

	version := w.Version

	if version == nil {
		s := ""
		version = &s
	}

	file := strings.ReplaceAll(path.Base(w.File), ".go", "")

	// https://docs.graylog.org/en/4.0/pages/gelf.html
	m = &Message{
		Version:  "1.1",
		Host:     w.HostName,
		Short:    string(short),
		Full:     string(full),
		TimeUnix: float64(time.Now().UnixNano()) / float64(time.Second),
		Level:    levelMap[parts.Level],

		// Facility is deprecated
		//Facility: w.Facility,
		Extra: map[string]interface{}{
			"_app":      appName,
			"_env":      env,
			"_filename": w.File,
			"_file":     file,
			"_function": parts.Func,
			"_id":       parts.Id,
			"_line":     w.Line,
			"_module":   parts.Module,
			"_pid":      os.Getpid(),
			"_pkg":      parts.Pkg,
			"_version":  version,
		},
	}

	for k, v := range w.Meta {
		m.Extra[k] = v
	}

	return m
}
