package main

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

func genUUID() string {
	id := uuid.NewV4()
	return strings.ReplaceAll(id.String(), "-", "")
}
