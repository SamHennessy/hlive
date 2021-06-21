document.addEventListener("DOMContentLoaded", function(event) {
    let conn = null;

    function sendMsg(msg) {
        queueMicrotask(function () {
            conn.send(JSON.stringify(msg));
        });
    }

    function log(message) {
        console.log(message)
        if (conn) {
            let msg = {
                t: "l",
                d: {m: message},
            };
            sendMsg(msg);
        }
    }

    function eventData(e) {
        let d = {};
        if (e.currentTarget.value !== undefined) {
            d.value = String(e.currentTarget.value);
        }

        if (e.key !== undefined) {
            d.key = e.key;
            d.charCode = String(e.charCode);
            d.keyCode = String(e.keyCode);
            d.shiftKey = String(e.shiftKey);
            d.altKey = String(e.altKey);
            d.ctrlKey = String(e.ctrlKey);
        }

        return d
    }

    function handlerHelper(e, attrName) {
        if (e.currentTarget && e.currentTarget.hasAttribute && e.currentTarget.hasAttribute(attrName)) {
            let msg = {
                t: "e",
                i: e.currentTarget.getAttribute(attrName),
            };
            sendMsg(msg);
        }
    }

    function handlerWithDataHelper(e, attrName) {
        if (e.currentTarget && e.currentTarget.hasAttribute && e.currentTarget.hasAttribute(attrName)) {
            let msg = {
                t: "e",
                i: e.currentTarget.getAttribute(attrName),
                d: eventData(e),
            };
            sendMsg(msg);
        }
    }

    function handlerHelperGeneric(e, attrName) {
        if (!e.currentTarget || !e.currentTarget.hasAttribute || !e.currentTarget.hasAttribute(attrName)) {
            return
        }

        const el = e.currentTarget;

        let msg = {
            t: "e",
            i: el.getAttribute(attrName),
        };

        let d = {};
        if (el.value !== undefined) {
            d.value = String(el.value);
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

                    sendMsg(msg);
                }

                return
            }
        }

        sendMsg(msg);

    }


    function handlerHelper2(e, handlerID) {
        const el = e.currentTarget;

        let msg = {
            t: "e",
            i: handlerID,
        };

        let d = {};
        if (el.value !== undefined) {
            d.value = String(el.value);
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

                    sendMsg(msg);
                }

                return
            }
        }

        sendMsg(msg);

    }

    const eventHandler = (e) => {
        if (!e.currentTarget || !e.currentTarget.getAttribute) {
            return
        }

        const el = e.currentTarget;

        const pairs = el.getAttribute("data-hlive-on").split(",");
        for (let i = 0; i < pairs.length; i++) {
            const parts = pairs[i].split("|");
            if (parts[1].toLowerCase() === e.type.toLowerCase()) {
                handlerHelper2(e, parts[0]);
            }
        }
    }

    const clickHandler = (e) => {
        e.preventDefault();
        handlerHelper(e,"data-hlive-onclick");
    }

    const focusHandler = (e) => {
        handlerHelper(e,"data-hlive-onfocus");
    }

    const keydownHandler = (e) => {
        handlerWithDataHelper(e,"data-hlive-onkeydown");
    }

    const keyupHandler = (e) => {
        handlerWithDataHelper(e,"data-hlive-onkeyup");
    }

    const animationendHandler = (e) => {
        handlerHelper(e,"data-hlive-onanimationend");
    }

    const animationcancelHandler = (e) => {
        handlerHelper(e,"data-hlive-onanimationcancel");
    }

    const mouseenterHandler = (e) => {
        handlerHelper(e,"data-hlive-onmouseenter");
    }

    const mouseleaveHandler = (e) => {
        handlerHelper(e,"data-hlive-onmouseleave");
    }

    const changeHandler = (e) => {
        handlerHelperGeneric(e, "data-hlive-onchange")
    }

    function setEventHandlers() {
        document.querySelectorAll("[data-hlive-on]").forEach(function (el) {
            const pairs = el.getAttribute("data-hlive-on").split(",");
            for (let i = 0; i < pairs.length; i++) {
                const parts = pairs[i].split("|");
                el.addEventListener(parts[1].toLowerCase(), eventHandler);
            }
        });

        document.querySelectorAll("[data-hlive-onclick]").forEach(function (value) {
            value.addEventListener("click", clickHandler);
        });

        document.querySelectorAll("[data-hlive-onkeydown]").forEach(function (value) {
            value.addEventListener("keydown", keydownHandler);
        });

        document.querySelectorAll("[data-hlive-onkeyup]").forEach(function (value) {
            value.addEventListener("keyup", keyupHandler);
        });

        document.querySelectorAll("[data-hlive-onfocus]").forEach(function (value) {
            value.addEventListener("focus", focusHandler);
        });

        document.querySelectorAll("[data-hlive-onanimationend]").forEach(function (value) {
            value.addEventListener("animationend", animationendHandler);
        });

        document.querySelectorAll("[data-hlive-onanimationcancel]").forEach(function (value) {
            value.addEventListener("animationcancel", animationcancelHandler);
        });

        document.querySelectorAll("[data-hlive-onmouseenter]").forEach(function (value) {
            value.addEventListener("mouseenter", mouseenterHandler);
        });

        document.querySelectorAll("[data-hlive-onchange]").forEach(function (value) {
            value.addEventListener("change", changeHandler);
        });
    }

    function removeEventHandlers(el) {
        if (!el.hasAttribute) {
            return;
        }

        if (el.hasAttribute("data-hlive-onclick")) {
            el.removeEventListener("click", clickHandler);
        }

        if (el.hasAttribute("data-hlive-onkeydown")) {
            el.removeEventListener("keydown", keydownHandler);
        }

        if (el.hasAttribute("data-hlive-onkeyup")) {
            el.removeEventListener("keyup", keyupHandler);
        }

        if (el.hasAttribute("data-hlive-onfocus")) {
            el.removeEventListener("focus", focusHandler);
        }

        if (el.hasAttribute("data-hlive-onanimationend")) {
            el.removeEventListener("animationend", animationendHandler);
        }

        if (el.hasAttribute("data-hlive-onanimationcancel")) {
            el.removeEventListener("animationcancel", animationcancelHandler);
        }

        if ( el.hasAttribute("data-hlive-onmouseenter")) {
            el.removeEventListener("mouseenter", mouseenterHandler);
        }

        if (el.hasAttribute("data-hlive-onmouseleave")) {
            el.removeEventListener("mouseleave", mouseleaveHandler);
        }
    }

    function postMessage() {
        setEventHandlers();

        // Give focus
        document.querySelectorAll("[data-hlive-focus]").forEach(function (el) {
            el.focus();
            if (el.selectionStart !== undefined) {
                setTimeout(function(){ el.selectionStart = el.selectionEnd = 10000; }, 0);
            }
        });

        // Start file upload
        document.querySelectorAll("[data-hlive-upload]").forEach(function (el) {
            if (el.hasAttribute("data-hlive-onupload")) {

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
                        i: el.getAttribute("data-hlive-onupload"),
                        file: fileMeta,
                    };

                    queueMicrotask(function () {
                        conn.send(new Blob([JSON.stringify(msg) + "\n\n", el.files[i]], {type: el.files[i].type}));
                    });
                }
            }
        });

        // Trigger ondiffapply, should always be last
        // TODO: make work with generic event handler
        //      maybe: [data-hlive-on*="diffapply"]
        document.querySelectorAll("[data-hlive-ondiffapply]").forEach(function (el) {
            let msg = {
                t: "e",
                i: el.getAttribute("data-hlive-ondiffapply"),
            };

            sendMsg(msg);
        });
    }

    let sessID = "1";

    if (window["WebSocket"]) {
        let ws = "ws";
        if (location.protocol === 'https:') {
            ws = "wss";
        }

        conn = new WebSocket(ws + "://" + window.location.host + window.location.pathname + "?ws=" + sessID);
        conn.onopen = function (evt) {
            log("con: open");
        };

        conn.onclose = function (evt) {
            log("con: closed");

            let cover = document.createElement("div");
            const s = "position:fixed;top:0;left:0;background:rgba(0,0,0,0.4);z-index:5;width:100%;height:100%;";
            cover.setAttribute("style", s);

            document.getElementsByTagName("body")[0].appendChild(cover);
        };

        conn.onmessage = function (evt) {
            processMsg(evt);
            postMessage();
        };

        function processMsg (evt) {
            const msgPart = {
                Type: 0,
            };

            const diffParts = {
                DiffType: 1,
                Root: 2,
                Path: 3,
                ContentType: 4,
                Content: 5,
            };

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
                    const aIndexOf = aParts[diffParts.Path].lastIndexOf(">");
                    const aParentPath = aParts[diffParts.Path].substring(0, aIndexOf);

                    const bParts =  messages[i].split("|");
                    const bIndexOf = bParts[diffParts.Path].lastIndexOf(">");
                    const bParentPath = bParts[diffParts.Path].substring(0, bIndexOf);

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
                if (parts[msgPart.Type] === "d") {
                    if (parts.length !== 6) {
                        log("invalid message format");
                        continue;
                    }

                    let target = document
                    if (parts[diffParts.Root] !== "doc") {
                        target = document.querySelector('[data-hlive-id="'+parts[diffParts.Root]+'"]');
                    }

                    if (target === null) {
                        log("root element not found: " + parts[diffParts.Root]);
                        continue;
                    }

                    const path = parts[diffParts.Path].split(">");

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
                            log("child not found " + parts[diffParts.Root] + ":" + parts[diffParts.Path]);

                            target = null;
                            break;
                        }

                        target = target.childNodes[path[j]];
                    }

                    if (target !== null) {
                        // Text
                        if (parts[diffParts.ContentType] === "t") {
                            if (parts[diffParts.DiffType] === "c") {
                                let element = document.createTextNode(window.atob(parts[diffParts.Content]));

                                const index = path[path.length - 1];
                                if (index < target.childNodes.length) {
                                    target.insertBefore(element.cloneNode(true), target.childNodes[index]);
                                } else {
                                    target.appendChild(element.cloneNode(true));
                                }
                            } else {
                                target.textContent = window.atob(parts[diffParts.Content]);
                            }
                        }

                        // Tag / HTML
                        if (parts[diffParts.DiffType] === "c" && parts[diffParts.ContentType] === "h") {
                            // Only a single root element is allowed
                            let template = document.createElement('template');
                            template.innerHTML = window.atob(parts[diffParts.Content]);

                            const index = path[path.length - 1];
                            if (index < target.childNodes.length) {
                                target.insertBefore(template.content.firstChild, target.childNodes[index]);
                            } else {
                                target.appendChild(template.content.firstChild);
                            }
                        } else if (parts[diffParts.DiffType] === "u" && parts[diffParts.ContentType] === "h") {
                            let template = document.createElement('template');
                            template.innerHTML = window.atob(parts[diffParts.Content]);
                            target.replaceWith(template.content.firstChild);
                        }

                        // Attributes
                        if (parts[diffParts.ContentType] === "a") {
                            const attrData = window.atob(parts[diffParts.Content]);
                            let attrParts = [];
                            const index = attrData.indexOf("=");
                            if (index === -1) {
                                attrParts[0] = attrData;
                                attrParts[1] = "";
                            } else {
                                attrParts[0] = attrData.substring(0, index);
                                attrParts[1] = attrData.substring(index + 1);
                            }

                            const name = attrParts[0].trim();
                            attrParts[0] = attrParts[0].trim()
                            if (attrParts.length === 2 && attrParts[1] !== "") {
                                if (attrParts[1].substring(0, 1) === '"') {
                                    attrParts[1] = attrParts[1].substring(1, attrParts[1].length - 1);
                                }
                            } else {
                                attrParts[1] = "";
                            }

                            if (parts[diffParts.DiffType] === "c" || parts[diffParts.DiffType] === "u" ) {
                                if (name === "value") {
                                    if (target === document.activeElement && attrParts[1] !== "") {
                                        // Don't update when someone is typing
                                    } else {
                                        target.value = attrParts[1];
                                    }
                                } else {

                                    target.setAttribute(name, attrParts[1]);
                                    if (name === "data-hlive-focus") {
                                        target.focus();
                                    }
                                }
                            } else if (parts[diffParts.DiffType] === "d") {
                                target.removeAttribute(name);
                            }
                        }

                        // Generic delete
                        if (parts[diffParts.DiffType] === "d" && parts[diffParts.ContentType] !== "a") {
                            removeEventHandlers(target);
                            target.remove();
                        }
                    }

                    // Sessions
                } else if (parts[msgPart.Type] === "s") {
                    if (parts.length === 3) {
                        sessID = parts[2];
                    }
                }
            }
        }

    } else {
        alert("Your browser does not support WebSockets");
    }
});
