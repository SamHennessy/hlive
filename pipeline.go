package hlive

import (
	"context"
	"errors"
	"fmt"
	"io"
)

var ErrDOMInvalidated = errors.New("dom invalidated")

type (
	PipeNodeHandler       func(ctx context.Context, w io.Writer, node any) (any, error)
	PipeNodegroupHandler  func(ctx context.Context, w io.Writer, node *NodeGroup) (*NodeGroup, error)
	PipeTaggerHandler     func(ctx context.Context, w io.Writer, tagger Tagger) (Tagger, error)
	PipeTagHandler        func(ctx context.Context, w io.Writer, tag *Tag) (*Tag, error)
	PipeAttributerHandler func(ctx context.Context, w io.Writer, tag Attributer) (Attributer, error)
)

type Pipeline struct {
	processors   []*PipelineProcessor
	processorMap map[string]*PipelineProcessor

	onSimpleNodeCache []*PipelineProcessor
	beforeWalkCache   []*PipelineProcessor
	afterWalkCache    []*PipelineProcessor
	beforeTaggerCache []*PipelineProcessor
	afterTaggerCache  []*PipelineProcessor
	beforeAttrCache   []*PipelineProcessor
	afterAttrCache    []*PipelineProcessor
	// Add new caches to RemoveAll
}

func NewPipeline(pps ...*PipelineProcessor) *Pipeline {
	p := &Pipeline{processorMap: map[string]*PipelineProcessor{}}
	p.Add(pps...)

	return p
}

func (p *Pipeline) Add(processors ...*PipelineProcessor) {
	p.processors = append(p.processors, processors...)

	for i := 0; i < len(processors); i++ {
		if processors[i].Key != "" {
			p.processorMap[processors[i].Key] = processors[i]
		}

		if processors[i].OnSimpleNode != nil {
			p.onSimpleNodeCache = append(p.onSimpleNodeCache, processors[i])
		}

		if processors[i].BeforeWalk != nil {
			p.beforeWalkCache = append(p.beforeWalkCache, processors[i])
		}

		if processors[i].AfterWalk != nil {
			p.afterWalkCache = append(p.afterWalkCache, processors[i])
		}

		if processors[i].BeforeTagger != nil {
			p.beforeTaggerCache = append(p.beforeTaggerCache, processors[i])
		}

		if processors[i].AfterTagger != nil {
			p.afterTaggerCache = append(p.afterTaggerCache, processors[i])
		}

		if processors[i].BeforeAttribute != nil {
			p.beforeAttrCache = append(p.beforeAttrCache, processors[i])
		}

		if processors[i].AfterAttribute != nil {
			p.afterAttrCache = append(p.afterAttrCache, processors[i])
		}
	}
}

func (p *Pipeline) RemoveAll() {
	p.processors = nil

	p.processorMap = map[string]*PipelineProcessor{}

	p.onSimpleNodeCache = nil
	p.beforeWalkCache = nil
	p.afterWalkCache = nil
	p.beforeTaggerCache = nil
	p.afterTaggerCache = nil
	p.beforeAttrCache = nil
	p.afterAttrCache = nil
}

func (p *Pipeline) AddAfter(processorKey string, processors ...*PipelineProcessor) {
	var newProcessors []*PipelineProcessor

	var hit bool

	for i := 0; i < len(p.processors); i++ {
		newProcessors = append(newProcessors, p.processors[i])

		if p.processors[i].Key == processorKey {
			hit = true

			newProcessors = append(newProcessors, processors...)
		}
	}

	if !hit {
		newProcessors = append(newProcessors, processors...)
	}

	p.RemoveAll()
	p.Add(newProcessors...)
}

func (p *Pipeline) AddBefore(processorKey string, processors ...*PipelineProcessor) {
	var newProcessors []*PipelineProcessor

	var hit bool

	for i := 0; i < len(p.processors); i++ {
		if p.processors[i].Key == processorKey {
			hit = true
			newProcessors = append(newProcessors, processors...)
		}

		newProcessors = append(newProcessors, p.processors[i])
	}

	if !hit {
		newProcessors = append(newProcessors, processors...)
	}

	p.RemoveAll()
	p.Add(newProcessors...)
}

func (p *Pipeline) onSimpleNode(ctx context.Context, w io.Writer, node any) (any, error) {
	for _, processor := range p.onSimpleNodeCache {
		if processor.Disabled {
			continue
		}

		newNode, err := processor.OnSimpleNode(ctx, w, node)
		if err != nil {
			return node, fmt.Errorf("onSimpleNode: %w", err)
		}

		node = newNode
	}

	return node, nil
}

func (p *Pipeline) afterWalk(ctx context.Context, w io.Writer, node *NodeGroup) (*NodeGroup, error) {
	for _, processor := range p.afterWalkCache {
		if processor.Disabled {
			continue
		}

		newNode, err := processor.AfterWalk(ctx, w, node)
		if err != nil {
			return node, fmt.Errorf("afterWalk: %w", err)
		}

		node = newNode
	}

	return node, nil
}

func (p *Pipeline) beforeWalk(ctx context.Context, w io.Writer, node *NodeGroup) (*NodeGroup, error) {
	for _, processor := range p.beforeWalkCache {
		if processor.Disabled || processor.BeforeWalk == nil {
			continue
		}

		newNode, err := processor.BeforeWalk(ctx, w, node)
		if err != nil {
			return node, fmt.Errorf("before: %w", err)
		}

		node = newNode
	}

	return node, nil
}

func (p *Pipeline) afterTagger(ctx context.Context, w io.Writer, node *Tag) (*Tag, error) {
	for _, processor := range p.afterTaggerCache {
		if processor.Disabled {
			continue
		}

		newNode, err := processor.AfterTagger(ctx, w, node)
		if err != nil {
			return node, fmt.Errorf("afterTagger: %w", err)
		}

		node = newNode
	}

	return node, nil
}

func (p *Pipeline) beforeTagger(ctx context.Context, w io.Writer, node Tagger) (Tagger, error) {
	for _, processor := range p.beforeTaggerCache {
		if processor.Disabled {
			continue
		}

		newNode, err := processor.BeforeTagger(ctx, w, node)
		if err != nil {
			return node, fmt.Errorf("beforeTagger: %w", err)
		}

		node = newNode
	}

	return node, nil
}

func (p *Pipeline) beforeAttr(ctx context.Context, w io.Writer, attr Attributer) (Attributer, error) {
	var err error
	for _, processor := range p.beforeAttrCache {
		if processor.Disabled {
			continue
		}

		attr, err = processor.BeforeAttribute(ctx, w, attr)
		if err != nil {
			return nil, fmt.Errorf("before attribute: %w", err)
		}
	}

	return attr, nil
}

func (p *Pipeline) afterAttr(ctx context.Context, w io.Writer, attr Attributer) (Attributer, error) {
	var err error
	for _, processor := range p.afterAttrCache {
		if processor.Disabled {
			continue
		}

		attr, err = processor.AfterAttribute(ctx, w, attr)
		if err != nil {
			return nil, fmt.Errorf("after attribute: %w", err)
		}
	}

	return attr, nil
}

// Run all the steps
func (p *Pipeline) run(ctx context.Context, w io.Writer, nodeGroup *NodeGroup) (*NodeGroup, error) {
	nodeGroup, err := p.beforeWalk(ctx, w, nodeGroup)
	if err != nil {
		return nil, fmt.Errorf("run: beforeWalk: %w", err)
	}

	newGroup := G()
	list := nodeGroup.Get()
	for i := 0; i < len(list); i++ {
		newNode, err := p.walk(ctx, w, list[i])
		if err != nil {
			return nil, fmt.Errorf("run: walk: %w", err)
		}

		newGroup.Add(newNode)
	}

	newGroup, err = p.afterWalk(ctx, w, newGroup)
	if err != nil {
		return nil, fmt.Errorf("run full tree: %w", err)
	}

	return newGroup, nil
}

// Skips some steps
func (p *Pipeline) runNode(ctx context.Context, w io.Writer, node any) (any, error) {
	return p.walk(ctx, w, node)
}

func (p *Pipeline) walk(ctx context.Context, w io.Writer, node any) (any, error) {
	switch v := node.(type) {
	case nil:
		return nil, nil
	// Single Node,
	// Not a Tagger
	case string, HTML,
		int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:

		return p.onSimpleNode(ctx, w, node)
	// All Taggers wil be converted to a Tag
	case Tagger:
		if v.IsNil() {
			return nil, nil
		}

		v, err := p.beforeTagger(ctx, w, v)
		if err != nil {
			return nil, err
		}

		kids, err := p.walk(ctx, w, v.GetNodes())
		if err != nil {
			return nil, err
		}

		oldAttrs := v.GetAttributes()
		var attrs []Attributer

		for i := 0; i < len(oldAttrs); i++ {
			attr := oldAttrs[i]

			attr, err = p.beforeAttr(ctx, w, attr)
			if err != nil {
				return nil, err
			}

			attr = oldAttrs[i].Clone()

			attr, err = p.afterAttr(ctx, w, attr)
			if err != nil {
				return nil, err
			}

			attrs = append(attrs, attr)
		}

		tag := T(v.GetName(), attrs, kids)
		tag.SetVoid(tag.IsVoid())

		tag, err = p.afterTagger(ctx, w, tag)
		if err != nil {
			return tag, err
		}

		return tag, nil
	//
	// Lists, the following will all eventually be sent to the above simple node or Tagger cases
	//
	case *NodeGroup:
		if v == nil || len(v.Get()) == 0 {
			return nil, nil
		}

		var (
			list     = v.Get()
			newGroup []any

			thisNodeStr   string
			thisNodeIsStr bool
			lastNodeStr   string
			lastNodeIsStr bool
		)

		for i := 0; i < len(list); i++ {
			if list[i] == nil {
				continue
			}

			node, err := p.walk(ctx, w, list[i])
			if err != nil {
				return nil, err
			}

			if node == nil {
				continue
			}

			// Combine consecutive strings
			thisNodeStr, thisNodeIsStr = node.(string)

			// Combine strings like a browser would
			if lastNodeIsStr && thisNodeIsStr && len(newGroup) > 0 {
				// update this in case we have another string
				thisNodeStr = lastNodeStr + thisNodeStr
				// replace last node
				newGroup[len(newGroup)-1] = thisNodeStr
			} else {
				newGroup = append(newGroup, node)
			}
			// Update state for the next loop
			lastNodeStr, lastNodeIsStr = thisNodeStr, thisNodeIsStr
		}

		return newGroup, nil
	case []Componenter:
		g := G()

		for i := 0; i < len(v); i++ {
			g.Add(v[i])
		}

		return p.walk(ctx, w, g)
	case []Tagger:
		g := G()

		for i := 0; i < len(v); i++ {
			g.Add(v[i])
		}

		return p.walk(ctx, w, g)
	case []UniqueTagger:
		g := G()

		for i := 0; i < len(v); i++ {
			g.Add(v[i])
		}

		return p.walk(ctx, w, g)
	case []*Component:
		g := G()

		for i := 0; i < len(v); i++ {
			g.Add(v[i])
		}

		return p.walk(ctx, w, g)
	case []*Tag:
		g := G()

		for i := 0; i < len(v); i++ {
			g.Add(v[i])
		}

		return p.walk(ctx, w, g)
	default:
		return nil, fmt.Errorf("pileline.walk: node: %#v: %w", v, ErrRenderElement)
	}
}
