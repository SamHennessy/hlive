package hlive

import (
	"errors"
	"time"
)

// Public errors
var (
	ErrRenderElement = errors.New("attempted to render an unrecognised element")
)

// HLive special attributes
const (
	AttrID     = "hid"
	AttrOn     = "hon"
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

type CtxKey string

// Context keys
const (
	CtxRender          CtxKey = "render"
	CtxRenderComponent CtxKey = "render_comp"
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
