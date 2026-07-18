import type { main } from "../wailsjs/skell-gui/models";

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
					SkellPresent(): Promise<boolean>;
				};
			};
		};
	}
}

export {};