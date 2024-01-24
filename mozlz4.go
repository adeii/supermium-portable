package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"

	"github.com/pierrec/lz4/v3"
	"github.com/pkg/errors"
)

const (
	mozLz4Header = "mozLz40\x00"
)

func mozLz4Decompress(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	header := make([]byte, len(mozLz4Header))
	_, err = file.Read(header)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't read header")
	}
	if string(header) != mozLz4Header {
		return nil, errors.New("no mozLz4 header")
	}

	var size uint32
	err = binary.Read(file, binary.LittleEndian, &size)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't read size")
	}

	src, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't read compressed data")
	}

	out := make([]byte, size)
	_, err = lz4.UncompressBlock(src, out)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decompress data")
	}

	return out, nil
}

func mozLz4Compress(src []byte) ([]byte, error) {
	out := new(bytes.Buffer)

	_, err := out.Write([]byte(mozLz4Header))
	if err != nil {
		return nil, errors.Wrap(err, "couldn't write header")
	}

	err = binary.Write(out, binary.LittleEndian, uint32(len(src)))
	if err != nil {
		return nil, errors.Wrap(err, "couldn't encode length")
	}

	buf := make([]byte, 10*len(src))
	sz, err := lz4.CompressBlockHC(src, buf, -1)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't compress data")
	}
	if sz == 0 {
		return nil, errors.New("data incompressible")
	}

	_, err = out.Write(buf[:sz])
	if err != nil {
		return nil, errors.Wrap(err, "couldn't write compressed data")
	}

	return out.Bytes(), nil
}
