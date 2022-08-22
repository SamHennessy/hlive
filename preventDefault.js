// Prevent Default
// Register plugin
if (hlive.beforeSendEvent.get("hpd") === undefined) {
    hlive.beforeSendEvent.set("hpd", function (e, msg) {
        if (!e.currentTarget || !e.currentTarget.getAttribute) {
            return msg;
        }

        if (!e.currentTarget.hasAttribute("data-hlive-pd")) {
            return msg;
        }

        e.preventDefault();

        return msg;
    })
}
