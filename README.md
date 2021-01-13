# GCS
Google Cloud Speech

Goroutine Parallel Speech Recognition
## Prerequisite 
1. Get Cloud API Key
https://console.cloud.google.com/apis/dashboard 
2. Activate Cloud Speech API
3. Set Environment Config

    Windows

 `set GOOGLE_APPLICATION_CREDENTIALS=C:\Users\username\Downloads\my-key.json`

    Mac/Linux    
 `export GOOGLE_APPLICATION_CREDENTIALS=/home/user/Downloads/my-key.json`

## Usage
16Bit Raw PCM audio

`./gcs [audio_path] `

## GCS makes result.json
