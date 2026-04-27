// Companion Bridge - syncs Tiled selection state with the asset editor's
// companion page via HTTP (using curl). The asset editor server at
// localhost:8090 acts as the message broker.
//
// Install: run sync.sh to copy to Tiled's extensions directory.
// Requires the asset editor server to be running.
//
// Note: Tiled's scripting engine (Qt QJSEngine) does not provide
// setTimeout/setInterval. The plugin polls for commands on each
// selection change. Use Edit > Poll Companion Commands (or the
// shortcut) to manually poll if commands were queued without a
// selection change.

var BRIDGE_URL = "http://localhost:8090/api/v1/tiled-bridge";
var bridgeEnabled = true;

function getMapFile() {
    var a = tiled.activeAsset;
    if (!a || !a.fileName) return "";
    var parts = a.fileName.split("/");
    var idx = parts.indexOf("tiled_maps");
    if (idx >= 0) return parts.slice(idx).join("/");
    return parts[parts.length - 1];
}

function serializeSelectedObjects() {
    var a = tiled.activeAsset;
    if (!a || !a.selectedObjects) return [];
    var objects = a.selectedObjects;
    var result = [];
    for (var i = 0; i < objects.length; i++) {
        var obj = objects[i];
        var props = {};
        var resolved = obj.resolvedProperties();
        for (var name in resolved) {
            var val = resolved[name];
            if (val !== undefined && val !== null && val !== "") {
                props[name] = String(val);
            }
        }
        result.push({
            id: obj.id,
            name: obj.name || "",
            className: obj.className || "",
            x: obj.x,
            y: obj.y,
            width: obj.width || 0,
            height: obj.height || 0,
            properties: props
        });
    }
    return result;
}

function httpPostWithResponse(path, payload) {
    var proc = new Process();
    var output = "";
    try {
        proc.exec("curl", [
            "-s", "-X", "POST",
            "-H", "Content-Type: application/json",
            "-d", payload,
            "--connect-timeout", "0.2",
            "--max-time", "1",
            BRIDGE_URL + path
        ]);
        output = proc.readStdOut();
    } catch (e) {
        // Server not running
    }
    proc.close();
    return output;
}

function httpPost(path, payload) {
    var proc = new Process();
    try {
        proc.exec("curl", [
            "-s", "-X", "POST",
            "-H", "Content-Type: application/json",
            "-d", payload,
            "--connect-timeout", "0.2",
            "--max-time", "1",
            BRIDGE_URL + path
        ]);
    } catch (e) {
        // Server not running
    }
    proc.close();
}

function pushSelection() {
    if (!bridgeEnabled) return;
    var payload = JSON.stringify({
        mapFile: getMapFile(),
        objects: serializeSelectedObjects()
    });
    // Selection POST returns pending commands in the response,
    // avoiding a second round trip for heartbeat/poll.
    var output = httpPostWithResponse("/selection", payload);
    if (output && output.trim() !== "" && output.trim() !== "[]") {
        try {
            var commands = JSON.parse(output);
            applyCommands(commands);
        } catch (e) {
            // parse error
        }
    }
}

function pollCommands() {
    if (!bridgeEnabled) return;
    var output = httpPostWithResponse("/heartbeat", "{}");
    if (!output || output.trim() === "") return;
    try {
        var commands = JSON.parse(output);
        applyCommands(commands);
    } catch (e) {
        // parse error or empty response
    }
}

function applyCommands(commands) {
    if (!commands || commands.length === 0) return;
    var a = tiled.activeAsset;
    if (!a) {
        sendAcks(commands.map(function (cmd) {
            return { id: cmd.id, ok: false, message: "No active asset in Tiled" };
        }));
        return;
    }

    var objectMap = {};
    for (var li = 0; li < a.layerCount; li++) {
        var layer = a.layerAt(li);
        if (layer.isObjectLayer) {
            var objs = layer.objects;
            for (var oi = 0; oi < objs.length; oi++) {
                objectMap[objs[oi].id] = objs[oi];
            }
        }
    }

    var acks = [];
    var errors = [];

    for (var i = 0; i < commands.length; i++) {
        var cmd = commands[i];
        var obj = objectMap[cmd.objectId];
        if (!obj) {
            var msg = "Object " + cmd.objectId + " not found in map";
            tiled.log("[companion] " + msg);
            acks.push({ id: cmd.id, ok: false, message: msg });
            errors.push(msg);
            continue;
        }
        if (cmd.action === "setProperty") {
            if (cmd.name === "class") {
                obj.className = cmd.value;
                tiled.log("[companion] Set className=" + cmd.value + " on object " + cmd.objectId);
            } else {
                obj.setProperty(cmd.name, cmd.value);
                tiled.log("[companion] Set " + cmd.name + "=" + cmd.value + " on object " + cmd.objectId);
            }
            acks.push({ id: cmd.id, ok: true });
        } else if (cmd.action === "removeProperty") {
            obj.removeProperty(cmd.name);
            tiled.log("[companion] Removed " + cmd.name + " from object " + cmd.objectId);
            acks.push({ id: cmd.id, ok: true });
        } else {
            var umsg = "Unknown action: " + cmd.action;
            acks.push({ id: cmd.id, ok: false, message: umsg });
            errors.push(umsg);
        }
    }

    if (errors.length > 0) {
        tiled.alert("Companion bridge failed to apply " + errors.length + " command(s):\n\n" + errors.join("\n"));
    }

    // Push updated state + acks in a single request
    var payload = JSON.stringify({
        mapFile: getMapFile(),
        objects: serializeSelectedObjects(),
        acks: acks
    });
    httpPostWithResponse("/selection", payload);
}

function sendAcks(acks) {
    if (!acks || acks.length === 0) return;
    httpPost("/ack", JSON.stringify(acks));
}

// Watch selection changes on the active asset
var currentAsset = null;

function connectAsset(asset) {
    if (currentAsset && currentAsset.selectedObjectsChanged) {
        currentAsset.selectedObjectsChanged.disconnect(pushSelection);
    }
    currentAsset = asset;
    if (asset && asset.selectedObjectsChanged) {
        asset.selectedObjectsChanged.connect(pushSelection);
    }
    pushSelection();
}

tiled.activeAssetChanged.connect(function () {
    connectAsset(tiled.activeAsset);
});

// Toggle action
var toggleAction = tiled.registerAction("ToggleCompanionBridge", function () {
    bridgeEnabled = !bridgeEnabled;
    toggleAction.checked = bridgeEnabled;
    if (bridgeEnabled) {
        connectAsset(tiled.activeAsset);
        tiled.log("[companion] Bridge enabled");
    } else {
        tiled.log("[companion] Bridge disabled");
    }
});
toggleAction.text = "Companion Bridge";
toggleAction.checkable = true;
toggleAction.checked = true;

// Open companion page action
var openCompanionAction = tiled.registerAction("OpenCompanionPage", function () {
    var proc = new Process();
    proc.exec("open", ["http://localhost:5173/tiled"]);
    proc.close();
});
openCompanionAction.text = "Open Companion Page";
openCompanionAction.shortcut = "Ctrl+Shift+T";

// Manual poll action
var pollAction = tiled.registerAction("PollCompanionCommands", function () {
    pollCommands();
    tiled.log("[companion] Polled for commands");
});
pollAction.text = "Poll Companion Commands";

tiled.extendMenu("Edit", [
    { separator: true },
    { action: "ToggleCompanionBridge" },
    { action: "OpenCompanionPage" },
    { action: "PollCompanionCommands" }
]);

// Start
connectAsset(tiled.activeAsset);
tiled.log("[companion] Companion bridge loaded");
