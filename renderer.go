package hlive

import (
	"fmt"
	"html"
	"io"
	"strconv"

	"github.com/rs/zerolog"
)

func NewRenderer() *Renderer {
	return &Renderer{
		log: zerolog.Nop(),
	}
}

type Renderer struct {
	log zerolog.Logger
}

func (r *Renderer) SetLogger(logger zerolog.Logger) {
	r.log = logger
}

// HTML renders items that can be render to valid HTML nodes
func (r *Renderer) HTML(w io.Writer, el any) error {
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
	case int:
		if _, err := w.Write([]byte(strconv.Itoa(v))); err != nil {
			return fmt.Errorf("html write: %w", err)
		}
	case int8:
		if _, err := w.Write([]byte(strconv.Itoa(int(v)))); err != nil {
			return fmt.Errorf("html write: %w", err)
		}
	case int16:
		if _, err := w.Write([]byte(strconv.Itoa(int(v)))); err != nil {
			return fmt.Errorf("html write: %w", err)
		}
	case int32:
		if _, err := w.Write([]byte(strconv.Itoa(int(v)))); err != nil {
			return fmt.Errorf("html write: %w", err)
		}
	case int64:
		if _, err := w.Write([]byte(strconv.FormatInt(v, base10))); err != nil {
			return fmt.Errorf("html write: %w", err)
		}
	case uint:
		if _, err := w.Write([]byte(strconv.FormatUint(uint64(v), base10))); err != nil {
			return fmt.Errorf("html write: %w", err)
		}
	case uint8:
		if _, err := w.Write([]byte(strconv.FormatUint(uint64(v), base10))); err != nil {
			return fmt.Errorf("html write: %w", err)
		}
	case uint16:
		if _, err := w.Write([]byte(strconv.FormatUint(uint64(v), base10))); err != nil {
			return fmt.Errorf("html write: %w", err)
		}
	case uint32:
		if _, err := w.Write([]byte(strconv.FormatUint(uint64(v), base10))); err != nil {
			return fmt.Errorf("html write: %w", err)
		}
	case uint64:
		if _, err := w.Write([]byte(strconv.FormatUint(v, base10))); err != nil {
			return fmt.Errorf("html write: %w", err)
		}
	case float32:
		if _, err := w.Write([]byte(strconv.FormatFloat(float64(v), 'f', -1, bit32))); err != nil {
			return fmt.Errorf("html write: %w", err)
		}
	case float64:
		if _, err := w.Write([]byte(strconv.FormatFloat(v, 'f', -1, bit64))); err != nil {
			return fmt.Errorf("html write: %w", err)
		}
	case *int:
		if v != nil {
			if _, err := w.Write([]byte(strconv.Itoa(*v))); err != nil {
				return fmt.Errorf("html write: %w", err)
			}
		}
	case *int8:
		if v != nil {
			if _, err := w.Write([]byte(strconv.Itoa(int(*v)))); err != nil {
				return fmt.Errorf("html write: %w", err)
			}
		}
	case *int16:
		if v != nil {
			if _, err := w.Write([]byte(strconv.Itoa(int(*v)))); err != nil {
				return fmt.Errorf("html write: %w", err)
			}
		}
	case *int32:
		if v != nil {
			if _, err := w.Write([]byte(strconv.Itoa(int(*v)))); err != nil {
				return fmt.Errorf("html write: %w", err)
			}
		}
	case *int64:
		if v != nil {
			if _, err := w.Write([]byte(strconv.FormatInt(*v, base10))); err != nil {
				return fmt.Errorf("html write: %w", err)
			}
		}
	case *uint:
		if v != nil {
			if _, err := w.Write([]byte(strconv.FormatUint(uint64(*v), base10))); err != nil {
				return fmt.Errorf("html write: %w", err)
			}
		}
	case *uint8:
		if v != nil {
			if _, err := w.Write([]byte(strconv.FormatUint(uint64(*v), base10))); err != nil {
				return fmt.Errorf("html write: %w", err)
			}
		}
	case *uint16:
		if v != nil {
			if _, err := w.Write([]byte(strconv.FormatUint(uint64(*v), base10))); err != nil {
				return fmt.Errorf("html write: %w", err)
			}
		}
	case *uint32:
		if v != nil {
			if _, err := w.Write([]byte(strconv.FormatUint(uint64(*v), base10))); err != nil {
				return fmt.Errorf("html write: %w", err)
			}
		}
	case *uint64:
		if v != nil {
			if _, err := w.Write([]byte(strconv.FormatUint(*v, base10))); err != nil {
				return fmt.Errorf("html write: %w", err)
			}
		}
	case *float32:
		if v != nil {
			if _, err := w.Write([]byte(strconv.FormatFloat(float64(*v), 'f', -1, bit32))); err != nil {
				return fmt.Errorf("html write: %w", err)
			}
		}
	case *float64:
		if v != nil {
			if _, err := w.Write([]byte(strconv.FormatFloat(*v, 'f', -1, bit64))); err != nil {
				return fmt.Errorf("html write: %w", err)
			}
		}
	case *HTML:
		if v == nil {
			return nil
		}

		if _, err := w.Write([]byte(*v)); err != nil {
			return err
		}
	case HTML:
		if _, err := w.Write([]byte(v)); err != nil {
			return err
		}
	case Tagger:
		if v != nil {
			if err := r.tag(v, w); err != nil {
				return err
			}
		}
	// I don't think this is possible anymore
	case []any:
		if err := r.HTML(w, G(v...)); err != nil {
			return err
		}
	case *NodeGroup:
		g := v.Get()
		for i := 0; i < len(g); i++ {
			if err := r.HTML(w, g[i]); err != nil {
				return err
			}
		}
	default:
		return ErrRenderElement
	}

	return nil
}

// Attribute renders an Attribute to it's HTML string representation
// While it's possible to have HTML attributes without values it simplifies things if we always have a value
func (r *Renderer) Attribute(attrs []Attributer, w io.Writer) error {
	for i := 0; i < len(attrs); i++ {
		attr := attrs[i]
		if attr.IsNoEscapeString() {
			if _, err := w.Write([]byte(fmt.Sprintf(` %s="%s"`, attr.GetName(), attr.GetValue()))); err != nil {
				return fmt.Errorf("write: %w", err)
			}
		} else {
			if _, err := w.Write([]byte(fmt.Sprintf(` %s="%s"`, attr.GetName(), html.EscapeString(attr.GetValue())))); err != nil {
				return fmt.Errorf("write: %w", err)
			}
		}
	}

	return nil
}

func (r *Renderer) text(text string, w io.Writer) error {
	if text == "" {
		return nil
	}

	if _, err := w.Write([]byte(html.EscapeString(text))); err != nil {
		return fmt.Errorf("write to writer: %w", err)
	}

	return nil
}

func (r *Renderer) tag(tag Tagger, w io.Writer) error {
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

	if err := r.HTML(w, tag.GetNodes()); err != nil {
		return fmt.Errorf("render nodes for: %s: %w", tag.GetName(), err)
	}

	if _, err := w.Write([]byte("</" + tag.GetName() + ">")); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}
