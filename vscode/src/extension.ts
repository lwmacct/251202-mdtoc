import * as vscode from "vscode";
import { Config } from "./config";
import { MdtocCli } from "./cli";
import { MdtocFormatter } from "./formatter";

let outputChannel: vscode.OutputChannel;

export async function activate(context: vscode.ExtensionContext): Promise<void> {
  outputChannel = vscode.window.createOutputChannel("mdtoc");

  const config = new Config();
  const cli = new MdtocCli(config, outputChannel);
  const formatter = new MdtocFormatter(cli, config);

  const cliAvailable = await cli.checkAvailability();
  if (!cliAvailable) {
    showInstallPrompt();
  }

  context.subscriptions.push(vscode.languages.registerDocumentFormattingEditProvider({ language: "markdown", scheme: "file" }, formatter));

  context.subscriptions.push(
    vscode.commands.registerCommand("mdtoc.updateToc", () => {
      formatter.formatActiveDocument();
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand("mdtoc.deleteToc", () => {
      formatter.deleteFromActiveDocument();
    }),
  );

  context.subscriptions.push(
    vscode.workspace.onWillSaveTextDocument(async (event) => {
      if (!config.get("enable") || !config.get("formatOnSave")) {
        return;
      }

      if (event.document.languageId !== "markdown") {
        return;
      }

      if (event.document.uri.scheme !== "file") {
        return;
      }

      const hasTocMarker = event.document.getText().includes("<!--TOC-->");
      if (!hasTocMarker) {
        return;
      }

      event.waitUntil(
        (async (): Promise<vscode.TextEdit[]> => {
          const filePath = event.document.uri.fsPath;
          await cli.updateToc(filePath);
          return [];
        })(),
      );
    }),
  );

  context.subscriptions.push(
    config.onDidChange(async () => {
      const available = await cli.checkAvailability();
      if (!available) {
        showInstallPrompt();
      }
    }),
  );

  context.subscriptions.push(outputChannel);

  outputChannel.appendLine("mdtoc extension activated");
}

export function deactivate(): void {
  outputChannel?.appendLine("mdtoc extension deactivated");
}

async function showInstallPrompt(): Promise<void> {
  const selection = await vscode.window.showWarningMessage("mdtoc CLI not found. Please install it to use this extension.", "Install with Go", "View Documentation", "Configure Path");

  switch (selection) {
    case "Install with Go": {
      const terminal = vscode.window.createTerminal("mdtoc install");
      terminal.show();
      terminal.sendText("go install github.com/lwmacct/251202-mdtoc/cmd/mdtoc@latest");
      vscode.window.showInformationMessage("Running installation command. After it completes, reload the window.", "Reload Window").then((action) => {
        if (action === "Reload Window") {
          vscode.commands.executeCommand("workbench.action.reloadWindow");
        }
      });
      break;
    }

    case "View Documentation":
      vscode.env.openExternal(vscode.Uri.parse("https://github.com/lwmacct/251202-mdtoc"));
      break;

    case "Configure Path":
      vscode.commands.executeCommand("workbench.action.openSettings", "mcMdtoc.cliPath");
      break;
  }
}
