package connection

import (
	"github.com/mindstand/go-bolt/errors"
	"github.com/mindstand/go-bolt/log"
	"github.com/mindstand/go-bolt/structures/messages"
)

type boltTransaction struct {
	conn *Connection
	closed bool
}

func (t *boltTransaction) Exec(query string, params QueryParams) (IResult, error) {
	return t.conn.Exec(query, params)
}

func (t *boltTransaction) Query(query string, params QueryParams) (IRows, error) {
	return t.conn.Query(query, params)
}

func (t *boltTransaction) Commit() error {
	if t.closed {
		return errors.New("Transaction already closed")
	}

	if t.conn.openQuery {
		//?
	}

	// send commit
	err := t.conn.sendMessage(t.conn.boltProtocol.GetTxCommitMessage())
	if err != nil {
		return err
	}

	err = t.conn.sendMessage(messages.PullAllMessage{})
	if err != nil {
		return err
	}

	runSucc, err := t.conn.consume()
	if err != nil {
		return err
	}

	pullSucc, err := t.conn.consume()
	if err != nil {
		return err
	}

	success, ok := runSucc.(messages.SuccessMessage)
	if !ok {
		return errors.New("Unrecognized response type committing transaction: %#v", success)
	}

	log.Infof("Got success message committing transaction: %#v", success)

	pull, ok := pullSucc.(messages.SuccessMessage)
	if !ok {
		return errors.New("Unrecognized response type pulling transaction:  %#v", pull)
	}

	log.Infof("Got success message pulling transaction: %#v", pull)

	t.conn.transaction = nil
	t.closed = true
	return err
}

func (t *boltTransaction) Rollback() error {
	if t.closed {
		return errors.New("Transaction already closed")
	}

	if t.conn.openQuery {
		//?
	}

	// send rollback
	err := t.conn.sendMessage(t.conn.boltProtocol.GetTxRollbackMessage())
	if err != nil {
		return err
	}

	err = t.conn.sendMessage(messages.PullAllMessage{})
	if err != nil {
		return err
	}

	runSucc, err := t.conn.consume()
	if err != nil {
		return err
	}

	pullSucc, err := t.conn.consume()
	if err != nil {
		return err
	}

	success, ok := runSucc.(messages.SuccessMessage)
	if !ok {
		return errors.New("Unrecognized response type rolling back transaction: %#v", success)
	}

	log.Infof("Got success message rolling back transaction: %#v", success)

	pull, ok := pullSucc.(messages.SuccessMessage)
	if !ok {
		return errors.New("Unrecognized response type pulling transaction:  %#v", pull)
	}

	log.Infof("Got success message pulling transaction: %#v", pull)

	t.conn.transaction = nil
	t.closed = true
	return err
}

func (t *boltTransaction) IsClosed() bool {
	return t.closed
}
