// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package parquet

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/golang/snappy"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
	"github.com/minio/minio/pkg/s3select/internal/parquet-go/gen-go/parquet"
	"github.com/pierrec/lz4"
)

type compressionCodec parquet.CompressionCodec

var zstdOnce sync.Once
var zstdEnc *zstd.Encoder
var zstdDec *zstd.Decoder

func initZstd() {
	zstdOnce.Do(func() {
		zstdEnc, _ = zstd.NewWriter(nil, zstd.WithZeroFrames(true))
		zstdDec, _ = zstd.NewReader(nil)
	})
}

func (c compressionCodec) compress(buf []byte) ([]byte, error) {
	switch parquet.CompressionCodec(c) {
	case parquet.CompressionCodec_UNCOMPRESSED:
		return buf, nil

	case parquet.CompressionCodec_SNAPPY:
		return snappy.Encode(nil, buf), nil

	case parquet.CompressionCodec_GZIP:
		byteBuf := new(bytes.Buffer)
		writer := gzip.NewWriter(byteBuf)
		n, err := writer.Write(buf)
		if err != nil {
			return nil, err
		}
		if n != len(buf) {
			return nil, fmt.Errorf("short writes")
		}

		if err = writer.Flush(); err != nil {
			return nil, err
		}

		if err = writer.Close(); err != nil {
			return nil, err
		}

		return byteBuf.Bytes(), nil

	case parquet.CompressionCodec_LZ4:
		byteBuf := new(bytes.Buffer)
		writer := lz4.NewWriter(byteBuf)
		n, err := writer.Write(buf)
		if err != nil {
			return nil, err
		}
		if n != len(buf) {
			return nil, fmt.Errorf("short writes")
		}

		if err = writer.Flush(); err != nil {
			return nil, err
		}

		if err = writer.Close(); err != nil {
			return nil, err
		}

		return byteBuf.Bytes(), nil
	case parquet.CompressionCodec_ZSTD:
		initZstd()
		return zstdEnc.EncodeAll(buf, nil), nil
	}

	return nil, fmt.Errorf("invalid compression codec %v", c)
}

func (c compressionCodec) uncompress(buf []byte) ([]byte, error) {
	switch parquet.CompressionCodec(c) {
	case parquet.CompressionCodec_UNCOMPRESSED:
		return buf, nil

	case parquet.CompressionCodec_SNAPPY:
		return snappy.Decode(nil, buf)

	case parquet.CompressionCodec_GZIP:
		reader, err := gzip.NewReader(bytes.NewReader(buf))
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return ioutil.ReadAll(reader)

	case parquet.CompressionCodec_LZ4:
		return ioutil.ReadAll(lz4.NewReader(bytes.NewReader(buf)))

	case parquet.CompressionCodec_ZSTD:
		initZstd()
		return zstdDec.DecodeAll(buf, nil)
	}

	return nil, fmt.Errorf("invalid compression codec %v", c)
}
