package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/Layer8Collective/vomic"
	"github.com/Layer8Collective/vomic/internal/logging"
)

func main() {
	ctrlPortPtr := flag.Int("c", 8125, "Control Port that is used in handshaking and connection-tracking process.")
	mediaPortPointer := flag.Int("m", 16001, "Media Port that is used for audio transfership.")

	flag.Parse()

	p := PactlConf{
		SampleRate: 44100,
		PipePath:   "/tmp/vomic-pipe",
		SourceName: "VO-Mic",
		Format:     vomic.Signed_le_16,
		Channel:    vomic.Mono,
	}

	id := createPipeSource(p)

	defer disposePipeSource(id)

	f, err := os.OpenFile(p.PipePath, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		log.Fatalf("Error when opening the file %s: %s", f.Name(), err)
	}
	defer f.Close()

	infoLogger := logging.NewInfoLogger()

	_, ok := os.LookupEnv("VOMIC_DEBUG")
	debugLogger := logging.NewDebugLogger(ok)

	server := vomic.VOMic{
		ControlPort: uint16(*ctrlPortPtr),
		MediaPort:   uint16(*mediaPortPointer),
		PipeSource:  f,
		InfoLogger:  infoLogger,
		DebugLogger: debugLogger,
	}

	err = server.Listen()

	if err != nil {
		fmt.Println(err)
	}

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
func createPipeSource(c PactlConf) string {
	_, err := exec.LookPath("pactl")

	if err != nil {
		log.Fatalln("Does the program `pactl` exists? We cannot find it.")
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
		log.Fatalf("Failed to create the pipe-source: %s \n output: %s", err.Error(), string(output))
	}

	fmt.Println("Created a pipe-source with module-index:", string(output))

	return string(output)
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
