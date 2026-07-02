import * as vscode from "vscode";
import { spawn, execSync } from "child_process";
import { Config } from "./config";

export interface CliResult {
  success: boolean;
  output: string;
  error?: string;
}

export class MdtocCli {
  private cliPath: string | undefined;

  constructor(
    private config: Config,
    private output: vscode.OutputChannel,
  ) {}

  async checkAvailability(): Promise<boolean> {
    const configPath = this.config.get("cliPath");
    if (configPath) {
      try {
        execSync(`"${configPath}" version`, { stdio: "pipe" });
        this.cliPath = configPath;
        this.log(`Using configured CLI path: ${configPath}`);
        return true;
      } catch {
        this.log(`Configured CLI path not working: ${configPath}`);
      }
    }

    try {
      const which = process.platform === "win32" ? "where" : "which";
      const result = execSync(`${which} mdtoc`, { encoding: "utf-8" }).trim();
      this.cliPath = result.split("\n")[0];
      this.log(`Found CLI in PATH: ${this.cliPath}`);
      return true;
    } catch {
      this.log("mdtoc CLI not found in PATH");
      return false;
    }
  }

  getCliPath(): string | undefined {
    return this.cliPath;
  }

  async updateToc(filePath: string): Promise<CliResult> {
    const args = this.buildArgs(["-i", filePath]);
    return this.execute(args);
  }

  async deleteToc(filePath: string): Promise<CliResult> {
    return this.execute(["-d", filePath]);
  }

  private buildArgs(baseArgs: string[]): string[] {
    const args: string[] = [];
    const cfg = this.config.getAll();

    if (cfg.globalMode) {
      args.push("-g");
    }

    if (cfg.minLevel !== 1) {
      args.push("-m", String(cfg.minLevel));
    }

    if (cfg.maxLevel !== 3) {
      args.push("-M", String(cfg.maxLevel));
    }

    if (cfg.ordered) {
      args.push("-o");
    }

    args.push(...baseArgs);
    return args;
  }

  private execute(args: string[]): Promise<CliResult> {
    return new Promise((resolve) => {
      if (!this.cliPath) {
        resolve({
          success: false,
          output: "",
          error: "CLI not found. Please install mdtoc.",
        });
        return;
      }

      this.log(`Executing: ${this.cliPath} ${args.join(" ")}`);

      const proc = spawn(this.cliPath, args, {
        stdio: ["pipe", "pipe", "pipe"],
      });

      let stdout = "";
      let stderr = "";

      proc.stdout.on("data", (data) => {
        stdout += data.toString();
      });

      proc.stderr.on("data", (data) => {
        stderr += data.toString();
      });

      proc.on("close", (code) => {
        if (stdout) {
          this.log(`stdout: ${stdout}`);
        }
        if (stderr) {
          this.log(`stderr: ${stderr}`);
        }
        this.log(`Exit code: ${code}`);

        resolve({
          success: code === 0,
          output: stdout,
          error: code !== 0 ? stderr || `Process exited with code ${code}` : undefined,
        });
      });

      proc.on("error", (err) => {
        this.log(`Error: ${err.message}`);
        resolve({
          success: false,
          output: "",
          error: err.message,
        });
      });
    });
  }

  private log(message: string): void {
    const timestamp = new Date().toISOString().substring(11, 19);
    this.output.appendLine(`[${timestamp}] ${message}`);
  }
}
