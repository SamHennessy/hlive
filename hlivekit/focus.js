// Give focus
// Register plugin
if (hlive.afterMessage.get("hfocue") === undefined) {
    hlive.afterMessage.set("hfocus", function () {
        document.querySelectorAll("[data-hlive-focus]").forEach(function (el) {
            el.focus();
            if (el.selectionStart !== undefined) {
                setTimeout(function () {
                    el.selectionStart = el.selectionEnd = 10000;
                }, 0);
            }
        });
    });
}
