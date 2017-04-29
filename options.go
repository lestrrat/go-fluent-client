package fluent

import "time"

type option struct {
	name  string
	value interface{}
}

func (o *option) Name() string {
	return o.name
}

func (o *option) Value() interface{} {
	return o.value
}

func WithNetwork(s string) Option {
	const name = "network"
	return &option{
		name:  name,
		value: s,
	}
}

func WithAddress(s string) Option {
	const name = "address"
	return &option{
		name:  name,
		value: s,
	}
}

func WithTimestamp(t time.Time) Option {
	const name = "timestamp"
	return &option{
		name:  name,
		value: t,
	}
}

func WithJSONMarshaler() Option {
	const name = "marshaler"
	return &option{
		name:  name,
		value: marshalFunc(jsonMarshal),
	}
}

func WithMsgpackMarshaler() Option {
	const name = "marshaler"
	return &option{
		name:  name,
		value: marshalFunc(msgpackMarshal),
	}
}

func WithTagPrefix(s string) Option {
	const name = "tag_prefix"
	return &option{
		name:  name,
		value: s,
	}
}

func WithSyncAppend(b bool) Option {
	const name = "sync_append"
	return &option{
		name:  name,
		value: b,
	}
}

func WithBufferLimit(v interface{}) Option {
	const name = "buffer_limit"
	return &option{
		name:  name,
		value: v,
	}
}
