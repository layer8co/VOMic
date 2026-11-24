# 🎙️ VOMic
Turn your phone to act as a microphone source. Exclusive for pipewire or
pulseaudio users.


> [!IMPORTANT]
> This program depends on the `pactl` command-line utility. Which should be
> available in all systems that are using the pipewire/pulseaudio sound servers.


## Installation
The program works in a client-server model. We start by setting up the server
side, which is our Linux workstation.

### Server(Linux workstation)

#### Method #1: The Python script (simplest)
Considering you have a Python interpreter already installed, you can use our Python
script which installs the program on your machine:
```
curl -fsSL https://github.com/layer8co/VOMic/raw/refs/heads/main/go/installer/main.py | python
```
The script adheres to *XDG Specification*, so it tries to install the command in
the `$PATH/.local/bin` directory.


#### Method #2: The go toolchain
Use the `go` toolchain to install the latest binary, Using the command below:
```bash
go install github.com/Layer8Collective/vomic/cmd@latest
```

### Client(Android)
Download and install the Android app from [Github releases](https://github.com/layer8co/VOMic/releases)

## Usage

Run the binary. You can also change some options by providing the command-line arguments:
```bash
$ vomic -h
  -c int
        Control Port that is used in handshaking and connection-tracking process. (default 8125)
  -m int
        Media Port that is used for audio transfership. (default 16001)

$ ./vomic-linux-amd64
Created a pipe-source with module-index: 536870916

INFO: 2026/01/04 12:11:32 🎤 VO-MIC Listening on TCP network:  [::]:8125
INFO: 2026/01/04 12:11:32 Waiting for new connection...
```

Then, on the android side, use your computer's IP/Port to initiate the
connection. 


Then, the `VOMic` input-source should be available to use in your system.

## FAQ

#### Cant we change the volume somehow through the android app?
No. This would be better achived using tools like `pavucontrol`.

#### What about noise cancellation, echo suppression and ....?
This would be achived using application level tools. E.g Discord's krisp


