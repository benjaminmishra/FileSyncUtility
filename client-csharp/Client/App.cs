using System.Text;
using System;

namespace Client;

public class App
{
    private readonly FilSyncUtilTcpClient _fileSyncUtilTcpClient;
    private readonly string _rootDirectoryPath;
    private readonly FileSystemWatcher _watcher;

    public App(FilSyncUtilTcpClient tcpClient, string rootDirectoryPath)
    {
        _fileSyncUtilTcpClient = tcpClient;
        _rootDirectoryPath = rootDirectoryPath;
        _watcher = new FileSystemWatcher(rootDirectoryPath);
        _watcher.Changed += OnFileSystemChanged;
        _watcher.EnableRaisingEvents = true;
    }

    public async Task ListenToServerUpdates()
    {
        Console.WriteLine("Listening to server updates");
        // wait infinitely , but break if we find DONE
        while(true) 
        {
            var receivedActionString =  await _fileSyncUtilTcpClient.ReadLineAsync();

            if (receivedActionString == "DONE")
                break;

            var receivedFileAction = ParseActionString(receivedActionString!) ?? throw new Exception("Received nothing from server");
            
            await ExecuteFileActionString(receivedFileAction);
        }
    }

    private void OnFileSystemChanged(object sender, FileSystemEventArgs e)
    {
        Console.WriteLine($"File: {e.FullPath} {e.ChangeType}");
    }

    private FileAction? ParseActionString(string actionStr)
    {
        var parts = actionStr.Split(',');

        if (parts.Length != 4)
        {
            throw new ArgumentException("Invalid action string" + actionStr);
        }

        var actionPart = parts[0].Trim().Split(" : ");
        if (actionPart.Length != 2)
        {
            throw new ArgumentException("Invalid action string");
        }

        var typePart = parts[1].Trim().Split(" : ");
        if (typePart.Length != 2)
        {
            throw new ArgumentException("Invalid type string");
        }

        var namePart = parts[2].Trim().Split(" : ");
        if (namePart.Length != 2)
        {
            throw new ArgumentException("Invalid name string");
        }

        var sizePart = parts[3].Trim().Split(" : ");
        if (sizePart.Length != 2)
        {
            throw new ArgumentException("Invalid size string");
        }

        return new FileAction(actionPart[1].Trim(), typePart[1].Trim(), namePart[1].Trim(), long.Parse(sizePart[1].Trim()));
    }


    private async Task ExecuteFileActionString(FileAction fileAction)
    {
        var path = Path.GetFullPath(_rootDirectoryPath);
        if (!Path.Exists(path))
        {
            Directory.CreateDirectory(path, UnixFileMode.UserWrite);
        }
        Directory.SetCurrentDirectory(path);

        if (fileAction.Action == "CREATE")
        {
            switch (fileAction.Type)
            {
                case "FILE":
                    // Send readiness confirmation after processing the line
                    await _fileSyncUtilTcpClient.SendAsync("READY");
                    // derive total number of bytes we need to read
                    await _fileSyncUtilTcpClient.StreamIntoFileAsync(Path.Combine(path,fileAction.Name), fileAction.Size);

                    break;
                case "DIR":
                    // Create a directory
                    Directory.CreateDirectory(fileAction.Name);
                    break;
                default:
                    Console.WriteLine($"No action taken for type {fileAction.Type}");
                    break;
            }

            // Send done confirmation after processing each line
            await _fileSyncUtilTcpClient.SendAsync("OK");
        }
    }
}