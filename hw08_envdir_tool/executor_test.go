package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_prepareCmd(t *testing.T) {
	const (
		toRemove    = "HW08-REMOVE"
		toRemoveVal = "MUST BE DELETED"

		toOverwrite    = "HW08-OVERWRITE"
		toOverwriteOld = "MUST BE OVERWRITTEN"
		toOverwriteNew = "NEW VALUE"

		newVar    = "HW08-NEW"
		newVarVal = "NEW VAR"
	)

	os.Setenv(toRemove, toRemoveVal)
	os.Setenv(toOverwrite, toOverwriteOld)

	t.Run("set vars", func(t *testing.T) {
		execCmd := prepareCmd(
			[]string{"test"},
			Environment{
				toRemove:    {toRemoveVal, true},
				toOverwrite: {toOverwriteNew, false},
				newVar:      {newVarVal, false},
			},
		)

		require.Contains(
			t,
			execCmd.Env,
			toOverwrite+"="+toOverwriteOld,
			"must contains "+toOverwriteOld)

		require.Contains(
			t,
			execCmd.Env,
			toOverwrite+"="+toOverwriteNew,
			"must contains "+toOverwriteNew)

		require.Contains(
			t,
			execCmd.Env,
			newVar+"="+newVarVal,
			"must contains "+newVarVal)

		require.NotContains(
			t,
			execCmd.Env,
			toRemove+"="+toRemoveVal,
			"must not contains "+toRemove)
	})

	t.Run("set no vars", func(t *testing.T) {
		execCmd := prepareCmd(
			[]string{"test"},
			Environment{},
		)

		require.NotContains(
			t,
			execCmd.Env,
			toOverwrite+"="+toOverwriteNew,
			"must not contains "+toOverwriteNew)

		require.NotContains(
			t,
			execCmd.Env,
			newVar+"="+newVarVal,
			"must not contains "+newVarVal)

		require.Contains(
			t,
			execCmd.Env,
			toRemove+"="+toRemoveVal,
			"must contains "+toRemove)
	})
}
