// Stop Propagation
// Register plugin
if (hlive.beforeSendEvent.get("hsp") === undefined) {
    hlive.beforeSendEvent.set("hsp", function (e, msg) {
        if (!e.currentTarget || !e.currentTarget.getAttribute) {
            return msg;
        }

        if (!e.currentTarget.hasAttribute("data-hlive-sp")) {
            return msg;
        }

        e.stopPropagation()

        return msg;
    })
}