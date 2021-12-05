package hlive_test

import l "github.com/SamHennessy/hlive"

type testTagger struct{}

func (t *testTagger) GetName() string {
	return ""
}

func (t *testTagger) GetAttributes() []l.Attributer {
	return nil
}

func (t *testTagger) GetNodes() *l.NodeGroup {
	return nil
}

func (t *testTagger) IsVoid() bool {
	return false
}

func (t *testTagger) IsNil() bool {
	return t == nil
}

type testUniqueTagger struct {
	testTagger
}

func (t *testUniqueTagger) GetID() string {
	return ""
}

type testComponenter struct {
	testUniqueTagger
}

func (c *testComponenter) GetEventBinding(id string) *l.EventBinding {
	return nil
}

func (c *testComponenter) GetEventBindings() []*l.EventBinding {
	return nil
}
func (c *testComponenter) RemoveEventBinding(id string) {}
func (c *testComponenter) IsAutoRender() bool {
	return false
}
