.PHONY: build clean update-vendor test

build:
	cd tools && go run ./extend --targets-dir ../targets

clean:
	rm -rf svgsJa graphicsJa.txt preview.html

update-vendor:
	git submodule update --remote vendor/animCJK

test:
	cd tools && go test ./...
