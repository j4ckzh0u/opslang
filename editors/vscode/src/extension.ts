import * as path from 'path';
import * as vscode from 'vscode';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    TransportKind
} from 'vscode-languageclient/node';

let client: LanguageClient | undefined;

export function activate(context: vscode.ExtensionContext) {
    const config = vscode.workspace.getConfiguration('opslang');
    const enableLSP = config.get<boolean>('enableLSP', true);

    if (!enableLSP) {
        console.log('OpsLang LSP 已禁用');
        return;
    }

    const serverPath = config.get<string>('serverPath', 'ops');

    // 启动 LSP 服务器
    const serverOptions: ServerOptions = {
        command: serverPath,
        args: ['lsp'],
        transport: TransportKind.stdio
    };

    // 客户端选项
    const clientOptions: LanguageClientOptions = {
        documentSelector: [
            { scheme: 'file', language: 'opslang' },
            { scheme: 'untitled', language: 'opslang' }
        ],
        synchronize: {
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.ops')
        },
        initializationOptions: {
            // 可传递给 LSP 服务器的初始配置
        }
    };

    // 创建客户端
    client = new LanguageClient(
        'opslang',
        'OpsLang Language Server',
        serverOptions,
        clientOptions
    );

    // 启动客户端（同时启动服务器）
    client.start().then(() => {
        console.log('OpsLang LSP 已启动');
    }).catch((err) => {
        console.error('OpsLang LSP 启动失败:', err);
        vscode.window.showWarningMessage(
            `OpsLang LSP 启动失败。请确保 "${serverPath}" 在 PATH 中，` +
            `或在设置中配置 opslang.serverPath。`
        );
    });

    // 注册命令
    context.subscriptions.push(
        vscode.commands.registerCommand('opslang.restartServer', async () => {
            if (client) {
                await client.stop();
            }
            client = new LanguageClient(
                'opslang',
                'OpsLang Language Server',
                serverOptions,
                clientOptions
            );
            await client.start();
            vscode.window.showInformationMessage('OpsLang LSP 已重启');
        })
    );
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    return client.stop();
}
