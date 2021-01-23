# Calcetto.IOT

copy `config.json.template` to `config.json` before running.

### Test Serial (system with ttypX devices)

send messages to serial `/dev/ttyp5` with:

    $ echo "GOAL_BLUE_1" > /dev/ptyp5 
    $ echo "GOAL_RED_1" > /dev/ptyp5 

### Test Serial (system without ttypX devices)
Run:

    $ socat -d -d pty,raw,echo=0 pty,raw,echo=0

and set the read devices that you get from the command output into `config.json` . Then run:

    $ echo "GOAL_BLUE_1" > /dev/pts/3 
    $ echo "GOAL_RED_1" > /dev/pts/3

Replace `/dev/pts/3` with the correct input device.

# Build

Build for the same system:

    $ go build -o build/calcetto-server *.go

or cross build for Raspberry Pi:

    $ env GOOS=linux GOARCH=arm GOARM=5 go build -o build/calcetto-server-rasp *.go