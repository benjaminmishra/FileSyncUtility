build:
	dotnet build client-csharp/Client.sln
	cd client-go && go build -o client-go main.go filewatcher.go 

