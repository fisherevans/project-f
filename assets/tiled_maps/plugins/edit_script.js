// Edit object's `script_src` property in an external editor (macOS-friendly).
// Opens Sublime via `open -a`, then waits for your confirmation before reading back.

const PROP_NAME = "script_src";
const MAC_EDITOR_APP = "Sublime Text"; // or "Visual Studio Code"

function oneSelectedObject() {
    const a = tiled.activeAsset;
    if (!a || a.isTileSet) return null;
    const s = a.selectedObjects;
    return s && s.length === 1 ? s[0] : null;
}

function readText(p) { const f = new TextFile(p, TextFile.ReadOnly); const s = f.readAll(); f.close(); return s; }
function writeText(p, d) { const f = new TextFile(p, TextFile.WriteOnly); f.write(d); f.close(); }

function tempJsPath(base = "tiled-prop") {
    const name = (base || "entity").replace(/[^\w.-]+/g, "_");
    const tmpDir = (tiled.tempPath || FileInfo.path(tiled.extensionsPath));
    return FileInfo.toNativeSeparators(`${tmpDir}/${name}-${Date.now()}.js`);
}

function launchEditor(filePath) {
    // Use macOS `open` so we don’t rely on CLI binaries being on PATH.
    // No -W (wait) here because that waits for app quit; instead we prompt after.
    return tiled.executeCommand("open", ["-a", MAC_EDITOR_APP, filePath], "", /*inTerminal*/ false);
}

const editScriptExternal = tiled.registerAction("editScriptExternalNoPath", function () {
    const obj = oneSelectedObject();
    if (!obj) { tiled.alert("Select exactly one object."); return; }

    const current = (obj.property(PROP_NAME) ?? "").toString();
    const filePath = tempJsPath(obj.name || "entity");
    writeText(filePath, current);

    if (!launchEditor(filePath)) {
        tiled.alert("Could not launch editor. Change MAC_EDITOR_APP to match your editor’s name.");
        return;
    }

    // Simple blocking prompt so you can save in editor before we read the file back.
    tiled.prompt("Editing in external editor.\nSave your file there, then click OK to import changes.", current);

    try {
        const edited = readText(filePath);
        obj.setProperty(PROP_NAME, edited);
        tiled.log(`Updated '${PROP_NAME}' from external editor (${FileInfo.baseName(filePath)}).`);
    } catch (e) {
        tiled.alert("Failed to read edited file: " + e);
    }
});
editScriptExternal.text = "Edit Script (External) – script_src";
editScriptExternal.shortcut = "Ctrl+Alt+E";

tiled.extendMenu("Edit", [
    { action: "editScriptExternalNoPath", before: "Preferences" },
    { separator: true }
]);