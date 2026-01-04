# 🎙️ VOMic
Turn your phone to act as a microphone source, for pipewire or pulseaudio.


> [!IMPORTANT]
> This program depends on the `pactl` command-line utility. So make sure its available in your $PATH. 


> [!CAUTION]
> Be aware this project is still a work in progress, so expect bugs, breakage and unexpected behaviour.

## Usage
The program works in a client/server model. We start by setting up the desktop side.
### Desktop
1. Use the `go` toolchain to install the latest binary in your $PATH, Using the command below:
```bash
go install github.com/Layer8Collective/vomic/cmd@latest
```
2. Pickup the latest binary, from a latest nightly release that we ship pariodically, in the [github releases](https://github.com/layer8co/VOMic/releases) page.

3. Run the binary. You can also change some options by providing the command-line arguments:
```bash
$ ./vomic-linux-amd64 -h
  -c int
        Control Port that is used in handshaking and connection-tracking process. (default 8125)
  -m int
        Media Port that is used for audio transfership. (default 16001)

$ ./vomic-linux-amd64
Created a pipe-source with module-index: 536870916

INFO: 2026/01/04 12:11:32 🎤 VO-MIC Listening on TCP network:  [::]:8125
INFO: 2026/01/04 12:11:32 Waiting for new connection...
```
## Android 
1. Download the Android app from the [github releases](https://github.com/layer8co/VOMic/releases)
2. Install. Specify the `host` (IP address) and `port`, then press Connect.



> [!IMPORTANT]
> You can also make things more convinent for youself by wrapping the application around a custom, handmade systemd service.





## FAQ

#### Is The app doesn't have the ability to mute the sound for some time?
It would be added in the near future!

#### Cant we change the volume somehow through the android app?
No. This would be better achived using tools like `pavucontrol`.

#### What about noise cancellation, echo suppression and ....?
This would be achived using application level tools. E.g Discord's krisp


