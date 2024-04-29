// Copyright 2022 Chris Palmer, https://noncombatant.org/
// SPDX-License-Identifier: Apache-2.0

package zreader

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZReader(t *testing.T) {
	expected, e := os.ReadFile("testdata/zreader.txt")
	assert.Nil(t, e)

	zfiles := []string{
		"testdata/zreader.txt",
		"testdata/zreader.txt.bz2",
		"testdata/zreader.txt.gz",
		"testdata/zreader.txt.zlib",
	}
	for _, zf := range zfiles {
		t.Run(zf, func(t *testing.T) {
			reader, e := Open(zf)
			assert.Nil(t, e)
			bytes, e := io.ReadAll(reader)
			assert.Nil(t, e)
			assert.Equal(t, expected, bytes)
		})
	}
}

func TestShort(t *testing.T) {
	expected, e := os.ReadFile("testdata/short.txt")
	assert.Nil(t, e)

	reader, e := Open("testdata/short.txt")
	assert.Nil(t, e)
	bytes, e := io.ReadAll(reader)
	assert.Nil(t, e)
	assert.Equal(t, expected, bytes)
}
