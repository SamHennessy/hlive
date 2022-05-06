// Stop Propagation
// Register plugin
hlive.beforeSendEvent.push(function (e, msg) {
    if (!e.currentTarget || !e.currentTarget.getAttribute) {
        return msg;
    }

    if (!e.currentTarget.hasAttribute("data-hlive-sp")) {
        return msg;
    }

    e.stopPropagation()

    return msg;
})