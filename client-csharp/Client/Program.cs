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

        // generate and send JWT
        var jwt = GenerateJWT(e.FullPath);
        var jwtBytes = Encoding.UTF8.GetBytes(jwt + "\n");
        stream.Write(jwtBytes, 0, jwtBytes.Length);

        // read and send file
        var fileBytes = File.ReadAllBytes(e.FullPath);
        stream.Write(fileBytes, 0, fileBytes.Length);
        
        // listen for reply
        var buffer = new byte[client.ReceiveBufferSize];
        var bytesRead = stream.Read(buffer, 0, client.ReceiveBufferSize);
        var response = Encoding.UTF8.GetString(buffer, 0, bytesRead);
        Console.WriteLine($"Response from server: {response}");
    }

    private static string GenerateJWT(string path)
    {
        var header = Convert.ToBase64String(Encoding.UTF8.GetBytes("{\"alg\":\"HS256\",\"typ\":\"JWT\"}"));
        var payload = Convert.ToBase64String(Encoding.UTF8.GetBytes("{\"path\":\"" + path + "\",\"iat\":1625529121,\"exp\":1685368627}"));
        var signatureRaw = ComputeHmacSha256(header + "." + payload, "mu9vTDxsLDZMfqP9NP+l81WjG6t4yYe8H8gLKH2X9wE=");
        var signature = Convert.ToBase64String(signatureRaw);
        return header + "." + payload + "." + signature;
    }

    private static byte[] ComputeHmacSha256(string data, string secret)
    {
        using var hmac = new HMACSHA256(Encoding.UTF8.GetBytes(secret));
        return hmac.ComputeHash(Encoding.UTF8.GetBytes(data));
    }
}
