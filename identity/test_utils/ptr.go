package test_utils

import "encore.dev/types/uuid"

func StringPointer(val string) *string {
	return &val
}

func UUIDPointer(val uuid.UUID) *uuid.UUID {
	return &val
}
