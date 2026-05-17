package config

import (
	"github.com/zetkey/waka3x/lib"
)

type SharedDataKey string

const (
	MiddlewareKeyPrincipal   = SharedDataKey("principal")
	MiddlewareKeyPrincipalId = SharedDataKey("principal_identity")
)

type SharedData struct {
	*lib.ConcurrentMap[SharedDataKey, interface{}]
}

func NewSharedData() *SharedData {
	return &SharedData{lib.NewConcurrentMap[SharedDataKey, interface{}]()}
}
