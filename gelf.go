package gelf

const (
	// LogFormat is known and expected log format callers should use so extra data can be parsed and imported into Seq.
	LogFormat = `%{id:03x}|%{shortfunc}|%{level:.4s}|%{message}`
)

// Settings contains Gelf writer settings
type Settings struct {
	Address string
	Env     *string
	AppName *string
}
