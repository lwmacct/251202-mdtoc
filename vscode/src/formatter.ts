import * as vscode from "vscode";
import * as fs from "fs";
import { MdtocCli } from "./cli";
import { Config } from "./config";

export class MdtocFormatter implements vscode.DocumentFormattingEditProvider {
  constructor(
    private cli: MdtocCli,
    private config: Config,
  ) {}

  async provideDocumentFormattingEdits(document: vscode.TextDocument, _options: vscode.FormattingOptions, _token: vscode.CancellationToken): Promise<vscode.TextEdit[]> {
    if (!this.config.get("enable")) {
      return [];
    }

    if (document.languageId !== "markdown") {
      return [];
    }

    const filePath = document.uri.fsPath;

    if (document.isDirty) {
      await document.save();
    }

    const contentBefore = fs.readFileSync(filePath, "utf-8");

    const result = await this.cli.updateToc(filePath);

    if (!result.success) {
      if (result.error) {
        vscode.window.showWarningMessage(`mdtoc: ${result.error}`);
      }
      return [];
    }

    const contentAfter = fs.readFileSync(filePath, "utf-8");

    if (contentBefore === contentAfter) {
      return [];
    }

    const fullRange = new vscode.Range(document.positionAt(0), document.positionAt(contentBefore.length));

    return [vscode.TextEdit.replace(fullRange, contentAfter)];
  }

  async formatActiveDocument(): Promise<boolean> {
    const editor = vscode.window.activeTextEditor;
    if (!editor || editor.document.languageId !== "markdown") {
      vscode.window.showInformationMessage("mdtoc: Please open a Markdown file first.");
      return false;
    }

    const document = editor.document;
    const filePath = document.uri.fsPath;

    if (document.isDirty) {
      await document.save();
    }

    const result = await this.cli.updateToc(filePath);

    if (result.success) {
      await vscode.commands.executeCommand("workbench.action.files.revert");
      vscode.window.showInformationMessage("mdtoc: TOC updated successfully.");
      return true;
    } else {
      vscode.window.showErrorMessage(`mdtoc: ${result.error || "Failed to update TOC"}`);
      return false;
    }
  }

  async deleteFromActiveDocument(): Promise<boolean> {
    const editor = vscode.window.activeTextEditor;
    if (!editor || editor.document.languageId !== "markdown") {
      vscode.window.showInformationMessage("mdtoc: Please open a Markdown file first.");
      return false;
    }

    const document = editor.document;
    const filePath = document.uri.fsPath;

    if (document.isDirty) {
      await document.save();
    }

    const result = await this.cli.deleteToc(filePath);

    if (result.success) {
      await vscode.commands.executeCommand("workbench.action.files.revert");
      vscode.window.showInformationMessage("mdtoc: TOC deleted successfully.");
      return true;
    } else {
      vscode.window.showErrorMessage(`mdtoc: ${result.error || "Failed to delete TOC"}`);
      return false;
    }
  }
}
