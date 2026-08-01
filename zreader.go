// Copyright 2024 Chris Palmer, https://noncombatant.org/
// SPDX-License-Identifier: Apache-2.0

// Package zreader provides an [io.ReadCloser] for a variety of compression
// formats.
package zreader

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz"
)

// TODO: Support compress/{flate,lzw}, brotli.
//
// TODO: Consider using more or all of the klauspost implementations.
//
// TODO: Add an OpenWithType.

type Type int

const (
	TypeNone Type = iota
	TypeBzip2
	TypeGzip
	TypeZip
	TypeZstd
	TypeZlib
	TypeLZ4Frame
	TypeLZ4Block // TODO: Can't use a magic byte approach for this
	TypeXZ
)

// Reader is an [io.ReadCloser] that decompresses and reads compressed files.
// It supports the compression formats enumerated by [Type].
type Reader struct {
	Type
	io.Reader
	io.Closer
}

// Open opens pathname and returns an appropriate Reader. See [NewReader] for
// guidance on its behavior.
func Open(pathname string) (*Reader, error) {
	file, e := os.Open(pathname)
	if e != nil {
		return nil, e
	}
	return NewReader(bufio.NewReader(file), file)
}

func NewReader(r *bufio.Reader, c io.Closer) (*Reader, error) {
	magicBlock, e := r.Peek(magicBytePrefixSize)
	if e != nil {
		if errors.Is(e, io.EOF) {
			return &Reader{Type: TypeNone, Reader: r, Closer: c}, nil
		}
		return nil, e
	}

	switch TypeFromBytes(magicBlock) {
	case TypeBzip2:
		return &Reader{Type: TypeBzip2, Reader: bzip2.NewReader(r), Closer: c}, nil
	case TypeGzip:
		g, e := gzip.NewReader(r)
		if e != nil {
			return nil, e
		}
		return &Reader{Type: TypeGzip, Reader: g, Closer: c}, nil
	case TypeZip:
		return nil, fmt.Errorf("TODO")
	case TypeZstd:
		z, e := zstd.NewReader(r)
		if e != nil {
			return nil, e
		}
		return &Reader{Type: TypeZstd, Reader: z, Closer: c}, nil
	case TypeLZ4Frame:
		return &Reader{Type: TypeLZ4Frame, Reader: lz4.NewReader(r), Closer: c}, nil
	case TypeXZ:
		x, e := xz.NewReader(r)
		if e != nil {
			return nil, e
		}
		return &Reader{Type: TypeXZ, Reader: x, Closer: c}, nil
	case TypeZlib:
		z, e := zlib.NewReader(r)
		if e != nil {
			return nil, e
		}
		return &Reader{Type: TypeZlib, Reader: z, Closer: c}, nil
	default:
		return &Reader{Type: TypeNone, Reader: r, Closer: c}, nil
	}
}

// Read reads from the appropriate ReadCloser.
func (r *Reader) Read(p []byte) (int, error) {
	return r.Reader.Read(p)
}

// Close closes the Reader, and will close the underlying reader if it has one.
func (r *Reader) Close() error {
	if e := r.Closer.Close(); e != nil {
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

func TypeFromBytes(magic []byte) Type {
	switch {
	case bytes.HasPrefix(magic, magicBzip2):
		return TypeBzip2
	case bytes.HasPrefix(magic, magicGzip):
		return TypeGzip
	case bytes.HasPrefix(magic, magicZip),
		bytes.HasPrefix(magic, magicZipEmptyArchive),
		bytes.HasPrefix(magic, magicZipSpannedArchive):
		return TypeZip
	case bytes.HasPrefix(magic, magicZstd):
		return TypeZstd
	case bytes.HasPrefix(magic, magicLZ4Frame):
		return TypeLZ4Frame
	case bytes.HasPrefix(magic, magicXZ):
		return TypeXZ
	case bytes.HasPrefix(magic, magicZlibNoComp),
		bytes.HasPrefix(magic, magicZlibBestSpeed),
		bytes.HasPrefix(magic, magicZlibDefault),
		bytes.HasPrefix(magic, magicZlibBestComp),
		bytes.HasPrefix(magic, magicZlibNoCompPreset),
		bytes.HasPrefix(magic, magicZlibBestSpeedPreset),
		bytes.HasPrefix(magic, magicZlibDefaultPreset),
		bytes.HasPrefix(magic, magicZlibBestCompPreset):
		return TypeZlib
	default:
		return TypeNone
	}
}
