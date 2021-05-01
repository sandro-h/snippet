# snippet

## Development

### Windows cross-compilation

```shell
sudo apt install gcc-mingw-w64-x86-64 libz-mingw-w64-dev

GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ go build
```
