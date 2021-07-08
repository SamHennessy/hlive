package main

import (
	"context"
	"net/http"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel).With().Timestamp().Logger()

	http.Handle("/", l.NewPageServer(home(logger)))

	logger.Info().Str("addr", ":3000").Msg("listing")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		logger.Err(err).Msg("http listen and serve")
	}
}

func home(logger zerolog.Logger) func() *l.Page {
	return func() *l.Page {
		page := l.NewPage()
		page.SetLogger(logger)
		page.Title.Add("Add Remove Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		page.Body.Add(newTodoApp().tree)

		return page
	}
}

type todoApp struct {
	newTask      string
	newTaskInput *l.Component
	taskList     *l.ComponentListTidy
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
		l.On("keyup", func(_ context.Context, e l.Event) {
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
		l.T("div", "Task Label"),
		a.newTaskInput,
		l.T("button", "Add"),
	)

	a.tree = append(a.tree, f)
}

func (a *todoApp) initList() {
	a.taskList = l.List("div")
	a.tree = append(a.tree,
		l.T("h1", "To Do"),
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

		labelInput,
		l.T("button", "Update"),
	)

	container.Add(
		// Delete button
		l.C("button", "🗑️", l.On("click", func(_ context.Context, _ l.Event) {
			a.taskList.RemoveItem(container)
		})),
		// Edit button
		l.C("button", "✏️", l.On("click", func(_ context.Context, _ l.Event) {
			labelSpan.Add(l.Style{"display": "none"})
			labelForm.Add(l.Style{"display": nil})
			labelInput.Add(l.Attrs{l.AttrFocus: ""}, l.OnOnce("focus", func(_ context.Context, _ l.Event) {
				labelInput.RemoveAttributes(l.AttrFocus)
			}))
		})),
		labelSpan,
		labelForm,
	)

	a.taskList.AddItem(container)
}
