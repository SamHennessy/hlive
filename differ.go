package hlive

import (
	_ "embed"
	"fmt"
	"strconv"

	"github.com/rs/zerolog"
)

//go:embed page.js
var PageJavaScript []byte

// Diff Diffs are from old to new
type Diff struct {
	// Root element, where to start the path search from, "doc" is a special case that means the browser document
	Root string
	// Position of each child
	Path      string
	Type      DiffType
	Tag       Tagger
	Text      *string
	Attribute *Attribute
	HTML      *HTML
	// Not used for render but for Lifecycle events
	Old interface{}
}

func NewDiffer() *Differ {
	return &Differ{
		logger:     zerolog.Nop(),
		JavaScript: PageJavaScript,
	}
}

type Differ struct {
	logger     zerolog.Logger
	JavaScript []byte
}

func (d *Differ) SetLogger(logger zerolog.Logger) {
	d.logger = logger
}

// Trees diff two node tress
//
// Path: childIndex>childIndex
// Path: 0>1>3
//
// After tree copy you only have Tagger (with []Attribute), HTML, and strings. Then can be grouped in a NodeGroup
func (d *Differ) Trees(selector, path string, oldNode, newNode interface{}) ([]Diff, error) {
	var diffs []Diff

	d.logger.Trace().Str("sel", selector).Str("path", path).Msg("diffTrees")

	// More nodes in new browserTree
	if oldNode == nil && newNode != nil {
		diffs = append(diffs, diffCreate(selector, path, newNode)...)

		return diffs, nil
	}

	// Old node doesn't exist in new browserTree
	if oldNode != nil && newNode == nil {
		diffs = append(diffs, Diff{
			Root: selector,
			Path: path,
			Type: DiffDelete,
			Old:  oldNode,
		})

		return diffs, nil
	}

	// Not the same type, remove current node and replace with new
	if !diffTreeNodeTypeMatch(oldNode, newNode) {

		diffs = append(diffs, Diff{
			Root: selector,
			Path: path,
			Type: DiffDelete,
			Old:  oldNode,
		})

		diffs = append(diffs, diffCreate(selector, path, newNode)...)

		return diffs, nil
	}

	switch v := oldNode.(type) {
	case *NodeGroup:
		oldList := v.Get()
		newNG, _ := newNode.(*NodeGroup)
		newList := newNG.list
		indexOffset := 0

		for i := 0; i < len(oldList); i++ {
			var n interface{}
			if len(newList) > i {
				n = newList[i]
			}

			subDiffs, err := d.Trees(selector, path+strconv.Itoa(i-indexOffset), oldList[i], n)
			if err != nil {
				return nil, fmt.Errorf("diff NodeGroup: %w", err)
			}

			diffs = append(diffs, subDiffs...)
		}
		// Any new elements?
	case string:
		newStr, _ := newNode.(string)

		// content doesn't match, update content
		if v != newStr {
			diffs = append(diffs, Diff{
				Root: selector,
				Path: path,
				Type: DiffUpdate,
				Text: &newStr,
			})
		}
	case HTML:
		newHTML, _ := newNode.(HTML)

		// content doesn't match, update content
		if v != newHTML {
			diffs = append(diffs, Diff{
				Root: selector,
				Path: path,
				Type: DiffUpdate,
				HTML: &newHTML,
			})
		}
	case Tagger:
		newTag, _ := newNode.(Tagger)

		// Different tag?
		if v.GetName() != newTag.GetName() || v.IsVoid() != newTag.IsVoid() {
			diffs = append(diffs, Diff{
				Root: selector,
				Path: path,
				Type: DiffUpdate,
				Tag:  newTag,
			})

			return diffs, nil
		}

		// Attributes
		// The browser doesn't care about the order as we use setAttribute and removeAttribute. It would be

		oldAttrs := v.GetAttributes()
		newAttrs := newTag.GetAttributes()

		// exits maps, helps us know if we should delete or update
		oldAttrsMap := map[string]Attributer{}
		for i := 0; i < len(oldAttrs); i++ {
			oldAttrsMap[oldAttrs[i].GetAttribute().Name] = oldAttrs[i]
		}

		newAttrsMap := map[string]Attributer{}

		for i := 0; i < len(newAttrs); i++ {
			newAttrsMap[newAttrs[i].GetAttribute().Name] = newAttrs[i]
		}

		// Update existing or create new
		for i := 0; i < len(newAttrs); i++ {
			oldAttr, exits := oldAttrsMap[newAttrs[i].GetAttribute().Name]

			if !exits || newAttrs[i].GetAttribute().GetValue() != oldAttr.GetAttribute().GetValue() {
				dt := DiffUpdate
				if !exits {
					dt = DiffCreate
				}

				diffs = append(diffs, Diff{
					Root:      selector,
					Path:      path,
					Type:      dt,
					Attribute: newAttrs[i].GetAttribute(),
				})
			}
		}

		// Delete old attrs that have been removed
		for i := 0; i < len(oldAttrs); i++ {
			_, exits := newAttrsMap[oldAttrs[i].GetAttribute().Name]
			if !exits {
				diffs = append(diffs, Diff{
					Root:      selector,
					Path:      path,
					Type:      DiffDelete,
					Attribute: oldAttrs[i].GetAttribute(),
				})
			}
		}

		// Is this tag a component?
		if attr, exits := newAttrsMap[AttrID]; exits {
			selector = attr.GetAttribute().GetValue()
			path = ""
		}

		oldKids := v.GetNodes().Get()
		newKids := newTag.GetNodes().Get()

		// Loop old kids
		i := 0
		for ; i < len(oldKids); i++ {
			var newKid interface{}
			if i < len(newKids) {
				newKid = newKids[i]
			}

			kidDiffs, err := d.Trees(selector, path+">"+strconv.Itoa(i), oldKids[i], newKid)
			if err != nil {
				return nil, fmt.Errorf("tag diff kids: %w", err)
			}

			diffs = append(diffs, kidDiffs...)
		}

		// Any extra new kids?
		for ; i < len(newKids); i++ {
			diffs = append(diffs, diffCreate(selector, path+">"+strconv.Itoa(i), newKids[i])...)
		}
	}

	return diffs, nil
}

func diffCreate(compID, path string, el interface{}) []Diff {
	switch v := el.(type) {
	case *NodeGroup:
		g := v.Get()
		var diffs []Diff
		for i := 0; i < len(g); i++ {
			diffs = append(diffs, diffCreate(compID, path, g[i])...)
		}

		return diffs
	case string:
		return []Diff{
			{
				Root: compID,
				Path: path,
				Type: DiffCreate,
				Text: &v,
			},
		}
	case HTML:
		return []Diff{
			{
				Root: compID,
				Path: path,
				Type: DiffCreate,
				HTML: &v,
			},
		}
	case Tagger:
		return []Diff{
			{
				Root: compID,
				Path: path,
				Type: DiffCreate,
				Tag:  v,
			},
		}
	case *Attribute:

		return []Diff{
			{
				Root:      compID,
				Path:      path,
				Type:      DiffCreate,
				Attribute: v,
			},
		}
	case nil:
		return nil
	default:
		panic(fmt.Errorf("unexpected type: %#v", el))
	}
}

func diffTreeNodeTypeMatch(oldNode, newNode interface{}) bool {
	switch oldNode.(type) {
	case *NodeGroup:
		_, ok := newNode.(*NodeGroup)

		return ok
	case string:
		_, ok := newNode.(string)

		return ok
	case HTML:
		_, ok := newNode.(HTML)

		return ok
	case Tagger:
		_, ok := newNode.(Tagger)

		return ok
	case nil:
		return false
	default:
		panic(fmt.Errorf("unexpected type: %#v", oldNode))
	}
}
