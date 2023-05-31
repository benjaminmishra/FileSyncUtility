using System;
using System.IO;
using System.Net.Sockets;
using System.Text;
using System.Security.Cryptography;
using System;

namespace Client;

class Program
{
    public static async Task Main(string[] args)
    {
        const string HOST = "localhost";
        const int PORT = 6060;

        // "4756c5c884744c9fb5d6e6a0f0c71d9b"
        var apiKey = args[0].Trim() ?? "4756c5c884744c9fb5d6e6a0f0c71d9b";
        var directoryPath = args[1].Trim();

        var tcpClient = new FilSyncUtilTcpClient(HOST, PORT);

        var autheticationSuccess = await ConnectToServer(apiKey, tcpClient);

        if (!autheticationSuccess)
        {
            Console.WriteLine("Failed to autheticate");
            return;
        }
        var app = new App(tcpClient,directoryPath);

        await app.ListenToServerUpdates();

        Console.WriteLine("Press 'q' to quit the sample.");
        while (Console.Read() != 'q') ;
    }

    private static void OnChanged(object sender, FileSystemEventArgs e)
    {
        
    }

    private async static Task<bool> ConnectToServer(string apiKey, FilSyncUtilTcpClient client)
    {
        // send api key
        await client.SendAsync(apiKey);

        var response = await client.ReadLineAsync();
        Console.WriteLine(response);

        if (response!="SUCCESS")
        {
            Console.WriteLine("Authetication Failed");
            return false;
        }
        return true;
    }
}
