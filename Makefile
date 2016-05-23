all: build

deps:
	glide up

build:
	go build -o site main.go

package: build
	sudo rm -rf build/
	mkdir -p build/opt/0x7ff
	mkdir -p build/opt/0x7ff/resources
	mkdir -p build/etc/systemd/system
	cp site build/opt/0x7ff/site
	cp -r resources build/opt/0x7ff
	cp systemd/site-0x7ff.service build/etc/systemd/system/site-0x7ff.service
	sudo chown -R gosig: build/opt
	sudo chown -R root: build/etc
	echo 2.0 > build/debian-binary
	echo "Package: site-0x7ff" > build/control
	echo "Version: 1.0" >> build/control
	echo "Architecture: all" >> build/control
	echo "Section: net" >> build/control
	echo "Maintainer: cubeee <cubeee.gh@gmail.com>" >> build/control
	echo "Priority: optional" >> build/control
	echo "Homepage: https://0x7ff.com/"
	echo "Description: 0x7ff.com" >> build/control
	tar cvzf build/data.tar.gz -C build etc opt
	tar cvzf build/control.tar.gz -C build control
	cd build && ar rc site-0x7ff.deb debian-binary control.tar.gz data.tar.gz && cd ..
