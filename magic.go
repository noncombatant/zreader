package zreader

import "bytes"

const magicBytePrefixSize = 6

var (
	magicBzip2             = []byte{0x42, 0x5a, 0x68}
	magicGzip              = []byte{0x1f, 0x8b}
	magicZip               = []byte{0x50, 0x4b, 0x03, 0x04}
	magicZipEmptyArchive   = []byte{0x50, 0x4b, 0x05, 0x06}
	magicZipSpannedArchive = []byte{0x50, 0x4b, 0x07, 0x08}
	magicZstd              = []byte{0x28, 0xb5, 0x2f, 0xfd}
	magicLz4Frame          = []byte{0x04, 0x22, 0x4d, 0x18}
	magicXz                = []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}

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
	case bytes.HasPrefix(magic, magicLz4Frame):
		return zLz4Frame
	case bytes.HasPrefix(magic, magicXz):
		return zXz
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
