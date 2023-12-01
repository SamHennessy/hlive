// HLive Preempt Disable
function hlivePreDisable (e) {
    if (!e.currentTarget) {
        return
    }

    const el = e.currentTarget;
    el.setAttribute("disabled", "");
}

if (hlive.afterMessage.get("hpreDis") === undefined) {
    hlive.afterMessage.set("hpreDis", function () {
        document.querySelectorAll("[data-hlive-pre-disable]").forEach(function (el) {
            el.addEventListener(el.getAttribute("data-hlive-pre-disable"), hlivePreDisable)
        });
    });
}
