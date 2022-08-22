// test-ack
if (hlive.beforeSendEvent.get("htack") === undefined) {
    const hliveTestAck = {
        received: {},
    }

    hlive.beforeSendEvent.set("htack", function (e, msg) {
        if (!e.currentTarget || !e.currentTarget.getAttribute) {
            return msg;
        }

        if (!e.currentTarget.getAttribute("data-hlive-test-ack-id")) {
            return msg;
        }

        if (!msg.e) {
            msg.e = {};
        }

        msg.e["test-ack-id"] = e.currentTarget.getAttribute("data-hlive-test-ack-id");
        e.currentTarget.removeAttribute("data-hlive-test-ack-id");

        return msg;
    })

    hlive.beforeProcessMessage.set("htack", function (msg) {
        if (msg.startsWith("ack|") === false) {
            return msg;
        }

        const parts = msg.split("|")
        console.log("ack beforeProcessMessage", parts);

        hliveTestAck.received[parts[1]] = true;

        return "";
    })
}



