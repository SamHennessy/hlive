// Set scrollTop
// Register plugin
if (hlive.afterMessage.get("hscrollIntoView") === undefined) {
    hlive.afterMessage.set("hscrollIntoView", function () {
        document.querySelectorAll("[data-scrollIntoView]").forEach(function (el) {
            el.scrollIntoView(el.getAttribute("data-scrollIntoView") === "true");
        });
    });
}
