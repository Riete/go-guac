package protocol

import (
	"strconv"
)

// Status represents the status code in Guacamole protocol
// https://guacamole.apache.org/doc/gug/protocol-reference.html#status-codes
type Status int64

const (
	// Success - The operation succeeded. No error.
	Success Status = 0

	// Unsupported - The requested operation is unsupported.
	Unsupported Status = 256

	// ServerError - An internal error occurred, and the operation could not be performed.
	ServerError Status = 512

	// ServerBusy - The operation could not be performed because the server is busy.
	ServerBusy Status = 513

	// UpstreamTimeout - The upstream server is not responding. In most cases, the upstream server is the remote desktop server.
	UpstreamTimeout Status = 514

	// UpstreamError - The upstream server encountered an error. In most cases, the upstream server is the remote desktop server.
	UpstreamError Status = 515

	// ResourceNotFound - An associated resource, such as a file or stream, could not be found, and thus the operation failed.
	ResourceNotFound Status = 516

	// ResourceConflict - A resource is already in use or locked, preventing the requested operation.
	ResourceConflict Status = 517

	// ResourceClosed - The requested operation cannot continue because the associated resource has been closed.
	ResourceClosed Status = 518

	// UpstreamNotFound - The upstream server does not appear to exist, or cannot be reached over the network. In most cases, the upstream server is the remote desktop server.
	UpstreamNotFound Status = 519

	// UpstreamUnavailable - The upstream server is refusing to service connections. In most cases, the upstream server is the remote desktop server.
	UpstreamUnavailable Status = 520

	// SessionConflict - The session within the upstream server has ended because it conflicts with another session. In most cases, the upstream server is the remote desktop server.
	SessionConflict Status = 521

	// SessionTimeout - The session within the upstream server has ended because it appeared to be inactive. In most cases, the upstream server is the remote desktop server.
	SessionTimeout Status = 522

	// SessionClosed - The session within the upstream server has been forcibly closed. In most cases, the upstream server is the remote desktop server.
	SessionClosed Status = 523

	// ClientBadRequest - The parameters of the request are illegal or otherwise invalid.
	ClientBadRequest Status = 768

	// ClientUnauthorized - Permission was denied, because the user is not logged in. Note that the user may be logged into Guacamole, but still not logged in with respect to the remote desktop server.
	ClientUnauthorized Status = 769

	// ClientForbidden - Permission was denied, and logging in will not solve the problem.
	ClientForbidden Status = 771

	// ClientTimeout - The client (usually the user of Guacamole or their browser) is taking too long to respond.
	ClientTimeout Status = 776

	// ClientOverrun - The client has sent more data than the protocol allows.
	ClientOverrun Status = 781

	// ClientBadType - The client has sent data of an unexpected or illegal type.
	ClientBadType Status = 783

	// ClientTooMany - The client has sent too many requests.
	ClientTooMany Status = 797
)

// String returns the string representation of the status code with the status code number
func (s Status) String() string {
	code := strconv.FormatInt(int64(s), 10)
	switch s {
	case Success:
		return code + "_SUCCESS"
	case Unsupported:
		return code + "_UNSUPPORTED"
	case ServerError:
		return code + "_SERVER_ERROR"
	case ServerBusy:
		return code + "_SERVER_BUSY"
	case UpstreamTimeout:
		return code + "_UPSTREAM_TIMEOUT"
	case UpstreamError:
		return code + "_UPSTREAM_ERROR"
	case ResourceNotFound:
		return code + "_RESOURCE_NOT_FOUND"
	case ResourceConflict:
		return code + "_RESOURCE_CONFLICT"
	case ResourceClosed:
		return code + "_RESOURCE_CLOSED"
	case UpstreamNotFound:
		return code + "_UPSTREAM_NOT_FOUND"
	case UpstreamUnavailable:
		return code + "_UPSTREAM_UNAVAILABLE"
	case SessionConflict:
		return code + "_SESSION_CONFLICT"
	case SessionTimeout:
		return code + "_SESSION_TIMEOUT"
	case SessionClosed:
		return code + "_SESSION_CLOSED"
	case ClientBadRequest:
		return code + "_CLIENT_BAD_REQUEST"
	case ClientUnauthorized:
		return code + "_CLIENT_UNAUTHORIZED"
	case ClientForbidden:
		return code + "_CLIENT_FORBIDDEN"
	case ClientTimeout:
		return code + "_CLIENT_TIMEOUT"
	case ClientOverrun:
		return code + "_CLIENT_OVERRUN"
	case ClientBadType:
		return code + "_CLIENT_BAD_TYPE"
	case ClientTooMany:
		return code + "_CLIENT_TOO_MANY"
	default:
		return code + "_UNKNOWN"
	}
}
