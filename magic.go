package zreader

import "bytes"

const magicBytePrefixSize = 4

var (
	magicBzip2             = []byte{0x42, 0x5a, 0x68}
	magicGzip              = []byte{0x1f, 0x8b}
	magicZip               = []byte{0x50, 0x4b, 0x03, 0x04}
	magicZipEmptyArchive   = []byte{0x50, 0x4b, 0x05, 0x06}
	magicZipSpannedArchive = []byte{0x50, 0x4b, 0x07, 0x08}
	magicZstd              = []byte{0x28, 0xb5, 0x2f, 0xfd}
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
	default:
		return zNone
	}
}
