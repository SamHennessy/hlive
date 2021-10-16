package hlive

import (
	"errors"
	"time"
)

// Public errors
var (
	ErrRenderElement    = errors.New("attempted to render an unrecognised element")
	ErrAttrValueCount   = errors.New("zero or one value allowed only")
	ErrInvalidNode      = errors.New("variable is not a valid node")
	ErrInvalidElement   = errors.New("variable is not a valid element")
	ErrInvalidAttribute = errors.New("variable is not a valid attribute")
	ErrRenderCtx        = errors.New("render not found in context")
	ErrRenderCompCtx    = errors.New("component render not found in context")
)

// HLive special attributes
const (
	AttrID     = "data-hlive-id"
	AttrOn     = "data-hlive-on"
	AttrUpload = "data-hlive-upload"
	base10     = 10
	bit32      = 32
	bit64      = 64
)

// Defaults
const (
	HTML5DocType                      HTML = "<!doctype html>"
	WebSocketDisconnectTimeoutDefault      = time.Second * 5
	PageSessionLimitDefault                = 1000
	PageSessionGarbageCollectionTick       = time.Second
)

type CtxKey int

// Context keys
const (
	CtxPageSess CtxKey = iota
	CtxRender
	CtxRenderComponent
)

type DiffType string

// Diff types
const (
	DiffUpdate DiffType = "u"
	DiffCreate DiffType = "c"
	DiffDelete DiffType = "d"
)

var newline = []byte{'\n'}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	// maxMessageSize = 512
)
