// Companion Bridge - syncs Tiled selection state with the asset editor's
// companion page via HTTP. The asset editor server at localhost:8090 acts
// as the message broker.
//
// Install: run sync.sh to copy to Tiled's extensions directory.
// Requires the asset editor server to be running.

const BRIDGE_URL = "http://localhost:8090/api/v1/tiled-bridge";
const POLL_INTERVAL_MS = 300;

var bridgeEnabled = true;
var pollTimer = null;

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
        var propNames = Object.keys(obj.resolvedProperties());
        for (var j = 0; j < propNames.length; j++) {
            var name = propNames[j];
            var val = obj.resolvedProperty(name);
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

function pushSelection() {
    if (!bridgeEnabled) return;
    var payload = JSON.stringify({
        mapFile: getMapFile(),
        objects: serializeSelectedObjects()
    });
    try {
        var xhr = new XMLHttpRequest();
        xhr.open("POST", BRIDGE_URL + "/selection", false);
        xhr.setRequestHeader("Content-Type", "application/json");
        xhr.send(payload);
    } catch (e) {
        // Server not running - silently ignore
    }
}

function pollCommands() {
    if (!bridgeEnabled) return;
    try {
        var xhr = new XMLHttpRequest();
        xhr.open("GET", BRIDGE_URL + "/commands", false);
        xhr.send();
        if (xhr.status === 200) {
            var commands = JSON.parse(xhr.responseText);
            applyCommands(commands);
        }
    } catch (e) {
        // Server not running - silently ignore
    }
}

function applyCommands(commands) {
    if (!commands || commands.length === 0) return;
    var a = tiled.activeAsset;
    if (!a) return;

    // Build object lookup from all object layers
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

    for (var i = 0; i < commands.length; i++) {
        var cmd = commands[i];
        var obj = objectMap[cmd.objectId];
        if (!obj) {
            tiled.log("[companion] Object " + cmd.objectId + " not found");
            continue;
        }
        if (cmd.action === "setProperty") {
            obj.setProperty(cmd.name, cmd.value);
            tiled.log("[companion] Set " + cmd.name + "=" + cmd.value + " on object " + cmd.objectId);
        } else if (cmd.action === "removeProperty") {
            obj.removeProperty(cmd.name);
            tiled.log("[companion] Removed " + cmd.name + " from object " + cmd.objectId);
        }
    }

    // Push updated selection state after applying commands
    pushSelection();
}

function startPolling() {
    if (pollTimer) return;
    pollTimer = setInterval(function () {
        pollCommands();
    }, POLL_INTERVAL_MS);
}

function stopPolling() {
    if (pollTimer) {
        clearInterval(pollTimer);
        pollTimer = null;
    }
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
        startPolling();
        connectAsset(tiled.activeAsset);
        tiled.log("[companion] Bridge enabled");
    } else {
        stopPolling();
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

tiled.extendMenu("Edit", [
    { separator: true },
    { action: "ToggleCompanionBridge" },
    { action: "OpenCompanionPage" }
]);

// Start
connectAsset(tiled.activeAsset);
startPolling();
tiled.log("[companion] Companion bridge loaded");
