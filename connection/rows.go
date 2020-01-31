package connection

import (
	"github.com/mindstand/go-bolt/errors"
	"github.com/mindstand/go-bolt/log"
	"github.com/mindstand/go-bolt/structures/messages"
	"io"
)

type boltRows struct {
	metadata        map[string]interface{}
	conn            *Connection
	closed          bool
	consumed        bool
	finishedConsume bool
}

func newRows(conn *Connection, metadata map[string]interface{}) *boltRows {
	return &boltRows{
		conn:     conn,
		metadata: metadata,
	}
}

func newQueryRows(conn *Connection, metadata map[string]interface{}) *boltRows {
	rows := newRows(conn, metadata)
	rows.consumed = true // Already consumed from pipeline with PULL_ALL
	return rows
}

func (b *boltRows) Columns() []string {
	fieldsInt, ok := b.metadata["fields"]
	if !ok {
		return []string{}
	}

	fields, ok := fieldsInt.([]interface{})
	if !ok {
		log.Errorf("Unrecognized fields from success message: %#v", fieldsInt)
		return []string{}
	}

	fieldsStr := make([]string, len(fields))
	for i, f := range fields {
		if fieldsStr[i], ok = f.(string); !ok {
			log.Errorf("Unrecognized fields from success message: %#v", fieldsInt)
			return []string{}
		}
	}
	return fieldsStr
}

func (b *boltRows) Metadata() map[string]interface{} {
	return b.metadata
}

func (b *boltRows) Close() error {
	if b.closed {
		return nil
	}

	if !b.consumed {
		// Discard all messages if not consumed
		respInt, err := b.conn.sendMessageConsume(b.conn.boltProtocol.GetDiscardAllMessage())
		if err != nil {
			return errors.Wrap(err, "An error occurred discarding messages on row close")
		}

		switch resp := respInt.(type) {
		case messages.SuccessMessage:
			log.Infof("Got success message: %#v", resp)
		default:
			return errors.New("Unrecognized response type discarding all rows: Value: %#v", resp)
		}

	} else if !b.finishedConsume {
		// Clear out all unconsumed messages if we
		// never finished consuming them.
		_, err := b.conn.consume()
		if err != nil {
			return errors.Wrap(err, "An error occurred clearing out unconsumed stream")
		}
	}

	b.closed = true

	return nil
}

func (b *boltRows) Next() ([]interface{}, map[string]interface{}, error) {
	if b.closed {
		return nil, nil, errors.New("Rows are already closed")
	}

	if !b.consumed {
		b.consumed = true
		if err := b.conn.sendMessage(b.conn.boltProtocol.GetPullAllMessage()); err != nil {
			b.finishedConsume = true
			return nil, nil, err
		}
	}

	respInt, err := b.conn.consume()
	if err != nil {
		return nil, nil, err
	}

	switch resp := respInt.(type) {
	case messages.SuccessMessage:
		log.Infof("Got success message: %#v", resp)
		b.finishedConsume = true
		return nil, resp.Metadata, io.EOF
	case messages.RecordMessage:
		log.Infof("Got record message: %#v", resp)
		return resp.Fields, nil, nil
	default:
		return nil, nil, errors.New("Unrecognized response type getting next runQuery row: %#v", resp)
	}
}

func (b *boltRows) All() ([][]interface{}, map[string]interface{}, error) {
	output := [][]interface{}{}
	for {
		row, metadata, err := b.Next()
		if err != nil || row == nil {
			if err == io.EOF {
				return output, metadata, nil
			}
			return output, metadata, err
		}
		output = append(output, row)
	}
}
