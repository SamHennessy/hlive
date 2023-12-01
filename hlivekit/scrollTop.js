// Set scrollTop
// Register plugin
if (hlive.afterMessage.get("hscrollTop") === undefined) {
    hlive.afterMessage.set("hscrollTop", function () {
        document.querySelectorAll("[data-scrollTop]").forEach(function (el) {
            el.scrollTop = Number(el.getAttribute("data-scrollTop"));
        });
    });
}
