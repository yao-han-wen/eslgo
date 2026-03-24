package eslgo

import "errors"

const OPT_EVENT_CHANNEL_CAPACITY = 8
const OPT_CMD_TIMEOUT = 10

const CONTENT_TYPE_AUTH_REQUEST = "auth/request"
const CONTENT_TYPE_COMMAND_REPLY = "command/reply"
const CONTENT_TYPE_API_RESPONSE = "api/response"

const CONTENT_TYPE_TEXT_EVENT_PLAIN = "text/event-plain"
const CONTENT_TYPE_TEXT_EVENT_XML = "text/event-xml"
const CONTENT_TYPE_TEXT_EVENT_JSON = "text/event-json"

var ErrNoResponseToAuthRequest = errors.New("No response to auth/request")
var ErrInvalidPassword = errors.New("Invalid password")
var ErrTimeout = errors.New("Timeout")
var ErrConnectClosed = errors.New("Connect closed")
