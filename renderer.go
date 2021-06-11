package hlive

import (
	"fmt"
	"html"
	"io"

	"github.com/rs/zerolog"
)

func NewRender() *Renderer {
	return &Renderer{
		Logger: zerolog.Nop(),
	}
}

type Renderer struct {
	Logger zerolog.Logger
}

// HTML renders items that can be render to valid HTML nodes
func (r *Renderer) HTML(w io.Writer, el interface{}) error {
	switch v := el.(type) {
	case nil:
		return nil
	case *string:
		if v != nil {
			if err := r.text(*v, w); err != nil {
				return err
			}
		}
	case string:
		if err := r.text(v, w); err != nil {
			return err
		}
	case *int: // TODO: *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64
		if v != nil {
			if err := r.text(fmt.Sprint(*v), w); err != nil {
				return err
			}
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		if err := r.text(fmt.Sprint(v), w); err != nil {
			return err
		}
	case HTML:
		if err := r.rawHTML(v, w); err != nil {
			return err
		}
	case TagInterface:
		if v != nil {
			if err := r.tag(v, w); err != nil {
				return err
			}
		}
	case []interface{}:
		for i := 0; i < len(v); i++ {
			if err := r.HTML(w, v[i]); err != nil {
				return err
			}
		}
	default:
		return ErrRenderElement
	}

	return nil
}

// Attribute renders an Attribute to it's HTML string representation
func (r *Renderer) Attribute(attrs []*Attribute, w io.Writer) error {
	if len(attrs) == 0 {
		return nil
	}

	for i := 0; i < len(attrs); i++ {
		attrStr := " " + attrs[i].Name
		if attrs[i].Value != nil {
			attrStr += fmt.Sprintf(`="%s"`, *attrs[i].Value)
		}

		if _, err := w.Write([]byte(attrStr)); err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}

	return nil
}

func (r *Renderer) text(text string, w io.Writer) error {
	_, err := w.Write([]byte(html.EscapeString(text)))

	return err
}

func (r *Renderer) rawHTML(rawHTML HTML, w io.Writer) error {
	_, err := w.Write([]byte(rawHTML))

	return err
}

func (r *Renderer) tag(tag TagInterface, w io.Writer) error {
	if _, err := w.Write([]byte("<" + tag.GetName())); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	if err := r.Attribute(tag.GetAttributes(), w); err != nil {
		return fmt.Errorf("render attributes: %w", err)
	}

	if _, err := w.Write([]byte(">")); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	if tag.IsVoid() {
		return nil
	}

	kids := tag.Render()
	for i := 0; i < len(kids); i++ {
		if err := r.HTML(w, kids[i]); err != nil {
			return fmt.Errorf("render child: %#v: %w", kids[i], err)
		}
	}

	if _, err := w.Write([]byte("</" + tag.GetName() + ">")); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}
