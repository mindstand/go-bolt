package connection

import (
	"fmt"
	"github.com/mindstand/go-bolt/errors"
	"github.com/mindstand/go-bolt/log"
	"github.com/mindstand/go-bolt/structures/messages"
)

type boltTransaction struct {
	conn   *Connection
	closed bool
}

func (t *boltTransaction) Exec(query string, params QueryParams) (IResult, error) {
	return t.conn.Exec(query, params)
}

func (t *boltTransaction) ExecWithDb(query string, params QueryParams, db string) (IResult, error) {
	if !t.conn.boltProtocol.SupportsMultiDatabase() && db != "" {
		return nil, fmt.Errorf("bolt protocol version [%v] does not have multi database support", t.conn.protocolVersion)
	}

	success, err := t.conn.runQuery(query, params, db, true)
	if err != nil {
		return nil, err
	}

	return newBoltResult(success.Metadata), nil
}

func (t *boltTransaction) Query(query string, params QueryParams) (IRows, error) {
	return t.QueryWithDb(query, params, "")
}

func (t *boltTransaction) QueryWithDb(query string, params QueryParams, db string) (IRows, error) {
	if !t.conn.boltProtocol.SupportsMultiDatabase() && db != "" {
		return nil, fmt.Errorf("bolt protocol version [%v] does not have multi database support", t.conn.protocolVersion)
	}

	success, err := t.conn.runQuery(query, params, db, true)
	if err != nil {
		return nil, err
	}

	return newQueryRows(t.conn, success.Metadata, t.conn.boltProtocol.GetResultAvailableAfterKey(), t.conn.boltProtocol.GetResultConsumedAfterKey()), nil
}

func (t *boltTransaction) Commit() error {
	if t.closed {
		return errors.New("Transaction already closed")
	}

	if t.conn.openQuery {
		//?
	}

	msg := t.conn.boltProtocol.GetTxCommitMessage()

	_, isCommitType := msg.(messages.CommitMessage)

	// send commit
	err := t.conn.sendMessage(t.conn.boltProtocol.GetTxCommitMessage())
	if err != nil {
		return err
	}

	if !isCommitType {
		err = t.conn.sendMessage(t.conn.boltProtocol.GetPullAllMessage())
		if err != nil {
			return err
		}
	}

	runSucc, err := t.conn.consume()
	if err != nil {
		return err
	}

	var pullSucc interface{}

	if !isCommitType {
		pullSucc, err = t.conn.consume()
		if err != nil {
			return err
		}
	}

	// todo if the message is a commit message, do not pull

	success, ok := runSucc.(messages.SuccessMessage)
	if !ok {
		return errors.New("Unrecognized response type committing transaction: %#v", success)
	}

	log.Tracef("Got success message committing transaction: %#v", success)

	if !isCommitType {
		pull, ok := pullSucc.(messages.SuccessMessage)
		if !ok {
			return errors.New("Unrecognized response type pulling transaction:  %#v", pull)
		}

		log.Tracef("Got success message pulling transaction: %#v", pull)
	}

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

	msg := t.conn.boltProtocol.GetTxRollbackMessage()

	_, isRollbackType := msg.(messages.RollbackMessage)

	// send rollback
	err := t.conn.sendMessage(msg)
	if err != nil {
		return err
	}

	if !isRollbackType {
		err = t.conn.sendMessage(t.conn.boltProtocol.GetPullAllMessage())
		if err != nil {
			return err
		}
	}

	runSucc, err := t.conn.consume()
	if err != nil {
		return err
	}

	var pullSucc interface{}

	if !isRollbackType {
		pullSucc, err = t.conn.consume()
		if err != nil {
			return err
		}
	}

	success, ok := runSucc.(messages.SuccessMessage)
	if !ok {
		return errors.New("Unrecognized response type rolling back transaction: %#v", success)
	}

	log.Tracef("Got success message rolling back transaction: %#v", success)

	if !isRollbackType {
		pull, ok := pullSucc.(messages.SuccessMessage)
		if !ok {
			return errors.New("Unrecognized response type pulling transaction:  %#v", pull)
		}

		log.Tracef("Got success message pulling transaction: %#v", pull)
	}

	t.conn.transaction = nil
	t.closed = true
	return err
}

func (t *boltTransaction) IsClosed() bool {
	return t.closed
}
