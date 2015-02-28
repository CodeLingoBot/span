TARGETS = span span-hspec span-convert

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test:
	go get -d && go test -v

cover:
	go test -cover ./...

span: imports
	go build -o span cmd/span/main.go

span-hspec: imports
	go build -o span-hspec cmd/span-hspec/main.go

span-convert: imports
	go build -o span-convert cmd/span-convert/main.go

imports:
	goimports -w .

clean:
	rm -f $(TARGETS)
	rm -f span_*deb
	rm -f span-*rpm
	rm -rf ./packaging/deb/span/usr

deb: $(TARGETS)
	mkdir -p packaging/deb/span/usr/sbin
	cp $(TARGETS) packaging/deb/span/usr/sbin
	cd packaging/deb && fakeroot dpkg-deb --build span .
	mv packaging/deb/span_*.deb .

rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/rpm/span.spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	./packaging/rpm/buildrpm.sh span
	cp $(HOME)/rpmbuild/RPMS/x86_64/span*.rpm .

cloc:
	cloc --max-file-size 1 .
