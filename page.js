let hlive = {
    debug: false,
    reconnectLimit: 0,
    reconnectCount: 0,
    conn: null,
    initSyncDone: false,
    sessID: 1,

    afterMessage: new Map(),
    beforeRemoveEventHandlers: new Map(),
    afterRemoveEventHandlers: new Map(),
    beforeSendEvent: new Map(),
    beforeProcessMessage: new Map(),
};

hlive.msgPart = {
    Type: 0,
}

hlive.diffParts = {
    DiffType: 1, Root: 2, Path: 3, ContentType: 4, Content: 5,
}

// Base64 decode with unicode support
// Ref: https://stackoverflow.com/questions/30106476/using-javascripts-atob-to-decode-base64-doesnt-properly-decode-utf-8-strings
hlive.base64Decode = (str) => {
    // Going backwards: from byte stream, to percent-encoding, to original string.
    return decodeURIComponent(atob(str).split('').map(function (c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));
}

hlive.eventHandler = (e) => {
    if (!e.currentTarget || !e.currentTarget.getAttribute) {
        return
    }

    const pairs = e.currentTarget.getAttribute("hon").split(",");
    for (let i = 0; i < pairs.length; i++) {
        const parts = pairs[i].split("|");
        if (parts[1].toLowerCase() === e.type.toLowerCase()) {
            hlive.eventHandlerHelper(e, parts[0], false);
        }
    }
}

hlive.removeEventHandlers = (el) => {
    hlive.beforeRemoveEventHandlers.forEach(function (fn) {
        fn(el);
    })

    hlive.removeHLiveEventHandlers(el);

    hlive.afterRemoveEventHandlers.forEach(function (fn) {
        fn(el);
    })
}

hlive.removeHLiveEventHandlers = (el) => {
    if (!el.getAttribute) {
        return;
    }

    const val = el.getAttribute("hon");

    if (val === null || val === "") {
        return;
    }

    const pairs = val.split(",");
    for (let i = 0; i < pairs.length; i++) {
        const parts = pairs[i].split("|");
        el.removeEventListener(parts[1].toLowerCase(), hlive.eventHandler);
    }
}

hlive.eventHandlerHelper = (e, handlerID, isInitial) => {
    const el = e.currentTarget;

    let msg = {
        t: "e", i: handlerID,
    };

    let d = {};
    if (el.value !== undefined) {
        d.value = String(el.value);
        if (isInitial) {
            d.init = "true";
        }
    }

    if (el.selected) {
        msg.s = true;
    }

    if (el.checked) {
        msg.s = true;
    }

    if (el.selectedOptions && el.selectedOptions.length !== 0) {
        msg.vm = [];
        for (let i = 0; i < el.selectedOptions.length; i++) {
            const opt = el.selectedOptions[i]
            msg.vm.push(opt.value || opt.text);
        }
    }

    if (e.key !== undefined) {
        d.key = e.key;
        d.charCode = String(e.charCode);
        d.keyCode = String(e.keyCode);
        d.shiftKey = String(e.shiftKey);
        d.altKey = String(e.altKey);
        d.ctrlKey = String(e.ctrlKey);
    }

    if (d.length !== 0) {
        msg.d = d;
    }

    // File
    // TODO: move out to plugin
    if (el.files) {
        // No files
        msg.file = {
            "name": "", "size": 0, "type": "", "index": 0, "total": 0,
        };
        // Single file
        if (el.files.length === 1) {
            msg.file = {
                "name": el.files[0].name, "size": el.files[0].size, "type": el.files[0].type, "index": 0, "total": 1,
            };
        }
        // Multiple files
        // Need to send multiple messages
        if (el.files.length > 1) {
            for (let i = 0; i < el.files.length; i++) {
                msg.file = {
                    "name": el.files[i].name,
                    "size": el.files[i].size,
                    "type": el.files[i].type,
                    "index": i,
                    "total": el.files.length,
                };

                hlive.sendMsg(msg);
            }

            return
        }
    }

    hlive.beforeSendEvent.forEach(function (fn) {
        msg = fn(e, msg);
    });

    hlive.sendMsg(msg);
}

hlive.sendMsg = (msg) => {
    queueMicrotask(function () {
        // https://developer.mozilla.org/en-US/docs/Web/API/WebSocket/readyState
        // TODO: maybe add to a retry queue?
        if (hlive.conn.readyState === 1) {
            hlive.conn.send(JSON.stringify(msg));
        }
    });
}

hlive.log = (message) => {
    if (hlive.debug) {
        console.log(message);
    }

    if (hlive.conn) {
        let msg = {
            t: "l", d: {m: message},
        };
        hlive.sendMsg(msg);
    }
}

hlive.setEventHandlers = () => {
    document.querySelectorAll("[hon]").forEach(function (el) {
        const pairs = el.getAttribute("hon").split(",");
        for (let i = 0; i < pairs.length; i++) {
            const parts = pairs[i].split("|");
            el.addEventListener(parts[1].toLowerCase(), hlive.eventHandler);
        }
    });
}

// Sync Initial Input Values
// Looks at the current value of the input and if needed triggers events to sync that value to the backend
// TODO: move out as plugin?
hlive.initSync = () => {
    document.querySelectorAll("[hon]").forEach(function (el) {
        // Radio
        const elType = el.getAttribute("type");

        if (elType === "radio" || elType === "checkbox") {
            // No sync needed
            if (el.checked === el.hasAttribute("checked")) {
                return;
            }
        } else if (elType === "select") {
            let hit = false;
            for (let i = 0; i < el.selectedOptions.length; i++) {
                if (el.selectedOptions[i].selected && el.hasAttribute("checked") === false) {
                    hit = true;
                    break;
                }
            }
            // No sync needed
            if (hit === false) {
                return;
            }
        } else {
            // How'd this happen :)
            if (el.value === undefined) {
                return;
            }
            // At default
            const aV = el.getAttribute("value")
            // Empty
            if (aV === null && el.value === "") {
                return;
            }
            // Match
            if (aV === el.value) {
                return;
            }
        }

        const pairs = el.getAttribute("hon").split(",");
        for (let i = 0; i < pairs.length; i++) {
            const parts = pairs[i].split("|");
            const name = parts[1].toLowerCase();

            if (name === "keyup" || name === "keydown" || name === "keypress" || name === "input" || name === "change") {
                const evt = {
                    currentTarget: el
                }

                hlive.eventHandlerHelper(evt, parts[0], true);
            }
        }
    });
}

hlive.postMessage = () => {
    hlive.setEventHandlers();

    if (!hlive.initSyncDone && document.querySelectorAll("[hon]").length !== 0) {
        hlive.initSyncDone = true;
        hlive.initSync();
    }

    // Start file upload
    document.querySelectorAll("[data-hlive-upload]").forEach(function (el) {
        const ids = hlive.getEventHandlerIDs(el);

        if (!ids["upload"]) {
            return
        }

        if (el.files.length !== 0) {
            let i = 0;
            const file = el.files[0];

            const fileMeta = {
                "name": file.name, "size": file.size, "type": file.type, "index": i, "total": el.files.length,
            };

            let msg = {
                t: "e", file: fileMeta,
            };

            queueMicrotask(function () {
                for (let j = 0; j < ids["upload"].length; j++) {
                    msg.i = ids["upload"][j];

                    hlive.conn.send(new Blob([JSON.stringify(msg) + "\n\n", el.files[i]], {type: el.files[i].type}));
                }
            });
        }
    });

    hlive.afterMessage.forEach(function (fn) {
        fn();
    })
}

hlive.getEventHandlerIDs = (el) => {
    let map = {};

    if (el.getAttribute && el.getAttribute("hon") !== null) {
        const pairs = el.getAttribute("hon").split(",");
        for (let i = 0; i < pairs.length; i++) {
            const parts = pairs[i].split("|");
            const eventName = parts[1].toLowerCase();
            const eventID = parts[0];

            if (!map[eventName]) {
                map[eventName] = [eventID];
            } else {
                map[eventName].push(eventID);
            }
        }
    }

    return map;
}

hlive.findDiffTarget = (diff) => {
    const parts = diff.split("|");

    let target = document;
    if (parts[hlive.diffParts.Root] !== "doc") {
        target = document.querySelector('[hid="' + parts[hlive.diffParts.Root] + '"]');
    }

    if (!target) {
        hlive.log("root element not found: " + parts[hlive.diffParts.Root]);
        return null
    }

    const path = parts[hlive.diffParts.Path].split(">");

    for (let j = 0; j < path.length; j++) {
        // Doesn't exist
        if (parts[1] === "c" && (parts[4] === "h" || parts[4] === "t") && j === path.length - 1) {
            continue;
        }

        // Happens when we start the path for a new component
        if (path[j] === "") {
            continue;
        }

        // Skip and child nodes found above the head.
        // Often added by browser plugins
        if (target.tagName !== undefined && target.tagName === "HTML") {
            for (let i = 0; i < target.childNodes.length; i++) {
                if (target.childNodes[i].tagName === undefined || target.childNodes[i].tagName !== "HEAD") {
                    path[j]++
                } else {
                    break;
                }
            }
        }

        if (path[j] >= target.childNodes.length) {
            hlive.log("child not found at section : " + j + " : for: " + diff);

            target = null;
            break;
        }

        target = target.childNodes[path[j]];
    }

    return target;
}

hlive.processMsg = (evt) => {
    let messages = evt.data.split('\n');

    for (let i = 0; i < messages.length; i++) {
        let msg = messages[i];

        if (msg === "") {
            continue;
        }

        hlive.beforeProcessMessage.forEach(function (fn) {
            msg = fn(msg);
        })

        if (msg === "") {
            continue;
        }

        const parts = msg.split("|");

        // DOM Diffs
        if (parts[hlive.msgPart.Type] === "d") {
            if (parts.length !== 6) {
                hlive.log("invalid diff message format");
                continue;
            }

            const target = hlive.findDiffTarget(messages[i]);

            if (target === null) {
                return;
            }

            const path = parts[hlive.diffParts.Path].split(">");

            // Text
            if (parts[hlive.diffParts.ContentType] === "t") {
                if (parts[hlive.diffParts.DiffType] === "c") {
                    let element = document.createTextNode(hlive.base64Decode(parts[hlive.diffParts.Content]));

                    const index = path[path.length - 1];
                    if (index < target.childNodes.length) {
                        target.insertBefore(element.cloneNode(true), target.childNodes[index]);
                    } else {
                        target.appendChild(element.cloneNode(true));
                    }
                } else {
                    target.textContent = hlive.base64Decode(parts[hlive.diffParts.Content]);
                }
            }

            // Tag / HTML
            if (parts[hlive.diffParts.DiffType] === "c" && parts[hlive.diffParts.ContentType] === "h") {
                // Only a single root element is allowed
                let template = document.createElement('template');
                template.innerHTML = hlive.base64Decode(parts[hlive.diffParts.Content]);

                exJS(template.content);

                const index = path[path.length - 1];
                if (index < target.childNodes.length) {
                    target.insertBefore(template.content.firstChild, target.childNodes[index]);
                } else {
                    target.appendChild(template.content.firstChild);
                }
            } else if (parts[hlive.diffParts.DiffType] === "u" && parts[hlive.diffParts.ContentType] === "h") {
                let template = document.createElement('template');
                template.innerHTML = hlive.base64Decode(parts[hlive.diffParts.Content]);
                target.replaceWith(template.content.firstChild);
            }

            // Attributes
            if (parts[hlive.diffParts.ContentType] === "a") {
                const attrData = hlive.base64Decode(parts[hlive.diffParts.Content]);

                // We strictly control this Attribute data format
                const index = attrData.indexOf("=");
                const attrName = attrData.substring(0, index).trim();
                const attrValue = attrData.substring(index + 2, attrData.length - 1);

                if (parts[hlive.diffParts.DiffType] === "c" || parts[hlive.diffParts.DiffType] === "u") {
                    if (attrName === "hon" && parts[hlive.diffParts.DiffType] === "u") {
                        // They'll be set again if only some were removed
                        hlive.removeEventHandlers(target);
                    }

                    if (attrName === "value") {
                        if (target === document.activeElement && attrValue !== "") {
                            // Don't update when someone is typing
                        } else {
                            target.value = attrValue;
                        }
                    } else {
                        target.setAttribute(attrName, attrValue);
                    }
                } else if (parts[hlive.diffParts.DiffType] === "d") {
                    // They'll be set again if only some were removed
                    hlive.removeEventHandlers(target);

                    target.removeAttribute(attrName);
                }
            }

            // Generic delete
            if (parts[hlive.diffParts.DiffType] === "d" && parts[hlive.diffParts.ContentType] !== "a") {
                hlive.removeEventHandlers(target);
                target.remove();
            }
            // Sessions
        } else if (parts[hlive.msgPart.Type] === "s") {
            if (parts.length === 3) {
                hlive.sessID = parts[2];
            }
        }
    }
}

hlive.onopen = (evt) => {
    hlive.log("con: open");
    hlive.reconnectCount = 0;
}

hlive.onmessage = (evt) => {
    hlive.processMsg(evt);
    hlive.postMessage();
}

hlive.onclose = (evt) => {
    hlive.log("con: closed: "+ evt.reason + " ("+ evt.code+") Clean: "+ evt.wasClean);

    if (hlive.reconnectCount < hlive.reconnectLimit) {
        hlive.log("con: reconnect");
        hlive.reconnectCount++;

        hlive.connect();

        return;
    }

    let cover = document.createElement("div");
    const s = "position:fixed;top:0;left:0;background:rgba(0,0,0,0.4);z-index:1000;width:100%;height:100%;";
    cover.setAttribute("style", s);

    document.getElementsByTagName("body")[0].appendChild(cover);
}

hlive.connect = () => {
    let ws = "ws";
    if (location.protocol === 'https:') {
        ws = "wss";
    }

    // Add Session ID to query params
    let q = window.location.search;
    if (q === "") {
        q = "?";
    } else {
        q += "&";
    }
    q += "hlive=" + hlive.sessID;

    const hhash = document.documentElement.getAttribute("data-hlive-hash")
    if (hhash != null) {
        if (q === "") {
            q = "?";
        } else {
            q += "&";
        }
        q += "hhash=" + hhash;
    }

    hlive.conn = new WebSocket(ws + "://" + window.location.host + window.location.pathname + q);
    hlive.conn.onopen = hlive.onopen;
    hlive.conn.onmessage = hlive.onmessage;
    hlive.conn.onclose = hlive.onclose;
}

hlive.connectWails = () => {
    hlive.conn = {
        readyState: 1,
        send: function (msg) {
            runtime.EventsEmit("out", msg);
        }
    };

    runtime.EventsOn("in", function(evt){
        hlive.processMsg({"data":evt});
        hlive.postMessage();
    })

    runtime.EventsEmit("connect", true);
}

document.addEventListener("DOMContentLoaded", function (evt) {
    if (window.runtime !== undefined) {
        hlive.connectWails()
    } else if (window["WebSocket"]) {
        hlive.log("init");
        hlive.connect();
    } else {
        // TODO: do something better?
        alert("Your browser does not support WebSockets");
    }
});

// Execute Script Elements
// https://stackoverflow.com/a/69190644/1269893
const exJS = (containerElement) => {
    Array.from(containerElement.querySelectorAll("script")).forEach((el) => {
        const clone = document.createElement("script");

        Array.from(el.attributes).forEach((attr) => {
            clone.setAttribute(attr.name, attr.value);
        });

        clone.text = el.text;

        el.parentNode.replaceChild(clone, el);
    });
}
