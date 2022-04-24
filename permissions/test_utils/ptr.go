package test_utils

import (
	"encore.dev/types/uuid"
)

func Int64Pointer(val int64) *int64 {
	return &val
}

func UUIDPointer(val uuid.UUID) *uuid.UUID {
	return &val
}
