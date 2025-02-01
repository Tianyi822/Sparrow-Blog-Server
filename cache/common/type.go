package common

type ValueType uint8

const (
	INT ValueType = iota
	UINT
	FLOAT
	STRING
	OBJ
)

const (
	SET     = "SET"
	DELETE  = "DELETE"
	INCR    = "INCR"
	CLEANUP = "CLEANUP"
)
