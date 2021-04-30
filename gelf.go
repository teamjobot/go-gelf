package gelf

const (
	// LogFormat is known and expected log format callers should use so extra data can be parsed and imported into Seq.
	LogFormat = `%{id:03x}|%{module}|%{shortpkg}|%{shortfunc}|%{level:.4s}|%{message}`

	// v1 format:
	// LogFormat = `%{id:03x}|%{shortfunc}|%{level:.4s}|%{message}`
)

// Settings contains Gelf writer settings
type Settings struct {
	Address string
	AppName *string
	Env     *string

	// Additional fields to add to each log statement
	Meta map[string]string

	Version *string
}
