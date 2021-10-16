// HLive Preempt Disable
function hlivePreDisable (e) {
    if (!e.currentTarget) {
        return
    }

    const el = e.currentTarget;
    el.setAttribute("disabled", "");
}

hlive.afterMessage.push(function() {
    document.querySelectorAll("[data-hlive-pre-disable]").forEach(function (el) {
        el.addEventListener(el.getAttribute("data-hlive-pre-disable"), hlivePreDisable)
    });
});
