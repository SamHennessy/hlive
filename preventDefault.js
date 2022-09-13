// Prevent Default
// Register plugin
if (hlive.beforeSendEvent.get("hpd") === undefined) {
    hlive.beforeSendEvent.set("hpd", function (e, msg) {
        if (!e.currentTarget || !e.currentTarget.hasAttribute) {
            return msg;
        }

        if (e.currentTarget.hasAttribute("data-hlive-pd")) {
            e.preventDefault();
        }

        return msg;
    })
}
