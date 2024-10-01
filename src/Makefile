# PodResourceCalculator

all: PodResourceCalculator

PodResourceCalculator: main.go
	go build -o PodResourceCalculator

install: PodResourceCalculator
	sudo install -m 755 PodResourceCalculator /usr/local/bin/

clean:
	@rm -f PodResourceCalculator
