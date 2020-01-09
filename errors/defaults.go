package errors

var (
	ErrConfiguration = New("bolt configuration error")
	ErrInvalidVersion = New("bolt version error")
	ErrInternal = New("bolt driver internal error")
	ErrClosed = New("resource is already closed")
	ErrPool = New("encountered error in connection pool")
	ErrConnection    = New("bolt connection error")
)
