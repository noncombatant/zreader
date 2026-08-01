// Copyright 2024 Chris Palmer, https://noncombatant.org/
// SPDX-License-Identifier: Apache-2.0

// Package zreader provides an [io.ReadCloser] for a variety of compression
// formats.
package zreader

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"io"
	"os"

	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz"
)

// TODO: Support compress/{flate,lzw}.
//
// TODO: Consider using more or all of the klauspost implementations.
//
// TODO: Add an OpenWithType.
//
// TODO: Make ZType public, and test for it being correct.

type zType int

// TODO: Rename these TypeFoo.
const (
	zNone zType = iota
	zBzip2
	zGzip
	zZip
	zZstd
	zZlib
	zLZ4Frame
	zLZ4Block // TODO: Can't use a magic byte approach for this
	zXZ
)

// ZReader is an [io.ReadCloser] that decompresses and reads compressed files.
//
// It currently supports bzip2, gzip, and zstd.
type ZReader struct {
	zType
	decompressor io.ReadCloser
}

// Open opens pathname and returns an appropriate ZReader. See [NewReader] for
// guidance on its behavior.
func Open(pathname string) (*ZReader, error) {
	file, e := os.Open(pathname)
	if e != nil {
		return nil, e
	}

	r, e := NewReader(file)
	if e != nil {
		file.Close()
		return nil, e
	}
	return r, nil
}

// NewReader returns a ZReader for the given io.ReadCloser. It selects a
// decompressor based on the first few bytes of data. If it does not have a
// decompressor to match the bytes, subsequent calls to [Read] will return the
// raw bytes of the reader. (That might, or might not, be what you want.)
func NewReader(r io.Reader) (*ZReader, error) {
	return fromBufferedReader(bufio.NewReader(r))
}

// TODO: FromReader, FromFile, FromReadCloser, FromBufferedReader; and then use
// them as appropriate

func fromBufferedReader(uncompressed *bufio.Reader) (*ZReader, error) {
	magicBlock, e := uncompressed.Peek(magicBytePrefixSize)
	if e != nil {
		if errors.Is(e, io.EOF) {
			return &ZReader{zType: zNone, decompressor: io.NopCloser(uncompressed)}, nil
		}
		return nil, e
	}

	switch zTypeFromBytes(magicBlock) {
	case zBzip2:
		return &ZReader{zType: zBzip2, decompressor: io.NopCloser(bzip2.NewReader(uncompressed))}, nil
	case zGzip:
		r, e := gzip.NewReader(uncompressed)
		if e != nil {
			return nil, e
		}

		return &ZReader{zType: zGzip, decompressor: r}, nil
	case zZip:
		// TODO
		return &ZReader{zType: zZip, decompressor: io.NopCloser(uncompressed)}, nil
	case zZstd:
		d, e := zstd.NewReader(uncompressed)
		if e != nil {
			return nil, e
		}
		return &ZReader{zType: zZstd, decompressor: io.NopCloser(d)}, nil
	case zLZ4Frame:
		return &ZReader{zType: zLZ4Frame, decompressor: io.NopCloser(lz4.NewReader(uncompressed))}, nil
	case zXZ:
		r, e := xz.NewReader(uncompressed)
		if e != nil {
			return nil, e
		}
		return &ZReader{zType: zXZ, decompressor: io.NopCloser(r)}, nil
	case zZlib:
		r, e := zlib.NewReader(uncompressed)
		if e != nil {
			return nil, e
		}
		return &ZReader{zType: zZlib, decompressor: r}, nil
	default:
		return &ZReader{zType: zNone, decompressor: io.NopCloser(uncompressed)}, nil
	}
}

// Read reads from the appropriate decompressor.
func (r *ZReader) Read(p []byte) (int, error) {
	return r.decompressor.Read(p)
}

// Close closes the ZReader, and will close the underlying reader if it has one.
func (r *ZReader) Close() error {
	if e := r.decompressor.Close(); e != nil {
		return e
	}
	return nil
}

const magicBytePrefixSize = 6

var (
	magicBzip2             = []byte{0x42, 0x5a, 0x68}
	magicGzip              = []byte{0x1f, 0x8b}
	magicZip               = []byte{0x50, 0x4b, 0x03, 0x04}
	magicZipEmptyArchive   = []byte{0x50, 0x4b, 0x05, 0x06}
	magicZipSpannedArchive = []byte{0x50, 0x4b, 0x07, 0x08}
	magicZstd              = []byte{0x28, 0xb5, 0x2f, 0xfd}
	magicLZ4Frame          = []byte{0x04, 0x22, 0x4d, 0x18}
	magicXZ                = []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}

	magicZlibNoComp          = []byte{0x78, 0x01}
	magicZlibBestSpeed       = []byte{0x78, 0x5e}
	magicZlibDefault         = []byte{0x78, 0x9c}
	magicZlibBestComp        = []byte{0x78, 0xda}
	magicZlibNoCompPreset    = []byte{0x78, 0x20}
	magicZlibBestSpeedPreset = []byte{0x78, 0x7d}
	magicZlibDefaultPreset   = []byte{0x78, 0xbb}
	magicZlibBestCompPreset  = []byte{0x78, 0xf9}
)

func zTypeFromBytes(magic []byte) zType {
	switch {
	case bytes.HasPrefix(magic, magicBzip2):
		return zBzip2
	case bytes.HasPrefix(magic, magicGzip):
		return zGzip
	case bytes.HasPrefix(magic, magicZip),
		bytes.HasPrefix(magic, magicZipEmptyArchive),
		bytes.HasPrefix(magic, magicZipSpannedArchive):
		return zZip
	case bytes.HasPrefix(magic, magicZstd):
		return zZstd
	case bytes.HasPrefix(magic, magicLZ4Frame):
		return zLZ4Frame
	case bytes.HasPrefix(magic, magicXZ):
		return zXZ
	case bytes.HasPrefix(magic, magicZlibNoComp),
		bytes.HasPrefix(magic, magicZlibBestSpeed),
		bytes.HasPrefix(magic, magicZlibDefault),
		bytes.HasPrefix(magic, magicZlibBestComp),
		bytes.HasPrefix(magic, magicZlibNoCompPreset),
		bytes.HasPrefix(magic, magicZlibBestSpeedPreset),
		bytes.HasPrefix(magic, magicZlibDefaultPreset),
		bytes.HasPrefix(magic, magicZlibBestCompPreset):
		return zZlib
	default:
		return zNone
	}
}
