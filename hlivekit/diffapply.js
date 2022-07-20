// Trigger diffapply, should always be last
function diffApply() {
    document.querySelectorAll("[data-hlive-on*=diffapply]").forEach(function (el) {
        const ids = hlive.getEventHandlerIDs(el);

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