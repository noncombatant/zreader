// Copyright 2024 Chris Palmer, https://noncombatant.org/
// SPDX-License-Identifier: Apache-2.0

package zreader

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	expected, e := os.ReadFile("testdata/zreader.txt")
	assert.Nil(t, e)

	type testCase struct {
		Type
		Pathname string
	}

	cases := []testCase{
		{Type: TypeNone, Pathname: "testdata/zreader.txt"},
		{Type: TypeBzip2, Pathname: "testdata/zreader.txt.bz2"},
		{Type: TypeGzip, Pathname: "testdata/zreader.txt.gz"},
		{Type: TypeLZ4Frame, Pathname: "testdata/zreader.txt.lz4"},
		{Type: TypeXZ, Pathname: "testdata/zreader.txt.xz"},
		{Type: TypeZlib, Pathname: "testdata/zreader.txt.zlib"},
	}

	for _, c := range cases {
		t.Run(c.Pathname, func(t *testing.T) {
			reader, e := Open(c.Pathname)
			assert.Nil(t, e)
			assert.Equal(t, c.Type, reader.Type)
			bytes, e := io.ReadAll(reader)
			assert.Nil(t, e)
			assert.Equal(t, expected, bytes)
			reader.Close()
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
	reader.Close()
}
