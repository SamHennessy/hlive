package hlivekit

import l "github.com/SamHennessy/hlive"

// DiffApply is a special event that will trigger when a diff is applied.
// This means that it will trigger itself when first added. This will allow you to know when a change in the tree has
// made it to the browser. You can then, if you wish, immediately remove it from the tree to prevent more triggers.
// You can also add it as a OnOnce and it wil remove itself.

func OnDiffApply(handler l.EventHandler) *l.ElementGroup {
	eb := l.On(DiffApplyEvent, handler)
	attr := &DiffApplyAttribute{
		l.NewAttribute(DiffApplyAttributeName, eb.ID),
	}

	return l.E(eb, attr)
}

func OnDiffApplyOnce(handler l.EventHandler) *l.ElementGroup {
	eb := l.On(DiffApplyEvent, handler)
	eb.Once = true
	attr := &DiffApplyAttribute{
		l.NewAttribute(DiffApplyAttributeName, eb.ID),
	}

	return l.E(eb, attr)
}

const (
	DiffApplyEvent         = "diffapply"
	DiffApplyAttributeName = "data-hlive-on-diffapply"
)

type DiffApplyAttribute struct {
	*l.Attribute
}

const diffApplyJS = `
// Trigger diffapply, should always be last
function diffApply() {
    document.querySelectorAll("[data-hlive-on*=diffapply]").forEach(function (el) {
        const ids = hlive.getEventHAndlerIDs(el);

        if (!ids["diffapply"]) {
            return;
        }

        for (let i = 0; i < ids["diffapply"].length; i++) {
            hlive.sendMsg({
                t: "e",
                i: ids["diffapply"][i],
            });
        }
    });
}

// Register plugin
hlive.afterMessage.push(diffApply);
`

func (a *DiffApplyAttribute) Initialize(page *l.Page) {
	page.Head.Add(l.T("script", l.HTML(diffApplyJS)))
}
