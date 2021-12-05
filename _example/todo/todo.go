package main

import (
	"context"
	"log"
	"net/http"

	l "github.com/SamHennessy/hlive"
	"github.com/SamHennessy/hlive/hlivekit"
)

func main() {
	http.Handle("/", home())

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve:", err)
	}
}

func home() *l.PageServer {
	f := func() *l.Page {
		page := l.NewPage()
		page.DOM.Title.Add("To Do Example")
		page.DOM.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}))

		page.DOM.Body.Add(
			l.T("header",
				l.T("h1", "To Do App Example"),
				l.T("p", "A simple app where you can add and remove elements"),
			),
			l.T("main",
				newTodoApp().tree,
			),
		)

		return page
	}

	return l.NewPageServer(f)
}

type todoApp struct {
	newTask      string
	newTaskInput *l.Component
	taskList     *hlivekit.ComponentList
	tree         []l.Tagger
}

func newTodoApp() *todoApp {
	app := &todoApp{}
	app.init()

	return app
}

func (a *todoApp) init() {
	a.initForm()
	a.initList()
}

func (a *todoApp) initForm() {
	a.newTaskInput = l.C("input", l.Attrs{"type": "text", "placeholder": "Task E.g: Buy Food, Walk dog, ..."})
	a.newTaskInput.Add(
		l.On("input", func(_ context.Context, e l.Event) {
			a.newTask = e.Value
			// This is needed to allow us to clear the input on submit
			// Without this there would be no difference in the tree to trigger a diff
			a.newTaskInput.Add(l.Attrs{"value": e.Value})
		}),
	)

	f := l.C("form",
		l.On("submit", func(ctx context.Context, _ l.Event) {
			a.addTask(a.newTask)
			// Clear input
			a.newTask = ""
			a.newTaskInput.Add(l.Attrs{"value": ""})
		}),
		l.T("p", "Task Label"),
		a.newTaskInput, " ",
		l.T("button", "Add"),
	)

	a.tree = append(a.tree, f)
}

func (a *todoApp) initList() {
	a.taskList = hlivekit.List("div")
	a.tree = append(a.tree,
		l.T("h3", "To Do List:"),
		a.taskList,
	)
}

func (a *todoApp) addTask(label string) {
	// This is a ComponentMountable. This allows the list to do clean up when we remove it.
	container := l.CM("div")
	labelSpan := l.T("span", &label)

	labelInput := l.C("input",
		l.Attrs{"type": "text", "value": &label},
		l.On("keyup", func(_ context.Context, e l.Event) {
			label = e.Value
		}),
	)

	// Prevent a server side render on each keypress as we don't have anything in the view that updates on keypress of
	// this input
	labelInput.AutoRender = false

	labelForm := l.C("form", l.Style{"display": "none"},
		l.On("submit", func(_ context.Context, e l.Event) {
			// You can get back to the bound component from the event
			lf, ok := e.Binding.Component.(*l.Component)
			if !ok {
				return
			}

			lf.Add(l.Style{"display": "none"})
			labelSpan.Add(l.Style{"display": nil})
		}),

		labelInput, " ",
		l.T("button", "Update"),
	)

	container.Add(
		// Delete button
		l.C("button", "🗑️", l.On("click", func(_ context.Context, _ l.Event) {
			a.taskList.RemoveItems(container)
		})), " ",
		// Edit button
		l.C("button", "✏️", l.On("click", func(_ context.Context, _ l.Event) {
			labelSpan.Add(l.Style{"display": "none"})
			labelForm.Add(l.Style{"display": nil})
			labelInput.Add(hlivekit.Focus(), l.OnOnce("focus", func(ctx context.Context, _ l.Event) {
				hlivekit.FocusRemove(labelInput)

				l.Render(ctx)
			}))
		})), " ",
		labelSpan,
		labelForm,
	)

	a.taskList.AddItem(container)
}
