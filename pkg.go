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
	ErrEventType        = errors.New("invalid event type")
)

// HLive special attributes
const (
	AttrID     = "data-hlive-id"
	AttrFocus  = "data-hlive-focus"
	AttrUpload = "data-hlive-upload"
)

// Event attribute keys
const (
	AttrOnClick           = "data-hlive-onclick"
	AttrOnKeyDown         = "data-hlive-onkeydown"
	AttrOnKeyUp           = "data-hlive-onkeyup"
	AttrOnFocus           = "data-hlive-onfocus"
	AttrOnAnimationEnd    = "data-hlive-onanimationend"
	AttrOnAnimationCancel = "data-hlive-onanimationcancel"
	AttrOnMouseEnter      = "data-hlive-onmouseenter"
	AttrOnMouseLeave      = "data-hlive-onmouseleave"
	AttrOnDiffApply       = "data-hlive-ondiffapply"
	AttrOnUpload          = "data-hlive-onupload"
	AttrOnChange          = "data-hlive-onchange"
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

type EventType int

// Event types
const (
	Click EventType = iota
	KeyDown
	KeyUp
	Focus
	AnimationEnd
	AnimationCancel
	MouseEnter
	MouseLeave
	DiffApply
	Upload
	Change
)
