package eslgo

import "errors"

const MSG_END = "\r\n\r\n"

const OPT_CONNECT_PASSWORD = "ClueCon"
const OPT_CONNECT_TIMEOUT = 10
const OPT_COMMAND_TIMEOUT = 10
const OPT_EVENT_CHANNEL_CAPACITY = 8

const CONTENT_TYPE_AUTH_REQUEST = "auth/request"
const CONTENT_TYPE_COMMAND_REPLY = "command/reply"
const CONTENT_TYPE_API_RESPONSE = "api/response"

const CONTENT_TYPE_TEXT_EVENT_PLAIN = "text/event-plain"
const CONTENT_TYPE_TEXT_EVENT_XML = "text/event-xml"
const CONTENT_TYPE_TEXT_EVENT_JSON = "text/event-json"
const CONTENT_TYPE_TEXT_DISCONNECT_NOTICE = "text/disconnect-notice"

var ErrNoResponseToAuthRequest = errors.New("No response to auth/request")
var ErrInvalidPassword = errors.New("Invalid password")
var ErrCommandTimeout = errors.New("Command Timeout")
var ErrConnectClosed = errors.New("Connect closed")
var ErrDisconnectNotice = errors.New("Disconnect notice")
var ErrContentTypeMismatch = errors.New("Content-Type mismatch")
