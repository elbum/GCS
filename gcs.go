package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

const usage = `Usage: caption <audiofile>
Audio file must be a 16-bit signed little-endian encoded
with a sample rate of 16000.
The path to the audio file may be a GCS URI (gs://...).
`

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}

	var runFunc func(io.Writer, string) error

	path := os.Args[1]
	if strings.Contains(path, "://") {
		runFunc = recognizeGCS
	} else {
		runFunc = recognize
	}

	// Perform the request.
	if err := runFunc(os.Stdout, os.Args[1]); err != nil {
		log.Fatal(err)
	}
}

// [START speech_transcribe_sync_gcs]

func recognizeGCS(w io.Writer, gcsURI string) error {
	ctx := context.Background()

	client, err := speech.NewClient(ctx)
	if err != nil {
		return err
	}

	// Send the request with the URI (gs://...)
	// and sample rate information to be transcripted.
	resp, err := client.Recognize(ctx, &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_LINEAR16,
			SampleRateHertz: 16000,
			LanguageCode:    "ko-KR",
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{Uri: gcsURI},
		},
	})

	// Print the results.
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			fmt.Fprintf(w, "\"%v\" (confidence=%3f)\n", alt.Transcript, alt.Confidence)
		}
	}
	return nil
}

// [END speech_transcribe_sync_gcs]

// [START speech_transcribe_sync]
type results map[string]string

var wg sync.WaitGroup

func recognize(w io.Writer, file string) error {
	runtime.GOMAXPROCS(runtime.NumCPU()) // Maximize CPU Utilization
	throttle := 64

	js := make(results, 10)

	matches, _ := filepath.Glob(os.Args[1] + "/*")
	for _, file := range matches {
		fmt.Println("Processing =", file)

		wg.Add(1)

		go func(file string) {

			// 작업종료 알림
			defer wg.Done()

			ctx := context.Background()

			client, err := speech.NewClient(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			data, err := ioutil.ReadFile(file)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Send the contents of the audio file with the encoding and
			// and sample rate information to be transcripted.
			resp, err := client.Recognize(ctx, &speechpb.RecognizeRequest{
				Config: &speechpb.RecognitionConfig{
					Encoding:        speechpb.RecognitionConfig_LINEAR16,
					SampleRateHertz: 16000,
					LanguageCode:    "ko-KR",
				},
				Audio: &speechpb.RecognitionAudio{
					AudioSource: &speechpb.RecognitionAudio_Content{Content: data},
				},
			})

			if err != nil {
				return
			}

			// Print the results.
			for _, result := range resp.Results {
				for i, alt := range result.Alternatives {
					if i == 0 {
						fmt.Fprintf(w, "\"%v\" (confidence=%3f)\n", alt.Transcript, alt.Confidence)
						js[file] = alt.Transcript
					}

				}
			}

		}(file)

		// Throttling
		if runtime.NumGoroutine() > throttle {
			time.Sleep(1 * time.Second)
		}

	}
	wg.Wait()

	fmt.Printf("result %+v", js)

	// Json Marshalling
	// Bytes -> write File
	marshalbytes, _ := json.MarshalIndent(js, "", "	")

	err := ioutil.WriteFile("result.json", marshalbytes, 0644)
	if err != nil {
		log.Print("Error is = ", err)
		os.Exit(1)
	}

	return nil
}
