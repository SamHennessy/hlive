package hlive

import (
	"fmt"
	"strconv"

	"github.com/rs/zerolog"
)

// Diff Diffs are from old to new
type Diff struct {
	// Root element, where to start the path search from, "doc" is a special case that means the browser document
	Root string
	// Position of each child
	Path      string
	Type      DiffType
	Tag       TagInterface
	Text      *string
	Attribute *Attribute
	HTML      *HTML
	// Not used for render but for Lifecycle events
	Old interface{}
}

func NewDiffer() *Differ {
	return &Differ{Logger: zerolog.Nop()}
}

type Differ struct {
	Logger zerolog.Logger
}

// Trees
//
// Path: childIndex>childIndex
// Path: 0>1>3
//
// After tree copy you only have TagInterface (with []Attribute), HTML, and strings. Then can be grouped in []interface{}
func (p *Differ) Trees(selector, path string, old, new interface{}) ([]Diff, error) {
	p.Logger.Trace().Str("sel", selector).Str("path", path).Msg("diffTrees")
	var diffs []Diff

	// More nodes in new tree
	if old == nil && new != nil {
		diffs = append(diffs, diffCreate(selector, path, new)...)

		return diffs, nil
	}

	// Old node doesn't exist in new tree
	if old != nil && new == nil {
		diffs = append(diffs, Diff{
			Root: selector,
			Path: path,
			Type: DiffDelete,
			Old:  old,
		})

		return diffs, nil
	}

	// Not the same type, remove current node and replace with new
	// TODO: won't work if old node is a text node
	if !diffTreeNodeTypeMatch(old, new) {

		diffs = append(diffs, Diff{
			Root: selector,
			Path: path,
			Type: DiffDelete,
			Old:  old,
		})

		diffs = append(diffs, diffCreate(selector, path, new)...)

		return diffs, nil
	}

	switch v := old.(type) {
	case []interface{}:
		newIS := new.([]interface{})
		indexOffset := 0
		for i := 0; i < len(v); i++ {
			// Browser don't recognise this in the doc
			// h5dt, ok := v[i].(HTML)
			// if ok && h5dt == HTML5DocType {
			// 	indexOffset++
			// 	continue
			// }

			// TODO: what to do with nils that would mess up the count in the browser?
			// If both elements are nil then increase offset but what if it's a []interface{} that contains
			// Nothing would be rendered, so this would throw off the index that's in the browser
			// Maybe exclude them when doing a tree copy?
			// if v[i] == nil {
			// 	indexOffset++
			// 	continue
			// }

			var n interface{}
			if len(newIS) > i {
				n = newIS[i]
			}

			subDiffs, err := p.Trees(selector, path+strconv.Itoa(i-indexOffset), v[i], n)
			if err != nil {
				return nil, fmt.Errorf("diff []interface{}: %w", err)
			}

			diffs = append(diffs, subDiffs...)
		}
		// Any new elements?
	case string:
		newStr := new.(string)

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
		newHTML := new.(HTML)

		// content doesn't match, update content
		if v != newHTML {
			diffs = append(diffs, Diff{
				Root: selector,
				Path: path,
				Type: DiffUpdate,
				HTML: &newHTML,
			})
		}
	case TagInterface:
		newTag := new.(TagInterface)

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
		//    The browser doesn't care about the order as we use setAttribute and removeAttribute. It would be
		//    possible to ignore the order by sort both sets of attributes first
		// Loop old attrs
		oldAttrs := v.GetAttributes()
		newAttrs := newTag.GetAttributes()

		i := 0
		for ; i < len(oldAttrs); i++ {
			// No new element exists
			if i >= len(newAttrs) {
				diffs = append(diffs, Diff{
					Root:      selector,
					Path:      path,
					Type:      DiffDelete,
					Attribute: oldAttrs[i],
				})

				continue
			}

			if oldAttrs[i].Name != newAttrs[i].Name {
				diffs = append(diffs, Diff{
					Root:      selector,
					Path:      path,
					Type:      DiffDelete,
					Attribute: oldAttrs[i],
				})

				diffs = append(diffs, diffCreate(selector, path, newAttrs[i])...)

				continue
			}

			if !eqStrPtr(oldAttrs[i].Value, newAttrs[i].Value) {
				diffs = append(diffs, Diff{
					Root:      selector,
					Path:      path,
					Type:      DiffUpdate,
					Attribute: newAttrs[i],
					Old:       oldAttrs[i],
				})
			}
		}
		// Any extra new attrs?
		for ; i < len(newAttrs); i++ {
			diffs = append(diffs, diffCreate(selector, path, newAttrs[i])...)
		}

		// Is new tag a component?
		for x := 0; x < len(newAttrs); x++ {
			if newAttrs[x].Name == AttrID && newAttrs[x].Value != nil {
				selector = *newAttrs[x].Value
				path = ""
			}
		}

		oldKids := v.Render()
		newKids := newTag.Render()
		// Loop old kids
		i = 0
		for ; i < len(oldKids); i++ {

			var newKid interface{}
			if i < len(newKids) {
				newKid = newKids[i]
			}

			kidDiffs, err := p.Trees(selector, path+">"+strconv.Itoa(i), oldKids[i], newKid)
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

func eqStrPtr(s1, s2 *string) bool {
	if s1 == nil && s2 == nil {
		return true
	}

	if s1 == nil || s2 == nil {
		return false
	}

	return *s1 == *s2
}

func diffCreate(compID, path string, el interface{}) []Diff {
	switch v := el.(type) {
	case []interface{}:
		var diffs []Diff
		for i := 0; i < len(v); i++ {
			diffs = append(diffs, diffCreate(compID, path, v[i])...)
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
	case TagInterface:
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

func diffTreeNodeTypeMatch(old, new interface{}) bool {
	switch old.(type) {
	case []interface{}:
		_, ok := new.([]interface{})
		return ok
	case string:
		_, ok := new.(string)
		return ok
	case HTML:
		_, ok := new.(HTML)
		return ok
	case TagInterface:
		_, ok := new.(TagInterface)
		return ok
	case nil:
		return false
	default:
		panic(fmt.Errorf("unexpected type: %#v", old))
	}
}
