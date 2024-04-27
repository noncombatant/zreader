// Copyright 2022 Chris Palmer, https://noncombatant.org/
// SPDX-License-Identifier: Apache-2.0

// Package zreader provides an [io.ReadCloser] for a variety of compression
// formats.
package zreader

import (
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
)

// TODO: https://github.com/ulikunitz/xz
//
// TODO: Support all of: compress/{bzip2,flate,gzip,lzw,zlib}.
//
// TODO: Consider using more or all of the klauspost implementations.

type zType int

const (
	zNone zType = iota
	zBzip2
	zGzip
	zZip
	zZstd
)

// ZReader is an [io.ReadCloser] that reads compressed files.
//
// It currently supports bzip2, gzip, and zstd.
type ZReader struct {
	file *os.File
	zType
	bzip2Reader io.Reader
	gzipReader  *gzip.Reader
	zstdDecoder *zstd.Decoder
}

// Open opens pathname and returns an appropriate ZReader. It selects a
// decompressor based on pathname's file extension. If it does not have a
// decompressor to match the extension, subsequent calls to [Read] will return
// the raw bytes of the file. (That might, or might not, be what you want.)
func Open(pathname string) (*ZReader, error) {
	file, e := os.Open(pathname)
	if e != nil {
		return nil, e
	}

	reader := &ZReader{file: file}

	fileType := strings.ToLower(filepath.Ext(pathname))
	switch fileType {
	case ".bz2":
		reader.zType = zBzip2
		reader.bzip2Reader = bzip2.NewReader(reader.file)
	case ".gz":
		reader.zType = zGzip
		r, e := gzip.NewReader(reader.file)
		if e != nil {
			reader.file.Close()
			return nil, e
		}
		reader.gzipReader = r
	case ".zip":
		reader.zType = zZip
		// TODO
	case ".zstd":
		reader.zType = zZstd
		d, e := zstd.NewReader(reader.file)
		if e != nil {
			reader.file.Close()
			return nil, e
		}
		reader.zstdDecoder = d
	default:
		reader.zType = zNone
	}
	return reader, nil
}

// Read reads up to len(bytes) from the File and stores them in bytes. It
// returns the number of bytes read and any error encountered. At end of file,
// Read returns 0, io.EOF.
func (zr *ZReader) Read(bytes []byte) (n int, err error) {
	switch zr.zType {
	case zBzip2:
		return zr.bzip2Reader.Read(bytes)
	case zGzip:
		return zr.gzipReader.Read(bytes)
	case zZip:
		// TODO
	case zZstd:
		return zr.zstdDecoder.Read(bytes)
	}
	return zr.file.Read(bytes)
}

// Close closes zr, rendering it unusable for I/O. Close will return an error if
// it has already been called.
func (zr *ZReader) Close() error {
	switch zr.zType {
	case zNone:
	case zBzip2:
	case zGzip:
		zr.gzipReader.Close()
	case zZip:
	case zZstd:
	}
	return zr.file.Close()
}
