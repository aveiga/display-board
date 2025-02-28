linux-arm-build:
	docker run --rm -v $(PWD):/go/src/app -w /go/src/app \
	debian:bullseye \
	sh -c 'apt-get update && \
	apt-get install -y gcc-arm-linux-gnueabihf libc6-dev-armhf-cross \
	pkg-config git wget && \
	wget https://go.dev/dl/go1.22.4.linux-amd64.tar.gz && \
	tar -C /usr/local -xzf go1.22.4.linux-amd64.tar.gz && \
	rm go1.22.4.linux-amd64.tar.gz && \
	dpkg --add-architecture armhf && \
	apt-get update && \
	apt-get install -y \
	libx11-dev:armhf libxrandr-dev:armhf libxinerama-dev:armhf libxcursor-dev:armhf \
	libxi-dev:armhf libgl1-mesa-dev:armhf libgles2-mesa-dev:armhf libxxf86vm-dev:armhf && \
	mkdir -p /usr/arm-linux-gnueabihf/lib/pkgconfig && \
	echo "Name: gl" > /usr/arm-linux-gnueabihf/lib/pkgconfig/gl.pc && \
	echo "Description: Mesa OpenGL library" >> /usr/arm-linux-gnueabihf/lib/pkgconfig/gl.pc && \
	echo "Version: 22.3.6" >> /usr/arm-linux-gnueabihf/lib/pkgconfig/gl.pc && \
	echo "Libs: -lGL" >> /usr/arm-linux-gnueabihf/lib/pkgconfig/gl.pc && \
	echo "Cflags: -I/usr/include" >> /usr/arm-linux-gnueabihf/lib/pkgconfig/gl.pc && \
	PATH=/usr/local/go/bin:$$PATH \
	GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc \
	PKG_CONFIG_PATH=/usr/arm-linux-gnueabihf/lib/pkgconfig \
	PKG_CONFIG_LIBDIR=/usr/lib/arm-linux-gnueabihf/pkgconfig:/usr/share/pkgconfig \
	CGO_LDFLAGS="-L/usr/lib/arm-linux-gnueabihf -lGLESv2" \
	/usr/local/go/bin/go build -tags "gles" -o dp -v cmd/display-board/main.go'

build:
	go build -o dp -v cmd/display-board/main.go

run:
	go run cmd/display-board/main.go


