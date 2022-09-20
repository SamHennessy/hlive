package main

import (
	"context"
	"log"
	"net/http"
	"regexp"

	l "github.com/SamHennessy/hlive"
	"github.com/SamHennessy/hlive/hlivekit"
)

func main() {
	http.Handle("/", l.NewPageServer(home))

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve: ", err)
	}
}

const (
	pstInputInvalid  = "input_invalid"
	pstInputValid    = "input_valid"
	pstFormValidate  = "form_validate"
	pstFormInvalid   = "form_invalid"
	pstFormSubmit    = "form_submit"
	pstFormSubmitted = "form_summited"
)

// Source: https://golangcode.com/validate-an-email-address/
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type inputValue struct {
	name  string
	value *l.NodeBox[string]
	error *l.NodeBox[string]
}

func newInputValue(name string) inputValue {
	return inputValue{
		name:  name,
		value: l.Box(""),
		error: l.Box(""),
	}
}

func home() *l.Page {
	pubSub := hlivekit.NewPubSub()

	page := l.NewPage()
	page.DOM().HTML().Add(hlivekit.InstallPubSub(pubSub))
	page.DOM().Title().Add("PubSub Example")
	page.DOM().Head().Add(l.T("link",
		l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

	page.DOM().Body().Add(
		l.T("h1", "PubSub"),
		l.T("blockquote", "Use the PubSub system to allow for decoupled components."),
		l.T("hr"),
		newErrorMessages(),
		newUserForm(
			newInputName(),
			newInputEmail(),
			l.C("button", "Submit"),
			l.T("p", "*Required"),
		),
		newFormOutput(),
	)

	return page
}

//
// Components
//

// Error messages

func newErrorMessages() *errorMessages {
	c := &errorMessages{
		Component:  l.C("div"),
		inputMap:   map[string]inputValue{},
		errMessage: l.Box(""),
	}

	c.initDOM()

	return c
}

type errorMessages struct {
	*l.Component

	pubSub     *hlivekit.PubSub
	errMessage *l.NodeBox[string]
	inputs     []string
	inputMap   map[string]inputValue
}

func (c *errorMessages) PubSubMount(_ context.Context, pubSub *hlivekit.PubSub) {
	c.pubSub = pubSub

	// Track input updates
	pubSub.Subscribe(hlivekit.NewSub(c.onInput), pstInputInvalid, pstInputValid)

	// Reset
	pubSub.Subscribe(hlivekit.NewSub(c.onFormValidate), pstFormValidate)
}

func (c *errorMessages) onFormValidate(_ hlivekit.QueueMessage) {
	c.inputs = nil
	c.inputMap = map[string]inputValue{}
	c.errMessage.Set("")
	c.Add(l.Attrs{"display": "none"})
}

func (c *errorMessages) onInput(message hlivekit.QueueMessage) {
	input, ok := message.Value.(inputValue)
	if !ok {
		return
	}

	_, exists := c.inputMap[input.name]

	if !exists {
		c.inputs = append(c.inputs, input.name)
	}

	c.inputMap[input.name] = input

	c.formatErrMessage()
}

func (c *errorMessages) initDOM() {
	c.Add(l.ClassBool{"card": true}, l.Style{"display": "none"},
		l.T("h4", "Errors"),
		l.T("hr"),
		l.T("p", c.errMessage),
	)
}

func (c *errorMessages) formatErrMessage() {
	c.errMessage.Set("")
	for i := 0; i < len(c.inputs); i++ {
		if c.inputMap[c.inputs[i]].error.Get() != "" {
			c.errMessage.Lock(func(v string) string {
				return v + c.inputMap[c.inputs[i]].error.Get() + " "
			})
		}
	}

	if c.errMessage.Get() == "" {
		c.Add(l.Style{"display": "none"})
	} else {
		c.Add(l.StyleOff{"display"})
	}
}

// User form

func newUserForm(nodes ...any) *userForm {
	c := &userForm{
		Component: l.C("form", nodes...),
	}

	c.initDOM()

	return c
}

type userForm struct {
	*l.Component

	isInvalid bool
	pubSub    *hlivekit.PubSub
}

func (c *userForm) PubSubMount(_ context.Context, pubSub *hlivekit.PubSub) {
	c.pubSub = pubSub

	// If any errors, then we can't submit
	pubSub.Subscribe(hlivekit.NewSub(func(message hlivekit.QueueMessage) {
		c.isInvalid = true
	}), pstInputInvalid)
}

func (c *userForm) initDOM() {
	// Revalidate the form, if invalid then submit
	c.Add(l.On("submit", c.onSubmit))
}

func (c *userForm) onSubmit(_ context.Context, _ l.Event) {
	c.isInvalid = false

	// Revalidate form
	c.pubSub.Publish(pstFormValidate, nil)

	if c.isInvalid {
		c.pubSub.Publish(pstFormInvalid, nil)

		return
	}

	c.pubSub.Publish(pstFormSubmit, nil)
}

// Input, name

func newInputName() *inputName {
	c := &inputName{
		Component: l.NewComponent("span"),
		input:     newInputValue("name"),
	}

	c.initDOM()

	return c
}

type inputName struct {
	*l.Component

	pubSub      *hlivekit.PubSub
	input       inputValue
	firstChange bool
}

func (c *inputName) PubSubMount(_ context.Context, pubSub *hlivekit.PubSub) {
	c.pubSub = pubSub

	c.pubSub.Subscribe(hlivekit.NewSub(c.onFormValidate), pstFormValidate)
}

func (c *inputName) initDOM() {
	c.Add(
		l.T("label", "Name*"),

		l.C("input",
			l.Attrs{"name": "name", "placeholder": "Your name"},
			l.AttrsLockBox{"value": c.input.value.LockBox},
			l.On("input", c.onInput),
			l.On("change", c.onChange),
		),

		l.T("p", l.Style{"color": "red"}, c.input.error),
	)
}

func (c *inputName) onFormValidate(_ hlivekit.QueueMessage) {
	c.firstChange = true
	c.validate()
}

func (c *inputName) onChange(ctx context.Context, e l.Event) {
	c.firstChange = true
	c.onInput(ctx, e)
}

func (c *inputName) onInput(_ context.Context, e l.Event) {
	c.input.value.Set(e.Value)

	if c.firstChange {
		c.validate()
	}
}

func (c *inputName) validate() {
	c.input.error.Set("")

	if c.input.value.Get() == "" {
		c.input.error.Set("Name is required.")
		c.pubSub.Publish(pstInputInvalid, c.input)

		return
	}

	if len([]rune(c.input.value.Get())) < 2 {
		c.input.error.Set("Name is too short.")
		c.pubSub.Publish(pstInputInvalid, c.input)

		return
	}

	c.pubSub.Publish(pstInputValid, c.input)
}

// Input, email

func newInputEmail() *inputEmail {
	c := &inputEmail{
		Component: l.NewComponent("span"),
		input:     newInputValue("email"),
	}

	c.initDOM()

	return c
}

type inputEmail struct {
	*l.Component

	pubSub      *hlivekit.PubSub
	input       inputValue
	firstChange bool
}

func (c *inputEmail) PubSubMount(_ context.Context, pubSub *hlivekit.PubSub) {
	c.pubSub = pubSub

	c.pubSub.Subscribe(hlivekit.NewSub(c.onFormValidate), pstFormValidate)
}

func (c *inputEmail) initDOM() {
	c.Add(
		l.T("label", "Email"),

		l.C("input",
			l.Attrs{"name": "email", "placeholder": "Your email address"},
			l.AttrsLockBox{"value": c.input.value.LockBox},
			l.On("input", c.onInput),
			l.On("change", c.onChange),
		),

		l.T("p", l.Style{"color": "red"}, c.input.error),
	)
}

func (c *inputEmail) onFormValidate(_ hlivekit.QueueMessage) {
	c.firstChange = true
	c.validate()
}

func (c *inputEmail) onChange(ctx context.Context, e l.Event) {
	c.firstChange = true
	c.onInput(ctx, e)
}

func (c *inputEmail) onInput(_ context.Context, e l.Event) {
	c.input.value.Set(e.Value)

	if c.firstChange {
		c.validate()
	}
}

func (c *inputEmail) validate() {
	c.input.error.Set("")

	if len(c.input.value.Get()) != 0 && !emailRegex.MatchString(c.input.value.Get()) {
		c.input.error.Set("Email address not valid.")
		c.pubSub.Publish(pstInputInvalid, c.input)

		return
	}

	c.pubSub.Publish(pstInputValid, c.input)
}

// Form output

func newFormOutput() *formOutput {
	c := &formOutput{
		Component: l.C("table"),
		inputs:    map[string]inputValue{},
		list:      hlivekit.List("tbody"),
	}

	c.Add(
		l.Style{"display": "none"},
		l.T("thead",
			l.T("tr",
				l.T("th", "Key"),
				l.T("th", "Value"),
			),
		),
		c.list,
	)

	return c
}

type formOutput struct {
	*l.Component

	list   *hlivekit.ComponentList
	pubSub *hlivekit.PubSub
	inputs map[string]inputValue
}

func (c *formOutput) PubSubMount(_ context.Context, pubSub *hlivekit.PubSub) {
	c.pubSub = pubSub

	c.pubSub.Subscribe(hlivekit.NewSub(c.onValidInput), pstInputValid)
	c.pubSub.Subscribe(hlivekit.NewSub(c.onSubmitForm), pstFormSubmit)
}

func (c *formOutput) onValidInput(item hlivekit.QueueMessage) {
	if input, ok := item.Value.(inputValue); ok {
		c.inputs[input.name] = input
	}
}

func (c *formOutput) onSubmitForm(_ hlivekit.QueueMessage) {
	c.Add(l.StyleOff{"display"})
	c.list.RemoveAllItems()
	for key, input := range c.inputs {
		c.list.Add(
			l.CM("tr",
				l.T("td", key),
				l.T("td", input.value),
			),
		)
	}

	c.pubSub.Publish(pstFormSubmitted, nil)
}
