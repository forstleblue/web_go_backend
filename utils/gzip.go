package utils

import (
	"bytes"
	"io/ioutil"

	"github.com/klauspost/compress/gzip" //this is faster than the default gzip
)

//Gzip compresses given bytes using klauspost/compress/gzip
func Gzip(data []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	defer w.Close()
	w.Write(data)
	w.Flush()
	return buf.Bytes()
}

//Gunzip decompresses given bytes using klauspost/compress/gzip and returns any error
func Gunzip(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	r, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	ud, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ud, nil
}
