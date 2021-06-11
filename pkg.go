package hlive

import "errors"

// Package errors
var (
	ErrRenderElement    = errors.New("attempted to render an unrecognised element")
	ErrAttrValueCount   = errors.New("zero or one value allowed only")
	ErrInvalidNode      = errors.New("variable is not a valid node")
	ErrInvalidElement   = errors.New("variable is not a valid element")
	ErrInvalidAttribute = errors.New("variable is not a valid attribute")
)

const (
	AttrID                = "data-hlive-id"
	AttrFocus             = "data-hlive-focus"
	AttrOnClick           = "data-hlive-onclick"
	AttrOnKeyDown         = "data-hlive-onkeydown"
	AttrOnKeyUp           = "data-hlive-onkeyup"
	AttrOnFocus           = "data-hlive-onfocus"
	AttrOnAnimationEnd    = "data-hlive-onanimationend"
	AttrOnAnimationCancel = "data-hlive-onanimationcancel"
	AttrOnMouseEnter      = "data-hlive-onmouseenter"
	AttrOnMouseLeave      = "data-hlive-onmouseleave"
	HTML5DocType          = HTML("<!doctype html>")
)

type CtxKey int

const (
	CtxPageSess CtxKey = iota
	CtxIsWS
	CtxRender
	CtxRenderComponent
)

type DiffType string

const (
	DiffUpdate DiffType = "u"
	DiffCreate DiffType = "c"
	DiffDelete DiffType = "d"
)

type EventType int

const (
	Click EventType = iota
	KeyDown
	KeyUp
	Focus
	AnimationEnd
	AnimationCancel
	MouseEnter
	MouseLeave
)
