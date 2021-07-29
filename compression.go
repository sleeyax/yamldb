package yamldb

import (
	"github.com/peterbourgon/diskv/v3"
)

type Compression uint8

const (
	None = iota
	Zlib
	Gzip
)

func getCompression(compression Compression) diskv.Compression {
	switch compression {
	case Zlib:
		return diskv.NewZlibCompression()
	case Gzip:
		return diskv.NewGzipCompression()
	case None:
	default:
		break
	}
	return nil
}
