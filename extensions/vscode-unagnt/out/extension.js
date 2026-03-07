"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.activate = activate;
exports.deactivate = deactivate;
const vscode = require("vscode");
const explorer_1 = require("./explorer");
function activate(context) {
    const config = vscode.workspace.getConfiguration("unagnt");
    const serverUrl = config.get("serverUrl") ?? "http://localhost:8080";
    const apiKey = config.get("apiKey") ?? "";
    const explorer = new explorer_1.UnagntExplorerProvider(serverUrl, apiKey);
    vscode.window.registerTreeDataProvider("unagntExplorer", explorer);
    context.subscriptions.push(vscode.commands.registerCommand("unagnt.refresh", () => explorer.refresh()));
    context.subscriptions.push(vscode.commands.registerCommand("unagnt.validateWorkflow", async () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) {
            vscode.window.showWarningMessage("No active editor");
            return;
        }
        const doc = editor.document;
        const text = doc.getText();
        try {
            const resp = await fetch(`${serverUrl}/v1/workflows/validate`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    ...(apiKey ? { Authorization: `Bearer ${apiKey}` } : {}),
                },
                body: JSON.stringify({ workflow: text }),
            });
            if (resp.ok) {
                vscode.window.showInformationMessage("Workflow is valid");
            }
            else {
                const err = await resp.text();
                vscode.window.showErrorMessage(`Validation failed: ${err}`);
            }
        }
        catch (e) {
            vscode.window.showErrorMessage(`Request failed: ${e}`);
        }
    }));
    context.subscriptions.push(vscode.commands.registerCommand("unagnt.runAgent", async () => {
        const agent = await vscode.window.showInputBox({
            prompt: "Agent name",
            placeHolder: "demo-agent",
        });
        if (!agent)
            return;
        const goal = await vscode.window.showInputBox({
            prompt: "Goal",
            placeHolder: "List files",
        });
        if (!goal)
            return;
        try {
            const resp = await fetch(`${serverUrl}/v1/runs`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    ...(apiKey ? { Authorization: `Bearer ${apiKey}` } : {}),
                },
                body: JSON.stringify({ agent_name: agent, goal }),
            });
            const data = (await resp.json());
            if (data.run_id) {
                vscode.window.showInformationMessage(`Run started: ${data.run_id}`);
                explorer.refresh();
            }
            else {
                vscode.window.showErrorMessage("Failed to start run");
            }
        }
        catch (e) {
            vscode.window.showErrorMessage(`Request failed: ${e}`);
        }
    }));
}
function deactivate() { }
//# sourceMappingURL=extension.js.map