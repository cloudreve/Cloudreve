package balancer

import "errors"

var (
	ErrInputNotSlice   = errors.New("Input value is not silice")
	ErrNoAvaliableNode = errors.New("No nodes avaliable")
)
