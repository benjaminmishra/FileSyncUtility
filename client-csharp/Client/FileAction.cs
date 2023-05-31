public struct FileAction
{
    public FileAction(string action, string type, string name, long size)
    {
        Action = action;
        Type = type;
        Name = name;
        Size = size;
    }

    public string Action { get; init; }
    public string Type { get; init; }
    public string Name { get; init; }
    public long Size { get; init; }

    public override string ToString()
    {
        return $"ACTION : {Action}, TYPE : {Type} , NAME : {Name}, SIZE : {Size}\n";
    }
}