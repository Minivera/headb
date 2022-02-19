package test_utils

import (
	"testing"

	"encore.dev/beta/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func CompareErrors(t *testing.T, expected, actual error) {
	convertedExpected, ok := expected.(*errs.Error)
	require.True(t, ok, "expected must be of type *errs.Error")

	convertedActual, ok := actual.(*errs.Error)
	require.True(t, ok, "actual must be of type *errs.Error")

	assert.Equal(t, convertedExpected.Message, convertedActual.Message, "Error messages should be equal")
	assert.Equal(t, convertedExpected.Code, convertedActual.Code, "Error code should be equal")
	assert.Equal(t, convertedExpected.Meta, convertedActual.Meta, "Error meta should be equal")
	assert.Equal(t, convertedExpected.Details, convertedActual.Details, "Error details should be equal")
}
