package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Layer8Collective/vomic"
	"github.com/Layer8Collective/vomic/internal/logging"
)

func main() {
	ctrlPortPtr := flag.Int("c", 8125, "TCP Control Port that is used in handshaking and connection-tracking purposes.")
	mediaPortPtr := flag.Int("m", 16001, "UDP Media Port that is used for audio transfership.")

	flag.Parse()

	infoLogger := logging.NewInfoLogger()

	_, ok := os.LookupEnv("VOMIC_DEBUG")
	debugLogger := logging.NewDebugLogger(ok)

	if err := vomic.VerifyPort(*ctrlPortPtr); err != nil {
		infoLogger.Println(err.Error())
		return
	}

	if err := vomic.VerifyPort(*mediaPortPtr); err != nil {
		infoLogger.Println(err.Error())
		return
	}

	if err := vomic.IsUDPPortAvailable("", *mediaPortPtr); err != nil {
		infoLogger.Printf("Cannot bind to udp port %d: %s \n", *mediaPortPtr, err)
		return
	}

	p := PactlConf{
		SampleRate: 44100,
		PipePath:   "/tmp/vomic-pipe",
		SourceName: "VO-Mic",
		Format:     vomic.Signed_le_16,
		Channel:    vomic.Mono,
	}

	id, err := createPipeSource(p)
	if err != nil {
		log.Println(err)
		return
	}

	defer disposePipeSource(id)
	f, err := os.OpenFile(p.PipePath, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		log.Printf("Error when opening the pipeSource %s: %s", f.Name(), err)
		return
	}
	defer f.Close()

	server := vomic.VOMic{
		ControlPort: uint16(*ctrlPortPtr),
		MediaPort:   uint16(*mediaPortPtr),
		PipeSource:  f,
		InfoLogger:  infoLogger,
		DebugLogger: debugLogger,
	}
	// The application context respects the Interrupt and SIGTERM from Linux.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	err = server.ListenAndServe(ctx)

	if err != nil {
		fmt.Println(err)
	}
	infoLogger.Println("Good bye!")

}

type PactlConf struct {
	SampleRate int32

	// The pipeSource filename
	PipePath string

	// It's the name that is displayed to the users when they want to select
	// this microphone source.
	SourceName string

	// The list of availble formats are listed in the url below (We are
	// using signed little endian 16 bit):
	// https://docs.pipewire.org/page_man_pipewire-pulse-modules_7.html
	Format vomic.PWFormat

	// Mono or Stereo.
	Channel vomic.PWChannel
}

// Creates the pipeSource and returns the module index.
func createPipeSource(c PactlConf) (string, error) {
	_, err := exec.LookPath("pactl")

	if err != nil {
		return "", fmt.Errorf("Does the program `pactl` exists? We cannot find it.")
	}

	arg1 := "load-module"
	arg2 := "module-pipe-source"
	arg3 := fmt.Sprintf("source_name=%s", c.SourceName)
	arg4 := fmt.Sprintf("file=%s", c.PipePath)
	arg5 := fmt.Sprintf("format=%s", c.Format)
	arg6 := fmt.Sprintf("rate=%d", c.SampleRate)
	arg7 := fmt.Sprintf("channels=%d", c.Channel)

	cmd := exec.Command("pactl", arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("Failed to create the pipe-source: %s \n output: %s", err.Error(), string(output))
	}

	fmt.Println("Created a pipe-source with module-index:", string(output))

	return string(output), nil
}

func disposePipeSource(id string) (string, error) {
	id = strings.TrimSpace(id)

	cmd := exec.Command("pactl", "unload-module", id)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("Failed to unload the pipe-source module: %s \n", output)
	}
	return string(output), nil
}
