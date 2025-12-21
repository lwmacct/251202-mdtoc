import * as vscode from "vscode";

export interface McmdtocConfig {
  enable: boolean;
  formatOnSave: boolean;
  cliPath: string;
  globalMode: boolean;
  minLevel: number;
  maxLevel: number;
  ordered: boolean;
}

export class Config {
  private get configuration(): vscode.WorkspaceConfiguration {
    return vscode.workspace.getConfiguration("mcMdtoc");
  }

  get<K extends keyof McmdtocConfig>(key: K): McmdtocConfig[K] {
    return this.configuration.get<McmdtocConfig[K]>(key) as McmdtocConfig[K];
  }

  getAll(): McmdtocConfig {
    return {
      enable: this.get("enable") ?? true,
      formatOnSave: this.get("formatOnSave") ?? false,
      cliPath: this.get("cliPath") ?? "",
      globalMode: this.get("globalMode") ?? false,
      minLevel: this.get("minLevel") ?? 1,
      maxLevel: this.get("maxLevel") ?? 3,
      ordered: this.get("ordered") ?? false,
    };
  }

  onDidChange(callback: () => void): vscode.Disposable {
    return vscode.workspace.onDidChangeConfiguration((e) => {
      if (e.affectsConfiguration("mcMdtoc")) {
        callback();
      }
    });
  }
}
