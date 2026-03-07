"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.UnagntExplorerProvider = void 0;
const vscode = require("vscode");
class RunItem extends vscode.TreeItem {
    runId;
    state;
    constructor(runId, state) {
        super(runId, vscode.TreeItemCollapsibleState.None);
        this.runId = runId;
        this.state = state;
        this.description = state;
    }
}
class SnapshotItem extends vscode.TreeItem {
    snapshotId;
    runId;
    constructor(snapshotId, runId) {
        super(snapshotId, vscode.TreeItemCollapsibleState.None);
        this.snapshotId = snapshotId;
        this.runId = runId;
        this.description = runId;
    }
}
class FolderItem extends vscode.TreeItem {
    constructor(label, children) {
        super(label, vscode.TreeItemCollapsibleState.Expanded);
        this.children = children;
    }
    children;
}
class UnagntExplorerProvider {
    serverUrl;
    apiKey;
    _onDidChangeTreeData = new vscode.EventEmitter();
    onDidChangeTreeData = this._onDidChangeTreeData.event;
    constructor(serverUrl, apiKey) {
        this.serverUrl = serverUrl;
        this.apiKey = apiKey;
    }
    refresh() {
        this._onDidChangeTreeData.fire();
    }
    getTreeItem(element) {
        return element;
    }
    async getChildren(element) {
        if (!element) {
            const runs = [];
            const snapshots = [];
            try {
                const headers = {};
                if (this.apiKey)
                    headers["Authorization"] = `Bearer ${this.apiKey}`;
                const rResp = await fetch(`${this.serverUrl}/v1/runs?limit=20`, { headers });
                if (rResp.ok) {
                    const rData = (await rResp.json());
                    for (const id of rData.run_ids ?? []) {
                        const gResp = await fetch(`${this.serverUrl}/v1/runs/${id}`, { headers });
                        const run = gResp.ok ? (await gResp.json()) : {};
                        runs.push(new RunItem(id, run.state ?? "?"));
                    }
                }
                const sResp = await fetch(`${this.serverUrl}/v1/replay/snapshots?limit=20`, { headers });
                if (sResp.ok) {
                    const sData = (await sResp.json());
                    for (const s of sData.snapshots ?? []) {
                        snapshots.push(new SnapshotItem(s.id, s.run_id));
                    }
                }
            }
            catch {
                runs.push(new RunItem("(connect to server)", ""));
            }
            return [
                new FolderItem("Runs", runs.length ? runs : [new RunItem("(no runs)", "")]),
                new FolderItem("Snapshots", snapshots.length ? snapshots : [new SnapshotItem("(no snapshots)", "")]),
            ];
        }
        if (element instanceof FolderItem) {
            return element.children ?? [];
        }
        return [];
    }
}
exports.UnagntExplorerProvider = UnagntExplorerProvider;
//# sourceMappingURL=explorer.js.map