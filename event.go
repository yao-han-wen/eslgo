package eslgo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
)

type EventHeader map[string]string

type Event struct {
	Header EventHeader
	Body   string
}

func (e *Event) ParsePlainToEvent(data []byte) error {
	if e.Header == nil {
		e.Header = make(map[string]string)
	}

	reader := bufio.NewReader(bytes.NewReader(data))

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if line == "\r\n" || line == "\n" {
			break
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return fmt.Errorf("ParsePlainToEvent, invalid MIME header line: %q", line)
		}

		e.Header[strings.TrimSpace(key)], err = url.QueryUnescape(strings.TrimSpace(value))
		if err != nil {
			return err
		}
	}

	//存在body
	if v := e.Header["Content-Length"]; v != "" {
		length, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		body := make([]byte, length)
		if _, err := io.ReadFull(reader, body); err != nil {
			return err
		}
		e.Body = string(body)
	}

	return nil
}

func (e *Event) ParseJsonToEvent(data []byte) error {
	if e.Header == nil {
		e.Header = make(map[string]string)
	}

	err := json.Unmarshal(data, &e.Header)
	if err != nil {
		return err
	}
	//存在body
	if v := e.Header["_body"]; v != "" {
		e.Body = v
		delete(e.Header, "_body")
	}

	return nil
}

func (e *Event) ParseXMLToEvent(data []byte) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	if e.Header == nil {
		e.Header = make(map[string]string)
	}

	var inHeaders bool
	var isBody bool
	var currentKey string
	var contentBuilder strings.Builder

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "headers" {
				inHeaders = true
			}
			if t.Name.Local == "body" {
				isBody = true
				currentKey = t.Name.Local
				contentBuilder.Reset()
			}
			if inHeaders {
				currentKey = t.Name.Local
				contentBuilder.Reset()
			}

		case xml.EndElement:
			if t.Name.Local == "headers" {
				inHeaders = false
			}
			if t.Name.Local == "body" {
				isBody = false
				e.Body, err = url.QueryUnescape(strings.TrimSpace(contentBuilder.String()))
				if err != nil {
					return err
				}
			}
			if inHeaders && t.Name.Local == currentKey && currentKey != "" {
				e.Header[currentKey], err = url.QueryUnescape(strings.TrimSpace(contentBuilder.String()))
				if err != nil {
					return err
				}
				currentKey = ""
			}

		case xml.CharData:
			if inHeaders && currentKey != "" {
				contentBuilder.Write(t)
			}
			if isBody {
				contentBuilder.Write(t)
			}
		}
	}

	return nil
}
