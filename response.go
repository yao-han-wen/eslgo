package eslgo

import (
	"errors"
)

type Response struct {
	Mime    map[string]string
	Content []byte
}

func (r Response) HasError() error {
	switch r.Mime["Content-Type"] {
	case CONTENT_TYPE_COMMAND_REPLY:
		rs := r.Mime["Reply-Text"]
		if len(rs) > 1 && rs[:2] == "-E" {
			return errors.New(rs)
		}
	case CONTENT_TYPE_API_RESPONSE:
		rs := string(r.Content)
		if len(rs) > 1 && rs[:2] == "-E" {
			return errors.New(rs)
		}
	default:
		return ErrContentTypeMismatch
	}
	return nil
}

func (r Response) ToEvent() (*Event, error) {
	var err error
	event := &Event{
		Header: make(EventHeader),
	}

	switch r.Mime["Content-Type"] {
	case CONTENT_TYPE_TEXT_EVENT_PLAIN:
		err = event.ParsePlainToEvent(r.Content)
		if err != nil {
			return nil, err
		}

	case CONTENT_TYPE_TEXT_EVENT_XML:
		err = event.ParseXMLToEvent(r.Content)
		if err != nil {
			return nil, err
		}

	case CONTENT_TYPE_TEXT_EVENT_JSON:
		err = event.ParseJsonToEvent(r.Content)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrContentTypeMismatch
	}
	return event, nil
}
