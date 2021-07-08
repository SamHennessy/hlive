let hlive = {
    reconnectLimit: 5,
    reconnectCount: 0,
    conn: null,
    isInitialSyncDone: false,
    sessID: 1,
};

hlive.msgPart = {
    Type: 0,
};

hlive.diffParts = {
    DiffType: 1,
    Root: 2,
    Path: 3,
    ContentType: 4,
    Content: 5,
};

// Ref: https://stackoverflow.com/questions/30106476/using-javascripts-atob-to-decode-base64-doesnt-properly-decode-utf-8-strings
hlive.b64DecodeUnicode = (str) => {
    // Going backwards: from byte stream, to percent-encoding, to original string.
    return decodeURIComponent(atob(str).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));
}

hlive.eventHandler = (e) => {
    if (!e.currentTarget || !e.currentTarget.getAttribute) {
        return
    }

    const el = e.currentTarget;

    const pairs = el.getAttribute("data-hlive-on").split(",");
    for (let i = 0; i < pairs.length; i++) {
        const parts = pairs[i].split("|");
        if (parts[1].toLowerCase() === e.type.toLowerCase()) {
            hlive.eventHandlerHelper(e, parts[0], false);
        }
    }
}

hlive.removeEventHandlers = (el) => {
    if (!el.getAttribute ) {
        return
    }

    const value = el.getAttribute("data-hlive-on");

    if (value === null || value === "") {
        return
    }

    const pairs = value.split(",");
    for (let i = 0; i < pairs.length; i++) {
        const parts = pairs[i].split("|");
        el.removeEventListener(parts[1].toLowerCase(), hlive.eventHandler)
    }
}

hlive.eventHandlerHelper = (e, handlerID, isInitial) => {
    if (e.preventDefault) {
        e.preventDefault();
    }

    const el = e.currentTarget;

    let msg = {
        t: "e",
        i: handlerID,
    };

    let d = {};
    if (el.value !== undefined) {
        d.value = String(el.value);
        if (isInitial) {
            d.init = "true";
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

    // File?
    if (el.files) {
        // No files
        msg.file = {
            "name": "",
            "size": 0,
            "type": "",
            "index": 0,
            "total": 0,
        };
        // Single file
        if (el.files.length === 1) {
            msg.file = {
                "name": el.files[0].name,
                "size": el.files[0].size,
                "type": el.files[0].type,
                "index": 0,
                "total": 1,
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
    console.log(message)
    if (hlive.conn) {
        let msg = {
            t: "l",
            d: {m: message},
        };
        hlive.sendMsg(msg);
    }
}

hlive.setEventHandlers = () => {
    document.querySelectorAll("[data-hlive-on]").forEach(function (el) {
        const pairs = el.getAttribute("data-hlive-on").split(",");
        for (let i = 0; i < pairs.length; i++) {
            const parts = pairs[i].split("|");
            el.addEventListener(parts[1].toLowerCase(), hlive.eventHandler);
        }
    });
}

// Looks at the current value of the input and if needed triggers events to sync that value to the backend
hlive.syncInitialInputValues = () => {
    document.querySelectorAll("[data-hlive-on]").forEach(function (el) {
        // Radio
        if (el.getAttribute("type") && el.getAttribute("type").toLowerCase() === "radio") {
           if (el.checked && el.hasAttribute("checked") === false) {

           } else {
               return
           }
        } else {
            if (el.value === undefined) {
                return;
            }

            if (el.value === el.getAttribute("value")) {
                return;
            }
        }

        const pairs = el.getAttribute("data-hlive-on").split(",");
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

    if (!hlive.isInitialSyncDone) {
        hlive.isInitialSyncDone = true;
        hlive.syncInitialInputValues();
    }

    // Give focus
    document.querySelectorAll("[data-hlive-focus]").forEach(function (el) {
        el.focus();
        if (el.selectionStart !== undefined) {
            setTimeout(function(){ el.selectionStart = el.selectionEnd = 10000; }, 0);
        }
    });

    // Start file upload
    document.querySelectorAll("[data-hlive-upload]").forEach(function (el) {
        const ids = hlive.getEventHAndlerIDs(el);

        if (!ids["upload"]) {
            return
        }

        if (el.files.length !== 0) {
            let i = 0;
            const file = el.files[0];

            const fileMeta = {
                "name": file.name,
                "size": file.size,
                "type": file.type,
                "index": i,
                "total": el.files.length,
            };

            let msg = {
                t: "e",
                file: fileMeta,
            };

            queueMicrotask(function () {
                for (let j = 0; j < ids["upload"].length; j++) {
                    msg.i = ids["upload"][j];

                    hlive.conn.send(new Blob([JSON.stringify(msg) + "\n\n", el.files[i]], {type: el.files[i].type}));
                }
            });
        }
    });

    // Trigger diffapply, should always be last
    document.querySelectorAll("[data-hlive-on*=diffapply]").forEach(function (el) {
        const ids = hlive.getEventHAndlerIDs(el);

        if (!ids["diffapply"]) {
            return;
        }

        for (let i = 0; i < ids["diffapply"].length; i++) {
            hlive.sendMsg({
                t: "e",
                i: ids["diffapply"][i],
            });
        }
    });
}

hlive.getEventHAndlerIDs = (el) => {
    let map = {};

    if (el.getAttribute && el.getAttribute("data-hlive-on") !== "") {
        const pairs = el.getAttribute("data-hlive-on").split(",");
        for (let i = 0; i < pairs.length; i++) {
            const parts = pairs[i].split("|");
            const eventName = parts[1].toLowerCase();
            const eventID =  parts[0];

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

    let target = document
    if (parts[hlive.diffParts.Root] !== "doc") {
        target = document.querySelector('[data-hlive-id="'+parts[hlive.diffParts.Root]+'"]');
    }

    if (target === null) {
        hlive.log("root element not found: " + parts[hlive.diffParts.Root]);
        return null
    }

    const path = parts[hlive.diffParts.Path].split(">");

    for (let j = 0; j < path.length; j++) {
        // Doesn't exist
        if (parts[1] === "c" && (parts[4] === "h" || parts[4] === "t" ) && j === path.length - 1) {
            continue;
        }

        // Happens when we start the path for a new component
        if (path[j] === "") {
            continue;
        }

        if (path[j] >= target.childNodes.length ) {
            hlive.log("child not found " + parts[hlive.diffParts.Root] + ":" + parts[hlive.diffParts.Path]);

            target = null;
            break;
        }

        target = target.childNodes[path[j]];
    }

    return target;
}

hlive.processMsg = (evt) => {
    let messages = evt.data.split('\n');

    let newMessages = [];
    let deleteMessageBuffer = [];

    // Re-order deletes
    // Example problem:
    // d|d|doc|1>1>0>0>1>2||
    // d|d|doc|1>1>0>0>1>3||

    for (let i = 0; i < messages.length; i++) {
        // Delete diff?
        if (messages[i].substring(0, 4) === "d|d|") {
            // If buffer empty, start the buffer
            if (deleteMessageBuffer.length === 0) {
                deleteMessageBuffer[deleteMessageBuffer.length] = messages[i];

                continue;
            }
            // Is this delete a child of the same parent in the buffer?
            const aParts =  deleteMessageBuffer[0].split("|");
            const aIndexOf = aParts[hlive.diffParts.Path].lastIndexOf(">");
            const aParentPath = aParts[hlive.diffParts.Path].substring(0, aIndexOf);

            const bParts =  messages[i].split("|");
            const bIndexOf = bParts[hlive.diffParts.Path].lastIndexOf(">");
            const bParentPath = bParts[hlive.diffParts.Path].substring(0, bIndexOf);

            // Same, add to buffer
            if (aParentPath === bParentPath) {
                deleteMessageBuffer[deleteMessageBuffer.length] = messages[i];

                continue;
            }

            // Not the same, flush buffer and add delete be after buffer
            for (let j = 0; j < deleteMessageBuffer.length; j++) {
                newMessages[newMessages.length] = deleteMessageBuffer[deleteMessageBuffer.length - (j+1)];
            }

            deleteMessageBuffer = [];
            // Not a delete, flush buffer if not empty
        } else if (deleteMessageBuffer.length !== 0) {
            for (let j = 0; j < deleteMessageBuffer.length; j++) {
                newMessages[newMessages.length] = deleteMessageBuffer[deleteMessageBuffer.length - (j+1)];
            }

            deleteMessageBuffer = [];
        }

        newMessages[newMessages.length] = messages[i];
    }

    messages = newMessages;

    for (let i = 0; i < messages.length; i++) {
        if (messages[i] === "") {
            continue;
        }

        const parts = messages[i].split("|");

        // DOM Diffs
        if (parts[hlive.msgPart.Type] === "d") {
            if (parts.length !== 6) {
                hlive.log("invalid message format");
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
                    let element = document.createTextNode(hlive.b64DecodeUnicode(parts[hlive.diffParts.Content]));

                    const index = path[path.length - 1];
                    if (index < target.childNodes.length) {
                        target.insertBefore(element.cloneNode(true), target.childNodes[index]);
                    } else {
                        target.appendChild(element.cloneNode(true));
                    }
                } else {
                    target.textContent = hlive.b64DecodeUnicode(parts[hlive.diffParts.Content]);
                }
            }

            // Tag / HTML
            if (parts[hlive.diffParts.DiffType] === "c" && parts[hlive.diffParts.ContentType] === "h") {
                // Only a single root element is allowed
                let template = document.createElement('template');
                template.innerHTML = hlive.b64DecodeUnicode(parts[hlive.diffParts.Content]);

                const index = path[path.length - 1];
                if (index < target.childNodes.length) {
                    target.insertBefore(template.content.firstChild, target.childNodes[index]);
                } else {
                    target.appendChild(template.content.firstChild);
                }
            } else if (parts[hlive.diffParts.DiffType] === "u" && parts[hlive.diffParts.ContentType] === "h") {
                let template = document.createElement('template');
                template.innerHTML = hlive.b64DecodeUnicode(parts[hlive.diffParts.Content]);
                target.replaceWith(template.content.firstChild);
            }

            // Attributes
            if (parts[hlive.diffParts.ContentType] === "a") {
                const attrData = hlive.b64DecodeUnicode(parts[hlive.diffParts.Content]);

                // We strictly control this Attribute data format
                const index = attrData.indexOf("=");
                const attrName = attrData.substring(0, index).trim();
                const attrValue = attrData.substring(index + 2, attrData.length - 1);

                if (parts[hlive.diffParts.DiffType] === "c" || parts[hlive.diffParts.DiffType] === "u" ) {
                    if (attrName === "data-hlive-on" && parts[hlive.diffParts.DiffType] === "u") {
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
    hlive.log("con: closed");

    if (hlive.reconnectCount < hlive.reconnectLimit) {
        hlive.log("con: reconnect");
        hlive.reconnectCount++;

        hlive.connect();

        return;
    }

    let cover = document.createElement("div");
    const s = "position:fixed;top:0;left:0;background:rgba(0,0,0,0.4);z-index:5;width:100%;height:100%;";
    cover.setAttribute("style", s);

    document.getElementsByTagName("body")[0].appendChild(cover);
}

hlive.connect = () => {
    let ws = "ws";
    if (location.protocol === 'https:') {
        ws = "wss";
    }

    hlive.conn = new WebSocket(ws + "://" + window.location.host + window.location.pathname + "?ws=" + hlive.sessID);
    hlive.conn.onopen = hlive.onopen
    hlive.conn.onmessage = hlive.onmessage;
    hlive.conn.onclose = hlive.onclose;
}

document.addEventListener("DOMContentLoaded", function(evt) {
    if (window["WebSocket"]) {
        hlive.connect()
    } else {
        // TODO: do something better
        alert("Your browser does not support WebSockets");
    }
});
