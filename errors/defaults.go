package errors

var (
	ErrConfiguration = New("bolt configuration error")
	ErrConnection    = New("bolt connection error")
)
