package message

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"sort"
	"strings"
)

// From https://golang.org/src/mime/multipart/writer.go?s=2140:2215#L76
func writeHeader(w io.Writer, header textproto.MIMEHeader) error {
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range header[k] {
			if _, err := fmt.Fprintf(w, "%s: %s\r\n", k, v); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintf(w, "\r\n")
	return err
}


func CreatePart(w io.Writer, header textproto.MIMEHeader) (io.WriteCloser, error) {
	if err := writeHeader(w, header); err != nil {
		return nil, err
	}

	return encodeEncoding(w, header.Get("Content-Transfer-Encoding")), nil
}

func CreateMultipart(w io.Writer, header textproto.MIMEHeader) (*multipart.Writer, error) {
	mw := multipart.NewWriter(w)

	mediaType, mediaParams, _ := mime.ParseMediaType(header.Get("Content-Type"))
	if !strings.HasPrefix(mediaType, "multipart/") {
		return nil, errors.New("invalid multipart MIME type")
	}
	if mediaParams["boundary"] != "" {
		mw.SetBoundary(mediaParams["boundary"])
	} else {
		mediaParams["boundary"] = mw.Boundary()
		header.Set("Content-Type", mime.FormatMediaType(mediaType, mediaParams))
	}

	header.Del("Content-Transfer-Encoding")
	w, err := CreatePart(w, header)
	if err != nil {
		return nil, err
	}

	return mw, nil
}