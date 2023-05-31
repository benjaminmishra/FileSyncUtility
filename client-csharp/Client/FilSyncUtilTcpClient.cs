using System.Net.Sockets;
using System.Text;

namespace Client;

public class FilSyncUtilTcpClient
{
    private readonly TcpClient _client;
    private readonly NetworkStream _stream;
    private readonly StreamReader _reader;

    public FilSyncUtilTcpClient(string host, int port)
    {
        _client = new TcpClient(host, port);
        _stream = _client.GetStream();
        _reader = new StreamReader(_stream, Encoding.UTF8);
    }

    public async Task<string?> ReadLineAsync()
    {
        return await _reader.ReadLineAsync();
    }

    public async Task SendAsync(string requestMsg)
    {
        // sends a newline character at the end of the message
        var requestBytes = Encoding.UTF8.GetBytes(requestMsg+"\n");
        await _stream.WriteAsync(requestBytes,0,requestBytes.Length);
    }

    public async Task StreamIntoFileAsync(string fileName, long fileSize)
    {
        using var file = new FileStream(fileName, FileMode.Create, FileAccess.Write);
                    
        // Stream data from the connection to the file
        byte[] buffer = new byte[fileSize];
        int bytesRead = await _stream.ReadAsync(buffer, 0, (int)fileSize);
        file.Write(buffer, 0, bytesRead);
    }
}