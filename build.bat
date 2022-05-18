SET CGO_ENABLED=0
SET GOARCH=amd64

SET GOOS=windows
go build -o target/forward_windows_amd64.exe forward/cmd/forward

SET GOOS=linux
go build -o target/forward_linux_amd64 forward/cmd/forward

SET GOOS=darwin
go build -o target/forward_macos_amd64 forward/cmd/forward