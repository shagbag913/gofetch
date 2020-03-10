gofetch:
	go build
	strip -s gofetch

clean:
	rm gofetch

install: 
	sudo cp gofetch /usr/bin
