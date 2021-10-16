// Give focus
// Register plugin
hlive.afterMessage.push(function() {
    document.querySelectorAll("[data-hlive-focus]").forEach(function (el) {
        el.focus();
        if (el.selectionStart !== undefined) {
            setTimeout(function(){ el.selectionStart = el.selectionEnd = 10000; }, 0);
        }
    });
});
