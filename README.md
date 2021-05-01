# snippet

## Installation

### Redhat

Prerequisites:

```shell
wget https://sourceforge.net/projects/libpng/files/libpng16/1.6.37/libpng-1.6.37.tar.gz
tar xvf libpng-1.6.37.tar.gz
cd libpng-1.6.37/
./configure 
make
sudo make install
sudo ln -s /usr/local/lib/libpng16.so.16 /usr/lib64/libpng16.so.16
```


## Development

### Windows cross-compilation

```shell
sudo apt install gcc-mingw-w64-x86-64 libz-mingw-w64-dev

GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ go build
```
