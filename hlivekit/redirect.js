// Client side redirect
// Register plugin
hlive.afterMessage.push(function() {
    document.querySelectorAll("[data-redirect]").forEach(function (el) {
        window.location.replace(el.getAttribute("data-redirect"));
    });
});
