import type { main } from "../wailsjs/go/models";

declare global {
	interface Window {
		go: {
			main: {
				App: {
					RunSkell(args: string[]): Promise<main.SkellResult>;
					ReadFileContent(path: string): Promise<string>;
					ListDirectory(path: string): Promise<Array<main.FileEntry>>;
					SkellVersion(): Promise<string>;
					SelectDirectory(): Promise<string>;
					AuditLogPath(): Promise<string>;
					GlobalRootDir(): Promise<string>;
				};
			};
		};
	}
}

export {};