// elementVisible
// Register plugin
if (hlive.afterMessage.get("hElementVisible") === undefined) {
    hlive.afterMessage.set("hElementVisible", function () {
        // set global up observer
        let options = {
            root: null,
            rootMargin: "0px",
            threshold: 0,
        };

        // Send new data to all observed elements
        const callback = function (entries, observer) {
            entries.forEach(function (entity) {
                const ids = hlive.getEventHandlerIDs(entity.target);

                if (!ids["helementvisible"]) {
                    return;
                }

                for (let i = 0; i < ids["helementvisible"].length; i++) {
                    hlive.sendMsg({
                        t: "e",
                        i: ids["helementvisible"][i],
                        e: {
                            "isIntersecting": entity.isIntersecting.toString(),
                            "intersectionRatio": entity.intersectionRatio.toString()
                        }
                    });
                }
            })
        }

        // add elements to observer after messages
        // TODO: disconect/unobserver on unmount?
        let observer = new IntersectionObserver(callback, options);
        return function () {
            // add to observer
            document.querySelectorAll("[hon*=helementvisible]").forEach(function (el) {
                observer.observe(el);
            })
        }
    }());
}