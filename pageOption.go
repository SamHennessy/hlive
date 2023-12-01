package hlive

import "github.com/cornelk/hashmap"

type PageOption func(*Page)

func PageOptionCache(cache Cache) func(*Page) {
	return func(p *Page) {
		p.cache = cache
	}
}

func PageOptionEventBindingCache(m *hashmap.Map[string, *EventBinding]) func(*Page) {
	return func(page *Page) {
		page.eventBindings = m
	}
}
