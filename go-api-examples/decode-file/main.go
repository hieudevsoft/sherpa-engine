package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	sherpa "github.com/k2-fsa/sherpa-ncnn-go/sherpa_ncnn"
	flag "github.com/spf13/pflag"
	"github.com/youpy/go-wav"
	"log"
	"os"
	"strings"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	config := sherpa.RecognizerConfig{}
	config.Feat = sherpa.FeatureConfig{SampleRate: 16000, FeatureDim: 80}

	flag.StringVar(&config.Model.EncoderParam, "encoder-param", "", "Path to the encoder.ncnn.param")
	flag.StringVar(&config.Model.EncoderBin, "encoder-bin", "", "Path to the encoder.ncnn.bin")
	flag.StringVar(&config.Model.DecoderParam, "decoder-param", "", "Path to the decoder.ncnn.param")
	flag.StringVar(&config.Model.DecoderBin, "decoder-bin", "", "Path to the decoder.ncnn.bin")
	flag.StringVar(&config.Model.JoinerParam, "joiner-param", "", "Path to the joiner.ncnn.param")
	flag.StringVar(&config.Model.JoinerBin, "joiner-bin", "", "Path to the joiner.ncnn.bin")

	flag.StringVar(&config.Model.Tokens, "tokens", "", "Path to the tokens file")
	flag.IntVar(&config.Model.NumThreads, "num-threads", 1, "Number of threads for computing")
	flag.StringVar(&config.Decoder.DecodingMethod, "decoding-method", "greedy_search", "Decoding method. Possible values: greedy_search, modified_beam_search")
	flag.IntVar(&config.Decoder.NumActivePaths, "num-active-paths", 4, "Used only when --decoding-method is modified_beam_search")

	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatalf("Please provide one wave file")
	}
	checkConfig(&config)

	log.Println("Reading", flag.Arg(0))

	samples, sampleRate := readWave(flag.Arg(0))

	log.Println("Initializing recognizer")
	recognizer := sherpa.NewRecognizer(&config)
	log.Println("Recognizer created!")
	defer sherpa.DeleteRecognizer(recognizer)

	log.Println("Start decoding!")
	stream := sherpa.NewStream(recognizer)
	defer sherpa.DeleteStream(stream)

	stream.AcceptWaveform(sampleRate, samples)

	tailPadding := make([]float32, int(float32(sampleRate)*0.3))
	stream.AcceptWaveform(sampleRate, tailPadding)

	for recognizer.IsReady(stream) {
		recognizer.Decode(stream)
	}

	log.Println("Decoding done!")
	result := recognizer.GetResult(stream)

	log.Println(strings.ToLower(result.Text))
	log.Printf("Wave duration: %v seconds", float32(len(samples))/float32(sampleRate))
}

func readWave(filename string) (samples []float32, sampleRate int) {
	file, _ := os.Open(filename)
	defer file.Close()

	reader := wav.NewReader(file)
	format, err := reader.Format()
	if err != nil {
		log.Fatalf("Failed to read wave format")
	}

	if format.AudioFormat != 1 {
		log.Fatalf("Support only PCM format. Given: %v\n", format.AudioFormat)
	}

	if format.NumChannels != 1 {
		log.Fatalf("Support only 1 channel wave file. Given: %v\n", format.NumChannels)
	}

	if format.BitsPerSample != 16 {
		log.Fatalf("Support only 16-bit per sample. Given: %v\n", format.BitsPerSample)
	}

	reader.Duration() // so that it initializes reader.Size

	buf := make([]byte, reader.Size)
	n, err := reader.Read(buf)
	if n != int(reader.Size) {
		log.Fatalf("Failed to read %v bytes. Returned %v bytes\n", reader.Size, n)
	}

	samples = samplesInt16ToFloat(buf)
	sampleRate = int(format.SampleRate)

	return
}

func samplesInt16ToFloat(inSamples []byte) []float32 {
	numSamples := len(inSamples) / 2
	outSamples := make([]float32, numSamples)

	for i := 0; i != numSamples; i++ {
		s := inSamples[i*2 : (i+1)*2]

		var s16 int16
		buf := bytes.NewReader(s)
		err := binary.Read(buf, binary.LittleEndian, &s16)
		if err != nil {
			log.Fatal("Failed to parse 16-bit sample")
		}
		outSamples[i] = float32(s16) / 32768
	}

	return outSamples
}

func checkConfig(config *sherpa.RecognizerConfig) {
	// --encoder-param
	if config.Model.EncoderParam == "" {
		log.Fatal("Please provide --encoder-param")
	}

	if _, err := os.Stat(config.Model.EncoderParam); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("--encoder-param %v does not exist", config.Model.EncoderParam)
	}

	// --encoder-bin
	if config.Model.EncoderBin == "" {
		log.Fatal("Please provide --encoder-bin")
	}

	if _, err := os.Stat(config.Model.EncoderBin); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("--encoder-bin %v does not exist", config.Model.EncoderBin)
	}

	// --decoder-param
	if config.Model.DecoderParam == "" {
		log.Fatal("Please provide --decoder-param")
	}

	if _, err := os.Stat(config.Model.DecoderParam); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("--decoder-param %v does not exist", config.Model.DecoderParam)
	}

	// --decoder-bin
	if config.Model.DecoderBin == "" {
		log.Fatal("Please provide --decoder-bin")
	}

	if _, err := os.Stat(config.Model.DecoderBin); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("--decoder-bin %v does not exist", config.Model.DecoderBin)
	}

	// --joiner-param
	if config.Model.JoinerParam == "" {
		log.Fatal("Please provide --joiner-param")
	}

	if _, err := os.Stat(config.Model.JoinerParam); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("--joiner-param %v does not exist", config.Model.JoinerParam)
	}

	// --joiner-bin
	if config.Model.JoinerBin == "" {
		log.Fatal("Please provide --joiner-bin")
	}

	if _, err := os.Stat(config.Model.JoinerBin); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("--joiner-bin %v does not exist", config.Model.JoinerBin)
	}

	// --tokens
	if config.Model.Tokens == "" {
		log.Fatal("Please provide --tokens")
	}

	if _, err := os.Stat(config.Model.Tokens); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("--tokens %v does not exist", config.Model.Tokens)
	}
}
