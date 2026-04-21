package hlive

import "sync"

type PageOption func(*Page)

func PageOptionCache(cache Cache) func(*Page) {
	return func(p *Page) {
		p.cache = cache
	}
}

func PageOptionEventBindingCache(m *sync.Map) func(*Page) {
	return func(page *Page) {
		page.eventBindings = m
	}
}
