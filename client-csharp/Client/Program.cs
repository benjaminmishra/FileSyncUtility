using System;
using System.IO;
using System.Net.Sockets;
using System.Text;
using System.Security.Cryptography;

namespace Client;

class Program
{
    static void Main(string[] args)
    {
        var watcher = new FileSystemWatcher(".");

        watcher.Changed += OnChanged;

        watcher.EnableRaisingEvents = true;

        Console.WriteLine("Press 'q' to quit the sample.");
        while (Console.Read() != 'q') ;
    }

    private static void OnChanged(object sender, FileSystemEventArgs e)
    {
        Console.WriteLine($"File: {e.FullPath} {e.ChangeType}");

        var client = new TcpClient("localhost", 8080);

        var stream = client.GetStream();

        // send api key
        stream.Write(new byte[1], 0, "".Length);

        // read and send file
        var fileBytes = File.ReadAllBytes(e.FullPath);
        stream.Write(fileBytes, 0, fileBytes.Length);
        
        // listen for reply
        var buffer = new byte[client.ReceiveBufferSize];
        var bytesRead = stream.Read(buffer, 0, client.ReceiveBufferSize);
        var response = Encoding.UTF8.GetString(buffer, 0, bytesRead);
        Console.WriteLine($"Response from server: {response}");
    }
}
