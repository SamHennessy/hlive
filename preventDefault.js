// Prevent Default
// Register plugin
hlive.beforeSendEvent.push(function (e, msg) {
    if (!e.currentTarget || !e.currentTarget.getAttribute) {
        return msg;
    }

    if (!e.currentTarget.hasAttribute("data-hlive-pd")) {
        return msg;
    }

    e.preventDefault()

    return msg;
})