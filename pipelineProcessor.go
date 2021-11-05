package hlive

import (
	"context"
	"io"
	"strconv"
	"strings"
	"sync"
)

const (
	PipelineProcessorKeyStripHLiveAttrs      = "hlive_strip_hlive_attr"
	PipelineProcessorKeyRenderer             = "hlive_renderer"
	PipelineProcessorKeyEventBindingCache    = "hlive_eb"
	PipelineProcessorKeyAttributePluginMount = "hlive_attr_mount"
	PipelineProcessorKeyPubSubMount          = "hlive_ps_mount"
	PipelineProcessorKeyMount                = "hlive_mount"
	PipelineProcessorKeyUnmount              = "hlive_unmount"
)

type PipelineProcessor struct {
	// Will replace an existing processor with the same key. An empty string won't error.
	Key             string
	Disabled        bool
	BeforeWalk      PipeNodeHandler
	OnSimpleNode    PipeNodeHandler
	BeforeTagger    PipeTaggerHandler
	BeforeAttribute PipeAttributerHandler
	AfterAttribute  PipeAttributerHandler
	AfterTagger     PipeTagHandler
	AfterWalk       PipeNodegroupHandler
}

func NewPipelineProcessor(key string) *PipelineProcessor {
	return &PipelineProcessor{Key: key}
}

func PipelineProcessorStripHLiveAttrs() *PipelineProcessor {
	pp := NewPipelineProcessor(PipelineProcessorKeyStripHLiveAttrs)

	pp.AfterTagger = func(ctx context.Context, w io.Writer, tag *Tag) (*Tag, error) {
		for _, attr := range tag.GetAttributes() {
			if strings.HasPrefix(attr.GetAttribute().Name, "data-hlive") {
				tag.RemoveAttributes(attr.GetAttribute().Name)
			}
		}

		return tag, nil
	}

	return pp
}

func PipelineProcessorEventBindingCache(cache map[string]*EventBinding) *PipelineProcessor {
	pp := NewPipelineProcessor(PipelineProcessorKeyEventBindingCache)

	var lock sync.Mutex

	pp.BeforeWalk = func(ctx context.Context, w io.Writer, node interface{}) (interface{}, error) {
		lock.Lock()
		for key := range cache {
			delete(cache, key)
		}
		lock.Unlock()

		return node, nil
	}

	pp.BeforeTagger = func(ctx context.Context, w io.Writer, tag Tagger) (Tagger, error) {
		if comp, ok := tag.(Componenter); ok {
			bindings := comp.GetEventBindings()

			lock.Lock()
			for i := 0; i < len(bindings); i++ {
				if _, exists := cache[bindings[i].ID]; !exists {
					cache[bindings[i].ID] = bindings[i]
				}
			}
			lock.Unlock()
		}

		return tag, nil
	}

	return pp
}

func PipelineProcessorMount() *PipelineProcessor {
	pp := NewPipelineProcessor(PipelineProcessorKeyMount)

	cache := map[string]struct{}{}

	// TODO: temp
	var lock sync.Mutex

	pp.BeforeTagger = func(ctx context.Context, w io.Writer, tag Tagger) (Tagger, error) {
		if comp, ok := tag.(Mounter); ok {
			lock.Lock()
			if _, exists := cache[comp.GetID()]; !exists {
				comp.Mount(ctx)

				cache[comp.GetID()] = struct{}{}
			}
			lock.Unlock()

			// A way to remove the key when you delete a Component
			if comp, ok := tag.(Teardowner); ok {
				comp.AddTeardown(func() {
					lock.Lock()
					delete(cache, comp.GetID())
					lock.Unlock()
				})
			}
		}

		return tag, nil
	}

	return pp
}

func PipelineProcessorUnmount(page *Page) *PipelineProcessor {
	cache := map[string]Unmounter{}

	// TODO: temp
	var lock sync.Mutex

	page.HookClose = append(page.HookClose, func(ctx context.Context, page *Page) {
		for _, c := range cache {
			if c == nil {
				continue
			}

			c.Unmount(ctx)
		}
	})

	pp := NewPipelineProcessor(PipelineProcessorKeyUnmount)

	pp.BeforeTagger = func(ctx context.Context, w io.Writer, tag Tagger) (Tagger, error) {
		if comp, ok := tag.(Unmounter); ok {
			lock.Lock()
			if _, exists := cache[comp.GetID()]; !exists {
				cache[comp.GetID()] = comp
			}
			lock.Unlock()

			// A way to remove the key when you delete a Component
			if comp, ok := tag.(Teardowner); ok {
				comp.AddTeardown(func() {
					lock.Lock()
					delete(cache, comp.GetID())
					lock.Unlock()
				})
			}
		}

		return tag, nil
	}

	return pp
}

func PipelineProcessorRenderer(renderer *Renderer) *PipelineProcessor {
	pp := NewPipelineProcessor(PipelineProcessorKeyRenderer)

	pp.AfterWalk = func(ctx context.Context, w io.Writer, node *NodeGroup) (*NodeGroup, error) {
		return node, renderer.HTML(w, node)
	}

	return pp
}

func PipelineProcessorConvertToString() *PipelineProcessor {
	pp := NewPipelineProcessor(PipelineProcessorKeyStripHLiveAttrs)

	pp.OnSimpleNode = func(ctx context.Context, w io.Writer, node interface{}) (interface{}, error) {
		switch v := node.(type) {
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
		default:
			return v, nil
		}
	}

	return pp
}

func PipelineProcessorAttributePluginMount(page *Page) *PipelineProcessor {
	pp := NewPipelineProcessor(PipelineProcessorKeyAttributePluginMount)

	// TODO: ????
	// cache, will plugins be okay with running for SSR, Simulated SSR and DOM init?
	// cache := map[string]struct{}{}

	pp.BeforeAttribute = func(ctx context.Context, w io.Writer, attr Attributer) (Attributer, error) {

		var err error
		if ap, ok := attr.(AttributePluginer); ok {
			if _, exits := page.attributePluginMountedMap[ap.GetAttribute().Name]; !exits {
				// TODO: need to be sure we have the page exclusively
				page.attributePluginMountedMap[ap.GetAttribute().Name] = struct{}{}

				ap.Initialize(page)

				err = ErrDOMInvalidated
			}
		}

		return attr, err
	}

	return pp
}
