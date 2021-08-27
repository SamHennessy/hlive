package hlive

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// CopyTree Simplify the tree and create a copy
func (p *Page) CopyTree(ctx context.Context, oldTree interface{}) (interface{}, error) {
	return p.copyTree(ctx, oldTree, true)
}

func (p *Page) copyTree(ctx context.Context, oldTree interface{}, lifeCycle bool) (interface{}, error) {
	switch v := oldTree.(type) {
	case nil:
		return nil, nil
	case *string:
		if v == nil || *v == "" {
			return nil, nil
		}

		return *v, nil
	case string:
		if v == "" {
			return nil, nil
		}

		return v, nil
	case *int:
		if v == nil {
			return nil, nil
		}

		return strconv.Itoa(*v), nil
	case *int16:
		if v == nil {
			return nil, nil
		}

		return strconv.FormatInt(int64(*v), base10), nil
	case *int8:
		if v == nil {
			return nil, nil
		}

		return strconv.FormatInt(int64(*v), base10), nil
	case *int32:
		if v == nil {
			return nil, nil
		}

		return strconv.FormatInt(int64(*v), base10), nil
	case *int64:
		if v == nil {
			return nil, nil
		}

		return strconv.FormatInt(*v, base10), nil
	case *uint:
		if v == nil {
			return nil, nil
		}

		return strconv.FormatUint(uint64(*v), base10), nil
	case *uint8:
		if v == nil {
			return nil, nil
		}

		return strconv.FormatUint(uint64(*v), base10), nil
	case *uint16:
		if v == nil {
			return nil, nil
		}

		return strconv.FormatUint(uint64(*v), base10), nil
	case *uint32:
		if v == nil {
			return nil, nil
		}

		return strconv.FormatUint(uint64(*v), base10), nil
	case *uint64:
		if v == nil {
			return nil, nil
		}

		return strconv.FormatUint(*v, base10), nil
	case *float32:
		if v == nil {
			return nil, nil
		}

		return strconv.FormatFloat(float64(*v), 'f', -1, bit32), nil
	case *float64:
		if v == nil {
			return nil, nil
		}

		return strconv.FormatFloat(*v, 'f', -1, bit64), nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, base10), nil
	case uint64:
		return strconv.FormatUint(v, base10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, bit64), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, bit32), nil
	case int8:
		return strconv.FormatInt(int64(v), base10), nil
	case int16:
		return strconv.FormatInt(int64(v), base10), nil
	case int32:
		return strconv.FormatInt(int64(v), base10), nil
	case uint:
		return strconv.FormatUint(uint64(v), base10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), base10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), base10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), base10), nil
	case *HTML:
		if v == nil || *v == "" {
			return nil, nil
		}

		return *v, nil
	case HTML:
		return v, nil
	case Tagger:
		if v == nil {
			return nil, nil
		}

		if lifeCycle {
			// Needs a lock higher up to be thread safe
			if comp, ok := v.(Componenter); ok {
				bindings := comp.GetEventBindings()
				for i := 0; i < len(bindings); i++ {
					p.currentBindings[bindings[i].ID] = bindings[i]
				}
			}

			if comp, ok := v.(Mounter); ok {
				if _, exists := p.mountables[comp.GetID()]; !exists {
					comp.Mount(ctx)

					p.mountables[comp.GetID()] = struct{}{}
				}
			}

			if compU, ok := v.(Unmounter); ok {
				if _, exists := p.unmountables[compU.GetID()]; !exists {
					p.unmountables[compU.GetID()] = compU
				}
			}

			if compTr, ok := v.(Teardowner); ok {
				compTr.SetTeardown(func() {
					delete(p.mountables, compTr.GetID())
					delete(p.unmountables, compTr.GetID())
				})
			}
		}

		kids, err := p.copyTree(ctx, v.GetNodes(), lifeCycle)
		if err != nil {
			return nil, fmt.Errorf("copy tree on tag children: %s: %w", v.GetName(), err)
		}

		// Call GetNodes before GetAttributes so users can set attributes in GetNodes
		var (
			els      []interface{}
			attrs    []*Attribute
			oldAttrs = v.GetAttributes()
		)

		for i := 0; i < len(oldAttrs); i++ {
			attrs = append(attrs, oldAttrs[i].Clone())
		}

		// Strip hlive attributes to trigger diff
		if !isWebSocket(ctx) {
			var newAttrs []*Attribute

			for i := 0; i < len(attrs); i++ {
				if strings.HasPrefix(attrs[i].Name, "data-hlive") {
					continue
				}

				newAttrs = append(newAttrs, attrs[i].Clone())
			}

			attrs = newAttrs
		}

		els = append(els, attrs)

		return T(v.GetName(), append(els, kids)), nil
	case RenderFunc:
		return p.copyTree(ctx, v(), lifeCycle)
	case []interface{}:
		var (
			newTree       []interface{}
			thisNodeStr   string
			thisNodeIsStr bool
			lastNodeStr   string
			lastNodeIsStr bool
		)

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			thisNodeStr, thisNodeIsStr = v[i].(string)

			if !thisNodeIsStr {
				val, ok := v[i].(*string)
				if ok && val == nil {
					thisNodeStr, thisNodeIsStr = "", true
				} else if ok {
					thisNodeStr, thisNodeIsStr = *val, true
				}
			}

			// Combine strings like a browser would
			if lastNodeIsStr && thisNodeIsStr && len(newTree) > 0 {
				// update this in case we have another string
				thisNodeStr = lastNodeStr + thisNodeStr
				// replace last node
				newTree[len(newTree)-1] = thisNodeStr
			} else {
				newTree = append(newTree, node)
			}

			lastNodeStr, lastNodeIsStr = thisNodeStr, thisNodeIsStr
		}

		return newTree, nil
	case []Componenter:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []Tagger:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []UniqueTagger:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []Component:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []*Component:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []Tag:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []*Tag:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	default:
		return nil, ErrRenderElement
	}
}
