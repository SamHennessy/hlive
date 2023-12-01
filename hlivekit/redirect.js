// Client side redirect
// Register plugin
if (hlive.afterMessage.get("hredi") === undefined) {
    hlive.afterMessage.set("hredi", function () {
        document.querySelectorAll("[data-redirect]").forEach(function (el) {
            window.location.replace(el.getAttribute("data-redirect"));
        });
    });
}
