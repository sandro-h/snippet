BASE_VERSION=0.4.0
REPO_OWNER=sandro-h
BUILD_CENTOS_IMAGE_VERSION=0.0.1
BUILD_CENTOS_IMAGE_TAG=ghcr.io/${REPO_OWNER}/snippet-centos-build:${BUILD_CENTOS_IMAGE_VERSION}
PUSH_IMAGE=false


.PHONY: docker-login
docker-login:
	echo $$DOCKER_PWD | docker login ghcr.io -u ${REPO_OWNER} --password-stdin

.PHONY: push-centos-image
push-centos-image:
ifeq ($(PUSH_IMAGE), true)
	docker push ${BUILD_CENTOS_IMAGE_TAG}
endif

.PHONY: build-centos-image
build-centos-image:
	docker build \
		--tag ${BUILD_CENTOS_IMAGE_TAG} \
		centos/

.PHONY: ensure-centos-image
ensure-centos-image:
	docker pull ${BUILD_CENTOS_IMAGE_TAG} || make build-centos-image push-centos-image

.PHONY: ensure-centos-image
build-centos: ensure-centos-image
	docker run --rm \
		-v $$(pwd):/src \
		-e DEV_UID=$$(id -u) \
		-w /src \
		${BUILD_CENTOS_IMAGE_TAG} \
		"export PATH=/usr/local/go/bin:\$$PATH && go build -o snippet-centos ${EXTRA_BUILD_ARGS} && chown \$$DEV_UID snippet-centos"

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ go build -ldflags "-H=windowsgui ${EXTRA_WIN_LDFLAGS}" ${EXTRA_WIN_BUILD_ARGS}

.PHONY: build-linux
build-linux:
	go build ${EXTRA_BUILD_ARGS}

.PHONY: test
test:
	go test -v -coverprofile="coverage.out" ./...

.PHONY: lint
lint:
	golint -set_exit_status ./...

.PHONY: install-sys-packages
install-sys-packages:
	sudo apt update && \
	sudo apt install gcc libc6-dev \
	libx11-dev xorg-dev libxtst-dev libpng++-dev \
	xcb libxcb-xkb-dev x11-xkb-utils libx11-xcb-dev libxkbcommon-x11-dev \
	libxkbcommon-dev xsel xclip \
	libgl1-mesa-dev \
	gcc-mingw-w64-x86-64 libz-mingw-w64-dev

###########################################################################################
# Releasing
###########################################################################################

.PHONY: release
release:
	git tag v${BASE_VERSION}
	git push origin v${BASE_VERSION}


.PHONY: print-version
print-version:
	echo "::set-output name=version::${BASE_VERSION}.$${VERSION_NUMBER:-0}"

.PHONY: build-all-optimized
build-all-optimized: EXTRA_BUILD_ARGS=-ldflags='-s -w'
build-all-optimized: EXTRA_WIN_LDFLAGS=-s -w
build-all-optimized: build-linux build-centos build-windows

upx:
	wget https://github.com/upx/upx/releases/download/v3.96/upx-3.96-amd64_linux.tar.xz
	tar xf upx-3.96-amd64_linux.tar.xz
	mv upx-3.96-amd64_linux/ upx

.PHONY: compress-binaries
compress-binaries: upx
	chmod +x snippet snippet-centos
	upx/upx -q --brute snippet
	upx/upx -q --brute snippet-centos
	upx/upx -q --brute snippet.exe