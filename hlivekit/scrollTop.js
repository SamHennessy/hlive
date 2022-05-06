// Set scrollTop

// Register plugin
hlive.afterMessage.push(function() {
    document.querySelectorAll("[data-scrollTop]").forEach(function (el) {
        el.scrollTop = Number(el.getAttribute("data-scrollTop"));
    });
});
