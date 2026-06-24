package hlive

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"

	"log/slog"

	"github.com/vmihailenco/msgpack/v5"
)

const (
	PageHashAttr     = "data-hlive-hash"
	PageHashAttrTmpl = "{data-hlive-hash}"
)

const (
	msgpackExtHTML int8 = iota
	msgpackExtTag
	msgpackExtAttr
	msgpackExtNodeGroup
)

func init() {
	msgpack.RegisterExt(msgpackExtHTML, (*HTML)(nil))
	msgpack.RegisterExt(msgpackExtTag, (*Tag)(nil))
	msgpack.RegisterExt(msgpackExtAttr, (*Attribute)(nil))
	msgpack.RegisterExt(msgpackExtNodeGroup, (*NodeGroup)(nil))
}

// Cache allow cache adapters to be used in HLive
type Cache interface {
	Get(key any) (value any, hit bool)
	Set(key any, value any)
}

// PipelineProcessorRenderHashAndCache that will cache the returned tree to support SSR
func PipelineProcessorRenderHashAndCache(logger *slog.Logger, renderer *Renderer, cache Cache) *PipelineProcessor {
	pp := NewPipelineProcessor(PipelineProcessorKeyRenderer)

	pp.AfterWalk = func(ctx context.Context, w io.Writer, node *NodeGroup) (*NodeGroup, error) {
		byteBuf := bytes.NewBuffer(nil)
		hasher := sha256.New()
		multiW := io.MultiWriter(byteBuf, hasher)

		if err := renderer.HTML(multiW, node); err != nil {
			return node, fmt.Errorf("renderer.HTML: %w", err)
		}

		doc := byteBuf.Bytes()
		hhash := fmt.Sprintf("%x", hasher.Sum(nil))
		// Add hhash to the output
		doc = bytes.Replace(doc, []byte(PageHashAttrTmpl), []byte(hhash), 1)

		if nodeBytes, err := msgpack.Marshal(node); err != nil {
			logger.Error("PipelineProcessorRenderHashAndCache: msgpack.Marshal", "error", err)
		} else {
			cache.Set(hhash, nodeBytes)
			logger.Debug("cache set", "hhash", hhash, "size", len(nodeBytes)/1024)
		}

		if _, err := w.Write(doc); err != nil {
			return node, fmt.Errorf("write doc: %w", err)
		}

		return node, nil
	}

	return pp
}
