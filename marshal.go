package fluent

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"

	pdebug "github.com/lestrrat/go-pdebug"
	"github.com/pkg/errors"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

var _ = codes.Ext8

type marshalFunc func(*Message) ([]byte, error)

func (f marshalFunc) Marshal(msg *Message) ([]byte, error) {
	return f(msg)
}

// UnmarshalJSON deserializes from a JSON buffer and populates
// a Message struct appropriately
func (m *Message) UnmarshalJSON(buf []byte) error {
	var l []json.RawMessage
	if err := json.Unmarshal(buf, &l); err != nil {
		return errors.Wrap(err, `failed to unmarshal JSON: expected array`)
	}

	var tag string
	if err := json.Unmarshal(l[0], &tag); err != nil {
		return errors.Wrap(err, `failed to unmarshal JSON: expected tag`)
	}

	var t int64
	if err := json.Unmarshal(l[1], &t); err != nil {
		return errors.Wrap(err, `failed to unmarshal JSON: expected timestamp`)
	}

	var r interface{}
	if err := json.Unmarshal(l[2], &r); err != nil {
		return errors.Wrap(err, `failed to unmarshal JSON: expected record`)
	}

	var o interface{}
	if err := json.Unmarshal(l[3], &o); err != nil {
		return errors.Wrap(err, `failed to unmarshal JSON: expected options`)
	}

	*m = Message{
		Tag:    tag,
		Time:   EventTime{Time: time.Unix(t, 0)},
		Record: r,
		Option: o,
	}

	return nil
}

// MarshalJSON serializes a Message to JSON format
func (m *Message) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteByte('[')
	buf.WriteString(strconv.Quote(m.Tag))
	buf.WriteByte(',')
	buf.WriteString(strconv.FormatInt(m.Time.Unix(), 10))
	buf.WriteByte(',')

	// XXX Encoder appends a silly newline at the end, so use
	// json.Marshal instead
	data, err := json.Marshal(m.Record)
	if err != nil {
		return nil, errors.Wrap(err, `failed to encode record`)
	}
	buf.Write(data)
	buf.WriteByte(',')

	data, err = json.Marshal(m.Option)
	if err != nil {
		return nil, errors.Wrap(err, `failed to encode option`)
	}
	buf.Write(data)
	buf.WriteByte(']')

	if pdebug.Enabled {
		pdebug.Printf("message marshaled to: %s", strconv.Quote(buf.String()))
	}

	return buf.Bytes(), nil
}

// EncodeMsgpack serializes a Message to msgpack format
func (m *Message) EncodeMsgpack(enc *msgpack.Encoder) error {
	if err := enc.EncodeArrayLen(4); err != nil {
		return errors.Wrap(err, `failed to encode msgpack: array len`)
	}

	if err := enc.EncodeString(m.Tag); err != nil {
		return errors.Wrap(err, `failed to encode msgpack: tag`)
	}

	if m.subsecond {
		if err := enc.Encode(&m.Time); err != nil {
			return errors.Wrap(err, `failed to encode msgpack: time (EventTime)`)
		}
	} else {
		if err := enc.EncodeInt64(m.Time.Unix()); err != nil {
			return errors.Wrap(err, `failed to encode msgpack: time`)
		}
	}

	if err := enc.Encode(m.Record); err != nil {
		return errors.Wrap(err, `failed to encode msgpack: record`)
	}
	if err := enc.Encode(m.Option); err != nil {
		return errors.Wrap(err, `failed to encode msgpack: option`)
	}
	return nil
}

// DecodeMsgpack deserializes from a msgpack buffer and populates
// a Message struct appropriately
func (m *Message) DecodeMsgpack(dec *msgpack.Decoder) error {
	var code byte
	var err error

	code, err = dec.PeekCode()
	if err != nil {
		return errors.Wrap(err, `failed to peek code`)
	}

	if !codes.IsFixedArray(code) {
		dec.Skip()
		return errors.Wrap(err, `expected array`)
	}

	l, err := dec.DecodeArrayLen()
	if err != nil {
		return errors.Wrap(err, `failed to decode msgpack: array len`)
	}

	if l != 4 {
		return errors.Errorf(`expected tuple with 4 elements, got %d`, l)
	}

	m.Tag, err = dec.DecodeString()
	if err != nil {
		return errors.Wrap(err, `failed to decode msgpack: tag`)
	}

	code, err = dec.PeekCode()
	if err != nil {
		return errors.Wrap(err, `failed to peek code`)
	}

	switch code {
	case codes.Ext8, codes.FixExt8:
		if err := dec.Decode(&m.Time); err != nil {
			return errors.Wrap(err, `failed to decode msgpack: time (EventTime)`)
		}
	case codes.Uint32, codes.Int64:
		t, err := dec.DecodeInt64()
		if err != nil {
			return errors.Wrap(err, `failed to decode msgpack: time`)
		}
		m.Time.Time = time.Unix(t, 0)
	default:
		return errors.Errorf(`expected ext8, fixedext8, or int64 for time: %b`, code)
	}

	m.Record, err = dec.DecodeInterface()
	if err != nil {
		return errors.Wrap(err, `failed to decode msgpack: record`)
	}

	m.Option, err = dec.DecodeInterface()
	if err != nil {
		return errors.Wrap(err, `failed to decode msgpack: option`)
	}

	return nil
}

func msgpackMarshal(m *Message) ([]byte, error) {
	var buf bytes.Buffer
	if err := msgpack.NewEncoder(&buf).Encode(m); err != nil {
		return nil, errors.Wrap(err, `failed to encode msgpack`)
	}
	return buf.Bytes(), nil
}

func jsonMarshal(m *Message) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(m); err != nil {
		return nil, errors.Wrap(err, `failed to encode json`)
	}
	buf.Truncate(buf.Len() - 1) // remove new line
	return buf.Bytes(), nil
}
