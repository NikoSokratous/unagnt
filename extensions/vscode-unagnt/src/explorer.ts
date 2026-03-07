import * as vscode from "vscode";

type TreeItem = RunItem | SnapshotItem | FolderItem;

class RunItem extends vscode.TreeItem {
  constructor(public runId: string, public state: string) {
    super(runId, vscode.TreeItemCollapsibleState.None);
    this.description = state;
  }
}

class SnapshotItem extends vscode.TreeItem {
  constructor(public snapshotId: string, public runId: string) {
    super(snapshotId, vscode.TreeItemCollapsibleState.None);
    this.description = runId;
  }
}

class FolderItem extends vscode.TreeItem {
  children: TreeItem[] = [];
  constructor(label: string, children: TreeItem[]) {
    super(label, vscode.TreeItemCollapsibleState.Expanded);
    this.children = children;
  }
}

export class UnagntExplorerProvider implements vscode.TreeDataProvider<TreeItem> {
  private _onDidChangeTreeData = new vscode.EventEmitter<TreeItem | undefined | void>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  constructor(
    private serverUrl: string,
    private apiKey: string
  ) {}

  refresh(): void {
    this._onDidChangeTreeData.fire();
  }

  getTreeItem(element: TreeItem): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: TreeItem): Promise<TreeItem[]> {
    if (!element) {
      const runs: TreeItem[] = [];
      const snapshots: TreeItem[] = [];
      try {
        const headers: Record<string, string> = {};
        if (this.apiKey) headers["Authorization"] = `Bearer ${this.apiKey}`;
        const rResp = await fetch(`${this.serverUrl}/v1/runs?limit=20`, { headers });
        if (rResp.ok) {
          const rData = (await rResp.json()) as { run_ids?: string[] };
          for (const id of rData.run_ids ?? []) {
            const gResp = await fetch(`${this.serverUrl}/v1/runs/${id}`, { headers });
            const run = gResp.ok ? ((await gResp.json()) as { state?: string }) : {};
            runs.push(new RunItem(id, run.state ?? "?"));
          }
        }
        const sResp = await fetch(`${this.serverUrl}/v1/replay/snapshots?limit=20`, { headers });
        if (sResp.ok) {
          const sData = (await sResp.json()) as { snapshots?: { id: string; run_id: string }[] };
          for (const s of sData.snapshots ?? []) {
            snapshots.push(new SnapshotItem(s.id, s.run_id));
          }
        }
      } catch {
        runs.push(new RunItem("(connect to server)", ""));
      }
      return [
        new FolderItem("Runs", runs.length ? runs : [new RunItem("(no runs)", "")] as TreeItem[]),
        new FolderItem("Snapshots", snapshots.length ? snapshots : [new SnapshotItem("(no snapshots)", "")] as TreeItem[]),
      ];
    }
    if (element instanceof FolderItem) {
      return element.children ?? [];
    }
    return [];
  }
}
