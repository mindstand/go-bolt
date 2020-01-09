package goBolt

import (
	"github.com/mindstand/go-bolt/errors"
	"github.com/mindstand/go-bolt/log"
	"github.com/mindstand/go-bolt/structures/messages"
)

// Tx represents a transaction
type Tx interface {
	// Commit commits the transaction
	Commit() error
	// Rollback rolls back the transaction
	Rollback() error
}

type boltTx struct {
	conn   IConnection
	closed bool
}

func newTx(conn IConnection) *boltTx {
	return &boltTx{
		conn: conn,
	}
}

// Commit commits and closes the transaction
func (t *boltTx) Commit() error {
	if t.closed {
		return errors.New("Transaction already closed")
	}
	if t.conn.getStatement() != nil {
		if err := t.conn.getStatement().Close(); err != nil {
			return errors.Wrap(err, "An error occurred closing open rows in transaction Commit")
		}
	}

	successInt, pullInt, err := t.conn.sendRunPullAllConsumeSingle("COMMIT", nil)
	if err != nil {
		return errors.Wrap(err, "An error occurred committing transaction")
	}

	success, ok := successInt.(messages.SuccessMessage)
	if !ok {
		return errors.New("Unrecognized response type committing transaction: %#v", success)
	}

	log.Infof("Got success message committing transaction: %#v", success)

	pull, ok := pullInt.(messages.SuccessMessage)
	if !ok {
		return errors.New("Unrecognized response type pulling transaction:  %#v", pull)
	}

	log.Infof("Got success message pulling transaction: %#v", pull)

	t.conn.setTx(nil)
	t.closed = true
	return err
}

// Rollback rolls back and closes the transaction
func (t *boltTx) Rollback() error {
	if t.closed {
		return errors.New("Transaction already closed")
	}
	if t.conn.getStatement() != nil {
		if err := t.conn.getStatement().Close(); err != nil {
			return errors.Wrap(err, "An error occurred closing open rows in transaction Rollback")
		}
	}

	successInt, pullInt, err := t.conn.sendRunPullAllConsumeSingle("ROLLBACK", nil)
	if err != nil {
		return errors.Wrap(err, "An error occurred rolling back transaction")
	}

	success, ok := successInt.(messages.SuccessMessage)
	if !ok {
		return errors.New("Unrecognized response type rolling back transaction: %#v", success)
	}

	log.Infof("Got success message rolling back transaction: %#v", success)

	pull, ok := pullInt.(messages.SuccessMessage)
	if !ok {
		return errors.New("Unrecognized response type pulling transaction: %#v", pull)
	}

	log.Infof("Got success message pulling transaction: %#v", pull)

	t.conn.setTx(nil)
	t.closed = true
	return err
}
