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
)

// HLive special attributes
const (
	AttrID     = "data-hlive-id"
	AttrOn     = "data-hlive-on"
	AttrFocus  = "data-hlive-focus"
	AttrUpload = "data-hlive-upload"
	// DiffApply is a special event that will trigger when a diff is applied.
	// This means that it will trigger itself when first added. This will allow you to know when a change in the tree has
	// made it to the browser. You can then, if you wish, immediately remove it from the tree to prevent more triggers.
	// You can also add it as a OnOnce and it wil remove itself.
	DiffApply = "diffapply"
	base10    = 10
	bit32     = 32
	bit64     = 64
)

// Defaults
const (
	HTML5DocType                      = HTML("<!doctype html>")
	WebSocketDisconnectTimeoutDefault = time.Second * 5
	PageSessionLimitDefault           = 1000
	PageSessionGarbageCollectionTick  = time.Second
)

type CtxKey int

// Context keys
const (
	CtxPageSess CtxKey = iota
	CtxIsWS
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
